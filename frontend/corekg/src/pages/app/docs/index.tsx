import { useState, useEffect, useRef, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, message, Modal, Skeleton, Input, Dropdown, Select } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useBoolean } from 'ahooks'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
import { match } from 'ts-pattern'
import { cn, formatFileSize } from '@/utils'
import {
  getKnowledgeBaseList,
  deleteKnowledgeBase,
  createKnowledgeBase,
  updateForestWithPerm,
  updateKnowledgeBaseName,
  syncKnowledgeBaseToCoze,
} from '@/api/knowledge'
import ApiKeyWarning2 from '@/assets/icons/apiKey-warning2.svg'
import DatabaseBigIcon from '@/assets/icons/docs/database-big.svg'
import DatabaseIcon from '@/assets/icons/docs/database-create.svg'
import DeleteIcon from '@/assets/icons/docs/delete-file.svg'
import EditIcon from '@/assets/icons/docs/edit-file.svg'
import MultimodalBigIcon from '@/assets/icons/docs/multimodal-big.svg'
import KnowledgeBaseIcon from '@/assets/icons/docs/multimodal.svg'
import MultimodalIcon from '@/assets/icons/docs/multimodal.svg'
import QABigIcon from '@/assets/icons/docs/qa-big.svg'
import QAIcon from '@/assets/icons/docs/qa-create.svg'
import SyncIcon from '@/assets/icons/docs/syncCoze.svg'
import TableBigIcon from '@/assets/icons/docs/table-big.svg'
import TableIcon from '@/assets/icons/docs/table.svg'
import { ItemCard } from '@/components/ItemCard'
import { scroll } from '@/styles/scroll'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useLoginGlobalData } from '@/utils/useLoginGlobalData'
import { useVersion } from '@/utils/useVersion'
import bgImg from '../../../assets/bg.png'
import KnowledgeBaseModal from './components/KnowledgeBaseModal'
import AddIcon from './images/add.svg?react'
import EmptyIcon from './images/empty.svg?react'
import GraphSuccess from './images/graph-success.svg'
import MoreIcon from './images/more.svg?react'
import MutimodalIcon from './images/mutimodalIcon.png'
import SearchIcon from './images/search.svg?react'
import styles from './styles.module.scss'

interface KnowledgeBase {
  ID: number
  name: string
  description: string
  is_admin: boolean
  forest_type: string
  file_count: number
  character_count?: number
  app_count?: number
  disk_storage?: string
  total_size?: number
  data_source_type: string
  data_source_subtype: string
  CreatedAt?: string | number
  UpdatedAt?: string | number
  manager_ids?: number[]
  public_scope?: string
  scope_ids?: number[]
  is_synced?: boolean // 是否已同步至Coze
  graph_status?: string
}

export default function Home() {
  const { version } = useDeployConfig()
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const { t: tM } = useTranslation('messages')

  const [loading, { toggle }] = useBoolean()
  const [knowledgeBases, setKnowledgeBases] = useState<KnowledgeBase[]>([])
  const [showTypeDropdown, setShowTypeDropdown] = useState(false)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)
  // 新建草稿态：选择类型后，先进入命名编辑，再创建
  const [draftType, setDraftType] = useState<(typeof KB_TYPES)[0] | null>(null)
  const [draftName, setDraftName] = useState('')
  const [creating, setCreating] = useState(false)

  const { refresh } = useVersion()
  const navigate = useNavigate()

  // 清理所有知识库详情页的表格存储状态
  const clearKnowledgeDetailStorage = () => {
    try {
      const TABLE_STORAGE_KEY = 'ai-yygu-table-storage'
      const localStorageData = localStorage.getItem(TABLE_STORAGE_KEY)
      if (!localStorageData) return

      const data = JSON.parse(localStorageData)
      const updatedData = { ...data }

      // 清除所有知识库详情页相关的存储键
      Object.keys(updatedData).forEach((key) => {
        if (key === 'knowledgeDetailTable' || key.startsWith('folderDetail-')) {
          delete updatedData[key]
        }
      })

      localStorage.setItem(TABLE_STORAGE_KEY, JSON.stringify(updatedData))
    } catch (error) {
      console.error('清理知识库详情页存储失败', error)
    }
  }

  // 知识库类型配置
  const KB_TYPES = [
    { key: 'file', label: t('app.docs.multiModal'), icon: MultimodalIcon },
    { key: 'excel', label: t('app.docs.tableData'), icon: TableIcon },
  ]
  if (version !== 'international') {
    KB_TYPES.push(
      { key: 'db', label: t('app.docs.database'), icon: DatabaseIcon },
      { key: 'qa', label: t('app.docs.qaPair'), icon: QAIcon },
    )
  }

  // 获取知识库列表
  const fetchKnowledgeBaseList = async () => {
    toggle()
    try {
      const res = await getKnowledgeBaseList({ offset: 0, limit: 9999 })
      if (res.Data) {
        // 按创建时间逆序排列（新的在前）
        const sortedData = res.Data.sort(
          (a: KnowledgeBase, b: KnowledgeBase) => {
            // 当created_at字段可用时，使用创建时间排序
            if (a.CreatedAt && b.CreatedAt) {
              return (
                new Date(b.CreatedAt).getTime() -
                new Date(a.CreatedAt).getTime()
              )
            }
            // 暂时用ID逆序代替（ID越大通常越新）
            return b.ID - a.ID
          },
        )
        setKnowledgeBases(sortedData)
      } else {
        setKnowledgeBases([])
      }
    } finally {
      toggle()
    }
  }

  // 选择类型后，进入草稿命名态
  const handleQuickCreate = (type: (typeof KB_TYPES)[0]) => {
    setShowTypeDropdown(false)
    setDraftType(type)
    setDraftName(type.label)
  }

  // 提交创建（回车或失焦）
  const submitCreate = async () => {
    if (!draftType) return
    const name = (draftName || '').trim()
    if (!name) {
      message.warning(t('app.docs.kbNameRequired'))
      return
    }
    if (creating) return
    setCreating(true)
    try {
      const params = {
        name,
        description: '',
        forest_type:
          draftType.key === 'excel' || draftType.key === 'db'
            ? 'data'
            : draftType.key,
        public_scope: 'company' as const,
        data_source_type: (draftType.key === 'excel'
          ? 'excel'
          : draftType.key === 'db'
            ? 'db'
            : 'standard') as 'standard' | 'excel' | 'db',
        data_source_subtype: (draftType.key === 'excel'
          ? 'excel'
          : draftType.key === 'db'
            ? 'mysql'
            : 'standard') as 'standard' | 'excel' | 'mysql',
      }

      const { forest_id } = await createKnowledgeBase(params)
      message.success(tM('kbCreationSuccess'))
      setDraftType(null)
      setDraftName('')
      setCreating(false)

      // 创建成功后，根据类型跳转详情页
      if (draftType.key === 'qa') {
        navigate(`/docs/qa/${forest_id}`, { state: { is_admin: true } })
        return
      }
      if (draftType.key === 'db') {
        navigate(`/docs/db/${forest_id}`)
        return
      }
      if (draftType.key === 'excel') {
        // 与多模态一致，进入通用详情页
        navigate(`/docs/detail/${forest_id}`)
        return
      }
      navigate(`/docs/detail/${forest_id}`)
    } catch (error: any) {
      // 名称重复或其它错误：保持编辑态，提示错误
      // const msg = typeof error === 'string' ? error : error?.message
      // if (msg) message.error(msg)
    } finally {
      setCreating(false)
    }
  }

  // 删除知识库
  const handleDelete = (ID: number) => {
    Modal.confirm({
      icon: null,
      centered: true,
      className: 'delete-api-key-modal !w-[30%]',
      content: (
        <div className='relative'>
          <div className='flex justify-between'>
            <div className='text-[22px] font-[500] mb-2 text-[#000000E5]'>
              {tC('button.confirmDelete')}
            </div>
            {/* <img src={ApiKeyWarning2} className='w-[26px] h-[26px]' alt='' /> */}
          </div>
          <div className='h-[0.5px] w-[calc(100%+60px)] bg-[#C9CDD4] mt-4 -mx-6'></div>
          <div className='mt-6 text-base text-[#616373] mb-22 font-medium fontFamily-pingFangSC'>
            {tM('kbDeleteNoRecoverWarn')}
          </div>
        </div>
      ),
      okText: tC('button.confirmDelete'),
      okButtonProps: {
        className:
          'bg-[#0C99FF] hover:bg-[#0C99FF] !w-26 !h-[42px] !rounded-lg !text-base !px-4 !py-1',
        danger: false,
      },
      cancelButtonProps: {
        className:
          '!bg-[#F4F9FF] text-[#616373] !w-22 !h-[42px] !rounded !text-base !border-none !px-4 !py-1 !font-medium',
      },
      cancelText: tC('button.cancel'),
      maskClosable: true,
      onOk: async () => {
        try {
          await deleteKnowledgeBase({ id: Number(ID) })
          message.success(tM('kbDeletionSuccess'))
          fetchKnowledgeBaseList()
          refresh()
        } catch (error) {
          console.log(error)
        }
      },
    })
  }

  // 获取知识库同步类型
  const getKnowledgeBaseSyncType = (item: KnowledgeBase): string => {
    // 如果是data类型，根据data_source_type返回具体类型
    if (item.forest_type === 'data') {
      return item.data_source_type // 'excel' 或 'db'
    }
    // 其他类型直接返回forest_type
    return item.forest_type // 'file', 'qa', 'cad'
  }

  const handleSyncToCoze = async (ID: number) => {
    try {
      const knowledgeBase = knowledgeBases.find((item) => item.ID === ID)
      if (!knowledgeBase) return

      // 同步coze接口调用
      const response = await syncKnowledgeBaseToCoze({
        detail_id: ID,
        name: knowledgeBase.name || '',
        forest_type: getKnowledgeBaseSyncType(knowledgeBase),
      })
      if (response.code === 0) {
        message.success(tM('syncToCozeSuccess'))
        fetchKnowledgeBaseList()
        refresh()
      } else {
        message.error(response.message || tM('syncCozeFailRetryLater'))
      }
    } catch (error) {
      message.error(tM('syncCozeFailRetryLater'))
    }
  }

  // 点击外部关闭下拉菜单
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setShowTypeDropdown(false)
      }
    }
    if (showTypeDropdown) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [showTypeDropdown])
  // const handleSyncToCoze = async (ID: number) => {
  //   // try {
  //   const knowledgeBase = knowledgeBases.find((item) => item.ID === ID)
  //   if (!knowledgeBase) return

  //   // 同步coze接口调用
  //   const response = await syncKnowledgeBaseToCoze({
  //     detail_id: ID,
  //     name: knowledgeBase.name || '',
  //     forest_type: getKnowledgeBaseSyncType(knowledgeBase),
  //   })
  //   if (response.code === 0) {
  //     message.success('同步至coze成功')
  //     refresh()
  //   }
  //   // } catch (error) {
  //   //   message.error('同步coze失败，请稍后重试')
  //   // }
  // }
  const [sortKey, setSortKey] = useState<string>('CreatedAt')
  const [filterKey, setFilterKey] = useState<string>('all')
  const [graphStatus, setGraphStatus] = useState<string>('all')
  const [inputValue, setInputValue] = useState<string>('')
  const [searchValue, setSearchValue] = useState<string>('')

  const filteredKnowledgeBases = useMemo<KnowledgeBase[]>(() => {
    if (!knowledgeBases?.length) return [] as KnowledgeBase[]
    let filtered = knowledgeBases
      .filter((item) => {
        if (['db', 'excel'].includes(filterKey)) {
          return (
            item.forest_type === 'data' && item.data_source_type === filterKey
          )
        }

        return filterKey === 'all' || item.forest_type === filterKey
      })
      .filter((item) => {
        return (
          graphStatus === 'all' ||
          ['success', 'updatable'].includes(item.graph_status!)
        )
      })
    if (searchValue) {
      filtered = filtered.filter((item) => {
        return item.name.toLocaleLowerCase().includes(searchValue?.trim?.())
      })
    }

    if (!filtered.length) return [] as KnowledgeBase[]
    return filtered.sort((a, b) => {
      if (sortKey === 'CreatedAt' && a.CreatedAt && b.CreatedAt) {
        return new Date(b.CreatedAt).getTime() - new Date(a.CreatedAt).getTime()
      }
      if (sortKey === 'UpdatedAt' && a.UpdatedAt && b.UpdatedAt) {
        return new Date(b.UpdatedAt).getTime() - new Date(a.UpdatedAt).getTime()
      }
      // 要是没有这些参数的默认方案
      return b.ID - a.ID
    })
  }, [knowledgeBases, searchValue, filterKey, graphStatus, sortKey])

  useEffect(() => {
    fetchKnowledgeBaseList()
    clearKnowledgeDetailStorage()
  }, [])

  const renderEmpty = () => {
    return (
      <div className='flex items-center flex-col justify-center'>
        <EmptyIcon />
        暂未找到相关知识库
      </div>
    )
  }

  return (
    <div className='w-full h-full flex flex-col'>
      {/* 顶部导航部分 - 面包屑 */}
      <div className='w-full h-12 bg-white flex items-center px-5 pt-[2px]'>
        <div className='flex items-center gap-2 text-sm'>
          <img src={KnowledgeBaseIcon} className='w-4 h-4' alt='' />
          <span className='text-[#2D2D2D] cursor-default font-medium'>
            {t('app.docs.knowledgeBase')}
          </span>
        </div>
      </div>
      {/* 顶部欢迎区域 */}
      <div className={cn('rounded-2xl ', 'mx-12 mt-[10px] relative')}>
        <img src={bgImg} className='w-full' />
        <div className='text-[#2A4C95] z-10 absolute ml-12 top-1/2 -translate-y-1/2'>
          <h1 className='text-[40px] font-semibold '>
            {t('app.docs.welcomeTitle')}
          </h1>
          <p className='text-base font-medium mt-2.5'>
            {t('app.docs.welcomeDesc')}
          </p>
        </div>
      </div>

      {/* 知识库卡片列表区域 - 可滚动 */}
      <div
        className={cn(
          'flex-1 px-25 pb-12 overflow-auto bg-[#ffffff] pt-8',
          scroll,
        )}
      >
        {loading ? (
          <Skeleton active paragraph={{ rows: 10 }} />
        ) : (
          <div className='w-full flex justify-center'>
            <div
              className='w-full  grid gap-x-10 gap-y-8 justify-center'
              style={{
                gridTemplateColumns: 'repeat(auto-fill, 300px)',
              }}
            >
              {/* 卡片列表上方按钮 为对齐置于此处 */}
              <div
                className='flex items-center whitespace-nowrap'
                style={{ gridColumn: '1/-1' }}
              >
                <div className='mr-20 flex gap-[16px] items-center'>
                  <div className='flex gap-[6px] items-center'>
                    <div className='font-[500] text-[14px] text-[#919497]'>
                      排序方式
                    </div>
                    <Select
                      defaultValue={sortKey}
                      style={{ width: 114 }}
                      popupMatchSelectWidth={false}
                      classNames={{
                        popup: {
                          root: styles.filterSelect,
                        },
                      }}
                      onChange={setSortKey}
                      options={[
                        { value: 'CreatedAt', label: '按最新创建' },
                        { value: 'UpdatedAt', label: '按最近更新' },
                      ]}
                    />
                  </div>
                  <div className='flex gap-[6px] items-center'>
                    <div className='font-[500] text-[14px] text-[#919497]'>
                      知识库类型
                    </div>
                    <Select
                      defaultValue={filterKey}
                      style={{ width: 114 }}
                      classNames={{
                        popup: {
                          root: styles.filterSelect,
                        },
                      }}
                      onChange={setFilterKey}
                      popupMatchSelectWidth={false}
                      options={[
                        { value: 'all', label: '全部' },
                        { value: 'file', label: '多模态' },
                        { value: 'excel', label: '表格' },
                        { value: 'db', label: '数据库' },
                        { value: 'qa', label: '问答对' },
                      ]}
                    />
                  </div>
                  <div className='flex gap-[6px] items-center'>
                    <div className='font-[500] text-[14px] text-[#919497]'>
                      图谱状态
                    </div>
                    <Select
                      defaultValue={graphStatus}
                      style={{ width: 114 }}
                      popupMatchSelectWidth={false}
                      classNames={{
                        popup: {
                          root: styles.filterSelect,
                        },
                      }}
                      onChange={setGraphStatus}
                      options={[
                        { value: 'all', label: '全部' },
                        { value: 'success', label: '构建成功' },
                      ]}
                    />
                  </div>
                </div>
                <div className='ml-auto flex justify-end'>
                  <div
                    className='relative flex items-center gap-[12px]'
                    ref={dropdownRef}
                  >
                    <Input
                      value={inputValue}
                      placeholder={tC('button.search')}
                      prefix={<SearchIcon />}
                      onChange={(e) => setInputValue(e.target.value)}
                      onBlur={() =>
                        !inputValue?.trim?.() && setSearchValue(inputValue)
                      }
                      onPressEnter={() => setSearchValue(inputValue)}
                      className={`w-[70px] h-[30px] border-[#0C99FF] shadow-none  ${styles.searchInputWrap} ${inputValue?.trim?.() ? styles.searchInputWrapSearching : ''}`}
                    />
                    <Button
                      className={`text-sm font-medium rounded-[6px] border border-[#0C99FF] hover:border-[#0C99FF] text-[#0C99FF] active:text-[#0C99FF] shadow-none ${styles.createBtn}`}
                      onClick={() => setShowCreateModal(true)}
                    >
                      <AddIcon className={styles.createBtnIcon} />{' '}
                      {t('app.docs.newKnowledgeBase')}
                    </Button>
                    {showTypeDropdown && (
                      <div className='absolute top-[calc(100%)] right-0 mt-2 w-[230px] p-2.5 bg-white rounded-lg shadow-lg border border-gray-200 z-10'>
                        {KB_TYPES.map((type, index) => (
                          <div
                            key={type.key}
                            onClick={() => handleQuickCreate(type)}
                            className='px-2.5 py-2 hover:bg-[#F7F7F7] cursor-pointer flex font-medium items-center gap-2 transition-colors rounded'
                          >
                            <img
                              src={type.icon}
                              className='w-4 h-4 flex-shrink-0'
                              alt=''
                            />
                            <span
                              className='text-sm text-[#2D2D2D] font-medium truncate flex-1 min-w-0'
                              title={String(type.label)}
                            >
                              {type.label}
                            </span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>
              {!filteredKnowledgeBases.length ? (
                <div
                  className='w-full h-60 flex items-center justify-center text-[#919497]'
                  style={{ gridColumn: '1/-1' }}
                >
                  {knowledgeBases?.length
                    ? renderEmpty()
                    : t('app.docs.noKbDataCreateNew')}
                </div>
              ) : null}
              {draftType && (
                <div
                  className='relative group transition-transform duration-200 bg-white rounded-lg shadow-md border border-gray-200'
                  style={{ width: '359px', height: '245px' }}
                >
                  <div className='p-4 h-full flex flex-col rounded-lg'>
                    <div className='h-[170px] rounded-lg overflow-hidden relative flex items-center justify-center bg-[#F7F7F7]'>
                      {draftType.key === 'file' && (
                        <img
                          src={MultimodalBigIcon}
                          className='w-full h-full object-cover opacity-60'
                          alt=''
                        />
                      )}
                      {draftType.key === 'qa' && (
                        <img
                          src={QABigIcon}
                          className='w-full h-full object-cover opacity-60'
                          alt=''
                        />
                      )}
                      {draftType.key === 'excel' && (
                        <img
                          src={TableBigIcon}
                          className='w-full h-full object-cover opacity-60'
                          alt=''
                        />
                      )}
                      {draftType.key === 'db' && (
                        <img
                          src={DatabaseBigIcon}
                          className='w-full h-full object-cover opacity-60'
                          alt=''
                        />
                      )}
                    </div>
                    <div className='mt-3'>
                      <Input
                        autoFocus
                        value={draftName}
                        onChange={(e) => setDraftName(e.target.value)}
                        onPressEnter={submitCreate}
                        onBlur={submitCreate}
                        disabled={creating}
                        placeholder={t('app.docs.inputContent', {
                          target: t('app.docs.knowledgeBase'),
                        })}
                        className='h-8'
                      />
                    </div>
                  </div>
                </div>
              )}
              {filteredKnowledgeBases?.map((item) => (
                <KnowledgeBaseCard
                  key={item.ID}
                  item={item}
                  onDelete={handleDelete}
                  onSuccess={fetchKnowledgeBaseList}
                  onSyncToCoze={handleSyncToCoze}
                />
              ))}
            </div>
          </div>
        )}
      </div>

      {/* 新建知识库弹窗*/}
      <KnowledgeBaseModal
        open={showCreateModal}
        onCancel={() => setShowCreateModal(false)}
      />
    </div>
  )
}

function KnowledgeBaseCard({
  item,
  onDelete,
  onSuccess,
  onSyncToCoze,
}: {
  item: KnowledgeBase
  onDelete: (ID: number) => void
  onSuccess: () => void
  onSyncToCoze: (ID: number) => void
}) {
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const [showMenu, setShowMenu] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [editName, setEditName] = useState(item.name)
  const navigate = useNavigate()
  const { license } = useLoginGlobalData()
  const { version } = useDeployConfig()
  const handleCardClick = () => {
    if (isEditing) return

    // 根据类型跳转到对应的详情页面
    if (item.data_source_type === 'db') {
      navigate(`/docs/db/${item.ID}`)
      return
    }

    if (item.forest_type === 'qa') {
      navigate(`/docs/qa/${item.ID}`, {
        state: { is_admin: item.is_admin },
      })
      return
    }

    if (item.forest_type === 'cad') {
      navigate(`/docs/cad/${item.ID}`)
      return
    }

    navigate(`/docs/detail/${item.ID}`)
  }

  const handleEdit = () => {
    setIsEditing(true)
    setShowMenu(false)
  }

  const handleSaveEdit = async () => {
    if (editName === item.name) {
      setIsEditing(false)
      return
    }

    try {
      await updateKnowledgeBaseName({
        id: item.ID,
        name: editName,
      })

      message.success(tM('kbUpdateSuccess'))
      setIsEditing(false)
      onSuccess()
    } catch (error) {
      // message.error(tM('kbUpdateFailed'))
    }
  }

  const formatDate = (timestamp: string | number | undefined) => {
    if (!timestamp) return '-'
    return dayjs(timestamp).format('YYYY-MM-DD HH:mm')
  }

  // 菜单项配置（无权限也展示，但禁用操作项）
  const isReadOnly = !item.is_admin
  const isSynced = Boolean(item.is_synced)
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
          title={t('app.docs.rename')}
        >
          {t('app.docs.rename')}
        </span>
      ),
      icon: (
        <img
          src={EditIcon}
          className={`w-4 h-4 ${isReadOnly ? 'opacity-40' : ''}`}
          alt=''
        />
      ),
      onClick: isReadOnly ? undefined : handleEdit,
      className: `${sharedItemClass} ${isReadOnly ? disabledItemClass : ''}`,
      disabled: isReadOnly,
    },
    {
      key: 'delete',
      label: (
        <span
          className='flex-1 min-w-0 truncate font-medium'
          title={t('app.docs.deleteKnowledgeBase')}
        >
          {t('app.docs.deleteKnowledgeBase')}
        </span>
      ),
      icon: (
        <img
          src={DeleteIcon}
          className={`w-4 h-4 ${isReadOnly ? 'opacity-40' : ''}`}
          alt=''
        />
      ),
      onClick: isReadOnly ? undefined : () => onDelete(item.ID),
      className: `${sharedItemClass} ${isReadOnly ? disabledItemClass : ''}`,
      disabled: isReadOnly,
    },
  ]
  // 数据库类型和问答对类型的知识库不显示"同步至Coze"
  const shouldShowSyncToCoze =
    version !== 'international' &&
    item.data_source_type !== 'db' &&
    item.forest_type !== 'qa'
  // if (shouldShowSyncToCoze) {
  //   menuItems.unshift({
  //     key: 'syncToCoze',
  //     label: (
  //       <span
  //         className='flex-1 min-w-0 truncate font-medium'
  //         title={t('app.docs.syncToCoze')}
  //       >
  //         {t('app.docs.syncToCoze')}
  //       </span>
  //     ),
  //     icon: (
  //       <img
  //         src={SyncIcon}
  //         className={`w-4 h-4 ${isReadOnly || isSynced ? 'opacity-40' : ''}`}
  //         alt=''
  //       />
  //     ),
  //     onClick: isReadOnly || isSynced ? undefined : () => onSyncToCoze(item.ID),
  //     className: `${sharedItemClass} ${
  //       isReadOnly || isSynced ? disabledItemClass : ''
  //     }`,
  //     disabled: isReadOnly || isSynced,
  //   })
  // }
  const avatar = match({
    forest_type: item.forest_type,
    data_source_type: item.data_source_type,
  })
    .with({ forest_type: 'file' }, () => MutimodalIcon)
    .with({ forest_type: 'qa' }, () => QABigIcon)
    .with({ data_source_type: 'excel' }, () => TableBigIcon)
    .with({ data_source_type: 'db' }, () => DatabaseBigIcon)
    .otherwise(() => MultimodalBigIcon)
  return (
    <ItemCard
      onClick={handleCardClick}
      avatar={avatar}
      operators={{
        items: menuItems,
        onClick: (e) => e.domEvent.stopPropagation(),
        className: 'px-[10px] py-[10px] min-w-[190px] max-w-[230px] bg-white',
      }}
      title={
        isEditing ? (
          <Input
            value={editName}
            onChange={(e) => setEditName(e.target.value)}
            onPressEnter={handleSaveEdit}
            onBlur={handleSaveEdit}
            onClick={(e) => e.stopPropagation()}
            maxLength={50}
            autoFocus
            className='mb-2 !shadow-none focus:!border-[rgb(204,93,232)] hover:!border-[rgb(204,93,232)]'
          />
        ) : (
          item.name
        )
      }
      desc={item.description || '知识库暂无描述~'}
      extra={
        match(item.forest_type)
          .with('qa', () => '')
          .otherwise(() => {
            return `${item.file_count} 文件数 | `
          }) +
        dayjs(item.UpdatedAt).format('YYYY.MM.DD') +
        '更新'
      }
    >
      {['success', 'updatable'].includes(item.graph_status!) &&
      (!license || license.modules.includes('graph')) ? (
        <img src={GraphSuccess} className=' absolute z-10 -top-6 -left-4' />
      ) : null}
    </ItemCard>
  )
}
