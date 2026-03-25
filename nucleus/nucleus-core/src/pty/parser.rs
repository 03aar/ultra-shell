use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum CommandCategory {
    FilesystemMutation,
    Network,
    ProcessSpawn,
    EnvChange,
    ReadOnly,
    PackageManager,
    Git,
    Docker,
    Build,
    Unknown,
}

impl std::fmt::Display for CommandCategory {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            CommandCategory::FilesystemMutation => write!(f, "filesystem_mutation"),
            CommandCategory::Network => write!(f, "network"),
            CommandCategory::ProcessSpawn => write!(f, "process_spawn"),
            CommandCategory::EnvChange => write!(f, "env_change"),
            CommandCategory::ReadOnly => write!(f, "read_only"),
            CommandCategory::PackageManager => write!(f, "package_manager"),
            CommandCategory::Git => write!(f, "git"),
            CommandCategory::Docker => write!(f, "docker"),
            CommandCategory::Build => write!(f, "build"),
            CommandCategory::Unknown => write!(f, "unknown"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Redirection {
    pub direction: String, // "in", "out", "append", "err"
    pub target: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PipeSegment {
    pub binary: String,
    pub args: Vec<String>,
    pub flags: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ParsedCommand {
    pub raw: String,
    pub binary: String,
    pub args: Vec<String>,
    pub flags: Vec<String>,
    pub pipes: Vec<PipeSegment>,
    pub redirections: Vec<Redirection>,
    pub env_assignments: HashMap<String, String>,
    pub category: CommandCategory,
    pub is_background: bool,
    pub is_chained: bool,
}

impl ParsedCommand {
    pub fn files_potentially_mutated(&self) -> Vec<String> {
        let mut files = Vec::new();
        for redir in &self.redirections {
            if redir.direction == "out" || redir.direction == "append" {
                files.push(redir.target.clone());
            }
        }
        match self.category {
            CommandCategory::FilesystemMutation => {
                for arg in &self.args {
                    if !arg.starts_with('-') && !arg.is_empty() {
                        files.push(arg.clone());
                    }
                }
            }
            _ => {}
        }
        files
    }
}

pub fn parse_command(raw: &str) -> ParsedCommand {
    let raw = raw.trim().to_string();
    let mut env_assignments: HashMap<String, String> = HashMap::new();
    let mut redirections: Vec<Redirection> = Vec::new();
    let is_background = raw.ends_with('&') && !raw.ends_with("&&");
    let is_chained = raw.contains("&&") || raw.contains("||") || raw.contains(';');

    let work = if is_background {
        raw.trim_end_matches('&').trim().to_string()
    } else {
        raw.clone()
    };

    // Split by pipes
    let pipe_parts: Vec<&str> = split_by_pipes(&work);

    let mut all_segments: Vec<PipeSegment> = Vec::new();

    for part in &pipe_parts {
        let part = part.trim();
        let tokens = tokenize(part);
        let (seg, redirs) = parse_tokens(&tokens);
        all_segments.push(seg);
        redirections.extend(redirs);
    }

    // First segment is the main command
    let main = if all_segments.is_empty() {
        PipeSegment {
            binary: String::new(),
            args: Vec::new(),
            flags: Vec::new(),
        }
    } else {
        all_segments[0].clone()
    };

    let pipes = if all_segments.len() > 1 {
        all_segments[1..].to_vec()
    } else {
        Vec::new()
    };

    // Parse env assignments from prefix: VAR=val command ...
    let mut binary = main.binary.clone();
    let mut args = main.args.clone();
    let mut flags = main.flags.clone();

    // Check for env var prefix assignments
    let tokens_first = tokenize(pipe_parts[0].trim());
    let mut cmd_start = 0;
    for tok in &tokens_first {
        if tok.contains('=') && !tok.starts_with('-') && !tok.starts_with('/') {
            let parts: Vec<&str> = tok.splitn(2, '=').collect();
            if parts.len() == 2 && is_valid_var_name(parts[0]) {
                env_assignments.insert(parts[0].to_string(), parts[1].to_string());
                cmd_start += 1;
            } else {
                break;
            }
        } else {
            break;
        }
    }

    if cmd_start > 0 && cmd_start < tokens_first.len() {
        let remaining: Vec<String> = tokens_first[cmd_start..].to_vec();
        let (seg, _) = parse_tokens(&remaining);
        binary = seg.binary;
        args = seg.args;
        flags = seg.flags;
    } else if cmd_start > 0 && cmd_start >= tokens_first.len() {
        // Pure env assignment like: export FOO=bar or FOO=bar
        binary = String::new();
    }

    // Handle export
    if binary == "export" {
        for arg in &args {
            if arg.contains('=') {
                let parts: Vec<&str> = arg.splitn(2, '=').collect();
                if parts.len() == 2 {
                    env_assignments.insert(parts[0].to_string(), parts[1].to_string());
                }
            }
        }
    }

    let category = classify_command(&binary, &args, &flags, &env_assignments);

    ParsedCommand {
        raw,
        binary,
        args,
        flags,
        pipes,
        redirections,
        env_assignments,
        category,
        is_background,
        is_chained,
    }
}

fn is_valid_var_name(s: &str) -> bool {
    if s.is_empty() {
        return false;
    }
    let first = s.chars().next().unwrap();
    if !first.is_alphabetic() && first != '_' {
        return false;
    }
    s.chars().all(|c| c.is_alphanumeric() || c == '_')
}

fn split_by_pipes(s: &str) -> Vec<&str> {
    let mut result = Vec::new();
    let mut start = 0;
    let mut in_single_quote = false;
    let mut in_double_quote = false;
    let chars: Vec<char> = s.chars().collect();
    let mut i = 0;

    while i < chars.len() {
        match chars[i] {
            '\'' if !in_double_quote => in_single_quote = !in_single_quote,
            '"' if !in_single_quote => in_double_quote = !in_double_quote,
            '|' if !in_single_quote && !in_double_quote => {
                if i + 1 < chars.len() && chars[i + 1] == '|' {
                    i += 1; // skip ||
                } else {
                    result.push(&s[start..i]);
                    start = i + 1;
                }
            }
            _ => {}
        }
        i += 1;
    }
    result.push(&s[start..]);
    result
}

fn tokenize(s: &str) -> Vec<String> {
    let mut tokens = Vec::new();
    let mut current = String::new();
    let mut in_single_quote = false;
    let mut in_double_quote = false;
    let mut escape_next = false;

    for ch in s.chars() {
        if escape_next {
            current.push(ch);
            escape_next = false;
            continue;
        }
        match ch {
            '\\' if !in_single_quote => {
                escape_next = true;
            }
            '\'' if !in_double_quote => {
                in_single_quote = !in_single_quote;
            }
            '"' if !in_single_quote => {
                in_double_quote = !in_double_quote;
            }
            ' ' | '\t' if !in_single_quote && !in_double_quote => {
                if !current.is_empty() {
                    tokens.push(current.clone());
                    current.clear();
                }
            }
            _ => {
                current.push(ch);
            }
        }
    }
    if !current.is_empty() {
        tokens.push(current);
    }
    tokens
}

fn parse_tokens(tokens: &[String]) -> (PipeSegment, Vec<Redirection>) {
    let mut binary = String::new();
    let mut args = Vec::new();
    let mut flags = Vec::new();
    let mut redirections = Vec::new();
    let mut skip_next = false;

    for (i, tok) in tokens.iter().enumerate() {
        if skip_next {
            skip_next = false;
            continue;
        }

        // Handle redirections
        if tok == ">" || tok == ">>" || tok == "<" || tok == "2>" || tok == "2>>" || tok == "&>" {
            let direction = match tok.as_str() {
                ">" => "out",
                ">>" => "append",
                "<" => "in",
                "2>" | "2>>" => "err",
                "&>" => "out",
                _ => "out",
            };
            if i + 1 < tokens.len() {
                redirections.push(Redirection {
                    direction: direction.to_string(),
                    target: tokens[i + 1].clone(),
                });
                skip_next = true;
            }
            continue;
        }

        // Handle inline redirections like >file or >>file
        if tok.starts_with(">>") {
            redirections.push(Redirection {
                direction: "append".to_string(),
                target: tok[2..].to_string(),
            });
            continue;
        }
        if tok.starts_with('>') {
            redirections.push(Redirection {
                direction: "out".to_string(),
                target: tok[1..].to_string(),
            });
            continue;
        }
        if tok.starts_with('<') {
            redirections.push(Redirection {
                direction: "in".to_string(),
                target: tok[1..].to_string(),
            });
            continue;
        }

        if binary.is_empty() {
            binary = tok.clone();
        } else if tok.starts_with('-') {
            flags.push(tok.clone());
        } else {
            args.push(tok.clone());
        }
    }

    (PipeSegment { binary, args, flags }, redirections)
}

fn classify_command(
    binary: &str,
    args: &[String],
    _flags: &[String],
    env_assignments: &HashMap<String, String>,
) -> CommandCategory {
    let bin = binary
        .rsplit('/')
        .next()
        .unwrap_or(binary)
        .to_lowercase();

    // Filesystem mutation commands
    let fs_mutators = [
        "rm", "rmdir", "mv", "cp", "mkdir", "touch", "chmod", "chown", "chgrp", "truncate",
        "shred", "dd", "mkfs", "format", "ln", "unlink", "install",
    ];
    if fs_mutators.contains(&bin.as_str()) {
        return CommandCategory::FilesystemMutation;
    }

    // Editors that mutate files
    let editors = ["sed", "awk", "tee", "patch"];
    if editors.contains(&bin.as_str()) {
        return CommandCategory::FilesystemMutation;
    }

    // Network commands
    let network = [
        "curl", "wget", "ssh", "scp", "rsync", "ping", "traceroute", "netstat", "ss", "nc",
        "ncat", "dig", "nslookup", "ifconfig", "ip", "iptables", "nmap", "telnet", "ftp",
        "sftp",
    ];
    if network.contains(&bin.as_str()) {
        return CommandCategory::Network;
    }

    // Process management
    let process = [
        "kill", "killall", "pkill", "nohup", "disown", "bg", "fg", "nice", "renice",
        "systemctl", "service",
    ];
    if process.contains(&bin.as_str()) {
        return CommandCategory::ProcessSpawn;
    }

    // Environment changes
    if bin == "export" || bin == "unset" || bin == "source" || bin == "." {
        return CommandCategory::EnvChange;
    }
    if !env_assignments.is_empty() && binary.is_empty() {
        return CommandCategory::EnvChange;
    }

    // Package managers
    let pkg_managers = [
        "apt",
        "apt-get",
        "yum",
        "dnf",
        "pacman",
        "brew",
        "pip",
        "pip3",
        "npm",
        "yarn",
        "pnpm",
        "cargo",
        "go",
        "gem",
        "composer",
        "conda",
    ];
    if pkg_managers.contains(&bin.as_str()) {
        // "go" can be build as well
        if bin == "go" {
            let sub = args.first().map(|s| s.as_str()).unwrap_or("");
            if sub == "build" || sub == "run" || sub == "test" {
                return CommandCategory::Build;
            }
            if sub == "get" || sub == "install" || sub == "mod" {
                return CommandCategory::PackageManager;
            }
        }
        return CommandCategory::PackageManager;
    }

    // Git
    if bin == "git" {
        return CommandCategory::Git;
    }

    // Docker
    if bin == "docker" || bin == "docker-compose" || bin == "podman" {
        return CommandCategory::Docker;
    }

    // Build tools
    let builders = [
        "make", "cmake", "gcc", "g++", "clang", "rustc", "javac", "mvn", "gradle", "bazel",
        "ninja", "meson", "tsc", "webpack", "vite", "esbuild", "rollup",
    ];
    if builders.contains(&bin.as_str()) {
        return CommandCategory::Build;
    }
    if bin == "cargo" {
        let sub = args.first().map(|s| s.as_str()).unwrap_or("");
        if sub == "build" || sub == "run" || sub == "test" || sub == "check" {
            return CommandCategory::Build;
        }
    }

    // Read-only commands
    let readonly = [
        "ls", "cat", "head", "tail", "less", "more", "find", "grep", "rg", "ag", "wc", "file",
        "stat", "du", "df", "which", "whereis", "type", "echo", "printf", "date", "cal", "uptime",
        "whoami", "id", "hostname", "uname", "pwd", "env", "printenv", "tree", "diff", "md5sum",
        "sha256sum", "true", "false", "test", "history", "man", "help", "info",
    ];
    if readonly.contains(&bin.as_str()) {
        return CommandCategory::ReadOnly;
    }

    // cd is read-only but changes env
    if bin == "cd" {
        return CommandCategory::EnvChange;
    }

    CommandCategory::Unknown
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_simple_command() {
        let parsed = parse_command("ls -la /tmp");
        assert_eq!(parsed.binary, "ls");
        assert_eq!(parsed.flags, vec!["-la"]);
        assert_eq!(parsed.args, vec!["/tmp"]);
        assert_eq!(parsed.category, CommandCategory::ReadOnly);
    }

    #[test]
    fn test_pipe_command() {
        let parsed = parse_command("cat file.txt | grep error | wc -l");
        assert_eq!(parsed.binary, "cat");
        assert_eq!(parsed.pipes.len(), 2);
        assert_eq!(parsed.pipes[0].binary, "grep");
        assert_eq!(parsed.pipes[1].binary, "wc");
    }

    #[test]
    fn test_redirect() {
        let parsed = parse_command("echo hello > output.txt");
        assert_eq!(parsed.binary, "echo");
        assert_eq!(parsed.redirections.len(), 1);
        assert_eq!(parsed.redirections[0].direction, "out");
        assert_eq!(parsed.redirections[0].target, "output.txt");
    }

    #[test]
    fn test_env_assignment() {
        let parsed = parse_command("export FOO=bar");
        assert_eq!(parsed.category, CommandCategory::EnvChange);
        assert_eq!(
            parsed.env_assignments.get("FOO"),
            Some(&"bar".to_string())
        );
    }

    #[test]
    fn test_rm_classified() {
        let parsed = parse_command("rm -rf /tmp/test");
        assert_eq!(parsed.category, CommandCategory::FilesystemMutation);
    }

    #[test]
    fn test_git_classified() {
        let parsed = parse_command("git push origin main");
        assert_eq!(parsed.category, CommandCategory::Git);
    }
}
