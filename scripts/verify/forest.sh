#!/usr/bin/env bash
# =============================================================================
# forest.sh — 新建 / 列出 / 删除知识库
# source 后提供：
#   create_forest <name> [user_token]  -> 输出 forest_id；失败输出空并返回1
#   list_forests <token>               -> 打印 JSON
#   delete_forest <forest_id> <token>
# =============================================================================
set -uo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

# create_forest <name> <token>
# 使用内部鉴权 forest.CreateForest；返回 forest_id。
create_forest() {
  local name="$1" tok="$2"
  local resp code forest
  resp="$(http_post forest.CreateForest \
    "{\"request\":{\"name\":\"$name\",\"description\":\"auto-verify\",\"public_scope\":\"company\",\"forest_type\":\"file\"}}" "$tok")"
  local rc=$?
  if [ "$rc" -ne 0 ]; then vfail "创建知识库HTTP失败: $VERIFY_HTTP_ERR"; echo; return 1; fi
  code="$(printf '%s' "$resp" | json_val ".code // empty")"
  if [ "$code" != "0" ]; then vfail "创建知识库失败 code=$code → $(printf '%s' "$resp" | head -c 300)"; echo; return 1; fi
  forest="$(printf '%s' "$resp" | json_val ".Response.forest_id // empty" || printf '%s' "$resp" | json_val ".Response.id // empty")"
  printf '%s' "$forest"
  return 0
}

# list_forests <token>
list_forests() {
  local tok="$1"
  http_post forest.ListForest '{"request":{"offset":0,"limit":20}}' "$tok"
}

# delete_forest <forest_id> <token>
delete_forest() {
  local fid="$1" tok="$2"
  local resp code
  resp="$(http_post forest.DeleteForest "{\"request\":{\"id\":$fid}}" "$tok")"
  code="$(printf '%s' "$resp" | json_val ".code // empty")"
  if [ "$code" == "0" ]; then vok "已清理测试知识库 id=$fid"; else vskip "测试知识库 id=$fid 未删除(code=$code)；如保留请记录"; fi
}
