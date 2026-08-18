POST /ke_0/_update_by_query
{
  "query": {
    "terms": {
      "type": [ "FQA" ]
    }
  },
  "script": {
    "source": "ctx._source.enable = params.new_value",
    "lang": "painless",
    "params": {
      "new_value": 1
    }
  }
}