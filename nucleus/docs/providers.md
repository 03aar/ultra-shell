# AI Provider Configuration

The Nucleus orchestrator supports multiple AI providers.

## Supported Providers

### Claude (Anthropic)
```bash
export ANTHROPIC_API_KEY=sk-ant-...
nucleus-agent run "goal" --provider claude --model claude-sonnet-4-5
```

Models: claude-opus-4-5, claude-sonnet-4-5, claude-haiku-4-5

### OpenAI
```bash
export OPENAI_API_KEY=sk-...
nucleus-agent run "goal" --provider openai --model gpt-4o
```

Models: gpt-4o, gpt-4o-mini, o1, o3-mini

### Gemini (Google)
```bash
export GOOGLE_API_KEY=AI...
nucleus-agent run "goal" --provider gemini --model gemini-2.0-flash
```

Models: gemini-2.0-flash, gemini-1.5-pro, gemini-1.5-flash

### Ollama (Local)
```bash
# Ensure Ollama is running at localhost:11434
nucleus-agent run "goal" --provider ollama --model llama3
```

Models: Any model installed in local Ollama instance

## Provider Features

| Provider | Tool Use | Streaming | Context Window |
|----------|----------|-----------|----------------|
| Claude | Yes | Yes | 200K |
| OpenAI | Yes | Yes | 128K |
| Gemini | Yes | Yes | 1M |
| Ollama | Varies | Yes | Varies |
