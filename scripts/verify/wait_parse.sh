#!/usr/bin/env bash
# =============================================================================
# wait_parse.sh — 轮询文件解析状态直至成功（拆 chunk + 向量 + 摘要 完成）
#
# 依赖 pipeline worker：corekg 上传后写 ke.prase_pdf_task -> ke.knowledge_task
# 到 core_task 表，pipeline 轮询 knowledge.GetPendingTask 处理并 TaskCallBack，
# corekg 回调推进 parse->knowledge->SuccessFile。knowledge_status 置 success 才可问答。
#
# source 后提供：
#   wait_file_parse <forest_id> <file_id> <token> [timeout_s] [interval_s]
#       轮询 forest.ListFile 的 data[] 中该 file_id 的 knowledge_status。
#       成功 return 0；超时 return 1 并给出排障提示。
# =============================================================================
set -uo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

# _file_knowledge_status <forest_id> <file_id> <token>  -> echo 状态；找不到文件返回空
_file_knowledge_status() {
  local forest_id="$1" file_id="$2" tok="$3"
  local resp
  resp="$(http_post forest.ListFile \
    "{\"request\":{\"forest_id\":$forest_id,\"offset\":0,\"limit\":50}}" "$tok")"
  # data 数组里定位文件，取 knowledge_status（ListFile 项用 file_id 字段标识文件 id，is_dir 为布尔）
  printf '%s' "$resp" | jq -r --argjson fid "$file_id" \
    '.Response.data[]? | select(.file_id==$fid or .id==$fid or .forest_file_id==$fid) | select(.is_dir == false or .is_dir == 0) | .knowledge_status' 2>/dev/null \
    | head -1
}

wait_file_parse() {
  local forest_id="$1" file_id="$2" tok="$3"
  local timeout_s="${4:-180}" interval_s="${5:-3}"
  vlog "等待文件解析完成 file_id=$file_id (timeout=${timeout_s}s)…"
  local start=$SECONDS last_status dataline

  while :; do
    last_status="$( (_file_knowledge_status "$forest_id" "$file_id" "$tok") 2>/dev/null )"
    case "$last_status" in
      success)
        vok "文件解析完成 (knowledge_status=success) 耗时 $((SECONDS-start))s"
        return 0
        ;;
      fail|failed|error)
        vfail "文件解析失败 (knowledge_status=$last_status)"
        return 1
        ;;
      ""|pending|running|waiting|none)
        # 尚未就绪，继续等待
        ;;
      *)
        # 未知状态，暂时视为未完成
        ;;
    esac
    if [ $((SECONDS-start)) -ge "$timeout_s" ]; then
      vfail "等待解析超时(${timeout_s}s)，最后状态=$last_status"
      echo "  排障提示："
      echo "    1) pipeline chunker 是否运行且连接到 corekg："
      echo "       本地: cd apps/pipeline && python chunk_worker_main.py（指向 localhost:8080）"
      echo "       compose: docker compose -f docker-compose.pipeline.yml up -d pipeline"
      echo "    2) core_task 表是否推进：SELECT task_type,task_status FROM core_task WHERE subject_id=$file_id;"
      echo "    3) ES 索引(ke_0)与 embedding 配置是否可达（apps/pipeline/config/chunk_config.yaml）。"
      return 1
    fi
    sleep "$interval_s"
  done
}
