use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sled::Db;
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SessionMetadata {
    pub session_id: String,
    pub name: String,
    pub shell_pid: u32,
    pub start_time: DateTime<Utc>,
    pub end_time: Option<DateTime<Utc>>,
    pub working_directory: String,
    pub git_repo: Option<String>,
    pub env_snapshot: HashMap<String, String>,
}

pub struct SessionMemory {
    db: Db,
}

impl SessionMemory {
    pub fn new(db: Db) -> Self {
        SessionMemory { db }
    }

    pub fn save_session(&self, session: &SessionMetadata) {
        let key = format!("session:{}", session.session_id);
        if let Ok(json) = serde_json::to_vec(session) {
            let _ = self.db.insert(key.as_bytes(), json);
            let _ = self.db.flush();
        }
    }

    pub fn get_session(&self, session_id: &str) -> Option<SessionMetadata> {
        let key = format!("session:{}", session_id);
        self.db
            .get(key.as_bytes())
            .ok()
            .flatten()
            .and_then(|bytes| serde_json::from_slice(&bytes).ok())
    }

    pub fn list_sessions(&self) -> Vec<SessionMetadata> {
        let mut sessions = Vec::new();
        let prefix = b"session:";
        for item in self.db.scan_prefix(prefix) {
            if let Ok((_key, value)) = item {
                if let Ok(session) = serde_json::from_slice::<SessionMetadata>(&value) {
                    sessions.push(session);
                }
            }
        }
        sessions.sort_by(|a, b| b.start_time.cmp(&a.start_time));
        sessions
    }

    pub fn end_session(&self, session_id: &str) {
        if let Some(mut session) = self.get_session(session_id) {
            session.end_time = Some(Utc::now());
            self.save_session(&session);
        }
    }

    pub fn get_failure_patterns(&self, db: &Db) -> Vec<(String, u32)> {
        let mut failures: HashMap<String, u32> = HashMap::new();
        let prefix = b"exec:";

        for item in db.scan_prefix(prefix) {
            if let Ok((_key, value)) = item {
                if let Ok(node) =
                    serde_json::from_slice::<crate::context::graph::ExecutionNode>(&value)
                {
                    if node.exit_code != 0 {
                        let count = failures.entry(node.command.raw.clone()).or_insert(0);
                        *count += 1;
                    }
                }
            }
        }

        let mut patterns: Vec<(String, u32)> = failures
            .into_iter()
            .filter(|(_, count)| *count >= 2)
            .collect();
        patterns.sort_by(|a, b| b.1.cmp(&a.1));
        patterns
    }

    pub fn detect_git_repo(cwd: &str) -> Option<String> {
        let mut path = std::path::PathBuf::from(cwd);
        loop {
            let git_dir = path.join(".git");
            if git_dir.exists() {
                return Some(path.to_string_lossy().to_string());
            }
            if !path.pop() {
                return None;
            }
        }
    }
}
