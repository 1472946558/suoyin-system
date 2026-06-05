#!/bin/bash
set -euo pipefail

ROOT="/Users/xiaoliao/Desktop/openclaw/workspace/projects/gold-recycle-miniapp"
RUN_DIR="${1:-$ROOT/08_go_live/parallel/runtime/current}"

if [ ! -d "$RUN_DIR" ]; then
  echo "Run directory not found: $RUN_DIR"
  exit 1
fi

for name in development account server domain; do
  pidfile="$RUN_DIR/pids/$name.pid"
  if [ -f "$pidfile" ]; then
    pid="$(cat "$pidfile")"
    if kill -0 "$pid" 2>/dev/null; then
      echo "[stop] $name pid=$pid"
      kill "$pid" || true
    else
      echo "[skip] $name already exited (pid=$pid)"
    fi
  else
    echo "[skip] $name no pid file"
  fi
done
