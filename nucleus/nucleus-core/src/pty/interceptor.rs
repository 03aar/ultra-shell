use crate::context::evaluator::{RiskEvaluator, RiskLevel};
use crate::context::graph::{ExecutionGraph, ExecutionNode};
use crate::context::memory::{SessionMemory, SessionMetadata};
use crate::ipc::redis::RedisPublisher;
use crate::pty::parser::{parse_command, CommandCategory};
use crate::rollback::snapshot::SnapshotEngine;
use chrono::Utc;
use portable_pty::{CommandBuilder, NativePtySystem, PtySize, PtySystem};
use std::collections::HashMap;
use std::io::{Read, Write};
use std::path::PathBuf;
use std::sync::{Arc, Mutex};
use std::time::Instant;
use uuid::Uuid;

/// Returns the default shell for the current platform.
fn default_shell() -> String {
    #[cfg(windows)]
    {
        std::env::var("COMSPEC").unwrap_or_else(|_| "cmd.exe".to_string())
    }
    #[cfg(not(windows))]
    {
        std::env::var("SHELL").unwrap_or_else(|_| "/bin/bash".to_string())
    }
}

/// Returns the shell and flag used for `execute_command` one-liners.
fn shell_exec_args() -> (&'static str, &'static str) {
    #[cfg(windows)]
    {
        ("cmd.exe", "/C")
    }
    #[cfg(not(windows))]
    {
        ("sh", "-c")
    }
}

pub struct NucleusShell {
    session_id: String,
    graph: Arc<Mutex<ExecutionGraph>>,
    memory: Arc<SessionMemory>,
    snapshot_engine: Arc<SnapshotEngine>,
    risk_evaluator: Arc<Mutex<RiskEvaluator>>,
    redis_publisher: Arc<RedisPublisher>,
}

impl NucleusShell {
    pub fn new(db: sled::Db, redis_url: &str) -> Self {
        let session_id = Uuid::new_v4().to_string();
        let graph = Arc::new(Mutex::new(ExecutionGraph::new(db.clone())));
        let memory = Arc::new(SessionMemory::new(db.clone()));
        let snapshot_engine = Arc::new(SnapshotEngine::new(db.clone()));
        let risk_evaluator = Arc::new(Mutex::new(RiskEvaluator::new()));
        let redis_publisher = Arc::new(RedisPublisher::new_with_fallback(redis_url));

        let cwd = std::env::current_dir()
            .unwrap_or_default()
            .to_string_lossy()
            .to_string();

        let git_repo = SessionMemory::detect_git_repo(&cwd);
        let env_snapshot: HashMap<String, String> = std::env::vars().collect();

        let session = SessionMetadata {
            session_id: session_id.clone(),
            name: format!("session-{}", &session_id[..8]),
            shell_pid: std::process::id(),
            start_time: Utc::now(),
            end_time: None,
            working_directory: cwd,
            git_repo,
            env_snapshot,
        };
        memory.save_session(&session);

        if let Ok(json) = serde_json::to_string(&serde_json::json!({
            "type": "session_started",
            "session_id": session_id,
            "timestamp": Utc::now().to_rfc3339(),
        })) {
            redis_publisher.publish_session_event(&json);
        }

        NucleusShell {
            session_id,
            graph,
            memory,
            snapshot_engine,
            risk_evaluator,
            redis_publisher,
        }
    }

    pub fn session_id(&self) -> &str {
        &self.session_id
    }

    /// Run the interactive shell with PTY interception
    pub fn run(&self) -> Result<(), Box<dyn std::error::Error>> {
        let shell = default_shell();

        let pty_system = NativePtySystem::default();
        let pair = pty_system.openpty(PtySize {
            rows: 24,
            cols: 80,
            pixel_width: 0,
            pixel_height: 0,
        })?;

        let mut cmd = CommandBuilder::new(&shell);
        cmd.env("NUCLEUS_SESSION", &self.session_id);
        cmd.env("NUCLEUS_ACTIVE", "1");

        let _child = pair.slave.spawn_command(cmd)?;
        drop(pair.slave);

        let master = pair.master;

        let command_buffer = Arc::new(Mutex::new(String::new()));
        let in_command = Arc::new(Mutex::new(false));

        // --- Output reader thread ---
        let mut reader = master.try_clone_reader()?;
        let cmd_buf_clone = command_buffer.clone();
        let in_cmd_clone = in_command.clone();
        let session_id = self.session_id.clone();
        let graph = self.graph.clone();
        let snapshot_engine = self.snapshot_engine.clone();
        let redis_publisher = self.redis_publisher.clone();

        let output_thread = std::thread::spawn(move || {
            let mut buf = [0u8; 4096];
            let mut output_capture = String::new();
            let mut capture_start: Option<Instant> = None;

            loop {
                match reader.read(&mut buf) {
                    Ok(0) => break,
                    Ok(n) => {
                        let data = &buf[..n];
                        let _ = std::io::stdout().write_all(data);
                        let _ = std::io::stdout().flush();

                        let text = String::from_utf8_lossy(data);
                        output_capture.push_str(&text);

                        if is_prompt_return(&text) {
                            let is_in_cmd = *in_cmd_clone.lock().unwrap();
                            if is_in_cmd {
                                let cmd_str = cmd_buf_clone.lock().unwrap().clone();
                                if !cmd_str.trim().is_empty() {
                                    let duration = capture_start
                                        .map(|s| s.elapsed().as_millis() as u64)
                                        .unwrap_or(0);

                                    let parsed = parse_command(&cmd_str);
                                    let exit_code = 0;

                                    let stdout = if output_capture.len() > 4096 {
                                        output_capture[..4096].to_string()
                                    } else {
                                        output_capture.clone()
                                    };

                                    let node = ExecutionNode {
                                        id: Uuid::new_v4().to_string(),
                                        timestamp: Utc::now(),
                                        command: parsed.clone(),
                                        exit_code,
                                        duration_ms: duration,
                                        stdout,
                                        stderr: String::new(),
                                        files_mutated: parsed
                                            .files_potentially_mutated()
                                            .iter()
                                            .map(PathBuf::from)
                                            .collect(),
                                        env_changes: parsed.env_assignments.clone(),
                                        processes_spawned: Vec::new(),
                                        risk_flags: Vec::new(),
                                        rollback_available: snapshot_engine
                                            .has_snapshot(&cmd_str),
                                        rolled_back: false,
                                        session_id: session_id.clone(),
                                    };

                                    if let Ok(mut g) = graph.lock() {
                                        g.add_node(&node);
                                    }

                                    if let Ok(json) = serde_json::to_string(&node) {
                                        redis_publisher.publish_execution(&json);
                                    }
                                }

                                *in_cmd_clone.lock().unwrap() = false;
                                cmd_buf_clone.lock().unwrap().clear();
                                output_capture.clear();
                                capture_start = None;
                            }
                        }
                    }
                    Err(_) => break,
                }
            }
        });

        // --- Input reader thread ---
        let mut writer = master.take_writer()?;
        let cmd_buf_input = command_buffer.clone();
        let in_cmd_input = in_command.clone();
        let risk_eval = self.risk_evaluator.clone();
        let snap_engine = self.snapshot_engine.clone();

        let input_thread = std::thread::spawn(move || {
            let stdin = std::io::stdin();
            let mut buf = [0u8; 1024];

            loop {
                match stdin.lock().read(&mut buf) {
                    Ok(0) => break,
                    Ok(n) => {
                        let data = &buf[..n];

                        if data.contains(&b'\n') || data.contains(&b'\r') {
                            let cmd_str = cmd_buf_input.lock().unwrap().clone();
                            if !cmd_str.trim().is_empty() {
                                let parsed = parse_command(&cmd_str);
                                let assessment = risk_eval.lock().unwrap().evaluate(&parsed);

                                if assessment.block_execution {
                                    let warning = format!(
                                        "\r\n\x1b[1;31m[NUCLEUS] BLOCKED: {}\x1b[0m\r\n",
                                        assessment.warnings.join("; ")
                                    );
                                    let _ =
                                        std::io::stdout().write_all(warning.as_bytes());
                                    let _ = std::io::stdout().flush();
                                    cmd_buf_input.lock().unwrap().clear();
                                    continue;
                                }

                                if assessment.risk_level >= RiskLevel::Medium {
                                    let warning = format!(
                                        "\r\n\x1b[1;33m[NUCLEUS] Risk {}: {}\x1b[0m\r\n",
                                        assessment.risk_level,
                                        assessment.warnings.join("; ")
                                    );
                                    let _ =
                                        std::io::stdout().write_all(warning.as_bytes());
                                    let _ = std::io::stdout().flush();
                                }

                                if parsed.category == CommandCategory::FilesystemMutation
                                    || parsed.category == CommandCategory::EnvChange
                                {
                                    let files: Vec<PathBuf> = parsed
                                        .files_potentially_mutated()
                                        .iter()
                                        .map(PathBuf::from)
                                        .collect();
                                    let env_vars: Vec<String> =
                                        parsed.env_assignments.keys().cloned().collect();
                                    let exec_id = Uuid::new_v4().to_string();
                                    snap_engine.take_snapshot(
                                        &exec_id, &files, &env_vars,
                                    );
                                }

                                *in_cmd_input.lock().unwrap() = true;
                            }
                            cmd_buf_input.lock().unwrap().clear();
                        } else {
                            let mut cmd_buf = cmd_buf_input.lock().unwrap();
                            for &byte in data {
                                if byte == 127 || byte == 8 {
                                    cmd_buf.pop();
                                } else if byte == 3 {
                                    cmd_buf.clear();
                                } else if byte >= 32 {
                                    cmd_buf.push(byte as char);
                                }
                            }
                        }

                        let _ = writer.write_all(data);
                        let _ = writer.flush();
                    }
                    Err(_) => break,
                }
            }
        });

        let _ = output_thread.join();
        let _ = input_thread.join();

        self.memory.end_session(&self.session_id);
        Ok(())
    }

    /// Execute a single command non-interactively and return the result.
    pub fn execute_command(&self, command: &str) -> ExecutionNode {
        let parsed = parse_command(command);
        let assessment = self.risk_evaluator.lock().unwrap().evaluate(&parsed);
        let exec_id = Uuid::new_v4().to_string();

        let rollback_available = if parsed.category == CommandCategory::FilesystemMutation
            || parsed.category == CommandCategory::EnvChange
        {
            let files: Vec<PathBuf> = parsed
                .files_potentially_mutated()
                .iter()
                .map(PathBuf::from)
                .collect();
            let env_vars: Vec<String> = parsed.env_assignments.keys().cloned().collect();
            self.snapshot_engine
                .take_snapshot(&exec_id, &files, &env_vars);
            true
        } else {
            false
        };

        let (shell, flag) = shell_exec_args();
        let start = Instant::now();
        let output = std::process::Command::new(shell)
            .arg(flag)
            .arg(command)
            .output();
        let duration_ms = start.elapsed().as_millis() as u64;

        let (exit_code, stdout, stderr) = match output {
            Ok(out) => {
                let code = out.status.code().unwrap_or(-1);
                let out_str = String::from_utf8_lossy(&out.stdout).to_string();
                let err_str = String::from_utf8_lossy(&out.stderr).to_string();
                if code != 0 {
                    self.risk_evaluator.lock().unwrap().record_failure(command);
                }
                (code, out_str, err_str)
            }
            Err(e) => (-1, String::new(), e.to_string()),
        };

        let stdout_truncated = if stdout.len() > 4096 {
            stdout[..4096].to_string()
        } else {
            stdout
        };

        let node = ExecutionNode {
            id: exec_id,
            timestamp: Utc::now(),
            command: parsed.clone(),
            exit_code,
            duration_ms,
            stdout: stdout_truncated,
            stderr,
            files_mutated: parsed
                .files_potentially_mutated()
                .iter()
                .map(PathBuf::from)
                .collect(),
            env_changes: parsed.env_assignments.clone(),
            processes_spawned: Vec::new(),
            risk_flags: assessment.flags,
            rollback_available,
            rolled_back: false,
            session_id: self.session_id.clone(),
        };

        if let Ok(mut g) = self.graph.lock() {
            g.add_node(&node);
        }
        if let Ok(json) = serde_json::to_string(&node) {
            self.redis_publisher.publish_execution(&json);
        }

        node
    }

    /// Rollback a specific execution, restoring snapshotted files and env vars.
    pub fn rollback(&self, execution_id: &str) -> crate::rollback::snapshot::RollbackResult {
        let result = self.snapshot_engine.rollback(execution_id);

        if result.success {
            if let Ok(mut g) = self.graph.lock() {
                g.mark_rolled_back(execution_id);
            }
            if let Ok(json) = serde_json::to_string(&serde_json::json!({
                "type": "rollback_complete",
                "execution_id": execution_id,
                "files_restored": result.files_restored,
                "timestamp": Utc::now().to_rfc3339(),
            })) {
                self.redis_publisher.publish_rollback(&json);
            }
        }

        result
    }

    pub fn get_graph(&self) -> Arc<Mutex<ExecutionGraph>> {
        self.graph.clone()
    }
}

fn is_prompt_return(text: &str) -> bool {
    let lines: Vec<&str> = text.lines().collect();
    if let Some(last_line) = lines.last() {
        let trimmed = last_line.trim();
        trimmed.ends_with('$')
            || trimmed.ends_with('#')
            || trimmed.ends_with('>')
            || trimmed.ends_with('%')
    } else {
        false
    }
}
