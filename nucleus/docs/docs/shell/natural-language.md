---
sidebar_position: 3
title: Natural Language Commands
---

# Natural Language Commands

Nucleus supports a `?` prefix that lets you describe what you want in plain English. Nucleus translates it to the appropriate shell command, shows you the plan, and executes it with your confirmation.

## Usage

Inside a Nucleus shell session, prefix any input with `?`:

```bash
$ ? find all python files larger than 1MB
  → find . -name "*.py" -size +1M
  Execute? [Y/n] y
```

```bash
$ ? show disk usage sorted by size
  → du -sh * | sort -rh
  Execute? [Y/n] y
```

```bash
$ ? kill the process using port 3000
  → lsof -ti:3000 | xargs kill -9
  Execute? [Y/n] y
```

## How It Works

1. You type `? <natural language description>`.
2. Nucleus sends the description plus your current context (working directory, OS, shell, recent commands) to the configured AI provider.
3. The provider returns a shell command.
4. Nucleus displays the command and asks for confirmation.
5. On confirmation, the command is executed through the normal Nucleus pipeline (parsed, risk-evaluated, snapshotted, recorded).

## Context Awareness

The AI provider receives context that helps it generate accurate commands:

- **Working directory** — so it knows which files and projects are available.
- **Operating system** — macOS vs Linux commands differ (e.g., `sed -i` syntax).
- **Shell** — bash vs zsh vs fish syntax differences.
- **Recent commands** — the last 10 commands for conversational continuity.
- **Git status** — current branch, modified files, repo root.
- **Environment** — relevant environment variables (secrets are masked).

This means you can ask follow-up questions:

```bash
$ git diff
  (shows changes)
$ ? commit these changes with a good message
  → git commit -m "Fix authentication timeout in session handler"
  Execute? [Y/n] y
```

## Configuration

Configure the NL provider in `~/.config/nucleus/config.toml`:

```toml
[natural_language]
enabled = true
provider = "claude"              # "claude", "openai", "gemini", "ollama"
model = "claude-sonnet-4-5"    # Model to use for NL translation
auto_execute = false             # If true, skip confirmation prompt
show_context = false             # If true, show what context was sent
```

### Using Ollama (Local, No API Key)

```toml
[natural_language]
enabled = true
provider = "ollama"
model = "llama3"
```

Ensure Ollama is running at `localhost:11434`.

## Auto-Execute Mode

For power users who trust the AI output:

```toml
[natural_language]
auto_execute = true
```

With `auto_execute = true`, the `?` prefix generates and immediately executes the command. The command is still risk-evaluated — critical-risk commands are still blocked regardless of this setting.

## Multi-Step Plans

For complex requests, Nucleus generates a multi-step plan:

```bash
$ ? set up a new React project with TypeScript and Tailwind
  Plan:
    1. npx create-react-app myapp --template typescript
    2. cd myapp
    3. npm install -D tailwindcss postcss autoprefixer
    4. npx tailwindcss init -p
  Execute all? [Y/n/step] step
  → npx create-react-app myapp --template typescript
  Execute? [Y/n] y
```

The `step` option lets you approve each command individually.

## Safety

- All NL-generated commands pass through the same risk evaluator as manually typed commands.
- Critical-risk commands are always blocked, even with `auto_execute = true`.
- The full generated command is shown before execution (unless `auto_execute` is on).
- NL commands are recorded in the execution DAG with a flag indicating they were AI-generated.
