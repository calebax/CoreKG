POST /ke_chat_history/_update_by_query?conflicts=proceed
{
  "script": {
    "source": "if (ctx._source.containsKey('chat_reference_list')) { def v = ctx._source.chat_reference_list; if (v != null && v instanceof Collection) { ctx._source.chat_reference_list = null; } }"
  }
}
