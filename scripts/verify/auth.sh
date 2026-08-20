#!/usr/bin/env bash
# =============================================================================
# auth.sh — account.LoginByPassword 获取 JWT
# source 后提供：
#   VERIFY_TOKEN / VERIFY_LOGIN_USER
#   login_admin()  -> 登录成功返回0并设置 VERIFY_TOKEN；失败打印并返回1
# =============================================================================
set -uo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

LOGIN_USER="${LOGIN_USER:-admin@admin.com}"
LOGIN_PASS="${LOGIN_PASS:-admin123456}"
DOMAIN_NAME="${DOMAIN_NAME:-localhost:30000}"

login_admin() {
  local resp token code
  resp="$(http_post account.LoginByPassword \
    "{\"request\":{\"domain_name\":\"$DOMAIN_NAME\",\"username\":\"$LOGIN_USER\",\"password\":\"$LOGIN_PASS\"}}" )"
  local rc=$?
  if [ "$rc" -ne 0 ]; then vfail "登录HTTP失败: $VERIFY_HTTP_ERR"; return 1; fi
  code="$(printf '%s' "$resp" | json_val ".code // empty")"
  token="$(printf '%s' "$resp" | json_val ".Response.jwt_token // empty")"
  if [ "$code" != "0" ] || [ -z "$token" ]; then
    vfail "登录失败 code=$code → $(printf '%s' "$resp" | head -c 400)"
    echo "  提示: 检查 DOMAIN_NAME($DOMAIN_NAME) 是否与 admin_login_setting.domain_name 一致；"
    echo "        core_daily_log 是否有 recent valid 的 license 记录。"
    return 1
  fi
  VERIFY_TOKEN="$token"
  vok "登录成功 user=$LOGIN_USER token=${token:0:12}..."
  return 0
}
