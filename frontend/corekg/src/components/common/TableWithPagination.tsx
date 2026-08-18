import { useMemo, useState } from 'react'
import { Table, Pagination, Select, Checkbox, Input, Button } from 'antd'
import type { TableProps, PaginationProps } from 'antd'
import { useTranslation } from 'react-i18next'
import PagePrevDisabledIcon from '@/assets/icons/docs/page-back-disabled.svg?react'
import PagePrevIcon from '@/assets/icons/docs/page-back.svg?react'
import PageNextDisabledIcon from '@/assets/icons/docs/page-next-disabled.svg?react'
import PageNextIcon from '@/assets/icons/docs/page-next.svg?react'
import TableEmpty from '@/assets/icons/docs/table-empty.svg'
import scrollStyles from '@/styles/scroll/styles.module.scss'
import '@/styles/tableWithPagination.css'

interface TableWithPaginationProps<T>
  extends Omit<TableProps<T>, 'pagination' | 'onChange'> {
  total: number
  current: number
  pageSize: number
  onPageChange: (page: number, pageSize?: number) => void
  paginationProps?: Omit<
    PaginationProps,
    'total' | 'current' | 'pageSize' | 'onChange'
  >
  onTableChange?: (sorter: any, filters?: any, extra?: any) => void
  pageSizeOptions?: number[]
  tableHeight?: {
    default?: string
    sm?: string
    lg?: string
    xl?: string
    '2xl'?: string
  }
}

const TableWithPagination = <T extends object>({
  total,
  current,
  pageSize,
  onPageChange,
  paginationProps,
  onTableChange,
  pageSizeOptions = [10, 20, 50, 100],
  tableHeight,
  ...tableProps
}: TableWithPaginationProps<T>) => {
  const { t } = useTranslation('pages')
  const [goToValue, setGoToValue] = useState('')

  const handlePageSizeChange = (value: number) => {
    onPageChange(1, value)
  }

  const handleGoToPage = () => {
    const page = parseInt(goToValue)
    const maxPage = Math.ceil(total / pageSize)
    if (page >= 1 && page <= maxPage) {
      onPageChange(page, pageSize)
      setGoToValue('')
    }
  }

  // 处理表格变化事件，确保正确传递排序参数
  const handleTableChange: TableProps<T>['onChange'] = (
    pagination,
    filters,
    sorter,
    extra,
  ) => {
    if (onTableChange) {
      onTableChange(sorter, filters, extra)
    }
  }

  // 默认的高度配置 - 使用固定高度而不是最大高度
  const defaultHeight = {
    default: 'h-[calc(100vh-426px)]',
    sm: 'sm:h-[calc(100vh-426px)]',
    lg: 'lg:h-[calc(100vh-400px)]',
    '2xl': '2xl:h-[calc(100vh-376px)]',
  }

  // 合并用户提供的配置与默认配置
  const heightConfig = {
    default: tableHeight?.default || defaultHeight.default,
    sm: tableHeight?.sm || defaultHeight.sm,
    lg: tableHeight?.lg || defaultHeight.lg,
    '2xl': tableHeight?.['2xl'] || defaultHeight['2xl'],
  }

  // 检查是否使用flex布局（h-full类型）
  const useFlexHeight = Object.values(heightConfig).some(
    (height) => height && height.includes('h-full'),
  )

  // 构建最终的class字符串
  const heightClass = useFlexHeight
    ? `flex-1 overflow-auto ${scrollStyles.scroll}`
    : `${heightConfig.default} ${heightConfig.sm} ${heightConfig.lg} ${heightConfig['2xl']} overflow-auto ${scrollStyles.scroll}`
  // 判断是否为空数据
  const isEmpty =
    !Array.isArray((tableProps as any).dataSource) ||
    ((tableProps as any).dataSource?.length || 0) === 0

  return (
    <div className='w-[calc(100%+6px)] flex flex-col h-full ml-[-6px]'>
      {/* 表格内容区域 */}
      <div className={heightClass + ' relative'}>
        {' '}
        <Table
          {...tableProps}
          pagination={false}
          className='mb-0 custom-table'
          onChange={handleTableChange}
          rowClassName={() =>
            'h-10 hover:bg-[#F2F3F5] transition-colors duration-500 ease-in-out'
          }
          showSorterTooltip={false}
          locale={{ emptyText: null }}
        />
        {isEmpty && (
          <div className='absolute inset-0 flex items-center justify-center'>
            <div className='flex flex-col items-center'>
              <img
                src={TableEmpty}
                alt=''
                className='w-28 h-28 mb-2 opacity-90'
              />
              <div className='text-[#919497] text-sm'>
                {t('app.docs.emptyUploadFirst')}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* 分页区域 - 左侧显示总数，右侧显示页码和页面大小选择器 */}
      {total > 0 && (
        <div className='flex justify-between items-center bg-white p-5 pb-0 flex-shrink-0'>
          {/* 左侧区域：总数显示 */}
          <div className='flex items-center'>
            {/* <div className='text-sm text-[#000000D9] font-normal'>
              {t('app.docs.detail.totalItems', { count: total })}
            </div> */}
          </div>

          {/* 右侧区域：页码导航 + 页面大小选择器 */}
          <div className='flex items-center gap-4'>
            <div className='text-sm text-[#000000D9] font-normal'>
              {t('app.docs.detail.totalItems', { count: total })}
            </div>
            {/* 页码导航 */}
            <div className='flex items-center'>
              <Pagination
                {...paginationProps}
                current={current}
                pageSize={pageSize}
                total={total}
                onChange={(page) => onPageChange(page, pageSize)}
                showSizeChanger={false}
                showQuickJumper={false}
                showLessItems={true}
                showTotal={() => null}
                className={`custom-clean-pagination ${paginationProps?.className || ''}`}
                itemRender={(page, type, originalElement) => {
                  if (type === 'prev') {
                    const isDisabled = current === 1
                    return (
                      <button
                        className='flex items-center justify-center w-8 h-8 border-none rounded-sm disabled:cursor-not-allowed transition-colors duration-200 mr-1'
                        disabled={isDisabled}
                      >
                        {isDisabled ? (
                          <PagePrevDisabledIcon className='w-4 h-4' />
                        ) : (
                          <PagePrevIcon className='w-4 h-4' />
                        )}
                      </button>
                    )
                  }
                  if (type === 'next') {
                    const isDisabled = current === Math.ceil(total / pageSize)
                    return (
                      <button
                        className='flex items-center justify-center w-8 h-8 border-none rounded-sm disabled:cursor-not-allowed transition-colors duration-200 ml-1'
                        disabled={isDisabled}
                      >
                        {isDisabled ? (
                          <PageNextDisabledIcon className='w-4 h-4' />
                        ) : (
                          <PageNextIcon className='w-4 h-4' />
                        )}
                      </button>
                    )
                  }
                  if (type === 'page') {
                    return (
                      <button
                        className={`flex items-center justify-center w-8 h-8 cursor-pointer transition-all duration-200 text-sm font-normal mx-[2px] rounded-sm border
                        ${
                          page === current
                            ? 'text-[#0C99FF] border-[#0C99FF]'
                            : 'border-none text-[#333333] hover:border-[#0C99FF] hover:text-[#0C99FF]'
                        }`}
                      >
                        {page}
                      </button>
                    )
                  }
                  if (type === 'jump-prev' || type === 'jump-next') {
                    return (
                      <span className='flex items-center justify-center w-8 h-8 text-sm text-[#cccccc] cursor-pointer hover:text-[#0C99FF] transition-colors duration-200 mx-[2px]'>
                        •••
                      </span>
                    )
                  }
                  return originalElement
                }}
              />
            </div>

            {/* 页面大小选择器 */}
            <Select
              value={pageSize}
              onChange={handlePageSizeChange}
              options={pageSizeOptions.map((size) => ({
                value: size,
                label: t('app.docs.detail.perPage', { count: size }),
              }))}
              className='min-w-[100px] h-[32px] text-sm custom-page-size-select'
              // dropdownClassName='custom-page-size-select-dropdown'
              classNames={{
                popup: {
                  root: 'custom-page-size-select-dropdown',
                },
              }}
              popupMatchSelectWidth={false}
              size='middle'
              bordered={true}
              style={{
                borderRadius: '4px',
                color: '#000000D9',
              }}
              styles={{
                popup: {
                  root: { borderColor: '#EFF1F4' },
                },
              }}
            />
          </div>
        </div>
      )}
    </div>
  )
}

export default TableWithPagination
