PUT /ke_0/_mapping
{
  "properties": {
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
    }
  }
}
