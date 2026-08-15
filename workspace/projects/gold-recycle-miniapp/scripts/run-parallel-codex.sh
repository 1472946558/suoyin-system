#!/bin/bash
/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: run-parallel-codex.sh
 * 功能描述: 业务模块实现
 * 作者: 廖心慈
 * 创建日期: 2026-05-23
 */

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
