# Nucleus Skills

Skills are reusable, parameterized command workflows.

## Built-in Skills

| Skill | Description |
|-------|-------------|
| `git_cleanup` | Clean up merged git branches |
| `docker_cleanup` | Remove unused Docker resources |
| `project_setup` | Set up a new project (node/python/rust/go) |
| `deploy_check` | Pre-deployment validation checks |
| `env_audit` | Audit environment for exposed secrets |
| `process_debug` | Debug a running or crashed process |

## Usage

### CLI
```bash
nuc skill list
nuc skill run git_cleanup
nuc skill run project_setup --params project_name=myapp type=node
```

### API
```bash
curl -X POST http://localhost:8080/api/v1/skills/git_cleanup/run \
  -H "X-Nucleus-Key: dev-nucleus-key-local" \
  -H "Content-Type: application/json" \
  -d '{"params": {}}'
```

### Python SDK
```python
from nucleus import NucleusClient
client = NucleusClient()
result = client.run_skill("project_setup", {"project_name": "myapp", "type": "node"})
```

## Custom Skills

Create YAML files in `~/.config/nucleus/skills/`:

```yaml
name: my_skill
description: My custom workflow
parameters:
  - name: target
    type: string
    required: true
steps:
  - command: "echo Processing {target}"
    description: Process the target
    risk: low
  - command: "some-tool --input {target}"
    description: Run tool
    risk: medium
    requires_confirmation: true
```
