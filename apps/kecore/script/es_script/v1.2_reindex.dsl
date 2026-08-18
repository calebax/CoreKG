PUT /ke_0
{
  "mappings": {
    "properties": {
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
      "content": {
        "type": "text",
        "analyzer": "ik_max_word",
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
        "analyzer": "ik_max_word",
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
        "analyzer": "ik_max_word",
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
        "analyzer": "ik_max_word",
        "search_analyzer": "ik_smart"
      },
      "image_url": {
        "type": "keyword"
      },
      "formula": {
        "type": "keyword"
      },
      "references": {
        "type": "nested",
        "properties": {
          "file_id": {
            "type": "integer"
          },
          "description": {
            "type": "text",
            "analyzer": "ik_max_word",
            "search_analyzer": "ik_smart"
          },
          "chunk_id": {
            "type": "keyword"
          }
        }
      }
    },
    "dynamic": "strict"
  },
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1
  }
}


PUT /ke_0/_mapping
{
  "properties": {
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
      "analyzer": "ik_max_word",
      "search_analyzer": "ik_smart"
    },
    "image_url": {
      "type": "keyword"
    },
    "formula": {
      "type": "keyword"
    }
  }
}

// 复制内部语句
POST ke_0/_update_by_query
{
  "query": {
    "bool": {
      "must": [
        { "terms": { "type": ["chunk", "table","image","table"] }}
      ]
    }
  },
  "script": {
    "lang": "painless",
    "source": """
      if (ctx._source.references != null && ctx._source.references.size() > 0) {
        def ref = ctx._source.references[0];
        if (ref.containsKey("description")) {
          ctx._source.description = ref.description;
        }
        if (ref.containsKey("image_url")) {
          ctx._source.image_url = ref.image_url;
        }
        if (ref.containsKey("formula")) {
          ctx._source.formula = ref.formula;
        }
        if (ref.containsKey("file_id")) {
          ctx._source.file_id = ref.file_id;
        }
        if (ref.containsKey("sequence")) {
          ctx._source.sequence = ref.sequence;
        }
        if (ref.containsKey("location")) {
          ctx._source.location = ref.location;
        }
      }
    """
  }
}
