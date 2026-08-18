/**
 * 图谱实体名称校验
 * - 约束：禁止输入以下非法字符（包括空格/换行/制表符/回车）
 *   | > < " ' . , ; ` (空格) - + = ( ) [ ] { } * / \ ? ! @ # $ % ^ ~ : \t \n \r
 * - 说明：`&gt;` 等同于 `>`
 */

/** 实体名称非法字符正则（命中即非法） */
export const GRAPH_NODE_NAME_ILLEGAL_CHAR_REGEXP =
  /[|<>"'.,;`\s\-\+=\(\)\[\]\{\}\*\/\\\?!@#\$%\^~:]/u

/** Tooltip 提示文案（用于 UI 说明） */
export const GRAPH_NODE_NAME_INVALID_TOOLTIP =
  '实体名称不可包含以下字符：| > < " \' . , ; ` 空格 - + = ( ) [ ] { } * / \\\\ ? ! @ # $ % ^ ~ :，以及换行(\\n)、制表符(\\t)、回车(\\r)（&gt; 等同于 >）'

/** 表单校验错误文案（用于 Form validator） */
export const GRAPH_NODE_NAME_INVALID_ERROR =
  '实体名称包含非法字符：| > < " \' . , ; ` 空格 - + = ( ) [ ] { } * / \\\\ ? ! @ # $ % ^ ~ : 或换行/制表符/回车'

export const validateGraphNodeName = (_: unknown, v: string) => {
  if (!v) return Promise.resolve()
  if (GRAPH_NODE_NAME_ILLEGAL_CHAR_REGEXP.test(v)) {
    return Promise.reject(new Error(GRAPH_NODE_NAME_INVALID_ERROR))
  }
  return Promise.resolve()
}


