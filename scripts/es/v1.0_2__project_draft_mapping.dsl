PUT /project_draft
{
	"settings": {
		"number_of_shards": 1,
		"number_of_replicas": 0,
		"analysis": {
			"analyzer": {
				"text_analyzer": {
					"type": "custom",
					"tokenizer": "standard",
					"filter": ["lowercase", "stop", "snowball"]
				},
				"smartcn": {
					"type": "smartcn"
				}
			}
		}
	},
	"mappings": {
		"dynamic": false,
		"properties": {
			"create_time": {
				"type": "long"
			},
			"has_published": {
				"type": "keyword"
			},
			"id": {
				"type": "keyword"
			},
			"name": {
				"type": "text",
				"analyzer": "smartcn",
				"search_analyzer": "smartcn",
				"fields": {
					"raw": {
						"type": "keyword"
					}
				}
			},
			"owner_id": {
				"type": "keyword"
			},
			"publish_time": {
				"type": "long"
			},
			"space_id": {
				"type": "keyword"
			},
			"status": {
				"type": "keyword"
			},
			"type": {
				"type": "keyword"
			},
			"update_time": {
				"type": "long"
			},
			"fav_time": {
				"type": "long"
			},
			"recently_open_time": {
				"type": "long"
			},
			"is_fav": {
				"type": "keyword"
			},
			"is_recently_open": {
				"type": "keyword"
			}
		}
	}
}
