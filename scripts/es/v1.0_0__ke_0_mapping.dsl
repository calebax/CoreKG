PUT /ke_0
{
  "mappings": {
    "properties": {
      "version": {
        "type": "keyword"
      },
      "forest_id": {
        "type": "integer"
      },
      "company_id": {
        "type": "integer"
      },
      "uin": {
        "type": "integer"
      },
      "type": {
        "type": "keyword"
      },
      "description_hash": {
        "type": "keyword"
      },
      "content": {
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
      "content_source": {
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
      "content_target": {
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
      "embedding": {
        "type": "dense_vector",
        "dims": 1024,
        "index": true
      },
      "tokens": {
        "type": "integer"
      },
      "mind_map": {
        "type": "text",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "abstract": {
        "type": "text",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "questions": {
        "type": "keyword"
      },
      "file_id": {
        "type": "integer"
      },
      "location": {
        "type": "integer",
        "index": false
      },
      "sequence": {
        "type": "integer"
      },
      "description": {
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
      "image_url": {
        "type": "keyword"
      },
      "formula": {
        "type": "keyword"
      },
      "qa_question": {
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
      "qa_answer": {
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
      "qa_lable": {
        "type": "keyword"
      },
      "qa_answer_id": {
        "type": "keyword"
      },
      "qa_main_id": {
        "type": "keyword"
      },
      "title_level": {
        "type": "keyword"
      },
      "title_level_ids": {
        "type": "keyword"
      },
      "level": {
        "type": "integer"
      },
      "created_at": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
      },
      "updated_at": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
      },
      "references": {
        "type": "nested",
        "properties": {
          "file_id": {
            "type": "integer"
          },
          "description": {
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
          "chunk_id": {
            "type": "keyword"
          },
          "relationship_id":{
            "type": "keyword"
          }
        }
      },
      "chunk_size": {
        "type": "integer"
      },
      "yg_location": {
        "type": "keyword"
      },
      "image_embedding": {
        "type": "dense_vector",
        "dims": 1024,
        "index": true,
        "similarity": "cosine",
        "index_options": {
          "type": "int8_hnsw",
          "m": 16,
          "ef_construction": 100
        }
      },
      "file_name": {
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
      "is_disable": {
        "type": "boolean",
        "null_value": false
      },
      "source_from": {
        "type": "keyword"
      },
      "enable": {
        "type": "integer"
      },
      "graph_info": {
        "type": "object",
        "enabled": false
      },
      "graph_status": {
        "type": "keyword"
      },
      "table": {
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
