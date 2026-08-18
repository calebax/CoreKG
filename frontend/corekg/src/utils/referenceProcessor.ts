/**
 * Reference格式处理工具
 * 将{Reference §xxx}格式转换为数字标注
 * 支持格式：
 * - {Reference §123[16]} - 旧格式，单个chunk index
 * - {Reference §123[chunkid:xxx]} - 旧格式，chunk id
 * - {Reference §file_123[16, 18], §file_456[24]} - 新格式，多个chunk和文件
 */

let globalAnnotationIndex = 1

export const processReferenceMarkdown = (content: string): string => {
  if (!content) return content

  globalAnnotationIndex = 0
  const chunkKeyMap = new Map<string, number>()

  // 使用正则表达式匹配 {Reference §xxx} 格式
  return content.replace(
    /\{Reference\s+([^}]+)\}/g,
    (match, referenceContent) => {
      const annotations: { index: number; content: string }[] = []

      // 解析引用内容中的章节信息，支持 file_ 前缀
      // 匹配格式：§file_123[...] 或 §123[...]
      const sectionPattern = /§(?:file_)?(\d+)\[([^\]]+)\]/g
      let sectionMatch

      while ((sectionMatch = sectionPattern.exec(referenceContent)) !== null) {
        const [, docId, chunkContent] = sectionMatch

        // 解析chunk列表，支持逗号分隔的多个chunk
        // 例如：[16, 18] 或 [chunkid:xxx] 或 [16, chunkid:abc]
        const chunks = chunkContent
          .split(',')
          .map((chunk) => chunk.trim())
          .filter((chunk) => chunk.length > 0)

        // 为每个去重后的 chunk 生成独立标注，避免同一文件下不同段落都显示成同一个数字
        chunks.forEach((chunk) => {
          let chunkId = ''
          let chunkIndex = ''

          if (chunk.startsWith('chunkid:')) {
            // chunkid格式：chunkid:xxx
            chunkId = chunk.slice(8)
          } else {
            // 数字格式：chunk index
            chunkIndex = chunk
          }

          // 以 file_id + sequence 为主键去重，兼容旧的 chunkid 格式
          const uniqueChunkKey = chunkId
            ? `${docId}-chunkid-${chunkId}`
            : `${docId}-sequence-${chunkIndex}`

          if (!chunkKeyMap.has(uniqueChunkKey)) {
            globalAnnotationIndex++
            chunkKeyMap.set(uniqueChunkKey, globalAnnotationIndex)
          }

          const currentAnnotationIndex = chunkKeyMap.get(uniqueChunkKey)!

          // 构建data属性
          const dataAttributes: string[] = [
            `data-index="${currentAnnotationIndex}"`,
            `data-doc-id="${docId}"`,
          ]

          dataAttributes.push(`data-chunk-id="${chunkId}"`)

          dataAttributes.push(`data-chunk-index="${chunkIndex}"`)

          const annotationHtml = `<span class="reference-annotation hover:bg-[#CC5DE8]/70 inline-block mx-[2px] !w-[16px] !h-[16px] text-[#ffffff] text-[12px] font-normal bg-[#bbb] relative rounded-[50%] cursor-pointer leading-[16px] text-center align-middle" ${dataAttributes.join(' ')}>${currentAnnotationIndex}</span>`
          annotations.push({
            index: currentAnnotationIndex,
            content: annotationHtml,
          })
        })
      }

      return annotations
        .sort((a, b) => a.index - b.index)
        .map((item) => item.content)
        .join('')
    },
  )
}
