#!/bin/bash
set -euo pipefail

ROOT="/Users/xiaoliao/Desktop/openclaw/workspace/projects/gold-recycle-miniapp"
PROMPT_DIR="$ROOT/08_go_live/parallel/prompts"
RUNTIME_ROOT="$ROOT/08_go_live/parallel/runtime"
RUN_ID="${1:-$(date +%Y%m%d-%H%M%S)}"
RUN_DIR="$RUNTIME_ROOT/$RUN_ID"
LOG_DIR="$RUN_DIR/logs"
OUT_DIR="$RUN_DIR/output"
PID_DIR="$RUN_DIR/pids"

mkdir -p "$LOG_DIR" "$OUT_DIR" "$PID_DIR"
ln -sfn "$RUN_DIR" "$RUNTIME_ROOT/current"

launch_one() {
  local name="$1"
  local prompt="$PROMPT_DIR/$name.md"
  local log="$LOG_DIR/$name.console.log"
  local out="$OUT_DIR/$name.final.md"
  local pidfile="$PID_DIR/$name.pid"

  echo "[launch] $name"
  nohup /opt/homebrew/bin/codex -a never -s danger-full-access exec \
    -C "$ROOT" \
    -o "$out" \
    < "$prompt" \
    > "$log" 2>&1 &
  echo $! > "$pidfile"
}

launch_one development
launch_one account
launch_one server
launch_one domain

echo
echo "Run ID: $RUN_ID"
echo "Current run: $RUNTIME_ROOT/current"
echo "Status: ./scripts/check-parallel-codex.sh"
echo "Stop:   ./scripts/stop-parallel-codex.sh"
