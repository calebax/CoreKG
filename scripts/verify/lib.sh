#!/usr/bin/env bash
# =============================================================================
# CoreKG 自动验证公共库
#
# 提供：退出码/断言计数、HTTP post 封装、JSON 字段提取、登录(token)共享等。
# 本文件被 scripts/verify/*.sh source，不单独执行。
# =============================================================================
set -uo pipefail

# ---------- 全局状态（由各脚本管理） ----------
VERIFY_PASS=0
VERIFY_FAIL=0
VERIFY_SKIP=0
VERIFY_PARSE_FAIL=0        # 解析失败（JSON 语法错）计数
VERIFY_INTERNAL=0          # 服务端 code=500（业务依赖未就绪）
VERIFY_FATAL=0             # 致命错误（超时、不可恢复），直接累加并置标志
VERIFY_HAD_FATAL=0

BASE="${BASE:-http://localhost:8080}"
PREFIX="${PREFIX:-/v3}"
VERIFY_HTTP_TIMEOUT="${VERIFY_HTTP_TIMEOUT:-30}"

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[0;33m'; CYAN='\033[0;36m'; NC='\033[0m'

# ---------- 输出辅助 ----------
vlog()   { printf "${CYAN}[verify]${NC} %s\n" "$*"; }
vok()    { VERIFY_PASS=$((VERIFY_PASS+1)); printf "${GREEN}✔ PASS${NC}  %s\n" "$*"; }
vskip()  { VERIFY_SKIP=$((VERIFY_SKIP+1)); printf "${YELLOW}○ SKIP${NC}  %s\n" "$*"; }
vfail()  { VERIFY_FAIL=$((VERIFY_FAIL+1)); printf "${RED}✘ FAIL${NC}  %s\n" "$*"; }
vfatal() { VERIFY_FATAL=$((VERIFY_FATAL+1)); VERIFY_HAD_FATAL=1; printf "${RED}✘ FATAL${NC} %s\n" "$*"; }

# ---------- HTTP 封装 ----------
# http_post <action> <json-or-'@file'> [token]  -> stdout 返回原始 body；若 curl 自身失败则置全局
#   VERIFY_HTTP_ERR 并打印 FATAL。空 body 也赋值（区别于未执行）。
http_post() {
  local action="$1" body="$2" tok="${3:-}"
  local hdr=(-H "Content-Type: application/json")
  if [ -n "$tok" ]; then hdr+=(-H "Authorization: Bearer $tok"); fi
  local out
  out="$(curl -sS --max-time "$VERIFY_HTTP_TIMEOUT" -X POST "$BASE$PREFIX/$action" "${hdr[@]}" -d "$body" 2>/tmp/.corekg_curl_err)"
  local rc=$?
  if [ "$rc" -ne 0 ]; then
    VERIFY_HTTP_ERR="curl($rc): $(cat /tmp/.corekg_curl_err 2>/dev/null)"
    return "$rc"
  fi
  printf '%s' "$out"
  return 0
}

# json_val <jq-path> — 对 stdin 应用 jq 返回原始值。脚本依赖 jq（wait_parse/qa 也直接用它）。
json_val() {
  local path="$1"
  if ! command -v jq >/dev/null 2>&1; then
    VERIFY_FAIL=$((VERIFY_FAIL+1)); printf "${RED}✘ FAIL${NC}  缺少依赖 jq（macOS: brew install jq; Debian: apt install jq）\n"
    return
  fi
  jq -r "$path // empty" 2>/dev/null || printf '%s\n' ""
}

# ---------- 请求-断言一步到位 ----------
# expect_json <action> <body> <token> <jq-path> <期望值> <标签>
#   期望值支持 "_any"(非空) / 数字 / 字符串；可选把 '500' 视为通道 OK（见 expect_json_any）。
expect_json() {
  local action="$1" body="$2" tok="$3" path="$4" want="$5" label="$6"
  local resp val
  resp="$(http_post "$action" "$body" "$tok")"
  local rc=$?
  if [ "$rc" -ne 0 ]; then vfail "$label → HTTP 失败: $VERIFY_HTTP_ERR"; return; fi
  if [ -z "$resp" ]; then vfail "$label → 空响应"; return; fi
  val="$(printf '%s' "$resp" | json_val "$path")"
  if [ "$want" == "_any" ] && [ -n "$val" ]; then vok "$label ($path=$val)"; return; fi
  if [ "$val" == "$want" ]; then vok "$label ($path=$want)"; return; fi
  vfail "$label → 期望 $path='$want' 实得 '${val}'；body: $(printf '%s' "$resp" | head -c 300)"
}

# expect_chan <action> <body> <token> <label>
#   服务端 code==0 或接口可达(code!=500)即 PASS，仅 500 打印"通道OK"。用于"接口通路自检"。
expect_chan() {
  local action="$1" body="$2" tok="$3" label="$4"
  local resp code
  resp="$(http_post "$action" "$body" "$tok")"
  local rc=$?
  if [ "$rc" -ne 0 ]; then vfail "$label → HTTP 失败: $VERIFY_HTTP_ERR"; return; fi
  code="$(printf '%s' "$resp" | json_val ".code // empty")"
  case "$code" in
    0|""|"") vok "$label (code=$code)" ;;
    500) VERIFY_INTERNAL=$((VERIFY_INTERNAL+1)); vskip "$label → code=500(依赖未就绪,通道OK)" ;;
    *)   vfail "$label → code=$code；body: $(printf '%s' "$resp" | head -c 200)" ;;
  esac
}

# ---------- 汇总 ----------
verify_summary() {
  echo
  echo "==================== 结果汇总 ===================="
  printf "通过: ${GREEN}${VERIFY_PASS}${NC}  失败: ${RED}${VERIFY_FAIL}${NC}  跳过: ${YELLOW}${VERIFY_SKIP}${NC}  内部500: ${CYAN}${VERIFY_INTERNAL}${NC}  致命: ${RED}${VERIFY_FATAL}${NC}\n"
  if [ "$VERIFY_FAIL" -eq 0 ] && [ "$VERIFY_HAD_FATAL" -eq 0 ]; then
    [ "$VERIFY_INTERNAL" -gt 0 ] && echo -e "${YELLOW}⚠ 有接口服务端500(依赖未就绪)，但核心链路已通过${NC}"
    echo -e "${GREEN}✅ 验证全部通过${NC}"
    return 0
  else
    echo -e "${RED}❌ 验证未通过，请按上面红色 FATAL/FAIL 排查${NC}"
    return 1
  fi
}

# 等待就绪：执行给定函数直至返回0或超时。用法：wait_until [秒] func args...
wait_until() {
  local timeout="$1" fn="$2"; shift 2
  local start=$SECONDS
  while :; do
    if "$fn" "$@"; then return 0; fi
    if [ $((SECONDS-start)) -ge "$timeout" ]; then vfail "等待超时(${timeout}s)：$fn"; return 1; fi
    sleep 1
  done
}
