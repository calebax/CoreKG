import { FC } from 'react'

const icons = new Map(
  Object.entries(
    import.meta.glob('./images/*.svg', {
      eager: true,
      import: 'default',
    }),
  ).map(([k, v]) => {
    const type = /^.*\/(.+)\.svg$/.exec(k)![1]
    return [type, v]
  }),
) as Map<string, string>

/** 不同类型知识的Icon */
export const KnowledgeIcon: FC<Style & { type: string }> = (props) => {
  const { className, style, type } = props
  const src: any = icons.get(type)
  return src ? <img src={src} className={className} style={style}></img> : null
}
