#!/usr/bin/env bash
# =============================================================================
# CoreKG 基础知识库闭环自动验证脚本（verify-kb.sh）
#
# 覆盖完整链路：
#   ① 登录(account.LoginByPassword)
#   ② 新建知识库(forest.CreateForest)
#   ③ 上传文件(forest.UploadFile, multipart)
#   ④ 等待文件解析完成(拆 chunk + 向量化 + ES 入库 / knowledge_status=success)
#   ⑤ 基于该文件问答(内部 chat.* 流式 RAG)，断言答案 + RAG 引用命中本知识库
#
# 使用（本地模式 / docker-compose 模式）：
#   本地:  先起中间件 + hostcorekg + pipeline chunker，再：
#     ./scripts/verify/verify-kb.sh --mode local
#   compose:
#     docker compose -f docker-compose.pipeline.yml up -d --build   # 全容器
#     ./scripts/verify/verify-kb.sh --mode compose
#
# 可选参数：
#   --mode local|compose   运行模式(默认 local；只影响排障提示与 BASE 默认值)
#   --file <路径>          上传的样例文件(默认 testdata/verify_sample.txt)
#   --question <文本>      验证问题(默认问样例文件内容)
#   --cleanup              结束后删除本次创建的知识库
#   -h/--help              帮助
#
# 公共配置由脚本同目录 lib.sh 提供，可用环境变量覆盖：
#   BASE / PREFIX / LOGIN_USER / LOGIN_PASS / DOMAIN_NAME
# =============================================================================
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"
# shellcheck source=auth.sh
source "$SCRIPT_DIR/auth.sh"
# shellcheck source=forest.sh
source "$SCRIPT_DIR/forest.sh"
# shellcheck source=upload.sh
source "$SCRIPT_DIR/upload.sh"
# shellcheck source=wait_parse.sh
source "$SCRIPT_DIR/wait_parse.sh"
# shellcheck source=qa.sh
source "$SCRIPT_DIR/qa.sh"

# ---------- 参数解析 ----------
MODE="${MODE:-local}"
VERIFY_FILE="${VERIFY_FILE:-$ROOT_DIR/testdata/verify_sample.txt}"
QUESTION="${QUESTION:-$(cat "$ROOT_DIR/testdata/verify_question.txt" 2>/dev/null || echo '这个文件主要讲了什么？')}"
DO_CLEANUP=0

usage() {
  awk 'NR==1{next} /^#/{sub(/^# ?/,""); print; next} {exit}' "$0"
}
while [ $# -gt 0 ]; do
  case "$1" in
    --mode) MODE="${2:-local}"; shift 2;;
    --file) VERIFY_FILE="$2"; shift 2;;
    --question) QUESTION="$2"; shift 2;;
    --cleanup) DO_CLEANUP=1; shift;;
    -h|--help) usage; exit 0;;
    *) vfail "未知参数: $1 (用 --help 查看)"; exit 2;;
  esac
done

# 模式决定默认 BASE 的排障提示/默认值
case "$MODE" in
  local)  BASE="${BASE:-http://localhost:8080}";;
  compose) BASE="${BASE:-http://localhost:8080}";;
  *) vfail "未知模式: $MODE (local|compose)"; exit 2;;
esac
PREFIX="${PREFIX:-/v3}"

echo "============================================================"
echo " CoreKG 知识库闭环自动验证"
echo " 模式: $MODE   BASE: $BASE  文件: $VERIFY_FILE"
echo " 问题: $QUESTION"
echo "============================================================"

# 前置：样例文件必须存在
if [ ! -f "$VERIFY_FILE" ]; then
  vfail "样例文件不存在: $VERIFY_FILE"
  echo "  可先运行: echo '...' > $ROOT_DIR/testdata/verify_sample.txt，或 --file 指定"
  verify_summary; exit 1
fi

# ---------- ① 登录 ----------
echo "==================== ① 登录 ===================="
login_admin || { verify_summary; exit 1; }

# ---------- ② 新建知识库 ----------
echo "==================== ② 新建知识库 ===================="
FOREST_NAME="verify-kb-$(date +%s)"
FOREST_ID="$(create_forest "$FOREST_NAME" "$VERIFY_TOKEN")"
if [ -z "$FOREST_ID" ] || [ "$FOREST_ID" == "None" ]; then
  verify_summary; exit 1
fi
vok "已创建知识库 name=$FOREST_NAME id=$FOREST_ID"

# ---------- ③ 上传文件 ----------
echo "==================== ③ 上传文件 ===================="
FILE_ID="$(upload_file "$FOREST_ID" "$VERIFY_FILE" "$VERIFY_TOKEN")"
if [ -z "$FILE_ID" ] || [ "$FILE_ID" == "None" ]; then
  verify_summary; exit 1
fi
vok "已上传文件 $VERIFY_FILE forest_file_id=$FILE_ID"

# ---------- ④ 等待文件解析完成 ----------
echo "==================== ④ 等待解析完成(拆 chunk/向量/入库) ===================="
if ! wait_file_parse "$FOREST_ID" "$FILE_ID" "$VERIFY_TOKEN" "${VERIFY_PARSE_TIMEOUT:-180}" 3; then
  if [ "$DO_CLEANUP" -eq 1 ]; then delete_forest "$FOREST_ID" "$VERIFY_TOKEN"; fi
  verify_summary; exit 1
fi

# ---------- ⑤ 基于文件问答 ----------
echo "==================== ⑤ 知识库问答 ===================="
if ! ask_file_qa "$FOREST_ID" "$FILE_ID" "$QUESTION" "$VERIFY_TOKEN"; then
  if [ "$DO_CLEANUP" -eq 1 ]; then delete_forest "$FOREST_ID" "$VERIFY_TOKEN"; fi
  verify_summary; exit 1
fi

vok "基础知识库闭环验证通过：登录 → 建库 → 上传 → 解析 → 问答"

# ---------- 清理 ----------
if [ "$DO_CLEANUP" -eq 1 ]; then
  echo "==================== 清理 ===================="
  delete_forest "$FOREST_ID" "$VERIFY_TOKEN"
fi

verify_summary
