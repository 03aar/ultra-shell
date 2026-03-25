use crate::context::evaluator::{RiskAssessment, RiskLevel};
use crate::context::graph::ExecutionNode;

const YELLOW: &str = "\x1b[33m";
const GREEN: &str = "\x1b[32m";
const DIM: &str = "\x1b[2m";
const BOLD: &str = "\x1b[1m";
const RED: &str = "\x1b[31m";
const CYAN: &str = "\x1b[36m";
const RESET: &str = "\x1b[0m";

pub struct AnnotationRenderer;

impl AnnotationRenderer {
    pub fn render_risk_warning(assessment: &RiskAssessment) -> Vec<u8> {
        let level_color = match assessment.risk_level {
            RiskLevel::Critical => RED,
            RiskLevel::High => RED,
            RiskLevel::Medium => YELLOW,
            _ => YELLOW,
        };
        let level_str = format!("{}", assessment.risk_level).to_uppercase();
        let warning_text = assessment
            .warnings
            .first()
            .cloned()
            .unwrap_or_else(|| "Potentially dangerous operation".to_string());
        let suggestion_text = assessment
            .suggestions
            .first()
            .map(|s| format!("  {DIM}│{RESET}  Suggestion: {GREEN}{s}{RESET}\r\n"))
            .unwrap_or_default();

        let blocked_line = if assessment.block_execution {
            format!("  {RED}{BOLD}│  EXECUTION BLOCKED{RESET}\r\n")
        } else {
            String::new()
        };

        let output = format!(
            "\r\n\
             {level_color}  ┌─ NUCLEUS WARNING ──────────────────────────────────┐{RESET}\r\n\
             {level_color}  │{RESET}  Risk: {level_color}{BOLD}{level_str}{RESET}\r\n\
             {level_color}  │{RESET}  {warning_text}\r\n\
             {suggestion_text}\
             {blocked_line}\
             {level_color}  └─────────────────────────────────────────────────────┘{RESET}\r\n"
        );
        output.into_bytes()
    }

    pub fn render_context_note(note: &str) -> Vec<u8> {
        format!("  {DIM}nucleus › {note}{RESET}\r\n").into_bytes()
    }

    pub fn render_rollback_hint(execution_id: &str) -> Vec<u8> {
        let short_id = &execution_id[..8.min(execution_id.len())];
        format!("  {GREEN}↩  nuc rollback {short_id}{RESET}\r\n").into_bytes()
    }

    pub fn render_success_annotation(node: &ExecutionNode) -> Vec<u8> {
        let files_count = node.files_mutated.len();
        let files_text = if files_count > 0 {
            format!(" · {files_count} file{} changed", if files_count == 1 { "" } else { "s" })
        } else {
            String::new()
        };
        format!(
            "  {DIM}✓ {}ms{files_text} · nuc graph{RESET}\r\n",
            node.duration_ms
        )
        .into_bytes()
    }

    pub fn render_natural_language_prompt(translated_command: &str) -> Vec<u8> {
        format!(
            "\r\n  {CYAN}nucleus ›{RESET} {GREEN}{translated_command}{RESET}\r\n  {DIM}Run this? [Y/n]{RESET} "
        )
        .into_bytes()
    }

    pub fn render_blocked_banner() -> Vec<u8> {
        format!(
            "\r\n{RED}{BOLD}  ██ BLOCKED ██{RESET} {RED}This command was blocked by Nucleus safety rules.{RESET}\r\n\
             {DIM}  Use --force to override (not recommended).{RESET}\r\n"
        )
        .into_bytes()
    }
}
