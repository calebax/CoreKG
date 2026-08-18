import { useState, useRef, useEffect, useMemo } from 'react'
import { Button, Select, Input } from 'antd'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { listResourceTag } from '@/api/knowledge'
import FolderIcon from '@/assets/icons/docs/newFolder.svg?react'
import SearchIcon from '@/assets/icons/docs/search.svg?react'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { UploadButton } from '../ActionButtons/UploadButton'
import type { KnowledgeBaseType } from '../ActionButtons/UploadButton/Uploader'

interface OperationBarProps {
  forest_id: number
  parent_id: number
  refreshTable: () => void
  disabled?: boolean
  knowledgeBaseType?: KnowledgeBaseType
  onCreateFolder?: () => void
  onSearch?: (keyword: string) => void
  onFilterChange?: (parseStatus: string) => void
  onTagFilterChange?: (tagId: string | null) => void
}

export default function OperationBar({
  forest_id,
  parent_id,
  refreshTable,
  disabled,
  knowledgeBaseType = 'file',
  onCreateFolder,
  onSearch,
  onFilterChange,
  onTagFilterChange,
}: OperationBarProps) {
  const { t } = useTranslation('pages')
  const { version, mode } = useDeployConfig()
  const [searchKeyword, setSearchKeyword] = useState('')
  const [isSearchFocused, setIsSearchFocused] = useState(false)
  const [parseStatus, setParseStatus] = useState('all')
  const [selectedTagId, setSelectedTagId] = useState<string | null>(null)
  const searchInputRef = useRef<any>(null)

  // 判断是否显示标签筛选：本地环境、测试环境、生产环境、或 custom+cimc/h3c 模式
  const isDevEnv = import.meta.env.MODE === 'development'
  const isTestEnv = import.meta.env.MODE === 'test'
  const isProdEnv = import.meta.env.MODE === 'production'
  const showTagFilter =
    isDevEnv ||
    isTestEnv ||
    isProdEnv ||
    (version === 'custom' && (mode === 'cimc' || mode === 'h3c'))

  // 获取标签列表
  const { data: tagData, loading: tagLoading } = useRequest(
    () =>
      listResourceTag({
        limit: 1000,
        offset: 0,
      }),
    {
      ready: showTagFilter,
    },
  )

  // 构建标签选项
  const tagOptions = useMemo(() => {
    if (!tagData?.list || !Array.isArray(tagData.list)) {
      return []
    }
    return [
      { label: '全部', value: null },
      ...tagData.list.map((tag: any) => ({
        label: tag.name || tag.tag_name || '',
        value: String(tag.tag_id || tag.id || ''),
      })),
    ]
  }, [tagData])

  // 解析状态选项
  const parseStatusOptions = useMemo(() => {
    const options = [
      { label: t('app.docs.detail.parseStatus.success'), value: 'success' },
      {
        label:
          knowledgeBaseType === 'excel'
            ? '暂未支持'
            : t('app.docs.detail.parseStatus.fail'),
        value: 'fail',
      },
    ]

    // 表格类型知识库不显示"资源索引中"选项和"资源解析中"选项
    if (knowledgeBaseType !== 'excel') {
      // 非表格类型：添加"资源解析中"和"资源索引中"选项
      options.splice(1, 0, { label: '资源索引中', value: 'indexing' })
      options.splice(1, 0, { label: '资源解析中', value: 'parsing' })
    }

    if (knowledgeBaseType === 'file') {
      options.unshift({
        label: t('app.docs.detail.parseStatus.pending'),
        value: 'pending',
      })
    }
    options.unshift({
      label: t('app.docs.detail.parseStatus.all'),
      value: 'all',
    })

    return options
  }, [knowledgeBaseType, t])

  // 处理搜索按钮点击
  const handleSearchClick = () => {
    setIsSearchFocused(true)
    setTimeout(() => {
      searchInputRef.current?.focus()
    }, 0)
  }

  // 处理搜索输入框失焦
  const handleSearchBlur = () => {
    if (!searchKeyword.trim()) {
      setIsSearchFocused(false)
      // 清空输入框时自动触发搜索空内容
      onSearch?.('')
    }
  }

  // 处理搜索输入
  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    if (value.length <= 50) {
      setSearchKeyword(value)
    }
  }

  // 处理搜索提交
  const handleSearchSubmit = () => {
    onSearch?.(searchKeyword.trim())
  }

  // 处理回车搜索
  const handleSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSearchSubmit()
    }
  }

  // 处理解析状态变化
  const handleParseStatusChange = (value: string) => {
    setParseStatus(value)
    onFilterChange?.(value)
  }

  // 处理标签筛选变化
  const handleTagFilterChange = (value: string | null) => {
    setSelectedTagId(value)
    onTagFilterChange?.(value)
  }

  return (
    <div className='bg-white rounded-lg'>
      <div className='flex items-center justify-between'>
        {/* 左侧：解析状态筛选和标签筛选 */}
        <div className='flex items-center gap-1.5'>
          <span className='text-sm font-medium text-[#919497]'>
            {t('app.docs.detail.parseStatus.label')}
          </span>
          <Select
            value={parseStatus}
            onChange={handleParseStatusChange}
            options={parseStatusOptions}
            className='w-[120px]'
            size='middle'
            style={{
              color: '#0C1F17',
            }}
            dropdownStyle={{
              minWidth: '230px',
              borderRadius: '8px',
              boxShadow:
                '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
              border: '1px solid #e5e7eb',
              backgroundColor: 'white',
              marginTop: '4px',
            }}
          />
          {/* 标签筛选（仅在指定环境下显示） */}
          {showTagFilter && (
            <>
              <span className='text-sm font-medium text-[#919497] ml-2'>
                标签
              </span>
              <Select
                value={selectedTagId}
                onChange={handleTagFilterChange}
                options={tagOptions}
                className='w-[120px]'
                size='middle'
                loading={tagLoading}
                placeholder='请选择标签'
                style={{
                  color: '#0C1F17',
                }}
                dropdownStyle={{
                  minWidth: '230px',
                  borderRadius: '8px',
                  boxShadow:
                    '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
                  border: '1px solid #e5e7eb',
                  backgroundColor: 'white',
                  marginTop: '4px',
                }}
              />
            </>
          )}
        </div>

        {/* 右侧：搜索和操作按钮 */}
        <div className='flex items-center gap-3'>
          {/* 搜索功能 */}
          <div className='flex items-center'>
            {!isSearchFocused ? (
              <Button
                type='text'
                className='flex items-center gap-1 h-[30px] border border-[#0C99FF] px-3 py-2 text-[#0C99FF] bg-white hover:bg-white'
                onClick={handleSearchClick}
              >
                <SearchIcon className='w-4 h-4' />
                <span className='text-sm'>{t('app.docs.detail.search')}</span>
              </Button>
            ) : (
              <Input
                ref={searchInputRef}
                value={searchKeyword}
                onChange={handleSearchChange}
                onBlur={handleSearchBlur}
                onKeyDown={handleSearchKeyDown}
                placeholder={t('app.docs.detail.searchPlaceholder')}
                className='w-64 border border-[#0C99FF]'
                maxLength={50}
              />
            )}
          </div>

          {/* 上传文件按钮 */}
          <UploadButton
            disabled={disabled}
            forest_id={forest_id}
            parent_id={parent_id}
            knowledgeBaseType={knowledgeBaseType}
            afterUpload={refreshTable}
          />

          {/* 新建文件夹按钮 */}
          <Button
            className={cn(
              'flex items-center h-[30px] font-medium bg-white gap-2 px-2.5 py-2 border-[#0C99FF] text-[#0C99FF] hover:border-[#0C99FF]',
              {
                'opacity-40 cursor-not-allowed grayscale': disabled,
              },
            )}
            icon={<FolderIcon className='w-4 h-4' />}
            onClick={onCreateFolder}
            disabled={disabled}
          >
            <span className='text-sm'>{t('app.docs.detail.newFolder')}</span>
          </Button>
        </div>
      </div>
    </div>
  )
}
