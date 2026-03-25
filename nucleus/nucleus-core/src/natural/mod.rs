use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TranslationResult {
    pub command: String,
    pub explanation: String,
    pub risk_level: String,
}

/// Translates natural language input (prefixed with "?") into a shell command
/// by calling the nucleus-api translation endpoint.
pub async fn translate_to_command(
    input: &str,
    cwd: &str,
    api_url: &str,
    api_key: &str,
) -> Result<TranslationResult, String> {
    let query = input.trim_start_matches('?').trim();
    if query.is_empty() {
        return Err("Empty query".to_string());
    }

    let body = serde_json::json!({
        "input": query,
        "context": {
            "cwd": cwd,
        }
    });

    let client = reqwest::Client::new();
    let resp = client
        .post(format!("{}/api/v1/natural/translate", api_url))
        .header("Content-Type", "application/json")
        .header("X-Nucleus-Key", api_key)
        .json(&body)
        .send()
        .await
        .map_err(|e| format!("API request failed: {}", e))?;

    if !resp.status().is_success() {
        // Fallback: simple pattern matching for common queries
        return Ok(translate_locally(query));
    }

    let json: serde_json::Value = resp
        .json()
        .await
        .map_err(|e| format!("Failed to parse response: {}", e))?;

    let data = &json["data"];
    Ok(TranslationResult {
        command: data["command"]
            .as_str()
            .unwrap_or("")
            .to_string(),
        explanation: data["explanation"]
            .as_str()
            .unwrap_or("")
            .to_string(),
        risk_level: data["risk_level"]
            .as_str()
            .unwrap_or("low")
            .to_string(),
    })
}

/// Fallback local translation for common patterns when API is unavailable.
fn translate_locally(query: &str) -> TranslationResult {
    let q = query.to_lowercase();

    let (command, explanation) = if q.contains("large file") || q.contains("big file") {
        ("find . -size +100M -type f", "Find all files larger than 100MB")
    } else if q.contains("what changed") || q.contains("recent change") {
        ("git diff --stat HEAD~5", "Show changes in the last 5 commits")
    } else if q.contains("port") && q.contains("3000") {
        ("lsof -i :3000", "Show processes using port 3000")
    } else if q.contains("port") && q.contains("8080") {
        ("lsof -i :8080", "Show processes using port 8080")
    } else if q.contains("port") {
        ("ss -tlnp", "Show all listening TCP ports")
    } else if q.contains("disk") || q.contains("space") {
        ("df -h", "Show disk usage for all mounted filesystems")
    } else if q.contains("memory") || q.contains("ram") {
        ("free -h", "Show memory usage")
    } else if q.contains("process") || q.contains("running") {
        ("ps aux --sort=-%cpu | head -20", "Show top 20 processes by CPU")
    } else if q.contains("undo") || q.contains("rollback") {
        ("nuc rollback --last", "Rollback the most recent command")
    } else if q.contains("git status") || q.contains("repo status") {
        ("git status", "Show current git repository status")
    } else if q.contains("git log") || q.contains("commit history") {
        ("git log --oneline -20", "Show last 20 commits")
    } else if q.contains("find") && q.contains("log") {
        ("find . -name '*.log' -type f", "Find all log files")
    } else if q.contains("network") || q.contains("connection") {
        ("ss -tunp", "Show all network connections")
    } else if q.contains("docker") && q.contains("container") {
        ("docker ps -a", "List all Docker containers")
    } else if q.contains("environment") || q.contains("env var") {
        ("env | sort", "Show all environment variables sorted")
    } else {
        ("echo 'Could not translate. Try being more specific.'", "Unable to translate query")
    };

    TranslationResult {
        command: command.to_string(),
        explanation: explanation.to_string(),
        risk_level: "low".to_string(),
    }
}

/// Check if input is a natural language query (starts with ?)
pub fn is_natural_language_query(input: &str) -> bool {
    input.trim().starts_with('?')
}
