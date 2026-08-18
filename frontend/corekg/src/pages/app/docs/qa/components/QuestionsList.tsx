import { Button, Dropdown, Modal, Switch, Tooltip, message } from 'antd'
import { ExclamationCircleOutlined, MoreOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import DeleteIcon from '@/assets/icons/docs/delete-file.svg?react'
import EditIcon from '@/assets/icons/docs/edit-file.svg?react'
import TableWithPagination from '@/components/common/TableWithPagination'
import HelpIcon from '../images/help-tip.svg?react'

interface QAPair {
  id: string
  qa_question: string
  qa_answer: string
  created_at: string
  updated_at: string
  source_from: string
}

interface QuestionsListProps {
  loading: boolean
  dataSource: QAPair[]
  total: number
  current: number
  pageSize: number
  onPageChange: (page: number, pageSize?: number) => void
  onEdit: (record: QAPair) => void
  onDelete: (ids: string[]) => void
  isAdmin?: boolean
  onEnableChange?: (v: boolean, file: any) => void
  enablingQaId?: string | null // 正在处理启用状态的问答对ID
}

export default function QuestionsList({
  loading,
  dataSource,
  total,
  current,
  pageSize,
  onPageChange,
  onEdit,
  onDelete,
  isAdmin = false,
  onEnableChange,
  enablingQaId = null,
}: QuestionsListProps) {
  console.log(dataSource)

  const { t: tC } = useTranslation('common')
  const { t: tM } = useTranslation('messages')
  const { t, i18n } = useTranslation('pages')
  const handleDelete = (record: QAPair) => {
    if (!isAdmin) {
      message.error(tM('noPermissionContactKbAdmin'))
      return
    }

    Modal.confirm({
      title: tC('button.confirmDelete'),
      icon: <ExclamationCircleOutlined />,
      content: tM('confirmDeleteQAPairIrreversible'),
      okText: tC('button.confirmDelete'),
      cancelText: tC('button.cancel'),
      onOk: () => onDelete([record.id]),
    })
  }

  const handleEdit = (record: QAPair) => {
    if (!isAdmin) {
      message.error(tM('noPermissionContactKbAdmin'))
      return
    }
    onEdit(record)
  }

  const handleEnableChange = (v: boolean, record: any) => {
    if (!isAdmin) {
      message.error(tM('noPermissionContactKbAdmin'))
      return
    }
    onEnableChange?.(v, record)
  }

  const columns = [
    {
      title: t('app.docs.question'),
      dataIndex: 'qa_question',
      key: 'qa_question',
      width: 300,
      render: (text: string) => (
        <div className='flex items-center'>
          <span className='truncate'>{text}</span>
        </div>
      ),
    },
    {
      title: t('app.docs.source'),
      dataIndex: 'source_from',
      key: 'source',
      width: 120,
      render: (v: string) => (v === 'manul_import' ? '手动新建' : '本地上传'),
    },
    {
      title: t('app.docs.creationTime'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (text: string) => {
        if (!text) return '-'
        return new Date(text)
          .toLocaleString(i18n.language, {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
          })
          .replace(/\//g, '-')
      },
    },
    // {
    //   title: t('app.docs.updatedTime'),
    //   dataIndex: 'updated_at',
    //   key: 'updated_at',
    //   width: 180,
    //   render: (text: string) => {
    //     if (!text) return '-'
    //     return new Date(text)
    //       .toLocaleString(i18n.language, {
    //         year: 'numeric',
    //         month: '2-digit',
    //         day: '2-digit',
    //         hour: '2-digit',
    //         minute: '2-digit',
    //       })
    //       .replace(/\//g, '-')
    //   },
    // },
    {
      title: (
        <div className='flex items-center gap-[4px] font-medium text-[#919497] text-sm leading-[22px]'>
          启用状态
          <Tooltip title='启用后，系统将在问答时检索并引用该资源。'>
            <HelpIcon />
          </Tooltip>
        </div>
      ),
      dataIndex: 'enable',
      key: 'enable',
      width: '15%',
      render: (enable: boolean, record: any) => {
        // 如果该问答对正在处理启用状态，禁用开关
        const isProcessing = enablingQaId === record.id
        return (
          <div onClick={(e) => e.stopPropagation()}>
            <Switch
              value={enable}
              size='small'
              onChange={(v) => handleEnableChange(v, record)}
              disabled={isProcessing}
            />
          </div>
        )
      },
    },
    {
      title: (
        <div className='font-medium text-[#919497] text-sm leading-[22px]'>
          {t('app.docs.detail.actions')}
        </div>
      ),
      key: 'actions',
      width: '15%',
      render: (_: any, record: any) => {
        const isReadOnly = !isAdmin
        const sharedItemClass =
          'text-[#2D2D2D] hover:!bg-[#F5F5F5] px-[10px] py-[7px] rounded'
        const disabledItemClass =
          'cursor-not-allowed !text-[#C0C4CC] hover:!bg-transparent'

        const menuItems = [
          {
            key: 'edit',
            label: (
              <span
                className='flex-1 min-w-0 truncate font-medium'
                title={t('app.docs.detail.edit')}
              >
                {tC('button.edit')}
              </span>
            ),
            icon: (
              <EditIcon
                className={`w-4 h-4 ${isReadOnly ? 'opacity-40' : ''}`}
              />
            ),
            onClick: isReadOnly ? undefined : () => handleEdit(record),
            className: `${sharedItemClass} ${isReadOnly ? disabledItemClass : ''}`,
            disabled: isReadOnly,
          },
          {
            key: 'delete',
            label: (
              <span
                className='flex-1 min-w-0 truncate font-medium'
                title={t('app.docs.detail.delete')}
              >
                {t('app.docs.detail.delete')}
              </span>
            ),
            icon: (
              <DeleteIcon
                className={`w-4 h-4 ${isReadOnly ? 'opacity-40' : ''}`}
              />
            ),
            onClick: isReadOnly ? undefined : () => handleDelete(record),
            className: `${sharedItemClass} ${isReadOnly ? disabledItemClass : ''}`,
            disabled: isReadOnly,
          },
        ]

        return (
          <Dropdown
            menu={{
              items: menuItems,
              onClick: (e) => e.domEvent.stopPropagation(),
              className:
                'px-[10px] py-[10px] min-w-[190px] max-w-[230px] bg-white',
            }}
            trigger={['click']}
            placement='bottomRight'
          >
            <Button
              type='text'
              icon={<MoreOutlined style={{ transform: 'rotate(90deg)' }} />}
              className='!px-2'
              onClick={(e) => e.stopPropagation()}
              aria-label={t('app.docs.detail.actions')}
            />
          </Dropdown>
        )
      },
    },
  ]

  return (
    <TableWithPagination
      loading={loading}
      columns={columns}
      dataSource={dataSource}
      rowKey='id'
      total={total}
      current={current}
      pageSize={pageSize}
      onPageChange={onPageChange}
      scroll={{ x: 1000 }}
      tableHeight={{
        default: 'h-full',
        sm: 'sm:h-full',
        lg: 'lg:h-full',
        '2xl': '2xl:h-full',
      }}
    />
  )
}
