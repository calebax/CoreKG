#!/usr/bin/env bash
# =============================================================================
# dev-up.sh — CoreKG 本地开发环境一键启停（local / docker 两种模式）
#
# 两种模式均为「中间件 docker 容器 + 宿主 corekg + 宿主 pipeline worker + 宿主 ketask doc2pdf worker」：
#   local   默认模式，中间件来自 docker-compose.yml（MySQL/ES/Redis/MinIO/NATS/Nebula）
#   docker  与 local 相同地以 docker compose 起中间件，标签/排障文案区分（便于扩展全容器编排）
#
# 宿主 worker 角色：
#   analyser   -> 消费 ke.prase_pdf_task（S3 下载 → MinerU 转 Markdown → 上传回 S3）
#   chunker    -> 消费 ke.knowledge_task（切块 + 向量化 + 写 ES）
#   doc2pdf    -> 消费 ke.doc_to_pdf_task（docx/ppt/ofd → PDF），复用 ketask 二进制
#   description-> 消费 ke.description_task（生成 file_description 摘要文档写 ES），复用 ketask 二进制
#
# 本脚本只管「把服务拉起来」，不自动跑验证脚本（验证请单独执行：
#   ./scripts/verify/verify-kb.sh --mode local --cleanup ）
#
# 用法：
#   ./scripts/dev-up.sh [--mode local|docker] [command]
#
# 命令（默认 up）：
#   up      启动中间件 + 宿主 corekg + 宿主 pipeline worker + 宿主 doc2pdf(ketask) + description(ketask)，并等待就绪
#   stop    停止宿主 corekg + pipeline worker + doc2pdf + description（--keep-compose 则保留中间件容器）
#   status  查看各进程 PID / 端口就绪情况
#   help    帮助
#
# 参数：
#   --mode local|docker   启动模式（默认 local）
#   --keep-compose        仅 stop 时有效：保留 docker 中间件容器不 down
# =============================================================================
set -uo pipefail

# ---------- 路径 ----------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
STATE_DIR="$ROOT_DIR/.dev-up"
PID_COREKG="$STATE_DIR/corekg.pid"
PID_ANALYSER="$STATE_DIR/analyser.pid"
PID_CHUNKER="$STATE_DIR/chunker.pid"
PID_DOC2PDF="$STATE_DIR/doc2pdf.pid"
PID_DESCRIPTION="$STATE_DIR/description.pid"
LOG_COREKG="$ROOT_DIR/logs/corekg-dev.log"
LOG_ANALYSER="$ROOT_DIR/logs/analyser-dev.log"
LOG_CHUNKER="$ROOT_DIR/logs/chunker-dev.log"
LOG_DOC2PDF="$ROOT_DIR/logs/doc2pdf-dev.log"
LOG_DESCRIPTION="$ROOT_DIR/logs/description-dev.log"

# ---------- 配置 ----------
MODE="local"
KEEP_COMPOSE=0
COMMAND="up"
COREKG_CONFIG="$ROOT_DIR/apps/corekg/conf/test/config.yaml"
COREKG_BIN="$ROOT_DIR/bundles/corekg"
KETASK_CONFIG="$ROOT_DIR/apps/ketask/conf/test/config.yaml"
KETASK_BIN="$ROOT_DIR/bundles/ketask"
PIPELINE_PY="$ROOT_DIR/apps/pipeline/.venv/bin/python"

# 中间件 compose 文件（local/docker 一致用仓库默认；如需扩展全容器改为 -f docker-compose.pipeline.yml）
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
# docker-compose.yml 顶层服务里作为「应用容器（宿主替代）」需跳过的服务
APP_SERVICES=(corekg chunker analyser doc2pdf description keask)
# 需拉起的中间件服务
MIDWARE_SERVICES=(mysql elasticsearch es-init redis minio minio-init nats metad0 storaged0 graphd nebula-activator)

# 就绪探活端口（宿主机映射）
PORTS_READY=(8080 3308 9202 6381 9002 4225 9669)

# ---------- 颜色 ----------
GREEN='\033[0;32m'; YELLOW='\033[0;33m'; CYAN='\033[0;36m'; RED='\033[0;31m'; NC='\033[0m'
dlog()  { printf "${CYAN}[dev-up]${NC} %s\n" "$*"; }
dok()   { printf "${GREEN}✔${NC} %s\n" "$*"; }
dinfo() { printf "${YELLOW}○${NC} %s\n" "$*"; }
dfail() { printf "${RED}✘${NC} %s\n" "$*"; }

# ---------- 工具 ----------
command_exists() { command -v "$1" >/dev/null 2>&1; }
port_open() { nc -z -w1 127.0.0.1 "$1" >/dev/null 2>&1; }

usage() {
  # 打印文件头部注释块（# 起始，去前缀），用于 --help
  awk 'NR>1{ if (substr($0,1,1) == "#") { sub(/^# ?/,""); print } else exit }' "$0"
  exit 0
}

# 解析参数
parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      --mode)
        MODE="${2:-local}"; shift 2 ;;
      --keep-compose)
        KEEP_COMPOSE=1; shift ;;
      up|stop|status|help)
        COMMAND="$1"; shift ;;
      -h|--help)
        usage ;;
      *)
        dfail "未知参数: $1 (用 --help 查看)"
        exit 2 ;;
    esac
  done
  case "$MODE" in
    local|docker) ;;
    *) dfail "未知模式: $MODE (local|docker)"; exit 2 ;;
  esac
}

# 检查依赖工具
check_deps() {
  for t in docker curl nc; do
    command_exists "$t" || { dfail "缺少依赖: $t"; return 1; }
  done
  command_exists jq || dinfo "未检测到 jq（仅状态展示用，非必需）"
  # docker compose 是 docker 子命令，非独立二进制，单独校验
  if ! docker compose version >/dev/null 2>&1; then
    dfail "缺少依赖: docker compose"
    return 1
  fi
  return 0
}

# ---------- 状态 ----------
_status_line() {
  local name="$1" pidfile="$2" port="$3"
  if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile" 2>/dev/null)" 2>/dev/null; then
    printf "  %-9s PID=%-7s %s: %s\n" "$name" "$(cat "$pidfile")" "$port" \
      "$(port_open "$port" && echo OPEN || echo CLOSED)"
  else
    printf "  %-9s 未运行 %s\n" "$name" "$port"
  fi
}

status() {
  echo "==================== 状态（模式: ${MODE}） ===================="
  echo "中间件（docker compose）:"
  docker compose -f "$COMPOSE_FILE" ps --format '  {{.Name}}\t{{.Status}}' 2>/dev/null \
    | grep -vE "corekg-app|pipeline|-chunker|-analyser" || echo "  无（未启动）"
  echo "宿主进程:"
  _status_line "corekg" "$PID_COREKG" "8080"
  _status_line "analyser" "$PID_ANALYSER" "-"
  _status_line "chunker" "$PID_CHUNKER" "-"
  _status_line "doc2pdf" "$PID_DOC2PDF" "-"
  _status_line "description" "$PID_DESCRIPTION" "-"
  echo "就绪检查:"
  for p in "${PORTS_READY[@]}"; do
    printf "  :%s %s\n" "$p" "$(port_open "$p" && echo OPEN || echo CLOSED)"
  done
}

# ---------- 中间件 ----------
middleware_up() {
  dlog "[$MODE] 启动中间件容器: $(IFS=,; echo "${MIDWARE_SERVICES[*]}")"
  # docker-compose.yml 顶层也可能注册了 app 容器（corekg/chunker/analyser），
  # 它们由宿主进程替代，这里只显式拉起中间件服务。
  docker compose -f "$COMPOSE_FILE" up -d "${MIDWARE_SERVICES[@]}" || return 1
  dok "中间件容器已启动"

  # 等待中间件端口就绪（排除 corekg 8080，corekg 稍后单独探活）
  local wait=90 start=$SECONDS p
  for p in "${PORTS_READY[@]}"; do
    [ "$p" = "8080" ] && continue
    while ! port_open "$p"; do
      dlog "等待 :$p 就绪…"
      if [ $((SECONDS-start)) -ge "$wait" ]; then dfail "等待 :$p 超时(${wait}s)"; return 1; fi
      sleep 2
    done
  done
  dok "中间件端口全部就绪"
}

# ---------- corekg（宿主二进制） ----------
corekg_up() {
  # 已有 pid 在跑则跳过
  if [ -f "$PID_COREKG" ] && kill -0 "$(cat "$PID_COREKG" 2>/dev/null)" 2>/dev/null; then
    dinfo "corekg 已在运行 (PID $(cat "$PID_COREKG"))，跳过启动"
    return 0
  fi
  if [ ! -x "$COREKG_BIN" ]; then
    dinfo "未找到 $COREKG_BIN，开始构建 corekg…"
    ( cd "$ROOT_DIR" && make local APP=corekg ) || { dfail "corekg 构建失败"; return 1; }
  fi
  dlog "启动宿主 corekg: $COREKG_BIN -c $COREKG_CONFIG"
  mkdir -p "$(dirname "$LOG_COREKG")"
  nohup "$COREKG_BIN" -c "$COREKG_CONFIG" >> "$LOG_COREKG" 2>&1 &
  echo $! > "$PID_COREKG"
  dok "corekg 已启动 (PID $(cat "$PID_COREKG"))，日志: $LOG_COREKG"
}

# ---------- ketask doc2pdf worker（宿主二进制） ----------
ketask_up() {
  # 已有 pid 在跑则跳过
  if [ -f "$PID_DOC2PDF" ] && kill -0 "$(cat "$PID_DOC2PDF" 2>/dev/null)" 2>/dev/null; then
    dinfo "doc2pdf(ketask) 已在运行 (PID $(cat "$PID_DOC2PDF"))，跳过启动"
    return 0
  fi
  if [ ! -x "$KETASK_BIN" ]; then
    dinfo "未找到 $KETASK_BIN，开始构建 ketask…"
    ( cd "$ROOT_DIR" && make local APP=ketask ) || { dfail "ketask 构建失败"; return 1; }
  fi
  # doc2pdf worker：HTTP 轮询 corekg 的 knowledge.GetPendingTask，处理 ke.doc_to_pdf_task，
  # 把 docx/ppt/ofd 转 PDF 写回预览路径后回报 knowledge.TaskCallBack。
  dlog "启动宿主 doc2pdf worker: $KETASK_BIN doc2pdf -c $KETASK_CONFIG -t ke.doc_to_pdf_task -b http://localhost:8080/"
  mkdir -p "$(dirname "$LOG_DOC2PDF")"
  nohup "$KETASK_BIN" doc2pdf \
    -c "$KETASK_CONFIG" \
    -t ke.doc_to_pdf_task \
    -b http://localhost:8080/ \
    -r 1 >> "$LOG_DOC2PDF" 2>&1 &
  echo $! > "$PID_DOC2PDF"
  dok "doc2pdf(ketask) 已启动 (PID $(cat "$PID_DOC2PDF"))，日志: $LOG_DOC2PDF"
}

# ---------- ketask description worker（宿主二进制） ----------
description_up() {
  # 已有 pid 在跑则跳过
  if [ -f "$PID_DESCRIPTION" ] && kill -0 "$(cat "$PID_DESCRIPTION" 2>/dev/null)" 2>/dev/null; then
    dinfo "description(ketask) 已在运行 (PID $(cat "$PID_DESCRIPTION"))，跳过启动"
    return 0
  fi
  if [ ! -x "$KETASK_BIN" ]; then
    dinfo "未找到 $KETASK_BIN，开始构建 ketask…"
    ( cd "$ROOT_DIR" && make local APP=ketask ) || { dfail "ketask 构建失败"; return 1; }
  fi
  # description worker：HTTP 轮询 corekg 的 knowledge.GetPendingTask，处理 ke.description_task，
  # 读取解析产物 markdown（校验知识入 FileURL 已指向 content.md），用 agent 生成
  # file_description/mindmap/abstract 等写 ES，支撑 knowledge_summary 策略检索。
  # 模型/endpoint 走 ketask 配置文件的 agent: 段（见 apps/ketask/conf/test|docker/config.yaml）。
  dlog "启动宿主 description worker: $KETASK_BIN description -c $KETASK_CONFIG -t ke.description_task -b http://localhost:8080/"
  mkdir -p "$(dirname "$LOG_DESCRIPTION")"
  nohup "$KETASK_BIN" description \
    -c "$KETASK_CONFIG" \
    -t ke.description_task \
    -b http://localhost:8080/ \
    -r 1 >> "$LOG_DESCRIPTION" 2>&1 &
  echo $! > "$PID_DESCRIPTION"
  dok "description(ketask) 已启动 (PID $(cat "$PID_DESCRIPTION"))，日志: $LOG_DESCRIPTION"
}

# ---------- pipeline worker（宿主 venv） ----------
_pipeline_ensure_venv() {
  if [ -x "$PIPELINE_PY" ]; then return 0; fi
  dinfo "未找到 pipeline venv，开始创建…"
  ( cd "$ROOT_DIR/apps/pipeline" && python3 -m venv .venv ) || { dfail "创建 venv 失败"; return 1; }
  dinfo "安装 pipeline 依赖…"
  ( cd "$ROOT_DIR/apps/pipeline" && "$PIPELINE_PY" -m pip install -r requirements.txt -r requirements_analyser.txt ) \
    || { dfail "安装依赖失败"; return 1; }
  return 0
}

_pipeline_worker_up() {
  local label="$1" script="$2" pidfile="$3" logfile="$4"
  if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile" 2>/dev/null)" 2>/dev/null; then
    dinfo "$label 已在运行 (PID $(cat "$pidfile"))，跳过启动"
    return 0
  fi
  dlog "启动 $label: $PIPELINE_PY $script"
  # worker 需在 apps/pipeline 目录下运行，依赖其相对 ./config 读取配置
  ( cd "$ROOT_DIR/apps/pipeline" && exec nohup "$PIPELINE_PY" "$script" >> "$logfile" 2>&1 ) &
  echo $! > "$pidfile"
  dok "$label 已启动 (PID $(cat "$pidfile"))，日志: $logfile"
}

pipeline_up() {
  _pipeline_ensure_venv || return 1
  _pipeline_worker_up "analyser" "doc_worker_main.py" "$PID_ANALYSER" "$LOG_ANALYSER"
  _pipeline_worker_up "chunker"  "chunk_worker_main.py" "$PID_CHUNKER" "$LOG_CHUNKER"
}

# ---------- 就绪等待 ----------
wait_corekg_ready() {
  local wait="${READY_TIMEOUT:-90}" start=$SECONDS
  dlog "等待 corekg :8080 就绪（最多 ${wait}s）…"
  while ! port_open 8080; do
    if [ $((SECONDS-start)) -ge "$wait" ]; then dfail "corekg :8080 等待超时，查看 $LOG_COREKG"; return 1; fi
    sleep 2
  done
  dok "corekg :8080 就绪"
}

# ---------- up ----------
up() {
  check_deps || exit 1
  mkdir -p "$STATE_DIR"
  echo "==================== CoreKG 开发环境启动（模式: ${MODE}） ===================="
  middleware_up || return 1
  corekg_up || return 1
  pipeline_up || return 1
  ketask_up || return 1
  description_up || return 1
  wait_corekg_ready || return 1
  echo
  ok=1
  for name in corekg analyser chunker doc2pdf description; do
    eval "pid=\$PID_$(echo "$name" | tr a-z A-Z)"
    [ -f "$pid" ] && kill -0 "$(cat "$pid" 2>/dev/null)" 2>/dev/null \
      && dok "$name 运行中 (PID $(cat "$pid"))" \
      || { dfail "$name 未运行"; ok=0; }
  done
  echo "------------------------------------------------------------"
  dinfo "服务已就绪。若需跑验证："
  dinfo "  ./scripts/verify/verify-kb.sh --mode $MODE --cleanup"
  return $((1-ok))
}

# ---------- stop ----------
_stop_proc() {
  local pidfile="$1" label="$2"
  if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile" 2>/dev/null)" 2>/dev/null; then
    kill "$(cat "$pidfile")" 2>/dev/null
    dinfo "已停止 $label (PID $(cat "$pidfile"))"
    rm -f "$pidfile"
  else
    dinfo "$label 未运行"
  fi
}

stop() {
  echo "==================== 停止开发环境（模式: ${MODE}） ===================="
  _stop_proc "$PID_COREKG" "corekg"
  _stop_proc "$PID_ANALYSER" "analyser"
  _stop_proc "$PID_CHUNKER" "chunker"
  _stop_proc "$PID_DOC2PDF" "doc2pdf"
  _stop_proc "$PID_DESCRIPTION" "description"
  if [ "$KEEP_COMPOSE" -eq 0 ]; then
    dlog "停止中间件容器…"
    docker compose -f "$COMPOSE_FILE" stop "${MIDWARE_SERVICES[@]}" 2>/dev/null
    dok "中间件容器已停止"
  else
    dinfo "按 --keep-compose，保留中间件容器"
  fi
  dok "完成"
}

# ---------- 入口 ----------
parse_args "$@"
case "$COMMAND" in
  up)     up ;;
  stop)   stop ;;
  status) status ;;
  help)   usage ;;
esac
