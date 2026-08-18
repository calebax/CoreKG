PUT ke_chat_history/_mapping
{
  "properties": {
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
      }
  }
}