import { useState, useEffect, FC, useMemo } from 'react'
import { App, Button, Form, Input, InputRef, Modal } from 'antd'
import { useRequest } from 'ahooks'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
import { match } from 'ts-pattern'
import { cn, hasModulePermission } from '@/utils'
import {
  getKnowledgeBaseDetail,
  updateKnowledgeBaseName,
} from '@/api/knowledge'
import DatabaseBigIcon from '@/assets/icons/docs/database-big.svg'
import MultimodalBigIcon from '@/assets/icons/docs/multimodal-big.svg'
import PermissionHoverIcon from '@/assets/icons/docs/permission-icon-hover.svg?react'
import PermissionIcon from '@/assets/icons/docs/permission-icon.svg?react'
import QABigIcon from '@/assets/icons/docs/qa-big.svg'
import TableBigIcon from '@/assets/icons/docs/table-big.svg'
import UploadRecordHoverIcon from '@/assets/icons/docs/upload-record-hover.svg?react'
import UploadRecordIcon from '@/assets/icons/docs/upload-record.svg?react'
import { useLoginGlobalData } from '@/utils/useLoginGlobalData'
import { UploadRecordsSidebar } from '../../components/ForestUploadBtn/UploadRecordsSidebar'
import DBBanner from '../banners/DBBanner.png'
import ExcelBanner from '../banners/ExcelBanner.png'
import MutimodalBanner from '../banners/MutimodalBanner.png'
import QABanner from '../banners/QABanner.png'
import EditIcon from '../images/edit.svg?react'
import type { KnowledgeBaseType } from './ActionButtons/UploadButton/Uploader'
import { MutimodalGraph } from './MutimodalGraph'
import PermissionSidebar from './PermissionSidebar'

interface KnowledgeBaseInfoProps {
  knowledgeBaseName: string
  knowledgeBaseId: number
  createdAt?: string
  isAdmin: boolean
  description?: string
  knowledgeBaseType?: KnowledgeBaseType
  knowledgeBaseImage?: string // 新增图片字段
  // 新增：操作按钮所需参数
  parent_id?: number
  refreshTable?: () => void
  disabled?: boolean
  onCreateFolder?: () => void
  onChangeName?: (v: string) => void
  onChangeDesc?: (v: string) => void
  knowledgeBaseSize?: number
  knowledgeBaseFileCount?: number
  graph_info?: any
  graph_updatable?: boolean
  is_admin?: boolean
  knowledge_status?: string
}

export default function KnowledgeBaseInfo({
  knowledgeBaseName,
  knowledgeBaseId,
  createdAt,
  isAdmin,
  description,
  knowledgeBaseType,
  knowledgeBaseImage,
  parent_id = 0,
  refreshTable,
  disabled,
  onCreateFolder,
  onChangeName,
  onChangeDesc,
  knowledgeBaseSize,
  knowledgeBaseFileCount,
  graph_info,
  graph_updatable,
  is_admin,
  knowledge_status,
}: KnowledgeBaseInfoProps) {
  const { t } = useTranslation('pages')
  const { license } = useLoginGlobalData()
  const hasGraphPermission = hasModulePermission(license, 'graph')
  const [permissionDrawerVisible, setPermissionDrawerVisible] = useState(false)
  const [uploadRecordsSidebarVisible, setUploadRecordsSidebarVisible] =
    useState(false)
  // 名称和描述更改弹窗
  const [nameEditing, setNameEditing] = useState(false)

  const [permissionData, setPermissionData] = useState<{
    manager_ids: number[]
    public_scope: 'company' | 'custom'
    scope_ids: number[]
    name: string
    description: string
  } | null>(null)

  const formatDate = (timestamp: string | undefined) => {
    if (!timestamp) return '-'
    return dayjs(timestamp).format('YYYY-MM-DD HH:mm')
  }

  // 获取知识库权限数据
  const fetchPermissionData = async () => {
    try {
      const res = await getKnowledgeBaseDetail({ id: knowledgeBaseId })

      if (res && res.data) {
        setPermissionData({
          manager_ids: res.data.manager_ids || [],
          public_scope: res.data.public_scope || 'company',
          scope_ids: res.data.scope_ids || [],
          name: res.data.name || '',
          description: res.data.description || '',
        })
      }
    } catch (error) {
      console.error('获取权限数据失败:', error)
    }
  }

  useEffect(() => {
    if (knowledgeBaseId && permissionDrawerVisible) {
      fetchPermissionData()
    }
  }, [knowledgeBaseId, permissionDrawerVisible])

  const handlePermissionSuccess = () => {
    // 权限更新成功后，重新获取数据
    fetchPermissionData()
  }

  function convertBytes(bytes: number): string {
    const kb = 1024
    const mb = kb * 1024
    const gb = mb * 1024

    if (bytes >= gb) {
      return (bytes / gb).toFixed(2) + ' GB'
    } else if (bytes >= mb) {
      return (bytes / mb).toFixed(2) + ' MB'
    } else if (bytes >= kb) {
      return (bytes / kb).toFixed(2) + ' KB'
    }
    return bytes + ' B'
  }

  const banner = useMemo(() => {
    switch (knowledgeBaseType) {
      case 'file':
        return MutimodalBanner
      case 'excel':
        return ExcelBanner
      case 'qa':
        return QABanner
      case 'data':
        return DBBanner
      default:
        return ''
    }
  }, [knowledgeBaseType])

  return (
    <>
      <div
        className='bg-[#FAFAFA] p-[25px] rounded-xl bg-no-repeat bg-cover'
        style={{
          backgroundImage: `url(${banner})`,
        }}
      >
        <div className='flex items-stretch gap-6'>
          {/* 左侧知识库图片 */}
          <img
            src={
              knowledgeBaseImage ||
              match(knowledgeBaseType)
                .with('excel', () => TableBigIcon)
                .with('data', () => DatabaseBigIcon)
                .with('qa', () => QABigIcon)
                .with('file', () => MultimodalBigIcon)
                .otherwise(() => '')
            }
            alt='knowledge base'
            className=' h-37.5 max-h-full  object-cover'
          />

          {/* 中间信息区 */}
          <div className='flex-1 overflow-hidden flex flex-col justify-between gap-2'>
            <div className='items-center '>
              <div className='flex gap-[10px] items-center mb-[8px]'>
                <h1 className='text-2xl font-semibold overflow-hidden whitespace-nowrap  text-ellipsis  text-[#0C1F17]'>
                  {knowledgeBaseName}
                </h1>
                <EditIcon
                  className={cn(
                    is_admin === false
                      ? 'cursor-not-allowed opacity-40'
                      : 'cursor-pointer',
                  )}
                  onClick={
                    is_admin === false
                      ? undefined
                      : () => setNameEditing(!nameEditing)
                  }
                />
              </div>
              <div className='flex h-20px items-center gap-[8px] mb-[10px]'>
                {!!knowledgeBaseSize && (
                  <div className='px-[8px] rounded-[4px] text-[12px] leading-[20px]  h-full bg-[#F2F4F6] text-[#576275]'>
                    {convertBytes(knowledgeBaseSize)}
                  </div>
                )}
                {!!knowledgeBaseFileCount && (
                  <div className='px-[8px] rounded-[4px] text-[12px] leading-[20px]  h-full bg-[#F2F4F6] text-[#576275]'>
                    {knowledgeBaseFileCount}
                    {knowledgeBaseType === 'excel'
                      ? '个表格'
                      : knowledgeBaseType === 'data'
                        ? '个数据表'
                        : knowledgeBaseType === 'qa'
                          ? '个问题'
                          : '个文档'}
                  </div>
                )}
              </div>
              <div className='flex overflow-hidden gap-[10px] '>
                <p className='text-[14px] font-normal text-[#919497] leading-relaxed text-ellipsis line-clamp-3'>
                  {description || '知识库暂无描述~'}
                </p>
              </div>
            </div>
            <div className='grid grid-cols-[max-content_1fr] items-center gap-x-5 gap-y-[10px] font-medium'>
              <span className='text-xs text-[#919497] whitespace-nowrap'>
                {t('app.docs.detail.knowledgeBaseType')}
              </span>
              <span className='text-xs text-[#919497]'>
                {knowledgeBaseType === 'excel'
                  ? '表格'
                  : knowledgeBaseType === 'data'
                    ? '数据库'
                    : knowledgeBaseType === 'qa'
                      ? '问答对'
                      : '多模态'}
              </span>
              <span
                className='text-xs text-[#919497] whitespace-nowrap'
                style={{ lineHeight: 1 }}
              >
                {t('app.docs.detail.creationTime')}
              </span>
              <span
                className='text-xs text-[#919497]'
                style={{ lineHeight: 1 }}
              >
                {formatDate(createdAt)}
              </span>
            </div>
          </div>

          {/* 右侧：顶部权限按钮 + 底部上传/新建 */}
          <div className='flex flex-col justify-between relative items-end flex-shrink-0 self-stretch'>
            <div className='flex items-center gap-3'>
              {/* 传输记录按钮 - 仅在多模态知识库和表格知识库中显示 */}
              {(knowledgeBaseType === 'file' ||
                knowledgeBaseType === 'excel') && (
                <Button
                  type='text'
                  className='flex items-center gap-1 h-8 pl-1 pr-0 py-2 bg-transparent text-[#0C1F17] hover:text-[#0C99FF] group'
                  onClick={() => setUploadRecordsSidebarVisible(true)}
                >
                  <UploadRecordIcon className='w-4 h-4 mt-[2px] group-hover:hidden' />
                  <UploadRecordHoverIcon className='w-4 h-4 mt-[2px] hidden group-hover:block' />
                  <span className='text-sm font-medium '>
                    {t('app.docs.detail.transferRecords')}
                  </span>
                </Button>
              )}

              {isAdmin && (
                <Button
                  type='text'
                  className=' flex items-center gap-1 h-8 pl-1 pr-0 py-2 bg-transparent text-[#0C1F17] hover:text-[#0C99FF] group'
                  onClick={() => setPermissionDrawerVisible(true)}
                >
                  <PermissionIcon className='w-4 h-4 mt-[2px] group-hover:hidden' />
                  <PermissionHoverIcon className='w-4 h-4 mt-[2px] hidden group-hover:block' />
                  <span className='text-sm font-medium '>
                    {t('app.docs.detail.permission')}
                  </span>
                </Button>
              )}
            </div>
            {knowledgeBaseType === 'file' &&
            graph_info !== undefined &&
            hasGraphPermission ? (
              <MutimodalGraph
                is_admin={is_admin}
                knowledgebase_id={knowledgeBaseId}
                graph_info={graph_info}
                graph_updatable={graph_updatable}
                knowledge_status={knowledge_status}
              />
            ) : null}
          </div>
        </div>
      </div>

      {/* 权限管理侧边栏 */}
      {permissionData && (
        <PermissionSidebar
          open={permissionDrawerVisible}
          onClose={() => setPermissionDrawerVisible(false)}
          knowledgeBaseId={knowledgeBaseId}
          initialData={permissionData}
          onSuccess={handlePermissionSuccess}
        />
      )}

      {/* 上传记录侧边栏 */}
      <UploadRecordsSidebar
        forest_id={knowledgeBaseId}
        open={uploadRecordsSidebarVisible}
        onClose={() => setUploadRecordsSidebarVisible(false)}
        reloadAnalyzeFiles={() => {
          // 由于侧边栏只展示上传状态，这里可以留空或调用刷新
          refreshTable?.()
        }}
      />
      {/* 名称和描述更改弹窗 */}
      {nameEditing ? (
        <EditModal
          open={nameEditing}
          onCancel={() => setNameEditing(false)}
          defaultData={{ name: knowledgeBaseName, desc: description }}
          onOk={async (val) => {
            onChangeName?.(val.name)
            onChangeDesc?.(val.desc ?? '')
          }}
        />
      ) : null}
    </>
  )
}

type BaseInfo = { name: string; desc?: string }
const EditModal: FC<{
  open: boolean
  onCancel: () => void
  onOk: (data: BaseInfo) => Promise<void>
  defaultData: BaseInfo
}> = (props) => {
  const { onCancel, onOk, defaultData } = props
  const [form] = Form.useForm<BaseInfo>()
  const name = Form.useWatch('name', form)
  const { run: submit, loading } = useRequest(
    async () => {
      const data = await form.validateFields()
      await onOk(data)
      onCancel()
    },
    { manual: true },
  )
  return (
    <Modal
      title='编辑知识库'
      open
      onCancel={onCancel}
      onOk={submit}
      okButtonProps={{ loading, disabled: !name?.trim() }}
      cancelButtonProps={{ disabled: loading }}
      maskClosable={!loading}
      keyboard={!loading}
    >
      <Form
        requiredMark={false}
        form={form}
        layout='vertical'
        initialValues={defaultData}
      >
        <Form.Item
          label={
            <span className='text-base font-medium text-[#0c1f17]'>
              知识库名称
            </span>
          }
          name='name'
          rules={[{ required: true, message: '请输入知识库名称' }]}
        >
          <Input placeholder='请输入知识库名称' maxLength={50} />
        </Form.Item>
        <Form.Item
          label={
            <span className='text-base font-medium text-[#0c1f17]'>
              知识库描述
            </span>
          }
          name='desc'
        >
          <Input.TextArea
            placeholder='请输入知识库描述'
            rows={4}
            maxLength={200}
            showCount
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}
