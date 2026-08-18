import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { App, Button, Empty, Skeleton } from 'antd'
import type { StepComponent } from '..'
import { useGraphInfo, type GraphTagWithId } from '../../GraphProvider'
import { GraphTagDetails } from './GraphTagDetails'
import { GraphTags } from './GraphTags'
import { ParseMode } from './ParseMode'

export const Step2: StepComponent<{
  editRules?: boolean
}> = (props) => {
  const { modal } = App.useApp()
  const { decrease, increase, editRules } = props
  const { data, loading, mutateTags } = useGraphInfo()
  const { tags, parse_mode } = data ?? {}

  const [activeTagId, setActiveTagId] = useState<number>()
  const activeTag = useMemo(() => {
    return tags?.find((item) => item.tag_id === activeTagId)
  }, [activeTagId, tags])

  useEffect(() => {
    // 尽可能选中一个tag
    if (activeTag) return
    if (!tags || tags.length === 0) return
    setActiveTagId(tags[0].tag_id)
  }, [activeTag, tags])
  if (loading) return <Skeleton active />
  if (!tags) return <Empty />

  return (
    <ActiveTagContext.Provider value={{ activeTag, setActiveTagId }}>
      <div className='h-full overflow-hidden flex flex-col gap-2 relative'>
        <ParseMode />
        <div className='flex-1 overflow-hidden flex gap-4'>
          <GraphTags className='flex-3' />
          <GraphTagDetails className='flex-4' />
        </div>
        <span className='flex items-center self-end gap-4'>
          <Button
            type={editRules ? 'primary' : 'default'}
            onClick={() => {
              mutateTags([])
              decrease()
            }}
          >
            {editRules ? '保存' : '上一步'}
          </Button>
          <Button
            type='primary'
            onClick={() => {
              if (parse_mode === 'rule') {
                if (
                  !tags ||
                  tags.length === 0 ||
                  tags.some((t) => !t.properties || t.properties.length === 0)
                ) {
                  modal.warning({
                    content:
                      '模型策略为"标准模式"时，实体类型不能为空，实体类型的属性不能为空',
                  })
                  return
                }
              }
              increase()
            }}
          >
            {editRules ? '保存并更新图谱' : '下一步'}
          </Button>
        </span>
      </div>
    </ActiveTagContext.Provider>
  )
}

const ActiveTagContext = createContext<{
  activeTag?: GraphTagWithId
  setActiveTagId: (id: number) => void
} | null>(null)

/** 获取和修改当前选中的tag */
// eslint-disable-next-line react-refresh/only-export-components
export const useActiveTag = () => {
  const contextValue = useContext(ActiveTagContext)
  if (!contextValue) throw new Error('需要被ActiveTagContext包裹')
  return contextValue
}
