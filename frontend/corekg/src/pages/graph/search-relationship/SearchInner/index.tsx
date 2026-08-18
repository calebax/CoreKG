import { FC, useMemo } from 'react'
import { Button, Empty } from 'antd'
import { ArrowLeftRight1Icon } from 'tdesign-icons-react'
import { useImmer } from 'use-immer'
import { cn } from '@/utils'
import { getKnowledgeGraph } from '@/api/graph'
import { Neo4jGraphWithFilter } from '../../components/Neo4jGraphWithFilter'
import { Filter, NodeFilter } from '../../components/NodeFilter'
import { SpinWithMask } from '../../components/SpinWithMask'
import useKnowledgeGraphData from '../../hooks/useKnowledgeGraphData'
import EmptyPic from './EmptyPic.svg?react'

export const SearchInner: FC<Style & { graph_id: number }> = (props) => {
  const { graph_id, className, style } = props
  const [filterCache, setCache] = useImmer<[Filter, Filter]>([{}, {}])
  const [filters, setFilters] = useImmer<[Filter, Filter]>([{}, {}])
  const btnDisabled = useMemo(() => {
    return !filterCache.map(Object.values).flat().some(Boolean)
  }, [filterCache])
  const graphDisabled = useMemo(() => {
    return !filters.map(Object.values).flat().some(Boolean)
  }, [filters])

  const { nodes, relationships, loading, key } = useKnowledgeGraphData({
    graph_id,
    source: filters[0],
    target: filters[1],
    disableFilter: graphDisabled,
  })

  const fetchGraphData = async () => {
    if (graphDisabled) {
      return { nodes: [], relationships: [] }
    }
    const { knowledge_graph } = await getKnowledgeGraph({
      graph_id,
      src_tag: filters[0]?.tag_name,
      src_name: filters[0]?.search,
      dst_tag: filters[1]?.tag_name,
      dst_name: filters[1]?.search,
    })
    return {
      nodes: knowledge_graph.nodes ?? [],
      relationships: knowledge_graph.edges ?? [],
    }
  }

  return (
    <div
      className={cn('flex flex-col overflow-hidden gap-4', className)}
      style={style}
    >
      <div className='flex items-center justify-between'>
        <NodeFilter
          title='源实体类型'
          inputPlaceholder='请输入源实体'
          value={filterCache[0]}
          onChange={(v) => {
            setCache((draft) => {
              draft[0] = v ?? {}
            })
          }}
          onClear={(v) => {
            setFilters((draft) => {
              draft[0] = v ?? []
            })
          }}
        />
        <ArrowLeftRight1Icon
          className='text-xl cursor-pointer'
          onClick={() => {
            setCache((draft) => {
              ;[draft[0], draft[1]] = [draft[1], draft[0]]
            })
          }}
        />
        <NodeFilter
          title='目标实体类型'
          inputPlaceholder='请输入目标实体'
          value={filterCache[1]}
          onChange={(v) => {
            setCache((draft) => {
              draft[1] = v ?? {}
            })
          }}
          onClear={(v) => {
            setFilters((draft) => {
              draft[1] = v ?? {}
            })
          }}
        />
        <Button
          type='primary'
          disabled={btnDisabled}
          onClick={() => setFilters(JSON.parse(JSON.stringify(filterCache)))}
        >
          开始搜索
        </Button>
        <Button
          onClick={() => {
            setCache([{}, {}])
          }}
        >
          重置条件
        </Button>
      </div>
      {graphDisabled ? (
        <Empty
          className='h-full flex flex-col items-center justify-center'
          image={<EmptyPic className='mx-auto' />}
          description={
            <span className='text-center'>
              请先输入查询条件
              <br />
              系统将为您生成关联图谱
            </span>
          }
        />
      ) : (
        <div className=' relative flex-1'>
          <SpinWithMask show={loading} />
          <Neo4jGraphWithFilter
            key={key}
            graph_id={graph_id}
            nodes={nodes}
            relationships={relationships}
            fetchGraphData={fetchGraphData}
            autoFitOnRefresh
          />
        </div>
      )}
    </div>
  )
}
