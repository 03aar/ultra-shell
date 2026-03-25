use serde::{Deserialize, Serialize};
use sled::Db;
use std::collections::HashMap;
use std::fs;
use std::path::PathBuf;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileSnapshot {
    pub path: PathBuf,
    pub content: Option<Vec<u8>>,
    pub existed: bool,
    pub permissions: Option<u32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionSnapshot {
    pub execution_id: String,
    pub files: Vec<FileSnapshot>,
    pub env_vars: HashMap<String, Option<String>>,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RollbackResult {
    pub success: bool,
    pub files_restored: Vec<String>,
    pub env_restored: HashMap<String, Option<String>>,
    pub message: String,
}

pub struct SnapshotEngine {
    db: Db,
}

impl SnapshotEngine {
    pub fn new(db: Db) -> Self {
        SnapshotEngine { db }
    }

    /// Take a snapshot of files and env vars before a command executes
    pub fn take_snapshot(
        &self,
        execution_id: &str,
        files_to_snapshot: &[PathBuf],
        env_vars_to_snapshot: &[String],
    ) -> ExecutionSnapshot {
        let mut file_snapshots = Vec::new();

        for path in files_to_snapshot {
            let existed = path.exists();
            let content = if existed {
                fs::read(path).ok()
            } else {
                None
            };
            let permissions = if existed {
                #[cfg(unix)]
                {
                    use std::os::unix::fs::PermissionsExt;
                    fs::metadata(path)
                        .ok()
                        .map(|m| m.permissions().mode())
                }
                #[cfg(not(unix))]
                {
                    None
                }
            } else {
                None
            };

            file_snapshots.push(FileSnapshot {
                path: path.clone(),
                content,
                existed,
                permissions,
            });
        }

        let mut env_snapshot = HashMap::new();
        for var in env_vars_to_snapshot {
            env_snapshot.insert(var.clone(), std::env::var(var).ok());
        }

        let snapshot = ExecutionSnapshot {
            execution_id: execution_id.to_string(),
            files: file_snapshots,
            env_vars: env_snapshot,
            timestamp: chrono::Utc::now(),
        };

        // Persist to sled
        let key = format!("snapshot:{}", execution_id);
        if let Ok(json) = serde_json::to_vec(&snapshot) {
            let _ = self.db.insert(key.as_bytes(), json);
            let _ = self.db.flush();
        }

        snapshot
    }

    /// Rollback an execution by restoring its snapshot
    pub fn rollback(&self, execution_id: &str) -> RollbackResult {
        let key = format!("snapshot:{}", execution_id);

        let snapshot = match self.db.get(key.as_bytes()) {
            Ok(Some(bytes)) => match serde_json::from_slice::<ExecutionSnapshot>(&bytes) {
                Ok(s) => s,
                Err(e) => {
                    return RollbackResult {
                        success: false,
                        files_restored: Vec::new(),
                        env_restored: HashMap::new(),
                        message: format!("Failed to parse snapshot: {}", e),
                    }
                }
            },
            _ => {
                return RollbackResult {
                    success: false,
                    files_restored: Vec::new(),
                    env_restored: HashMap::new(),
                    message: "No snapshot found for this execution".to_string(),
                }
            }
        };

        let mut files_restored = Vec::new();
        let mut errors = Vec::new();

        for file_snap in &snapshot.files {
            let path = &file_snap.path;
            if file_snap.existed {
                if let Some(ref content) = file_snap.content {
                    match fs::write(path, content) {
                        Ok(()) => {
                            // Restore permissions
                            #[cfg(unix)]
                            if let Some(mode) = file_snap.permissions {
                                use std::os::unix::fs::PermissionsExt;
                                let _ = fs::set_permissions(
                                    path,
                                    fs::Permissions::from_mode(mode),
                                );
                            }
                            files_restored.push(path.to_string_lossy().to_string());
                        }
                        Err(e) => {
                            errors.push(format!(
                                "Failed to restore {}: {}",
                                path.display(),
                                e
                            ));
                        }
                    }
                }
            } else {
                // File didn't exist before — remove it if it was created
                if path.exists() {
                    match fs::remove_file(path) {
                        Ok(()) => {
                            files_restored
                                .push(format!("{} (removed)", path.to_string_lossy()));
                        }
                        Err(e) => {
                            errors.push(format!(
                                "Failed to remove {}: {}",
                                path.display(),
                                e
                            ));
                        }
                    }
                }
            }
        }

        // Restore env vars
        let mut env_restored = HashMap::new();
        for (var, value) in &snapshot.env_vars {
            match value {
                Some(val) => {
                    std::env::set_var(var, val);
                    env_restored.insert(var.clone(), Some(val.clone()));
                }
                None => {
                    std::env::remove_var(var);
                    env_restored.insert(var.clone(), None);
                }
            }
        }

        let success = errors.is_empty();
        let message = if success {
            format!(
                "Rollback successful: {} files restored",
                files_restored.len()
            )
        } else {
            format!(
                "Rollback partial: {} files restored, {} errors: {}",
                files_restored.len(),
                errors.len(),
                errors.join("; ")
            )
        };

        RollbackResult {
            success,
            files_restored,
            env_restored,
            message,
        }
    }

    pub fn has_snapshot(&self, execution_id: &str) -> bool {
        let key = format!("snapshot:{}", execution_id);
        self.db
            .get(key.as_bytes())
            .ok()
            .flatten()
            .is_some()
    }

    pub fn get_snapshot(&self, execution_id: &str) -> Option<ExecutionSnapshot> {
        let key = format!("snapshot:{}", execution_id);
        self.db
            .get(key.as_bytes())
            .ok()
            .flatten()
            .and_then(|bytes| serde_json::from_slice(&bytes).ok())
    }
}
