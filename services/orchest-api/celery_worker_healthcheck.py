#!/usr/bin/env python3

import subprocess
import sys

EXPECTED_WORKERS = {
    "worker-builds",
    "worker-deliveries",
    "worker-other-tasks",
    "worker-interactive",
    "worker-jobs",
}

try:
    # Execute the Celery status command
    result = subprocess.run(
        ["celery", "-A", "app.core.tasks", "status"],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        timeout=10,
    )

    if result.returncode != 0:
        print(f"Celery status command failed: {result.stderr.strip()}")
        sys.exit(1)

    # Parse the output to extract worker names
    responding_workers = set()
    for line in result.stdout.strip().splitlines():
        line = line.strip()
        if line.startswith("->"):
            parts = line.split("@")
            if len(parts) == 2:
                worker_name = parts[1].split(":")[0].strip()
                responding_workers.add(worker_name)

    # Determine if any expected workers are missing
    missing_workers = EXPECTED_WORKERS - responding_workers
    if missing_workers:
        print(f"Missing workers: {', '.join(sorted(missing_workers))}")
        sys.exit(1)
    else:
        print("All expected workers are responding.")
        sys.exit(0)

except subprocess.TimeoutExpired:
    print("Celery status command timed out.")
    sys.exit(1)
except Exception as e:
    print(f"Health check failed: {e}")
    sys.exit(1)
