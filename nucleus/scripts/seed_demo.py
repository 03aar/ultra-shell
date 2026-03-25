#!/usr/bin/env python3
"""
NUCLEUS Demo Data Seeder
Creates a realistic demo session with 40 execution nodes
for a compelling first-launch dashboard experience.
"""

import json
import time
import sys
from datetime import datetime, timedelta, timezone
from urllib.request import Request, urlopen
from urllib.error import URLError
import uuid

API_URL = "http://localhost:8080/api/v1"
API_KEY = "dev-nucleus-key-local"

def api_post(path, body):
    data = json.dumps(body).encode("utf-8")
    req = Request(f"{API_URL}{path}", data=data, method="POST")
    req.add_header("Content-Type", "application/json")
    req.add_header("X-Nucleus-Key", API_KEY)
    try:
        with urlopen(req, timeout=5) as resp:
            return json.loads(resp.read())
    except Exception as e:
        print(f"  Warning: {e}")
        return None

def seed():
    print("NUCLEUS Demo Data Seeder")
    print("=" * 40)

    # Wait for API to be ready
    print("Waiting for API...")
    for i in range(30):
        try:
            req = Request(f"{API_URL.rsplit('/api', 1)[0]}/health")
            with urlopen(req, timeout=2) as resp:
                if resp.status == 200:
                    print("  API is ready!")
                    break
        except Exception:
            pass
        time.sleep(2)
    else:
        print("  API not available after 60s. Exiting.")
        sys.exit(1)

    # Create a demo session
    print("\nCreating demo session...")
    result = api_post("/sessions", {"name": "demo-dev-workflow"})
    session_id = result["data"]["id"] if result and result.get("data") else str(uuid.uuid4())
    print(f"  Session: {session_id}")

    # Demo execution sequence - a realistic Node.js dev workflow
    demo_commands = [
        {"cmd": "pwd", "exit": 0, "dur": 5, "cat": "read_only", "stdout": "/home/dev/projects"},
        {"cmd": "git clone https://github.com/example/webapp.git", "exit": 0, "dur": 8500, "cat": "git", "stdout": "Cloning into 'webapp'...\nReceiving objects: 100% (1847/1847), done."},
        {"cmd": "cd webapp", "exit": 0, "dur": 3, "cat": "env_change", "stdout": ""},
        {"cmd": "ls -la", "exit": 0, "dur": 8, "cat": "read_only", "stdout": "total 120\ndrwxr-xr-x  15 dev dev 4096 Mar 25 10:00 .\ndrwxr-xr-x   5 dev dev 4096 Mar 25 10:00 ..\n-rw-r--r--   1 dev dev  456 Mar 25 10:00 package.json\n-rw-r--r--   1 dev dev  234 Mar 25 10:00 tsconfig.json\ndrwxr-xr-x   8 dev dev 4096 Mar 25 10:00 src\ndrwxr-xr-x   4 dev dev 4096 Mar 25 10:00 tests"},
        {"cmd": "cat package.json", "exit": 0, "dur": 5, "cat": "read_only", "stdout": '{\n  "name": "webapp",\n  "version": "1.0.0",\n  "scripts": {\n    "dev": "next dev",\n    "build": "next build",\n    "test": "jest"\n  }\n}'},
        {"cmd": "node --version", "exit": 0, "dur": 20, "cat": "read_only", "stdout": "v20.11.0"},
        {"cmd": "npm install", "exit": 0, "dur": 45000, "cat": "package_manager", "stdout": "added 1247 packages in 45s\n\n187 packages are looking for funding\n  run `npm fund` for details", "risk": "medium"},
        {"cmd": "git branch -a", "exit": 0, "dur": 15, "cat": "git", "stdout": "* main\n  remotes/origin/main\n  remotes/origin/feature/auth\n  remotes/origin/feature/dashboard"},
        {"cmd": "git checkout -b feature/api-v2", "exit": 0, "dur": 25, "cat": "git", "stdout": "Switched to a new branch 'feature/api-v2'"},
        {"cmd": "export API_PORT=3001", "exit": 0, "dur": 2, "cat": "env_change", "stdout": ""},
        {"cmd": "npm test", "exit": 1, "dur": 12000, "cat": "build", "stdout": "FAIL tests/api.test.ts\n  API Routes\n    ✓ GET /health (15 ms)\n    ✗ GET /users returns list (45 ms)\n    ✓ POST /auth/login (120 ms)\n\nTests: 1 failed, 2 passed, 3 total", "stderr": "FAIL: Expected status 200 but got 500"},
        {"cmd": "cat tests/api.test.ts", "exit": 0, "dur": 5, "cat": "read_only", "stdout": "import { describe, test, expect } from '@jest/globals';\n\ndescribe('API Routes', () => {\n  test('GET /users returns list', async () => {\n    const res = await fetch('/api/users');\n    expect(res.status).toBe(200);\n  });\n});"},
        {"cmd": "vim src/routes/users.ts", "exit": 0, "dur": 30000, "cat": "filesystem_mutation", "stdout": ""},
        {"cmd": "npm test", "exit": 0, "dur": 11000, "cat": "build", "stdout": "PASS tests/api.test.ts\n  API Routes\n    ✓ GET /health (12 ms)\n    ✓ GET /users returns list (38 ms)\n    ✓ POST /auth/login (115 ms)\n\nTests: 3 passed, 3 total"},
        {"cmd": "npm run build", "exit": 0, "dur": 18000, "cat": "build", "stdout": "   Creating an optimized production build...\n   Compiled successfully.\n\n   Route (app)                    Size     First Load JS\n   ─ /                            5.2 kB         89 kB\n   ─ /api/users                   1.1 kB         85 kB\n   ─ /dashboard                   12 kB          96 kB"},
        {"cmd": "docker build -t webapp:v2 .", "exit": 0, "dur": 35000, "cat": "docker", "stdout": "[+] Building 35.2s (14/14) FINISHED\n => [internal] load build definition from Dockerfile\n => [build 1/5] FROM node:20-alpine\n => [build 5/5] RUN npm run build\n => exporting to image\n => => naming to docker.io/library/webapp:v2"},
        {"cmd": "docker images | grep webapp", "exit": 0, "dur": 50, "cat": "docker", "stdout": "webapp   v2   a3b2c1d0e5f6   Just now   245MB\nwebapp   v1   f6e5d4c3b2a1   3 days ago 238MB"},
        {"cmd": "docker run -d --name webapp-test -p 3001:3000 webapp:v2", "exit": 0, "dur": 2000, "cat": "docker", "stdout": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6"},
        {"cmd": "curl -s http://localhost:3001/health", "exit": 0, "dur": 500, "cat": "network", "stdout": '{"status":"ok","version":"2.0.0","uptime":"2s"}'},
        {"cmd": "curl -s http://localhost:3001/api/users", "exit": 0, "dur": 200, "cat": "network", "stdout": '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]'},
        {"cmd": "docker logs webapp-test --tail 5", "exit": 0, "dur": 100, "cat": "docker", "stdout": "Server listening on port 3000\nGET /health 200 2ms\nGET /api/users 200 15ms"},
        {"cmd": "docker stop webapp-test && docker rm webapp-test", "exit": 0, "dur": 3000, "cat": "docker", "stdout": "webapp-test\nwebapp-test"},
        {"cmd": "git add -A", "exit": 0, "dur": 50, "cat": "git", "stdout": ""},
        {"cmd": "git commit -m 'feat: implement users API v2'", "exit": 0, "dur": 200, "cat": "git", "stdout": "[feature/api-v2 a1b2c3d] feat: implement users API v2\n 3 files changed, 47 insertions(+), 12 deletions(-)"},
        {"cmd": "git push origin feature/api-v2", "exit": 0, "dur": 5000, "cat": "git", "stdout": "Enumerating objects: 8, done.\nCounting objects: 100% (8/8), done.\nDelta compression using up to 8 threads\nCompressing objects: 100% (4/4), done.\nremote: Create a pull request: https://github.com/example/webapp/pull/new/feature/api-v2"},
        {"cmd": "rm -rf /tmp/old-backups", "exit": 0, "dur": 100, "cat": "filesystem_mutation", "risk": "high", "stdout": "", "risk_flags": [{"code": "RECURSIVE_DELETE", "message": "Recursive forced deletion", "level": "high"}]},
        {"cmd": "echo 'Deployment ready'", "exit": 0, "dur": 3, "cat": "read_only", "stdout": "Deployment ready"},
        {"cmd": "mkdir -p /tmp/rollback-test && echo 'important data' > /tmp/rollback-test/data.txt", "exit": 0, "dur": 10, "cat": "filesystem_mutation", "stdout": "", "rollback": True},
        {"cmd": "rm /tmp/rollback-test/data.txt", "exit": 0, "dur": 5, "cat": "filesystem_mutation", "risk": "medium", "stdout": "", "rollback": True, "risk_flags": [{"code": "FILE_DELETE", "message": "File deletion", "level": "medium"}]},
        {"cmd": "env | grep -c NODE", "exit": 0, "dur": 10, "cat": "read_only", "stdout": "3"},
        {"cmd": "df -h /", "exit": 0, "dur": 15, "cat": "read_only", "stdout": "Filesystem      Size  Used Avail Use% Mounted on\n/dev/sda1       100G   45G   55G  45% /"},
        {"cmd": "uptime", "exit": 0, "dur": 5, "cat": "read_only", "stdout": " 14:32:01 up 3 days, 2:15,  2 users,  load average: 0.52, 0.48, 0.41"},
        {"cmd": "docker ps", "exit": 0, "dur": 80, "cat": "docker", "stdout": "CONTAINER ID   IMAGE          COMMAND   STATUS          PORTS\na1b2c3d4e5f6   postgres:15    postgres  Up 3 hours      5432/tcp\nf6e5d4c3b2a1   redis:7        redis     Up 3 hours      6379/tcp"},
        {"cmd": "git log --oneline -5", "exit": 0, "dur": 20, "cat": "git", "stdout": "a1b2c3d feat: implement users API v2\n4e5f6g7 fix: auth middleware token check\nb8c9d0e feat: add dashboard layout\n1a2b3c4 chore: update dependencies\n5d6e7f8 initial commit"},
        {"cmd": "npm audit", "exit": 0, "dur": 8000, "cat": "package_manager", "stdout": "found 0 vulnerabilities"},
        {"cmd": "wc -l src/**/*.ts", "exit": 0, "dur": 50, "cat": "read_only", "stdout": "   45 src/routes/users.ts\n   32 src/routes/auth.ts\n   28 src/middleware/cors.ts\n   15 src/index.ts\n  120 total"},
        {"cmd": "git status", "exit": 0, "dur": 15, "cat": "git", "stdout": "On branch feature/api-v2\nnothing to commit, working tree clean"},
        {"cmd": "echo 'Session complete - all checks passed'", "exit": 0, "dur": 3, "cat": "read_only", "stdout": "Session complete - all checks passed"},
    ]

    print(f"\nSeeding {len(demo_commands)} execution records...")
    base_time = datetime.now(timezone.utc) - timedelta(hours=1)

    for i, cmd_data in enumerate(demo_commands):
        ts = base_time + timedelta(seconds=i * 30)
        exec_data = {
            "command": cmd_data["cmd"],
            "agent_id": f"demo-{session_id[:8]}",
            "dry_run": False,
        }
        result = api_post("/agent/execute", exec_data)
        status = "OK" if result else "SKIP"
        print(f"  [{i+1:2d}/{len(demo_commands)}] {status} {cmd_data['cmd'][:50]}")

    print(f"\nDemo data seeded successfully!")
    print(f"Open http://localhost:3000 to see the dashboard.")

if __name__ == "__main__":
    seed()
