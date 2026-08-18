import { FC, useMemo } from 'react'
import {
  Popover,
  Button,
  Skeleton,
  Empty,
  Tree,
  Tabs,
  Tag,
  Typography,
} from 'antd'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import globalConfig from '@/config'
import { cn } from '@/utils'
import { getTagTree } from '@/api/knowledge'
import LastModel from '@/pages/project/components/ModelSelect/images/last-model.svg?react'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { KnowledgeNode } from '..'

export type KnowledgeSelectProps = ValueController<string[]> & {
  knowledge: KnowledgeNode[]
  treeData?: KnowledgeNode[]
  loadData?: (val: KnowledgeNode) => Promise<any>
  // 模式切换
  mode: 'knowledge' | 'tag'
  onModeChange: (mode: 'knowledge' | 'tag') => void
  // 标签选择
  selectedTags: number[]
  onTagsChange: (tags: number[]) => void
  tagFileCount: number
}

export const KnowledgeSelect: FC<KnowledgeSelectProps> = (props) => {
  const { t } = useTranslation('pages')
  const {
    value,
    onChange,
    treeData,
    loadData,
    knowledge,
    mode,
    onModeChange,
    selectedTags,
    onTagsChange,
    tagFileCount,
  } = props

  const { version, mode: deployMode } = useDeployConfig()
  const showTagTab =
    globalConfig.env === 'development' ||
    globalConfig.apiEnv === 'test' ||
    (version === 'custom' &&
      (deployMode === 'cimc' || deployMode === 'h3c'))

  // 获取标签数据
  const { data: tagData, loading: tagLoading } = useRequest(getTagTree, {
    ready: showTagTab,
  })

  // 格式化标签树数据
  const formattedTagTree = useMemo(() => {
    if (!tagData?.group_list) return []
    return tagData.group_list.map((group: any) => ({
      title: group.tag_group_name,
      key: `group-${group.tag_group_id}`,
      // 根节点现在也可以选中了
      children: (group.tag_list || []).map((tag: any) => ({
        title: tag.tag_name,
        key: `tag-${tag.tag_id}`,
        isLeaf: true,
      })),
    }))
  }, [tagData])

  // 映射 ID 到 Key 的辅助逻辑
  const getCheckedKeys = (ids: number[]) => {
    const keys: string[] = []
    ids.forEach((id) => {
      if (id >= 0) {
        keys.push(`tag-${id}`)
      } else {
        keys.push(`group-${-id - 1}`)
      }
    })
    return keys
  }

  // 更新父节点选中状态的辅助函数
  const updateParentGroupStatus = (
    tagIds: number[],
    changedTagId: number,
  ): number[] => {
    const tagKeys = tagIds.filter((id) => id >= 0).map((id) => `tag-${id}`)
    const parentGroup = tagData?.group_list?.find((g: any) =>
      g.tag_list?.some((t: any) => t.tag_id === changedTagId),
    )

    if (!parentGroup) return tagIds

    const groupId = parentGroup.tag_group_id
    const childTagIds = (parentGroup.tag_list || []).map((t: any) => t.tag_id)
    const childTagKeys = childTagIds.map((id: number) => `tag-${id}`)
    const allChildrenSelected = childTagKeys.every((key) =>
      tagKeys.includes(key),
    )

    const encodedGroupId = -groupId - 1
    const hasGroupId = tagIds.includes(encodedGroupId)

    if (allChildrenSelected && !hasGroupId) {
      // 所有子标签都被选中，添加父节点
      return [...tagIds, encodedGroupId]
    } else if (!allChildrenSelected && hasGroupId) {
      // 有子标签未选中，移除父节点
      return tagIds.filter((id) => id !== encodedGroupId)
    }

    return tagIds
  }

  const popoverContent = (
    <div className='w-80 flex flex-col max-h-[60vh] overflow-hidden'>
      <Tabs
        activeKey={mode}
        onChange={(key) => onModeChange(key as any)}
        className='px-3 flex-1 overflow-hidden flex flex-col'
        items={[
          {
            key: 'knowledge',
            label: t('app.home.knowledgeBase'),
            children: (
              <div className='max-h-[50vh] overflow-y-auto pb-2'>
                {!treeData ? (
                  <Skeleton active className='p-4' />
                ) : treeData.length === 0 ? (
                  <Empty className='m-4' />
                ) : (
                  <Tree
                    className='p-2'
                    treeData={treeData}
                    loadData={loadData}
                    selectable={false}
                    showIcon
                    checkStrictly
                    checkable
                    checkedKeys={value}
                    onCheck={(checked, e) => {
                      const checkedKeys = (
                        'checked' in checked ? checked.checked : checked
                      ) as string[]
                      const targetNode = e.node
                      if (!e.checked || !value || value.length === 0) {
                        onChange?.(checkedKeys)
                        return
                      }
                      const { forest_id: currentForestID, type: currentType } =
                        knowledge.find((item) => item.key === checkedKeys[0])!
                      if (
                        targetNode.forest_id !== currentForestID ||
                        targetNode.type !== currentType
                      ) {
                        onChange?.([targetNode.key])
                        return
                      }
                      onChange?.(checkedKeys)
                    }}
                  />
                )}
              </div>
            ),
          },
          ...(showTagTab
            ? [
                {
                  key: 'tag',
                  label: t('app.home.tag'),
                  children: (
                    <div className='max-h-[50vh] overflow-y-auto pb-2'>
                      {tagLoading ? (
                        <Skeleton active className='p-4' />
                      ) : (
                        <div className='flex flex-col gap-4 p-2'>
                          {/* 最近使用 - 参考 ModelSelect 样式 */}
                          <div className='flex flex-col gap-2'>
                            <div className='flex items-center gap-1.5 text-[#616373] text-sm font-medium'>
                              <LastModel className='w-4 h-4' />
                              <span>{t('app.home.recentUsed')}</span>
                            </div>
                            {tagData?.recent_tag_list?.length > 0 ? (
                              <div className='flex flex-wrap gap-2'>
                                {tagData.recent_tag_list.map((tag: any) => (
                                  <Typography.Paragraph
                                    key={tag.tag_id}
                                    className={cn(
                                      'cursor-pointer bg-[#0C99FF1A] text-[#0C99FF] rounded-full',
                                      'px-2.5 py-0.5 m-0 text-xs transition-colors',
                                      selectedTags.includes(tag.tag_id)
                                        ? 'bg-[#0C99FF] text-white'
                                        : 'hover:bg-[#0C99FF33]',
                                    )}
                                    onClick={() => {
                                      const isSelected = selectedTags.includes(
                                        tag.tag_id,
                                      )
                                      let newTags = isSelected
                                        ? selectedTags.filter(
                                            (id) => id !== tag.tag_id,
                                          )
                                        : [...selectedTags, tag.tag_id]
                                      // 更新父节点状态
                                      newTags = updateParentGroupStatus(
                                        newTags,
                                        tag.tag_id,
                                      )
                                      onTagsChange(newTags)
                                    }}
                                    ellipsis={{
                                      rows: 1,
                                      tooltip: tag.tag_name,
                                    }}
                                  >
                                    {tag.tag_name}
                                  </Typography.Paragraph>
                                ))}
                              </div>
                            ) : (
                              // 为空时留出位置
                              <div className='h-6' />
                            )}
                            <div className='h-px bg-[#F2F3F5] mt-2' />
                          </div>
                          {/* 树结构 */}
                          <Tree
                            checkable
                            checkStrictly
                            selectable={false}
                            className='tag-tree'
                            treeData={formattedTagTree}
                            checkedKeys={getCheckedKeys(selectedTags)}
                            onCheck={(checked, info) => {
                              const nodeKey = info.node.key as string
                              const isChecked = info.checked
                              let newKeys = [...getCheckedKeys(selectedTags)]

                              if (nodeKey.startsWith('group-')) {
                                // 选中/取消选中分类及其所有子标签
                                const groupId = Number(
                                  nodeKey.replace('group-', ''),
                                )
                                const group = tagData.group_list.find(
                                  (g: any) => g.tag_group_id === groupId,
                                )
                                const childTagIds = (group?.tag_list || []).map(
                                  (t: any) => t.tag_id,
                                )

                                if (isChecked) {
                                  newKeys = Array.from(
                                    new Set([
                                      ...newKeys,
                                      nodeKey,
                                      ...childTagIds.map(
                                        (id: number) => `tag-${id}`,
                                      ),
                                    ]),
                                  )
                                } else {
                                  newKeys = newKeys.filter(
                                    (k) =>
                                      k !== nodeKey &&
                                      !childTagIds.some(
                                        (id: number) => `tag-${id}` === k,
                                      ),
                                  )
                                }
                              } else {
                                // 选中/取消选中单个标签
                                if (isChecked) {
                                  newKeys.push(nodeKey)
                                } else {
                                  newKeys = newKeys.filter((k) => k !== nodeKey)
                                }

                                // 检查并更新父节点（分类）的选中状态
                                const tagId = Number(
                                  nodeKey.replace('tag-', ''),
                                )
                                const parentGroup = tagData.group_list.find(
                                  (g: any) =>
                                    g.tag_list?.some(
                                      (t: any) => t.tag_id === tagId,
                                    ),
                                )

                                if (parentGroup) {
                                  const groupKey = `group-${parentGroup.tag_group_id}`
                                  const childTagIds = (
                                    parentGroup.tag_list || []
                                  ).map((t: any) => t.tag_id)
                                  const childTagKeys = childTagIds.map(
                                    (id: number) => `tag-${id}`,
                                  )
                                  // 检查该分类下的所有子标签是否都被选中
                                  const allChildrenSelected =
                                    childTagKeys.every((key) =>
                                      newKeys.includes(key),
                                    )

                                  if (allChildrenSelected) {
                                    // 所有子标签都被选中，自动选中父节点
                                    if (!newKeys.includes(groupKey)) {
                                      newKeys.push(groupKey)
                                    }
                                  } else {
                                    // 有子标签未选中，取消选中父节点
                                    newKeys = newKeys.filter(
                                      (k) => k !== groupKey,
                                    )
                                  }
                                }
                              }

                              // 转回 ID
                              const newIds = newKeys.map((k) => {
                                if (k.startsWith('tag-'))
                                  return Number(k.replace('tag-', ''))
                                return -Number(k.replace('group-', '')) - 1
                              })

                              onTagsChange(Array.from(new Set(newIds)))
                            }}
                          />
                        </div>
                      )}
                    </div>
                  ),
                },
              ]
            : []),
        ]}
      />
    </div>
  )

  const buttonText = useMemo(() => {
    if (mode === 'knowledge') {
      return value && value.length > 0
        ? t('app.home.selectedResource', { count: value.length })
        : t('app.home.selectAll')
    }
    return selectedTags.length > 0
      ? t('app.home.selectedResource', { count: tagFileCount })
      : t('app.home.selectAll')
  }, [mode, value, selectedTags, tagFileCount, t])

  return (
    <Popover
      placement='bottom'
      trigger={['click']}
      arrow={false}
      content={popoverContent}
      overlayClassName='resource-select-popover'
    >
      <Button
        className={cn(
          'w-auto rounded-[28px] px-3 h-8',
          ' text-[#653ec4] border-none bg-[#F2F0FF]',
          'hover:bg-[#c0b3ff] hover:text-[#341f68]',
          'focus:bg-[#c0b3ff] focus:text-[#341f68]',
        )}
      >
        {buttonText}
      </Button>
    </Popover>
  )
}
