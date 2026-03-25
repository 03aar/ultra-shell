---
sidebar_position: 2
title: TypeScript SDK
---

# TypeScript SDK

The Nucleus TypeScript SDK provides a fully typed client for Node.js and browser environments.

## Installation

```bash
npm install @nucleus/sdk
```

Or from the monorepo:

```bash
cd nucleus-sdk/typescript
npm install
npm run build
```

## Quick Start

```typescript
import { NucleusClient } from '@nucleus/sdk';

const client = new NucleusClient({
  apiUrl: 'http://localhost:8080',
  apiKey: 'dev-nucleus-key-local',
});

// Execute a command
const result = await client.execute('ls -la');
console.log(result.exitCode);    // 0
console.log(result.stdout);      // file listing
console.log(result.riskLevel);   // "low"

// Get environment context
const ctx = await client.getContext();
console.log(ctx.workingDirectory);  // "/home/user/project"
console.log(ctx.git.branch);       // "main"
```

## API Reference

### Constructor

```typescript
const client = new NucleusClient({
  apiUrl: string;     // Default: "http://localhost:8080"
  apiKey: string;     // Default: "dev-nucleus-key-local"
  timeout?: number;   // Default: 30000 (ms)
});
```

### `client.execute(command, options?)`

```typescript
interface ExecuteOptions {
  dryRun?: boolean;
  agentId?: string;
  workingDirectory?: string;
}

const result = await client.execute('npm test');

// Result type
interface ExecutionResult {
  id: string;
  command: ParsedCommand;
  exitCode: number;
  stdout: string;
  stderr: string;
  durationMs: number;
  riskLevel: RiskLevel;
  riskFlags: string[];
  rollbackAvailable: boolean;
  sessionId: string;
  timestamp: string;
}

// Dry run
const assessment = await client.execute('rm -rf /tmp/data', { dryRun: true });
console.log(assessment.riskLevel);     // "high"
console.log(assessment.wouldSnapshot); // true
```

### `client.getContext()`

```typescript
const ctx = await client.getContext();

interface Context {
  workingDirectory: string;
  shell: string;
  user: string;
  hostname: string;
  os: { name: string; version: string; arch: string };
  git: {
    isRepo: boolean;
    branch: string;
    repoRoot: string;
    modifiedFiles: string[];
    stagedFiles: string[];
    ahead: number;
    behind: number;
  };
  environment: Record<string, string>;
  processes: Process[];
}
```

### `client.getHistory(options?)`

```typescript
const history = await client.getHistory({ limit: 10 });
const results = await client.getHistory({ search: 'docker' });
```

### `client.rollback(executionId)`

```typescript
const result = await client.rollback('550e8400-...');
for (const file of result.filesRestored) {
  console.log(`${file.action}: ${file.path}`);
}
```

### `client.createPlan(goal, context?)`

```typescript
const plan = await client.createPlan('Set up CI/CD');
for (const step of plan.steps) {
  console.log(`${step.command} [${step.riskLevel}]`);
}
```

### `client.runSkill(name, params?)`

```typescript
await client.runSkill('git_cleanup');
await client.runSkill('project_setup', {
  project_name: 'myapp',
  type: 'node',
});
```

### `client.createSession(name)`

```typescript
const session = await client.createSession('deploy-v2.1');
console.log(session.id);
```

## WebSocket Streaming

```typescript
const stream = client.watch(['executions', 'warnings']);

stream.on('execution', (event) => {
  console.log(`[EXEC] ${event.command} -> ${event.exitCode}`);
});

stream.on('warning', (event) => {
  console.log(`[WARN] ${event.command} risk:${event.riskLevel}`);
});

stream.on('error', (err) => {
  console.error('Stream error:', err);
});

// Close when done
stream.close();
```

## Error Handling

```typescript
import { NucleusError, CommandBlockedError } from '@nucleus/sdk';

try {
  await client.execute('rm -rf /');
} catch (err) {
  if (err instanceof CommandBlockedError) {
    console.log(`Blocked: ${err.riskLevel}`);
    console.log(`Flags: ${err.riskFlags.join(', ')}`);
  } else if (err instanceof NucleusError) {
    console.log(`API error: ${err.code} ${err.message}`);
  }
}
```

## Browser Usage

The SDK works in browsers for building custom dashboards:

```typescript
// In a React component
import { NucleusClient } from '@nucleus/sdk';

const client = new NucleusClient({
  apiUrl: 'http://localhost:8080',
  apiKey: 'dev-nucleus-key-local',
});

function ExecutionFeed() {
  const [executions, setExecutions] = useState([]);

  useEffect(() => {
    client.getHistory({ limit: 20 }).then(setExecutions);

    const stream = client.watch(['executions']);
    stream.on('execution', (event) => {
      setExecutions((prev) => [event, ...prev].slice(0, 20));
    });

    return () => stream.close();
  }, []);

  return (
    <ul>
      {executions.map((e) => (
        <li key={e.id}>
          {e.command} — exit {e.exitCode}
        </li>
      ))}
    </ul>
  );
}
```
