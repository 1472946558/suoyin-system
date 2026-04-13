#!/bin/bash
# 黄金交易日报 - 定时任务执行脚本
#
# 使用方法:
#   chmod +x run-gold-report.sh
#   ./run-gold-report.sh
#
# 配置 crontab:
#   crontab -e
#   # 每天 9:00, 14:00, 20:00 执行
#   0 9,14,20 * * 1-6 /path/to/run-gold-report.sh >> /path/to/logs/gold-trader.log 2>&1
#
# 环境变量:
#   FEISHU_GOLD_WEBHOOK - 飞书群 Webhook URL

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="${SCRIPT_DIR}/logs"
LOG_FILE="${LOG_DIR}/gold-report-$(date +%Y%m%d).log"

# 创建日志目录
mkdir -p "${LOG_DIR}"

# 执行报告
echo "========== $(date '+%Y-%m-%d %H:%M:%S') ==========" >> "${LOG_FILE}"
node "${SCRIPT_DIR}/gold-report.mjs" >> "${LOG_FILE}" 2>&1
echo "" >> "${LOG_FILE}"
