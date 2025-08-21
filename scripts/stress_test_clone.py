#!/usr/bin/env python3
"""Stress test for the workspace repository clone endpoint.

Usage:
  export FINN_API_KEY="your-api-key"
  python scripts/stress_test_clone.py --base-url http://localhost:8000 \\
    --workspace-id <workspaceID> --repository-id <repositoryID> [options]

Options:
  --requests N       Total number of clone requests (default: 50)
  --concurrency N    Max concurrent clone requests (default: 10)
  --poll-interval S  Seconds between task status polls (default: 1)
  --task-timeout S   Max seconds to wait for a task to finish (default: 120)
"""
from __future__ import annotations

import argparse
import os
import sys
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from dataclasses import dataclass
from typing import Optional

try:
    import requests
except ImportError:
    print("Install requests: pip install requests", file=sys.stderr)
    sys.exit(1)


@dataclass
class TaskOutcome:
    task_id: str
    status: str
    success: bool
    duration_sec: Optional[float] = None
    error: Optional[str] = None


def clone(
    base_url: str,
    api_key: str,
    workspace_id: str,
    repository_id: str,
) -> tuple[Optional[str], Optional[str]]:
    """POST clone, return (task_id, error_message)."""
    url = f"{base_url.rstrip('/')}/api/v1/workspaces/{workspace_id}/repositories/{repository_id}/clone"
    headers = {"X-API-Key": api_key, "Content-Type": "application/json"}
    try:
        r = requests.post(url, headers=headers, timeout=30)
    except requests.RequestException as e:
        return None, str(e)
    if r.status_code != 202:
        try:
            msg = r.json().get("errorMessage", r.text)
        except Exception:
            msg = r.text or str(r.status_code)
        return None, f"HTTP {r.status_code}: {msg}"
    data = r.json()
    task_id = data.get("id")
    if not task_id:
        return None, "Response missing task id"
    return str(task_id), None


def get_task_status(
    base_url: str,
    api_key: str,
    workspace_id: str,
    repository_id: str,
    task_id: str,
) -> tuple[Optional[str], Optional[str]]:
    """GET task, return (status, error_message)."""
    url = f"{base_url.rstrip('/')}/api/v1/workspaces/{workspace_id}/repositories/{repository_id}/tasks/{task_id}"
    headers = {"X-API-Key": api_key}
    try:
        r = requests.get(url, headers=headers, timeout=10)
    except requests.RequestException as e:
        return None, str(e)
    if r.status_code != 200:
        try:
            msg = r.json().get("errorMessage", r.text)
        except Exception:
            msg = r.text or str(r.status_code)
        return None, f"HTTP {r.status_code}: {msg}"
    data = r.json()
    return data.get("status"), None


def wait_for_task(
    base_url: str,
    api_key: str,
    workspace_id: str,
    repository_id: str,
    task_id: str,
    poll_interval: float,
    task_timeout: float,
) -> TaskOutcome:
    """Poll until task is success/failed or timeout."""
    start = time.perf_counter()
    while True:
        elapsed = time.perf_counter() - start
        if elapsed >= task_timeout:
            return TaskOutcome(
                task_id=task_id,
                status="timeout",
                success=False,
                duration_sec=elapsed,
                error="Task did not complete within timeout",
            )
        status, err = get_task_status(base_url, api_key, workspace_id, repository_id, task_id)
        if err:
            return TaskOutcome(
                task_id=task_id,
                status="error",
                success=False,
                duration_sec=elapsed,
                error=err,
            )
        if status in ("success", "failed"):
            return TaskOutcome(
                task_id=task_id,
                status=status or "unknown",
                success=(status == "success"),
                duration_sec=time.perf_counter() - start,
            )
        time.sleep(poll_interval)


def run_one(
    base_url: str,
    api_key: str,
    workspace_id: str,
    repository_id: str,
    poll_interval: float,
    task_timeout: float,
    index: int,
) -> TaskOutcome:
    """Trigger one clone and wait for it; return outcome."""
    task_id, err = clone(base_url, api_key, workspace_id, repository_id)
    if err:
        return TaskOutcome(
            task_id=task_id or f"req-{index}",
            status="error",
            success=False,
            error=err,
        )
    return wait_for_task(
        base_url, api_key, workspace_id, repository_id,
        task_id, poll_interval, task_timeout,
    )


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Stress test the workspace repository clone endpoint.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    parser.add_argument(
        "--base-url",
        default=os.environ.get("FINN_BASE_URL", "http://localhost:8080"),
        help="API base URL (default: FINN_BASE_URL or http://localhost:8080)",
    )
    parser.add_argument(
        "--workspace-id",
        required=True,
        help="Workspace ID",
    )
    parser.add_argument(
        "--repository-id",
        required=True,
        help="Repository ID",
    )
    parser.add_argument(
        "--requests",
        type=int,
        default=50,
        help="Total number of clone requests (default: 50)",
    )
    parser.add_argument(
        "--concurrency",
        type=int,
        default=10,
        help="Max concurrent clone + wait operations (default: 10)",
    )
    parser.add_argument(
        "--poll-interval",
        type=float,
        default=1.0,
        help="Seconds between task status polls (default: 1)",
    )
    parser.add_argument(
        "--task-timeout",
        type=float,
        default=120.0,
        help="Max seconds to wait per task (default: 120)",
    )
    args = parser.parse_args()

    api_key = os.environ.get("FINN_API_KEY")
    if not api_key:
        print("Set FINN_API_KEY environment variable.", file=sys.stderr)
        sys.exit(1)

    if args.requests < 1 or args.concurrency < 1:
        print("--requests and --concurrency must be >= 1.", file=sys.stderr)
        sys.exit(1)

    base_url = args.base_url
    workspace_id = args.workspace_id
    repository_id = args.repository_id
    n_requests = args.requests
    concurrency = min(args.concurrency, n_requests)
    poll_interval = args.poll_interval
    task_timeout = args.task_timeout

    print(f"Stress test: {n_requests} clone(s), concurrency={concurrency}")
    print(f"  Base URL: {base_url}")
    print(f"  Workspace: {workspace_id}, Repository: {repository_id}")
    print(f"  Poll interval: {poll_interval}s, Task timeout: {task_timeout}s")
    print()

    outcomes: list[TaskOutcome] = []
    start_wall = time.perf_counter()

    with ThreadPoolExecutor(max_workers=concurrency) as executor:
        futures = {
            executor.submit(
                run_one,
                base_url,
                api_key,
                workspace_id,
                repository_id,
                poll_interval,
                task_timeout,
                i,
            ): i
            for i in range(n_requests)
        }
        for future in as_completed(futures):
            try:
                outcome = future.result()
                outcomes.append(outcome)
                status_icon = "✓" if outcome.success else "✗"
                print(f"  [{len(outcomes)}/{n_requests}] {status_icon} task={outcome.task_id} status={outcome.status}", end="")
                if outcome.duration_sec is not None:
                    print(f" duration={outcome.duration_sec:.1f}s", end="")
                if outcome.error:
                    print(f" error={outcome.error[:60]}", end="")
                print()
            except Exception as e:
                outcomes.append(TaskOutcome(
                    task_id="?",
                    status="exception",
                    success=False,
                    error=str(e),
                ))
                print(f"  [{len(outcomes)}/{n_requests}] ✗ exception: {e}")

    wall_sec = time.perf_counter() - start_wall
    success_count = sum(1 for o in outcomes if o.success)
    failed_count = len(outcomes) - success_count
    success_rate = (success_count / len(outcomes) * 100) if outcomes else 0.0

    durations = [o.duration_sec for o in outcomes if o.duration_sec is not None]
    avg_duration = sum(durations) / len(durations) if durations else 0
    min_duration = min(durations) if durations else 0
    max_duration = max(durations) if durations else 0

    print()
    print("--- Results ---")
    print(f"  Total:     {len(outcomes)}")
    print(f"  Success:   {success_count}")
    print(f"  Failed:    {failed_count}")
    print(f"  Success rate: {success_rate:.1f}%")
    print(f"  Wall time: {wall_sec:.1f}s")
    if durations:
        print(f"  Task duration (success/failed): min={min_duration:.1f}s avg={avg_duration:.1f}s max={max_duration:.1f}s")

    status_counts: dict[str, int] = {}
    for o in outcomes:
        status_counts[o.status] = status_counts.get(o.status, 0) + 1
    print(f"  By status: {status_counts}")

    sys.exit(0 if failed_count == 0 else 1)


if __name__ == "__main__":
    main()
