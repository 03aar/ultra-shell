use crate::pty::parser::{CommandCategory, ParsedCommand};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::Path;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, PartialOrd)]
pub enum RiskLevel {
    None,
    Low,
    Medium,
    High,
    Critical,
}

impl std::fmt::Display for RiskLevel {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            RiskLevel::None => write!(f, "none"),
            RiskLevel::Low => write!(f, "low"),
            RiskLevel::Medium => write!(f, "medium"),
            RiskLevel::High => write!(f, "high"),
            RiskLevel::Critical => write!(f, "critical"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RiskFlag {
    pub code: String,
    pub message: String,
    pub level: RiskLevel,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RiskAssessment {
    pub risk_level: RiskLevel,
    pub warnings: Vec<String>,
    pub suggestions: Vec<String>,
    pub block_execution: bool,
    pub flags: Vec<RiskFlag>,
}

impl RiskAssessment {
    pub fn safe() -> Self {
        RiskAssessment {
            risk_level: RiskLevel::None,
            warnings: Vec::new(),
            suggestions: Vec::new(),
            block_execution: false,
            flags: Vec::new(),
        }
    }
}

pub struct RiskEvaluator {
    failure_history: HashMap<String, u32>,
}

impl RiskEvaluator {
    pub fn new() -> Self {
        RiskEvaluator {
            failure_history: HashMap::new(),
        }
    }

    pub fn record_failure(&mut self, command: &str) {
        let count = self.failure_history.entry(command.to_string()).or_insert(0);
        *count += 1;
    }

    pub fn evaluate(&self, cmd: &ParsedCommand) -> RiskAssessment {
        let mut assessment = RiskAssessment::safe();

        // Check for destructive commands
        self.check_destructive(cmd, &mut assessment);

        // Check for repeated failures
        self.check_failure_history(cmd, &mut assessment);

        // Check for dangerous flags
        self.check_dangerous_flags(cmd, &mut assessment);

        // Check unmet dependencies
        self.check_dependencies(cmd, &mut assessment);

        // Compute overall risk level
        assessment.risk_level = assessment
            .flags
            .iter()
            .map(|f| f.level.clone())
            .max_by(|a, b| a.partial_cmp(b).unwrap_or(std::cmp::Ordering::Equal))
            .unwrap_or(RiskLevel::None);

        assessment
    }

    fn check_destructive(&self, cmd: &ParsedCommand, assessment: &mut RiskAssessment) {
        let bin = cmd.binary.rsplit('/').next().unwrap_or(&cmd.binary);

        // rm -rf / or rm -rf /*
        if bin == "rm" {
            let has_rf = cmd.flags.iter().any(|f| f.contains('r') && f.contains('f'))
                || (cmd.flags.iter().any(|f| f.contains('r'))
                    && cmd.flags.iter().any(|f| f.contains('f')));

            if has_rf {
                for arg in &cmd.args {
                    if arg == "/" || arg == "/*" || arg == "/." {
                        assessment.block_execution = true;
                        assessment.flags.push(RiskFlag {
                            code: "CRITICAL_RM_ROOT".to_string(),
                            message: "rm -rf on root filesystem — BLOCKED".to_string(),
                            level: RiskLevel::Critical,
                        });
                        assessment
                            .warnings
                            .push("This command would destroy the entire filesystem.".to_string());
                        return;
                    }
                }

                assessment.flags.push(RiskFlag {
                    code: "RECURSIVE_DELETE".to_string(),
                    message: format!(
                        "Recursive forced deletion: rm {} {}",
                        cmd.flags.join(" "),
                        cmd.args.join(" ")
                    ),
                    level: RiskLevel::High,
                });
                assessment
                    .warnings
                    .push("Recursive forced delete — files cannot be recovered.".to_string());
                assessment
                    .suggestions
                    .push("Consider using trash-cli instead of rm for safer deletion.".to_string());
            } else if cmd.flags.iter().any(|f| f.contains('r')) {
                assessment.flags.push(RiskFlag {
                    code: "RECURSIVE_DELETE".to_string(),
                    message: "Recursive deletion without force flag".to_string(),
                    level: RiskLevel::Medium,
                });
            }
        }

        // chmod 777
        if bin == "chmod" && cmd.args.iter().any(|a| a == "777") {
            assessment.flags.push(RiskFlag {
                code: "CHMOD_777".to_string(),
                message: "chmod 777 makes files world-writable".to_string(),
                level: RiskLevel::High,
            });
            assessment.warnings.push(
                "chmod 777 is a security risk — makes files readable/writable by everyone."
                    .to_string(),
            );
        }

        // kill -9
        if bin == "kill" && cmd.flags.iter().any(|f| f == "-9" || f == "-KILL") {
            assessment.flags.push(RiskFlag {
                code: "FORCE_KILL".to_string(),
                message: "Force kill prevents graceful shutdown".to_string(),
                level: RiskLevel::Medium,
            });
            assessment
                .suggestions
                .push("Consider SIGTERM (-15) first, then SIGKILL if needed.".to_string());
        }

        // truncate
        if bin == "truncate" {
            assessment.flags.push(RiskFlag {
                code: "TRUNCATE".to_string(),
                message: "File truncation is irreversible".to_string(),
                level: RiskLevel::Medium,
            });
        }

        // dd
        if bin == "dd" {
            assessment.flags.push(RiskFlag {
                code: "DD_WRITE".to_string(),
                message: "dd can overwrite disk partitions".to_string(),
                level: RiskLevel::High,
            });
            assessment.warnings.push(
                "dd writes raw bytes to devices — double-check of= target.".to_string(),
            );
        }

        // SQL DROP
        if (bin == "psql" || bin == "mysql" || bin == "sqlite3") {
            let full = cmd.raw.to_uppercase();
            if full.contains("DROP") {
                assessment.flags.push(RiskFlag {
                    code: "SQL_DROP".to_string(),
                    message: "SQL DROP statement detected".to_string(),
                    level: RiskLevel::High,
                });
            }
        }

        // Filesystem mutation general
        if cmd.category == CommandCategory::FilesystemMutation {
            if assessment.flags.is_empty() {
                assessment.flags.push(RiskFlag {
                    code: "FS_MUTATION".to_string(),
                    message: format!("Filesystem mutation: {}", bin),
                    level: RiskLevel::Low,
                });
            }
        }
    }

    fn check_failure_history(&self, cmd: &ParsedCommand, assessment: &mut RiskAssessment) {
        if let Some(count) = self.failure_history.get(&cmd.raw) {
            if *count >= 2 {
                assessment.flags.push(RiskFlag {
                    code: "REPEATED_FAILURE".to_string(),
                    message: format!(
                        "This command has failed {} times in this session",
                        count
                    ),
                    level: RiskLevel::Medium,
                });
                assessment.warnings.push(format!(
                    "Command has failed {} times previously. Check if conditions have changed.",
                    count
                ));
            }
        }
    }

    fn check_dangerous_flags(&self, cmd: &ParsedCommand, assessment: &mut RiskAssessment) {
        let bin = cmd.binary.rsplit('/').next().unwrap_or(&cmd.binary);

        // git push --force
        if bin == "git" && cmd.args.contains(&"push".to_string()) {
            if cmd.flags.iter().any(|f| f == "--force" || f == "-f") {
                assessment.flags.push(RiskFlag {
                    code: "GIT_FORCE_PUSH".to_string(),
                    message: "Force push can overwrite remote history".to_string(),
                    level: RiskLevel::High,
                });
            }
        }

        // sudo
        if bin == "sudo" {
            assessment.flags.push(RiskFlag {
                code: "SUDO".to_string(),
                message: "Command running with elevated privileges".to_string(),
                level: RiskLevel::Medium,
            });
        }
    }

    fn check_dependencies(&self, cmd: &ParsedCommand, assessment: &mut RiskAssessment) {
        // Check if target files exist for read operations
        if cmd.category == CommandCategory::ReadOnly {
            for arg in &cmd.args {
                if !arg.starts_with('-') && !arg.is_empty() {
                    let path = Path::new(arg);
                    if !path.exists() && path.extension().is_some() {
                        assessment.flags.push(RiskFlag {
                            code: "FILE_NOT_FOUND".to_string(),
                            message: format!("Target file may not exist: {}", arg),
                            level: RiskLevel::Low,
                        });
                    }
                }
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::pty::parser::parse_command;

    #[test]
    fn test_rm_rf_root_blocked() {
        let evaluator = RiskEvaluator::new();
        let cmd = parse_command("rm -rf /");
        let assessment = evaluator.evaluate(&cmd);
        assert!(assessment.block_execution);
        assert_eq!(assessment.risk_level, RiskLevel::Critical);
    }

    #[test]
    fn test_safe_command() {
        let evaluator = RiskEvaluator::new();
        let cmd = parse_command("ls -la");
        let assessment = evaluator.evaluate(&cmd);
        assert!(!assessment.block_execution);
        assert_eq!(assessment.risk_level, RiskLevel::None);
    }

    #[test]
    fn test_chmod_777_warning() {
        let evaluator = RiskEvaluator::new();
        let cmd = parse_command("chmod 777 /tmp/test");
        let assessment = evaluator.evaluate(&cmd);
        assert_eq!(assessment.risk_level, RiskLevel::High);
    }

    #[test]
    fn test_repeated_failure() {
        let mut evaluator = RiskEvaluator::new();
        evaluator.record_failure("npm install");
        evaluator.record_failure("npm install");
        let cmd = parse_command("npm install");
        let assessment = evaluator.evaluate(&cmd);
        assert!(assessment.flags.iter().any(|f| f.code == "REPEATED_FAILURE"));
    }
}
