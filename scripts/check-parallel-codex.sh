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
  logfile="$RUN_DIR/logs/$name.console.log"
  outfile="$RUN_DIR/output/$name.final.md"

  echo "=== $name ==="

  if [ -f "$pidfile" ]; then
    pid="$(cat "$pidfile")"
    if kill -0 "$pid" 2>/dev/null; then
      echo "status: running (pid=$pid)"
    else
      echo "status: exited (pid=$pid)"
    fi
  else
    echo "status: no pid file"
  fi

  if [ -f "$outfile" ]; then
    echo "final: $outfile"
  else
    echo "final: not written yet"
  fi

  if [ -f "$logfile" ]; then
    echo "-- tail log --"
    tail -n 12 "$logfile"
  else
    echo "log: missing"
  fi

  echo
done
