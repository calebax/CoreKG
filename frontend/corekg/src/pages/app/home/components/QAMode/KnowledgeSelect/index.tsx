import { FC } from 'react'
import { Button, Input } from 'antd'
import { useClickAway } from 'ahooks'
import { cn } from '@/utils'
import SearchIcon from '@/assets/icons/search.svg?react'
import { DialogInitData } from '../../../QA/type'
import { DocsSelect } from './DocsSelect'
import { FileSelect } from './FileSelect'

export type Knowledge = DialogInitData['knowledge']
export type KnowledgeSelect = {
  type: DialogInitData['type']
  selectedKnoledge: Knowledge
  onSelectKnowledge: (type: DialogInitData['type'], val: Knowledge) => void
  onClose: () => void
  onDestory: () => void
}
export const KnowledgeSelect: FC<KnowledgeSelect> = (props) => {
  const {
    type,
    selectedKnoledge: knowledge,
    onSelectKnowledge: setKnowledge,
    onClose,
    onDestory,
  } = props
  const [search, setSearch] = useState<string>()
  // const [knowledge, setKnowledge] = useState<Knowledge>([])
  const selectContainer = useRef<HTMLDivElement>(null)
  /** 如果已有选择知识仅关闭浮窗 否则直接销毁重新选择类型 */
  useClickAway(() => {
    if (knowledge.length > 0) {
      onClose()
    } else {
      onDestory()
    }
  }, selectContainer)
  return (
    <div
      ref={selectContainer}
      className={cn(
        'w-100 ml-[-16px] max-h-[40vh] overflow-hidden',
        'bg-white rounded-[10px] shadow-[0px_2px_12px_0px_rgba(0,0,0,0.1)]',
        'p-[10px] flex flex-col gap-2 mt-3.5',
      )}
    >
      <Input
        prefix={<SearchIcon className='w-3.5 h-3.5' />}
        allowClear
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder='请输入内容'
        className={cn('h-8 rounded-[19px] px-2 text-sm text-[#1e1f28]')}
        style={{ borderColor: search ? '#a895fc' : '#d7d9e5' }}
      />
      {/* setType改变当前组件选择的类型 onSelectKnowledge的type控制关联知识的类型 */}
      {type === 'forest' ? (
        <DocsSelect
          value={knowledge}
          setValue={(val) => setKnowledge('forest', val)}
          search={search}
        />
      ) : (
        <FileSelect
          value={knowledge}
          setValue={(val) => setKnowledge('file_list', val)}
          search={search}
        />
      )}
    </div>
  )
}
