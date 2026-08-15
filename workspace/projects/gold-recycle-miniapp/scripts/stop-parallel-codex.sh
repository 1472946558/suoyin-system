#!/bin/bash
/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: stop-parallel-codex.sh
 * 功能描述: 业务模块实现
 * 作者: 廖心慈
 * 创建日期: 2026-05-23
 */

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
