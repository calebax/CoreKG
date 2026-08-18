PUT /ke_0/_mapping
{
  "properties": {
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
    }
  }
}