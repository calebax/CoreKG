import { FC, useMemo, useState } from 'react'
import { Empty } from 'antd'
import { DoubleLeftOutlined, DoubleRightOutlined } from '@ant-design/icons'
import { useBoolean } from 'ahooks'
import { cn } from '@/utils'
import { GraphTagWithId } from '../../GraphProvider'
import { GraphNode } from '../../common'
import ArrowDown from './arrow.svg?react'

type NodeSidebarProps = {
  nodes: any[]
  tags: GraphTagWithId[]
  onClickNode?: (tag_name: string, node: any) => void
}

/** 实体列表侧边栏组件 */
const NodeSidebar: FC<NodeSidebarProps> = (props) => {
  const { tags, nodes, onClickNode } = props
  const [isExpanded, { toggle }] = useBoolean()

  const tagNames = useMemo(() => {
    return tags.map((tag) => tag.tag_name)
  }, [tags])

  // 按照 tag_name 分组 nodes
  const nodesByTag = useMemo(() => {
    const grouped: Record<string, GraphNode[]> = {}
    tagNames.forEach((t) => {
      grouped[t] = []
    })
    nodes.forEach((node) => {
      node.tags.forEach((t: any) => {
        if (tagNames.includes(t.name)) {
          grouped[t.name].push(node)
        }
      })
    })
    return Object.entries(grouped)
      .map((item) => {
        const [tag_name, nodes] = item
        return {
          tag_name,
          nodes,
        }
      })
      .filter((item) => {
        return item.nodes.length !== 0
      })
  }, [nodes, tagNames])

  return (
    <div
      className={cn(
        ' border border-[#EFF1F4] bg-white',
        'absolute right-0 top-0 bottom-0',
        'flex flex-col z-[10]',
        isExpanded ? 'w-60' : 'w-11',
      )}
    >
      {/* 收起状态 */}
      <div
        className={cn(
          'absolute inset-0 flex flex-col',
          !isExpanded ? '' : 'hidden',
        )}
      >
        {/* 顶部按钮 */}
        <div className='flex items-center justify-center p-2 border-b border-gray-200'>
          <DoubleLeftOutlined
            onClick={toggle}
            className='text-xs text-[#616373] cursor-pointer'
          />
        </div>
        {/* 垂直文字 */}
        <div className='flex-1 flex items-center justify-center py-4'>
          <span
            className='text-sm font-medium text-[#3C4149]'
            style={{
              writingMode: 'vertical-rl',
              textOrientation: 'upright',
            }}
          >
            实体列表
          </span>
        </div>
      </div>

      {/* 展开状态 */}
      <div
        className={cn(
          'absolute inset-0 flex flex-col',
          isExpanded ? '' : 'hidden',
        )}
      >
        {/* 顶部标题和收起按钮 */}
        <div className='flex items-center justify-between p-2 border-b border-gray-200'>
          <DoubleRightOutlined
            onClick={toggle}
            className='text-xs text-[#616373] cursor-pointer'
          />
          <span className='text-sm font-medium text-[#3C4149]'>实体列表</span>
        </div>
        {/* 可折叠面板列表 */}
        <div className='flex-1 overflow-y-auto'>
          {tags.length === 0 ? (
            <Empty description='暂无实体' className='mt-4' />
          ) : (
            <div className='flex flex-col'>
              {nodesByTag.map((item) => {
                const { tag_name, nodes } = item
                return (
                  <CollapsiblePanel key={tag_name} label={tag_name}>
                    {nodes.map((node) => (
                      <div
                        key={node.id}
                        className='text-[#3C4149] py-1 px-2 hover:bg-[#F5F7FA] rounded cursor-pointer'
                        onClick={() => onClickNode?.(tag_name, node)}
                      >
                        {node.name}
                      </div>
                    ))}
                  </CollapsiblePanel>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default NodeSidebar

type CollapsiblePanelProps = {
  label: string
  children: React.ReactNode
  defaultExpanded?: boolean
}

/** 可折叠面板组件 */
const CollapsiblePanel: FC<CollapsiblePanelProps> = ({
  label,
  children,
  defaultExpanded = true,
}) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded)

  const handleToggle = () => {
    setIsExpanded(!isExpanded)
  }

  return (
    <div className='border-b border-gray-100'>
      {/* 面板标题 */}
      <div
        className='flex items-center px-2 py-2 cursor-pointer hover:bg-[#F5F7FA]'
        onClick={handleToggle}
      >
        <span className='text-base font-medium text-[#3C4149]'>{label}</span>
        <ArrowDown
          className={cn('text-xs text-[#919497]', { 'rotate-180': isExpanded })}
        />
      </div>
      {/* 面板内容 */}
      {isExpanded && (
        <div className='flex flex-col gap-0.5 pl-2 pb-2'>{children}</div>
      )}
    </div>
  )
}
