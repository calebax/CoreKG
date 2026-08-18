import { useState, useMemo, useEffect } from 'react'
import { Modal, Tree, Spin, message } from 'antd'
import { useRequest } from 'ahooks'
import { getTagTree, setResourceTag } from '@/api/knowledge'
import { FileItem } from '../../types'

interface TagSelectModalProps {
  open: boolean
  file: FileItem | null
  onCancel: () => void
  onSuccess: () => void
}

export default function TagSelectModal({
  open,
  file,
  onCancel,
  onSuccess,
}: TagSelectModalProps) {
  const [checkedKeys, setCheckedKeys] = useState<string[]>([])

  // 获取标签树数据
  const { data: tagData, loading: tagLoading } = useRequest(getTagTree, {
    ready: open,
  })

  // 格式化标签树数据
  const formattedTagTree = useMemo(() => {
    if (!tagData?.group_list) return []
    return tagData.group_list.map((group: any) => ({
      title: group.tag_group_name,
      key: `group-${group.tag_group_id}`,
      children: (group.tag_list || []).map((tag: any) => ({
        title: tag.tag_name,
        key: `tag-${tag.tag_id}`,
        isLeaf: true,
      })),
    }))
  }, [tagData])

  // 根据文件的 tag_list 回显已选择的标签
  useEffect(() => {
    if (open && file?.tag_list && tagData?.group_list) {
      const fileTagIds = file.tag_list.map((tag: any) => tag.tag_id)
      const keys: string[] = []

      // 遍历所有标签组，检查哪些标签被选中
      tagData.group_list.forEach((group: any) => {
        const groupTagIds = (group.tag_list || []).map((t: any) => t.tag_id)
        const selectedTagIds = groupTagIds.filter((id: number) =>
          fileTagIds.includes(id),
        )

        // 如果该组下的所有标签都被选中，则选中组节点
        if (
          selectedTagIds.length > 0 &&
          selectedTagIds.length === groupTagIds.length
        ) {
          keys.push(`group-${group.tag_group_id}`)
        }

        // 添加选中的标签节点
        selectedTagIds.forEach((id: number) => {
          keys.push(`tag-${id}`)
        })
      })

      setCheckedKeys(keys)
    } else if (open && (!file?.tag_list || file.tag_list.length === 0)) {
      setCheckedKeys([])
    }
  }, [open, file, tagData])

  // 处理树节点选中变化
  const handleCheck = (checked: any, info: any) => {
    const nodeKey = info.node.key as string
    const isChecked = info.checked
    let newKeys = [...checkedKeys]

    if (nodeKey.startsWith('group-')) {
      // 选中/取消选中分类及其所有子标签
      const groupId = Number(nodeKey.replace('group-', ''))
      const group = tagData?.group_list?.find(
        (g: any) => g.tag_group_id === groupId,
      )
      const childTagIds = (group?.tag_list || []).map((t: any) => t.tag_id)
      const childTagKeys = childTagIds.map((id: number) => `tag-${id}`)

      if (isChecked) {
        newKeys = Array.from(new Set([...newKeys, nodeKey, ...childTagKeys]))
      } else {
        newKeys = newKeys.filter(
          (k) => k !== nodeKey && !childTagKeys.includes(k),
        )
      }
    } else {
      // 选中/取消选中单个标签
      if (isChecked) {
        if (!newKeys.includes(nodeKey)) {
          newKeys.push(nodeKey)
        }
      } else {
        newKeys = newKeys.filter((k) => k !== nodeKey)
      }

      // 检查并更新父节点（分类）的选中状态
      const tagId = Number(nodeKey.replace('tag-', ''))
      const parentGroup = tagData?.group_list?.find((g: any) =>
        g.tag_list?.some((t: any) => t.tag_id === tagId),
      )

      if (parentGroup) {
        const groupKey = `group-${parentGroup.tag_group_id}`
        const childTagIds = (parentGroup.tag_list || []).map(
          (t: any) => t.tag_id,
        )
        const childTagKeys = childTagIds.map((id: number) => `tag-${id}`)
        // 检查该分类下的所有子标签是否都被选中
        const allChildrenSelected = childTagKeys.every((key) =>
          newKeys.includes(key),
        )

        if (allChildrenSelected) {
          // 所有子标签都被选中，自动选中父节点
          if (!newKeys.includes(groupKey)) {
            newKeys.push(groupKey)
          }
        } else {
          // 有子标签未选中，取消选中父节点
          newKeys = newKeys.filter((k) => k !== groupKey)
        }
      }
    }

    setCheckedKeys(newKeys)
  }

  // 处理确认
  const handleConfirm = async () => {
    if (!file) return

    // 从选中的 keys 中提取标签 ID（只取 tag- 开头的）
    const tagIds = checkedKeys
      .filter((key) => key.startsWith('tag-'))
      .map((key) => Number(key.replace('tag-', '')))

    try {
      await setResourceTag({
        resource_id: file.id,
        resource_type: 'file',
        tag_ids: tagIds,
      })
      message.success('标签设置成功')
      onSuccess()
      onCancel()
    } catch (error) {
      console.error('设置标签失败:', error)
      // message.error('标签设置失败，请稍后再试')
    }
  }

  return (
    <Modal
      title='标签'
      open={open}
      onCancel={onCancel}
      footer={null}
      width={400}
      className='tag-select-modal'
      styles={{
        body: {
          padding: '16px',
        },
      }}
    >
      <div className='flex flex-col'>
        <div className='flex-1 overflow-y-auto max-h-[400px]'>
          {tagLoading ? (
            <div className='flex justify-center py-8'>
              <Spin />
            </div>
          ) : (
            <Tree
              checkable
              checkStrictly
              selectable={false}
              className='tag-tree'
              treeData={formattedTagTree}
              checkedKeys={checkedKeys}
              onCheck={handleCheck}
            />
          )}
        </div>
        <div className='flex gap-2 justify-end items-center pt-4 border-t border-gray-200 mt-4'>
          <button
            className='px-6 py-2 bg-[#F5F5F5] text-[#0C1F17] rounded-md text-sm cursor-pointer font-medium hover:bg-[#F5F5F5]'
            onClick={onCancel}
          >
            取消
          </button>
          <button
            className='px-6 py-2 bg-[#0C99FF] text-white rounded-md text-sm cursor-pointer font-medium hover:bg-[#0C99FF]'
            onClick={handleConfirm}
          >
            确定
          </button>
        </div>
      </div>
    </Modal>
  )
}
