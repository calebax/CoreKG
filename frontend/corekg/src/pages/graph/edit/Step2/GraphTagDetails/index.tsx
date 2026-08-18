import { FC, useMemo, useState } from 'react'
import { App, Button, Tabs, Tag } from 'antd'
import { useBoolean, useCounter } from 'ahooks'
import { produce } from 'immer'
import { AddCircleIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import { createEdge, updateTag } from '@/api/graph'
import { useGraphInfo } from '@/pages/graph/GraphProvider'
import { useActiveTag } from '..'
import { Properties, PropertyModal } from './Properties'
import { RelationshipModal, Relationships } from './Relationships'

type ContentType = 'prop' | 'rel'
/// 属性保存在本地 关系总是从后端获取
export const GraphTagDetails: FC<Style> = (props) => {
  const { activeTag } = useActiveTag()
  const { className, style } = props
  const [tab, setTab] = useState<ContentType>('prop')
  const [refreshKey, { inc: setKey }] = useCounter()
  const content = useMemo(() => {
    switch (tab) {
      case 'prop':
        return <Properties />
      case 'rel':
        return <Relationships key={refreshKey} />
    }
  }, [refreshKey, tab])

  return (
    <div
      className={cn(
        'rounded-xl border border-[#D7D9E5] bg-[#FCFCFE]',
        'overflow-hidden p-4 flex flex-col',
        {
          invisible: !activeTag,
        },
        className,
      )}
      style={style}
    >
      <span className='flex gap-2 items-center'>
        <span className='text-base font-medium'>{activeTag?.tag_name}</span>
        <Tag className='rounded-full text-[#885BD2] bg-[#F3EDF9]'>实体类型</Tag>
      </span>
      <span className='text-description'>{activeTag?.description}</span>
      <div className='flex justify-between'>
        <Tabs
          activeKey={tab}
          onChange={(v) => setTab(v as any)}
          items={[
            { key: 'prop', label: '属性定义' },
            { key: 'rel', label: '关系定义' },
          ]}
        />
        <CreateBtn type={tab} reloadRelationship={setKey} />
      </div>
      <div className='flex-1 flex flex-col overflow-hidden'>{content}</div>
    </div>
  )
}

const CreateBtn: FC<{ type: ContentType; reloadRelationship: () => void }> = (
  props,
) => {
  const { message } = App.useApp()
  const { type, reloadRelationship } = props
  const { data, mutateTags } = useGraphInfo()
  const { activeTag } = useActiveTag()
  const { tag_name } = activeTag ?? {}
  const [open, { toggle }] = useBoolean()
  const text = useMemo(() => {
    switch (type) {
      case 'prop':
        return '新建属性'
      case 'rel':
        return '新建关系'
    }
  }, [type])
  const modal = (() => {
    if (!open) return null
    switch (type) {
      case 'prop':
        return (
          <PropertyModal
            title='新建属性'
            onCancel={toggle}
            onSubmit={async (val) => {
              await updateTag({
                graph_id: data!.id,
                tag_id: activeTag!.tag_id,
                tag_name: tag_name!,
                properties: (activeTag?.properties ?? []).concat(val),
              })
              mutateTags(
                produce(data!.tags, (draft) => {
                  const currentTag = draft.find(
                    (item) => item.tag_name === tag_name,
                  )!
                  if (!currentTag.properties) {
                    currentTag.properties = []
                  }
                  currentTag.properties.push(val)
                }),
              )
              message.success('添加成功')
            }}
          />
        )
      case 'rel':
        return (
          <RelationshipModal
            title='新建关系'
            onCancel={toggle}
            onSubmit={async (val) => {
              await createEdge({
                graph_id: data!.id,
                ...val,
                egde_name: val.edge_name,
              })
              reloadRelationship()
              message.success('添加成功')
            }}
          ></RelationshipModal>
        )
    }
  })()
  return (
    <>
      <Button
        onClick={toggle}
        icon={<AddCircleIcon />}
        className='text-[#0C99FF] border-[#0C99FF]'
      >
        {text}
      </Button>
      {modal}
    </>
  )
}
