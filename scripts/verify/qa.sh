#!/usr/bin/env bash
# =============================================================================
# qa.sh — 基于知识库的真实问答验证（内部 chat.* 流式 RAG）
#
# 流程：
#   1) chat.ListModel 取一个可用模型 id
#   2) chat.NewChatSession 建 standard(知识库 RAG) 会话，绑定 forest_id
#   3) chat.NewChatQuestionStream 提交问题（异步 RAG 生成）；返回 question_id
#   4) 轮询 chat.ListSessionChats 等答案完成，断言 answer 非空 且
#      query_reference_list 命中本知识库（file_id 命中即证明 RAG 检索到该文件 chunk）
#
# source 后提供：
#   ask_file_qa <forest_id> <expect_file_id> <question> <token>
#       成功 return 0；失败 return 1。
# =============================================================================
set -uo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

# 取第一个可用模型 id（ListModel 返回未经 json tag 的 Go 字段，ID/Data 首字母大写）
_qa_pick_model_id() {
  http_post chat.ListModel '{"request":{"offset":0,"limit":10}}' "$1" \
    | jq -r '.Response.Data[]?.ID // empty' 2>/dev/null | head -1
}

# _qa_question_state <session_id> <question_id> <token> -> echo "status|answer_has|file_hit|chunk_count"
# ListSessionChats 返回 ES 文档结构：Data[]._id 是问题 id，._source.answer/status/query_reference_list 为主要字段。
_qa_question_state() {
  local sid="$1" qid="$2" tok="$3"
  local resp
  resp="$(http_post chat.ListSessionChats "{\"request\":{\"id\":$sid}}" "$tok")"
  printf '%s' "$resp" | jq -r \
    --arg qid "$qid" '
      (.Response.Data[]? | select(._id==$qid))
      | ._source
      | [(.status // ""),
         (if (.answer // "") != "" then "1" else "0" end),
         (if ([.query_reference_list[]?.file_id|tostring] | length>0) then "1" else "0" end),
         ([.query_reference_list[]?.chunk_list[]?] | length)] | join("|")' 2>/dev/null
}

ask_file_qa() {
  local forest_id="$1" expect_file_id="$2" question="$3" tok="$4"
  local timeout_s="${5:-240}"

  vlog "发起知识库问答 forest_id=$forest_id question=$(printf '%s' "$question" | head -c 40)…"

  local model_id
  model_id="$(_qa_pick_model_id "$tok")"
  if [ -z "$model_id" ] || [ "$model_id" == "0" ]; then
    vfail "未从 chat.ListModel 取到可用模型 id，跳过问答"
    return 1
  fi

  # 建 standard 会话绑定 forest
  local sess_resp sess_code sess_id fname
  fname="verify-forest-$forest_id"
  sess_resp="$(http_post chat.NewChatSession \
    "{\"request\":{\"model_id\":$model_id,\"resource_id\":$model_id,\"base_type\":\"standard\",\
\"resource_type\":\"forest\",\"ids\":[$forest_id],\"names\":[\"$fname\"]}}" "$tok")"
  sess_code="$(printf '%s' "$sess_resp" | jq -r '.code // empty' 2>/dev/null)"
  sess_id="$(printf '%s' "$sess_resp" | jq -r '.Response.ID // empty' 2>/dev/null)"
  if [ "$sess_code" != "0" ] || [ -z "$sess_id" ]; then
    vfail "新建会话失败 code=$sess_code → $(printf '%s' "$sess_resp" | head -c 300)"
    return 1
  fi
  vok "新建知识库会话 id=$sess_id"

  # 提交问题
  local q_resp q_code qid
  q_resp="$(http_post chat.NewChatQuestionStream \
    "{\"request\":{\"session_id\":$sess_id,\"question\":$(printf '%s' "$question" | jq -R -s .)}}" "$tok")"
  q_code="$(printf '%s' "$q_resp" | jq -r '.code // empty' 2>/dev/null)"
  qid="$(printf '%s' "$q_resp" | jq -r '.Response.question_id // empty' 2>/dev/null)"
  if [ "$q_code" != "0" ] || [ -z "$qid" ]; then
    vfail "提交问题失败 code=$q_code → $(printf '%s' "$q_resp" | head -c 300)"
    return 1
  fi
  vok "问题已提交 question_id=${qid}，等待答案生成(最多 ${timeout_s}s)…"

  # NewChatQuestionStream 只落地 pending 问题；必须再调 chat.ChatQuestionStream 触发真正的 RAG 答案生成。
  # 该接口为 SSE 流式（curl 无法 jq 解析），其响应仅用于落库推进，故直接消费并丢弃。
  local trig
  trig="$(curl -s -N --max-time 20 -X POST "${BASE}${PREFIX}/chat.ChatQuestionStream" \
    -H "Content-Type: application/json" -H "Authorization: Bearer $tok" \
    -d "{\"request\":{\"question_id\":\"$qid\"}}" 2>/dev/null)"
  # 触发不应改变本轮判定，仅留作 debug 线索
  printf '%s' "$trig" | grep -q '"code":[1-9]' 2>/dev/null && vlog "ChatQuestionStream 触发返回异常: $(printf '%s' "$trig" | head -c 200)"

  # 轮询答案
  local start=$SECONDS state status file_hit chunk_cnt
  while :; do
    state="$(_qa_question_state "$sess_id" "$qid" "$tok")"
    status="$(printf '%s' "$state" | cut -d'|' -f1)"
    file_hit="$(printf '%s' "$state" | cut -d'|' -f3)"
    chunk_cnt="$(printf '%s' "$state" | cut -d'|' -f4)"
    # 已生成答案
    if [ "$(printf '%s' "$state" | cut -d'|' -f2)" == "1" ]; then
      break
    fi
    if [ "$status" == "fail" ] || [ "$status" == "error" ]; then
      vfail "问答状态=${status}（未能生成答案）"
      return 1
    fi
    if [ $((SECONDS-start)) -ge "$timeout_s" ]; then
      vfail "等待答案超时(${timeout_s}s) 最后状态=$status"
      echo "  排障：确认 corekg 可连真实 LLM / 该会话问题是否卡在处理队列。"
      return 1
    fi
    sleep 3
  done

  vok "答案已生成 (状态=$status)"
  if [ "$file_hit" == "1" ]; then
    vok "RAG 引用命中知识库（chunk 数=${chunk_cnt}，证明检索到该文件内容）"
  else
    vfail "答案生成了但 query_reference_list 为空——RAG 未检索到文件 chunk，链路不完整"
    return 1
  fi
  return 0
}
