import remarkParse from 'remark-parse'
import remarkStringify from 'remark-stringify'
import stripMarkdown from 'strip-markdown'
import { unified } from 'unified'

export const markdownToText = (markdown: string): string => {
  const result = unified()
    .use(remarkParse)
    .use(stripMarkdown)
    .use(remarkStringify)
    .processSync(markdown)

  return result.toString()
}
