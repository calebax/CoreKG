PUT /ke_chat_history
{
  "mappings": {
    "properties": {
      "req_id": {   // 一次请求中的reqid为同一个，内部调用时使用内部的当前的req_id
        "type": "keyword"
      },
      "agent_step": {   // 这是当前请求的第几步，相同的reqid，step为当前question+1
        "type": "integer"
      },
      "company_id": {
        "type": "integer"
      },
      "uin": {
        "type": "integer"
      },
      "api_key_id": {
        "type": "integer"
      },
      "total_tokens": {
        "type": "integer"
      },
      "out_token": {
        "type": "integer"
      },
      "cache_hit_token": {
        "type": "integer"
      },
      "cache_miss_token": {
        "type": "integer"
      },
      "is_charged": {   // 是否完成计费，或收费
        "type": "boolean",
        "null_value": false
      },
      "base_agent_id": {  // 外部通过api调用存储参数
        "type": "integer"
      },
      "agent_version": {  // 外部通过api调用存储参数
        "type": "integer"
      },
      "model_id": {   // 外部通过api调用存储参数 
        "type": "integer"
      },
      "cost_seconds": {
        "type": "integer"
      },
      "reasoning_seconds": {
        "type": "integer"
      },
      "image_url_list": {
        "type": "keyword"
      },
      "status": {
        "type": "keyword"
      },
      "external_id": {
        "type": "keyword"
      },
      "session_id": {
        "type": "integer"
      },
      "created_at": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
      },
      "question": {
        "type": "text",
        "analyzer": "ik_smart",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "answer": {
        "type": "text",
        "analyzer": "ik_smart",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "reasoning": {
        "type": "text",
        "analyzer": "ik_smart",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "image_content": {   // 问答上传图片多模态分析后问答
        "type": "text",
        "analyzer": "ik_smart",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "query_reference_list": {
        "type": "object",
        "enabled": false
      },
      "chat_reference_list": {
        "type": "object",
        "enabled": false
      },
       "user_input": {
        "type": "object",
        "enabled": false
      },
      "agent_name": {
        "type": "keyword"
      }
    },
    "dynamic": "strict"
  },
  "settings": {
    "number_of_shards": 2,
    "number_of_replicas": 1
  }
}