use crate::context::evaluator::RiskFlag;
use crate::pty::parser::ParsedCommand;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sled::Db;
use std::collections::HashMap;
use std::path::PathBuf;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionNode {
    pub id: String,
    pub timestamp: DateTime<Utc>,
    pub command: ParsedCommand,
    pub exit_code: i32,
    pub duration_ms: u64,
    pub stdout: String,
    pub stderr: String,
    pub files_mutated: Vec<PathBuf>,
    pub env_changes: HashMap<String, String>,
    pub processes_spawned: Vec<u32>,
    pub risk_flags: Vec<RiskFlag>,
    pub rollback_available: bool,
    pub rolled_back: bool,
    pub session_id: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Edge {
    pub from: String,
    pub to: String,
    pub dependency_type: String, // "file_read_write", "process_lifecycle", "env_dependency"
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GraphJson {
    pub nodes: Vec<ExecutionNode>,
    pub edges: Vec<Edge>,
}

pub struct ExecutionGraph {
    db: Db,
    edges: Vec<Edge>,
    file_writers: HashMap<String, String>,   // filepath -> execution_id
    env_setters: HashMap<String, String>,    // var_name -> execution_id
    process_spawners: HashMap<u32, String>,  // pid -> execution_id
}

impl ExecutionGraph {
    pub fn new(db: Db) -> Self {
        ExecutionGraph {
            db,
            edges: Vec::new(),
            file_writers: HashMap::new(),
            env_setters: HashMap::new(),
            process_spawners: HashMap::new(),
        }
    }

    pub fn add_node(&mut self, node: &ExecutionNode) {
        // Build edges
        self.build_edges(node);

        // Store node
        let key = format!("exec:{}", node.id);
        if let Ok(json) = serde_json::to_vec(node) {
            let _ = self.db.insert(key.as_bytes(), json);
        }

        // Store in session index
        let session_key = format!(
            "session_exec:{}:{}:{}",
            node.session_id,
            node.timestamp.timestamp_millis(),
            node.id
        );
        let _ = self.db.insert(session_key.as_bytes(), node.id.as_bytes());

        // Track file writers
        for file in &node.files_mutated {
            self.file_writers
                .insert(file.to_string_lossy().to_string(), node.id.clone());
        }

        // Track env setters
        for (var, _val) in &node.env_changes {
            self.env_setters.insert(var.clone(), node.id.clone());
        }

        // Track process spawners
        for pid in &node.processes_spawned {
            self.process_spawners.insert(*pid, node.id.clone());
        }

        let _ = self.db.flush();
    }

    fn build_edges(&mut self, node: &ExecutionNode) {
        // Check if this command reads files written by a previous command
        for arg in &node.command.args {
            if let Some(writer_id) = self.file_writers.get(arg) {
                if writer_id != &node.id {
                    self.edges.push(Edge {
                        from: writer_id.clone(),
                        to: node.id.clone(),
                        dependency_type: "file_read_write".to_string(),
                    });
                    // Store edge
                    let edge_key = format!("edge:{}:{}", writer_id, node.id);
                    let _ = self.db.insert(
                        edge_key.as_bytes(),
                        b"file_read_write",
                    );
                }
            }
        }

        // Check if this command uses env vars set by a previous command
        // (simplified: check if any env var in the command string matches)
        for (var, setter_id) in &self.env_setters {
            if node.command.raw.contains(&format!("${}", var))
                || node.command.raw.contains(&format!("${{{}}}", var))
            {
                if setter_id != &node.id {
                    self.edges.push(Edge {
                        from: setter_id.clone(),
                        to: node.id.clone(),
                        dependency_type: "env_dependency".to_string(),
                    });
                    let edge_key = format!("edge:{}:{}", setter_id, node.id);
                    let _ = self.db.insert(
                        edge_key.as_bytes(),
                        b"env_dependency",
                    );
                }
            }
        }

        // Check process lifecycle dependencies (kill commands)
        if node.command.binary == "kill" || node.command.binary == "killall" {
            for arg in &node.command.args {
                if let Ok(pid) = arg.parse::<u32>() {
                    if let Some(spawner_id) = self.process_spawners.get(&pid) {
                        self.edges.push(Edge {
                            from: spawner_id.clone(),
                            to: node.id.clone(),
                            dependency_type: "process_lifecycle".to_string(),
                        });
                    }
                }
            }
        }
    }

    pub fn get_node(&self, id: &str) -> Option<ExecutionNode> {
        let key = format!("exec:{}", id);
        self.db
            .get(key.as_bytes())
            .ok()
            .flatten()
            .and_then(|bytes| serde_json::from_slice(&bytes).ok())
    }

    pub fn get_recent(&self, limit: usize) -> Vec<ExecutionNode> {
        let mut nodes: Vec<ExecutionNode> = Vec::new();
        let prefix = b"exec:";

        for item in self.db.scan_prefix(prefix) {
            if let Ok((_key, value)) = item {
                if let Ok(node) = serde_json::from_slice::<ExecutionNode>(&value) {
                    nodes.push(node);
                }
            }
        }

        nodes.sort_by(|a, b| b.timestamp.cmp(&a.timestamp));
        nodes.truncate(limit);
        nodes
    }

    pub fn get_session_executions(&self, session_id: &str) -> Vec<ExecutionNode> {
        let mut nodes: Vec<ExecutionNode> = Vec::new();
        let prefix = format!("session_exec:{}:", session_id);

        for item in self.db.scan_prefix(prefix.as_bytes()) {
            if let Ok((_key, value)) = item {
                let exec_id = String::from_utf8_lossy(&value).to_string();
                if let Some(node) = self.get_node(&exec_id) {
                    nodes.push(node);
                }
            }
        }

        nodes.sort_by(|a, b| a.timestamp.cmp(&b.timestamp));
        nodes
    }

    pub fn get_graph_json(&self) -> GraphJson {
        let nodes = self.get_recent(1000);
        GraphJson {
            nodes,
            edges: self.edges.clone(),
        }
    }

    pub fn get_session_graph_json(&self, session_id: &str) -> GraphJson {
        let nodes = self.get_session_executions(session_id);
        let node_ids: Vec<String> = nodes.iter().map(|n| n.id.clone()).collect();
        let edges: Vec<Edge> = self
            .edges
            .iter()
            .filter(|e| node_ids.contains(&e.from) || node_ids.contains(&e.to))
            .cloned()
            .collect();
        GraphJson { nodes, edges }
    }

    pub fn search(&self, query: &str) -> Vec<ExecutionNode> {
        let query_lower = query.to_lowercase();
        self.get_recent(1000)
            .into_iter()
            .filter(|node| {
                node.command.raw.to_lowercase().contains(&query_lower)
                    || node.stdout.to_lowercase().contains(&query_lower)
                    || node.stderr.to_lowercase().contains(&query_lower)
            })
            .collect()
    }

    pub fn mark_rolled_back(&mut self, execution_id: &str) -> bool {
        if let Some(mut node) = self.get_node(execution_id) {
            node.rolled_back = true;
            node.rollback_available = false;
            let key = format!("exec:{}", execution_id);
            if let Ok(json) = serde_json::to_vec(&node) {
                let _ = self.db.insert(key.as_bytes(), json);
                let _ = self.db.flush();
                return true;
            }
        }
        false
    }
}
