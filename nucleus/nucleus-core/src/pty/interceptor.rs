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

pub struct NucleusShell {
    session_id: String,
    graph: Arc<Mutex<ExecutionGraph>>,
    memory: Arc<SessionMemory>,
    snapshot_engine: Arc<SnapshotEngine>,
    risk_evaluator: Arc<Mutex<RiskEvaluator>>,
    redis_publisher: Arc<RedisPublisher>,
}

impl NucleusShell {
    pub fn new(
        db: sled::Db,
        redis_url: &str,
    ) -> Self {
        let session_id = Uuid::new_v4().to_string();
        let graph = Arc::new(Mutex::new(ExecutionGraph::new(db.clone())));
        let memory = Arc::new(SessionMemory::new(db.clone()));
        let snapshot_engine = Arc::new(SnapshotEngine::new(db.clone()));
        let risk_evaluator = Arc::new(Mutex::new(RiskEvaluator::new()));
        let redis_publisher = Arc::new(RedisPublisher::new_with_fallback(redis_url));

        // Save session metadata
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

        // Publish session started event
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
        let shell = std::env::var("SHELL").unwrap_or_else(|_| "/bin/bash".to_string());

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

        let child = pair.slave.spawn_command(cmd)?;
        drop(pair.slave);

        let mut master = pair.master;

        // Shared state for command buffer
        let command_buffer = Arc::new(Mutex::new(String::new()));
        let in_command = Arc::new(Mutex::new(false));

        // Read from master (child output) and forward to stdout
        let mut reader = master.try_clone_reader()?;
        let cmd_buf_clone = command_buffer.clone();
        let in_cmd_clone = in_command.clone();
        let session_id = self.session_id.clone();
        let graph = self.graph.clone();
        let risk_evaluator = self.risk_evaluator.clone();
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
                        // Write to real stdout
                        let _ = std::io::stdout().write_all(data);
                        let _ = std::io::stdout().flush();

                        // Capture output for the current command
                        let text = String::from_utf8_lossy(data);
                        output_capture.push_str(&text);

                        // Detect prompt return (simplified: look for $ or # or > at line start)
                        if is_prompt_return(&text) {
                            let is_in_cmd = *in_cmd_clone.lock().unwrap();
                            if is_in_cmd {
                                let cmd_str = cmd_buf_clone.lock().unwrap().clone();
                                if !cmd_str.trim().is_empty() {
                                    let duration = capture_start
                                        .map(|s| s.elapsed().as_millis() as u64)
                                        .unwrap_or(0);

                                    // Process the completed command
                                    let parsed = parse_command(&cmd_str);

                                    // Determine exit code from output (simplified)
                                    let exit_code = 0; // In real impl, would parse $?

                                    // Truncate stdout to 4KB
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

                                    // Add to graph
                                    if let Ok(mut g) = graph.lock() {
                                        g.add_node(&node);
                                    }

                                    // Publish to Redis
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

        // Read from stdin and forward to master (with interception)
        let mut writer = master.take_writer()?;
        let cmd_buf_input = command_buffer.clone();
        let in_cmd_input = in_command.clone();
        let risk_eval = self.risk_evaluator.clone();
        let snap_engine = self.snapshot_engine.clone();
        let redis_pub = self.redis_publisher.clone();

        let input_thread = std::thread::spawn(move || {
            let stdin = std::io::stdin();
            let mut buf = [0u8; 1024];

            // Set stdin to raw mode for proper terminal forwarding
            loop {
                match stdin.lock().read(&mut buf) {
                    Ok(0) => break,
                    Ok(n) => {
                        let data = &buf[..n];
                        let text = String::from_utf8_lossy(data);

                        // Detect enter key (newline)
                        if data.contains(&b'\n') || data.contains(&b'\r') {
                            let cmd_str = cmd_buf_input.lock().unwrap().clone();
                            if !cmd_str.trim().is_empty() {
                                let parsed = parse_command(&cmd_str);

                                // Risk evaluation
                                let assessment = risk_eval.lock().unwrap().evaluate(&parsed);

                                if assessment.block_execution {
                                    // Print warning and don't forward
                                    let warning = format!(
                                        "\r\n\x1b[1;31m[NUCLEUS] BLOCKED: {}\x1b[0m\r\n",
                                        assessment.warnings.join("; ")
                                    );
                                    let _ = std::io::stdout()
                                        .write_all(warning.as_bytes());
                                    let _ = std::io::stdout().flush();
                                    cmd_buf_input.lock().unwrap().clear();
                                    continue;
                                }

                                // Show risk warnings
                                if assessment.risk_level >= RiskLevel::Medium {
                                    let warning = format!(
                                        "\r\n\x1b[1;33m[NUCLEUS] Risk {}: {}\x1b[0m\r\n",
                                        assessment.risk_level,
                                        assessment.warnings.join("; ")
                                    );
                                    let _ = std::io::stdout()
                                        .write_all(warning.as_bytes());
                                    let _ = std::io::stdout().flush();
                                }

                                // Take snapshot for mutating commands
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
                                        &exec_id,
                                        &files,
                                        &env_vars,
                                    );
                                }

                                *in_cmd_input.lock().unwrap() = true;
                            }
                            cmd_buf_input.lock().unwrap().clear();
                        } else {
                            // Accumulate command characters
                            let mut buf = cmd_buf_input.lock().unwrap();
                            for &byte in data {
                                if byte == 127 || byte == 8 {
                                    // Backspace
                                    buf.pop();
                                } else if byte == 3 {
                                    // Ctrl+C
                                    buf.clear();
                                } else if byte >= 32 {
                                    buf.push(byte as char);
                                }
                            }
                        }

                        // Forward to the shell
                        let _ = writer.write_all(data);
                        let _ = writer.flush();
                    }
                    Err(_) => break,
                }
            }
        });

        let _ = output_thread.join();
        let _ = input_thread.join();

        // End session
        self.memory.end_session(&self.session_id);

        Ok(())
    }

    pub fn execute_command(&self, command: &str) -> ExecutionNode {
        let parsed = parse_command(command);
        let assessment = self.risk_evaluator.lock().unwrap().evaluate(&parsed);

        let exec_id = Uuid::new_v4().to_string();

        // Take snapshot if mutating
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

        // Execute command
        let start = Instant::now();
        let output = std::process::Command::new("sh")
            .arg("-c")
            .arg(command)
            .output();

        let duration_ms = start.elapsed().as_millis() as u64;

        let (exit_code, stdout, stderr) = match output {
            Ok(out) => {
                let exit_code = out.status.code().unwrap_or(-1);
                let stdout = String::from_utf8_lossy(&out.stdout).to_string();
                let stderr = String::from_utf8_lossy(&out.stderr).to_string();

                if exit_code != 0 {
                    self.risk_evaluator.lock().unwrap().record_failure(command);
                }

                (exit_code, stdout, stderr)
            }
            Err(e) => (-1, String::new(), e.to_string()),
        };

        // Truncate stdout to 4KB
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

        // Add to graph
        if let Ok(mut g) = self.graph.lock() {
            g.add_node(&node);
        }

        // Publish to Redis
        if let Ok(json) = serde_json::to_string(&node) {
            self.redis_publisher.publish_execution(&json);
        }

        node
    }

    pub fn rollback(&self, execution_id: &str) -> crate::rollback::snapshot::RollbackResult {
        let result = self.snapshot_engine.rollback(execution_id);

        if result.success {
            if let Ok(mut g) = self.graph.lock() {
                g.mark_rolled_back(execution_id);
            }

            // Publish rollback event
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
            || trimmed.contains("❯")
            || trimmed.contains("➜")
    } else {
        false
    }
}
