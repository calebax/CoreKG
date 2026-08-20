PUT /ke_chat_history
{
  "mappings": {
    "properties": {
      "req_id": {
        "type": "keyword"
      },
      "agent_step": {
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
      "is_charged": {
        "type": "boolean",
        "null_value": false
      },
      "base_agent_id": {
        "type": "integer"
      },
      "agent_version": {
        "type": "integer"
      },
      "model_id": {
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
      "image_content": {
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
      },
      "sub_question": {
        "type": "keyword"
      },
      "graph_reference": {
        "properties": {
          "edges": {
            "type": "object",
            "enabled": false
          },
          "nodes": {
            "type": "object",
            "enabled": false
          }
        }
      },
      "graph_chat_reference": {
        "properties": {
          "edges": {
            "type": "object",
            "enabled": false
          },
          "nodes": {
            "type": "object",
            "enabled": false
          }
        }
      },
      "react_agent_service": {
        "type": "object",
        "enabled": false
      },
      "extra": {
        "type": "flattened"
      }
    },
    "dynamic": "strict"
  },
  "settings": {
    "number_of_shards": 2,
    "number_of_replicas": 1
  }
}
