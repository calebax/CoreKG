import { FC } from 'react'
import { Button, Empty, List, Skeleton, Tree } from 'antd'
import { cn } from '@/utils'
import DeleteIcon from '@/assets/icons/home-delete-1.svg?react'
import DeleteIcon2 from '@/assets/icons/home-delete-2.svg?react'
import { scroll } from '@/styles/scroll'
import { Knowledge } from '..'
import { useKnowledge } from '../../../KnowledgeContext'
import { arrayToTree, TreeNode, isFileNode } from './utils'

export type FileSelect = {
  value: Knowledge
  setValue: (val: Knowledge) => void
  search?: string
}
export const FileSelect: FC<FileSelect> = (props) => {
  const { value, setValue, search } = props
  const [showSelectedValue, setShowSelectedValue] = useState(false)
  const { fileList } = useKnowledge()
  /** 已选择的文件 */
  const selectedFile = useMemo(() => {
    if (!fileList) return []
    return fileList.filter((file) => {
      return (
        !file.is_dir && // 非文件夹
        file.parent_id && // 非知识库
        value.some((item) => item.id === file.id)
      )
    })
  }, [value, fileList])
  const filteredSelectedFile = useMemo(() => {
    if (!search) return selectedFile
    return selectedFile.filter((item) => item.name.includes(search))
  }, [search, selectedFile])

  const fileTree = useMemo(() => {
    if (!fileList) return []
    return arrayToTree(fileList, search)
  }, [fileList, search])

  if (!fileList) return <Skeleton active />
  if (fileList.length === 0)
    return (
      <Empty
        description='暂无知识库'
        className='mb-2 text-xs'
        image={Empty.PRESENTED_IMAGE_SIMPLE}
      />
    )
  return (
    <div className={cn('overflow-hidden flex flex-col gap-2')}>
      <span className='flex items-center border-b justify-between border-[#EEF0F5] p-2'>
        <div className='flex items-center'>
          <Button
            type={!showSelectedValue ? 'link' : 'text'}
            className={cn(
              'justify-center h-9 hover:bg-transparent p-0',
              !showSelectedValue
                ? 'font-medium text-[#1e1f28]'
                : 'text-[#616373] font-normal',
            )}
            onClick={() => setShowSelectedValue(false)}
          >
            文件列表
            {!showSelectedValue && fileList && fileList.length > 0
              ? `（${fileList.filter((f) => !f.is_dir && f.parent_id).length}）`
              : ''}
          </Button>
        </div>
        <Button
          type={showSelectedValue ? 'link' : 'text'}
          className={cn(
            'px-0 h-10 justify-start hover:bg-transparent',
            showSelectedValue
              ? 'font-medium text-[#1e1f28]'
              : 'text-[#616373] font-normal',
          )}
          onClick={() => setShowSelectedValue(true)}
        >
          已选择文件
          {selectedFile.length > 0 ? `（${selectedFile.length}）` : ''}
        </Button>
        <Button
          aria-label='删除已选择的文件'
          size='small'
          type='text'
          className='h-6 w-6 p-0 flex items-center justify-center rounded group'
          onClick={() => {
            setValue([])
            setShowSelectedValue(false)
          }}
        >
          <span className='relative inline-flex'>
            <DeleteIcon className='w-6 h-6 group-hover:hidden' />
            <DeleteIcon2 className='w-6 h-6 hidden group-hover:inline' />
          </span>
        </Button>
      </span>
      <div className={cn('overflow-auto break-all max-h-[28vh] pt-1', scroll)}>
        <List
          className={cn({ hidden: !showSelectedValue })}
          dataSource={filteredSelectedFile}
          renderItem={(item, index) => (
            <List.Item className='!py-0'>
              <div
                className={cn(
                  'h-8 flex items-center w-full text-sm text-[#1e1f28] rounded px-2 cursor-default',
                  index === 0 ? 'bg-[#f8f9fd]' : 'hover:bg-[#F8F9FD]',
                )}
              >
                {item.name}
              </div>
            </List.Item>
          )}
        ></List>
        <Tree
          className={cn({ hidden: showSelectedValue })}
          treeData={fileTree}
          selectable={false}
          checkable
          checkedKeys={value.map((item) => `file-${item.id}`)}
          onCheck={(_, info) => {
            const allChangedNodeIds: number[] = []
            const addNode = (node: TreeNode) => {
              if (isFileNode(node)) {
                allChangedNodeIds.push(node.id)
              } else {
                node.children.forEach((child) => addNode(child))
              }
            }
            addNode(info.node)
            if (info.checked) {
              const newNode: Knowledge = []
              allChangedNodeIds.forEach((id) => {
                if (value.some((item) => item.id === id)) return
                newNode.push(fileList.find((item) => item.id === id)!)
              })
              setValue([...value, ...newNode])
            } else {
              setValue(value.filter((v) => !allChangedNodeIds.includes(v.id)))
            }
          }}
        ></Tree>
      </div>
    </div>
  )
}
