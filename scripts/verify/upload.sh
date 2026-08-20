#!/usr/bin/env bash
# =============================================================================
# upload.sh — 上传验证文件到知识库（multipart forest.UploadFile）
# source 后提供：
#   upload_file <forest_id> <file_path> <token>  -> 输出 forest_file_id；失败输出空返回1
# =============================================================================
set -uo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

# upload_file <forest_id> <file_path> <token>
# 上传 txt/md 样例。成功输出 file_id。
upload_file() {
  local forest_id="$1" file_path="$2" tok="$3"
  if [ ! -f "$file_path" ]; then vfail "上传文件不存在: $file_path"; echo; return 1; fi
  local resp code fid
  resp="$(curl -sS --max-time "$VERIFY_HTTP_TIMEOUT" \
    -X POST "$BASE$PREFIX/forest.UploadFile" \
    -H "Authorization: Bearer $tok" \
    -F "forest_id=$forest_id" \
    -F "parent_id=0" \
    -F "file=@$file_path")"
  local rc=$?
  if [ "$rc" -ne 0 ]; then vfail "上传HTTP失败: $resp"; echo; return 1; fi
  code="$(printf '%s' "$resp" | json_val ".code // empty")"
  if [ "$code" != "0" ]; then vfail "上传失败 code=$code → $(printf '%s' "$resp" | head -c 300)"; echo; return 1; fi
  fid="$(printf '%s' "$resp" | json_val ".Response.forest_file_id // empty")"
  printf '%s' "$fid"
  return 0
}
