import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useLocation } from 'react-router-dom'
import { Button, message } from 'antd'
import { useTranslation } from 'react-i18next'
import {
  listQAPair,
  deleteQAPair,
  updateKnowledgeBaseName,
  updateResourceEnable,
  updateKnowledgeBaseDesc,
} from '@/api/knowledge'
import { getKnowledgeBaseDetail } from '@/api/knowledge'
import PermissionIcon from '@/assets/icons/docs/permission-icon.svg?react'
import PermissionSidebar from '@/pages/app/docs/detail/components/PermissionSidebar'
import { KnowledgeBaseType } from '../detail/components/ActionButtons/UploadButton/Uploader'
import KnowledgeBaseInfo from '../detail/components/KnowledgeBaseInfo'
import ActionButtonGroup from './components/ActionButtonGroup'
import Nav from './components/Nav'
import QuestionsList from './components/QuestionsList'
import QuestionsModal from './components/QuestionsModal'
import SearchInput from './components/SearchInput'

interface QAPairAPIResponse {
  main: {
    id: string
    qa_question: string
    qa_answer: string
    created_at: string
    updated_at: string
    forest_id: number
    type: string
    uin: number
    company_id: number
    qa_answer_id: string
    source_from: string
    enable: number
  }
  child: Array<{
    id: string
    question: string
    created_at: string
    is_deleted: boolean
  }>
}

interface QAPair {
  id: string
  qa_question: string
  qa_answer: string
  created_at: string
  updated_at: string
  source_from: string
  enable: boolean
  sub_questions?: Array<{
    id: string
    question: string
    created_at: string
  }>
}

interface ModalState {
  visible: boolean
  mode: 'add' | 'edit'
  data?: QAPair
}

export default function QALibraryDetail() {
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const { id } = useParams<{ id: string }>()
  const forestId = Number(id)

  const [loading, setLoading] = useState(false)
  const [qaList, setQAList] = useState<QAPair[]>([])
  const [isAdmin, setIsAdmin] = useState<boolean>(false)
  const [adminLoading, setAdminLoading] = useState(true)
  const [total, setTotal] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [searchText, setSearchText] = useState('')
  const [modalState, setModalState] = useState<ModalState>({
    visible: false,
    mode: 'add',
  })

  // 权限侧边栏状态
  const [permissionDrawerVisible, setPermissionDrawerVisible] = useState(false)
  const [permissionData, setPermissionData] = useState<{
    manager_ids: number[]
    public_scope: 'company' | 'custom'
    scope_ids: number[]
    name: string
    description: string
  } | null>(null)

  // 启用状态切换loading状态（记录正在处理的问答对ID）
  const [enablingQaId, setEnablingQaId] = useState<string | null>(null)

  // 获取管理员权限
  const location = useLocation()

  const fetchAdminStatus = async () => {
    const routeState = location.state as { is_admin?: boolean } | null
    console.log(location, 'location')

    if (routeState?.is_admin !== undefined) {
      setIsAdmin(routeState.is_admin)
      setAdminLoading(false)
      return
    }

    // 如果没有路由状态，通过API获取
    try {
      setAdminLoading(true)
      const res = await getKnowledgeBaseDetail({ id: forestId })
      if (res?.data) {
        setIsAdmin(res.data.is_admin || false)
      }
    } catch (error) {
      console.error('获取知识库权限失败:', error)
      setIsAdmin(false)
    } finally {
      setAdminLoading(false)
    }
  }

  // 获取问答对列表
  const fetchQAList = async () => {
    if (!forestId) return

    setLoading(true)
    try {
      const params: any = {
        forest_id: forestId,
        limit: pageSize,
        offset: (currentPage - 1) * pageSize,
        orderBy: ['updated_at desc'],
      }

      if (searchText) {
        params.filters = [
          {
            field: 'qa_question',
            value: [searchText],
          },
        ]
      }

      const res = await listQAPair(params)

      if (res.data && Array.isArray(res.data)) {
        const transformedData: QAPair[] = res.data.map(
          (item: QAPairAPIResponse) => ({
            id: item.main.id,
            qa_question: item.main.qa_question,
            qa_answer: item.main.qa_answer,
            created_at: item.main.created_at,
            updated_at: item.main.updated_at,
            source_from: item.main.source_from,
            sub_questions:
              item.child
                ?.filter((child) => !child.is_deleted)
                .map((child) => ({
                  id: child.id,
                  question: child.question,
                  created_at: child.created_at,
                })) || [],
            enable: item.main.enable === 1,
          }),
        )

        console.log('Transformed Data:', transformedData)
        setQAList(transformedData)
        setTotal(res.total || 0)
      } else {
        setQAList([])
        setTotal(0)
      }
    } catch (error) {
      console.error('获取问答对列表失败:', error)
      message.error(tM('getQAPairListFail'))
      setQAList([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }

  const handlePageChange = (page: number, size?: number) => {
    setCurrentPage(page)
    if (size) setPageSize(size)
  }

  const handleSearch = (value: string) => {
    setSearchText(value)
    setCurrentPage(1)
  }

  const handleAdd = () => {
    setModalState({
      visible: true,
      mode: 'add',
    })
  }

  const handleEdit = (record: QAPair) => {
    setModalState({
      visible: true,
      mode: 'edit',
      data: record,
    })
  }

  const handleDelete = async (ids: string[]) => {
    try {
      await deleteQAPair({
        forest_id: forestId,
        question_ids: ids,
      })
      message.success(tM('deleteSuccess'))
      fetchQAList()
    } catch (error) {
      console.error('删除失败:', error)
      message.error(tM('deleteFailed'))
    }
  }

  // 关闭模态框
  const handleModalClose = () => {
    setModalState({
      visible: false,
      mode: 'add',
      data: undefined,
    })
  }

  // 操作成功回调
  const handleModalSuccess = () => {
    fetchQAList()
    handleModalClose()
  }

  // 统一类型到白名单
  const normalizeKnowledgeBaseType = (value: any): KnowledgeBaseType => {
    const allowed: KnowledgeBaseType[] = ['file', 'excel', 'qa', 'data']
    if (allowed.includes(value as KnowledgeBaseType))
      return value as KnowledgeBaseType
    return 'file'
  }

  const [knowledgeBaseName, setKnowledgeBaseName] = useState<string>('')
  const [knowledgeBaseDesc, setKnowledgeBaseDesc] = useState<string>('')
  const [knowledgeBaseFileCount, setKnowledgeBaseFileCount] =
    useState<number>(0)
  const [knowledgeBaseCreatedAt, setKnowledgeBaseCreatedAt] = useState<string>()
  const [knowledgeBaseType, setKnowledgeBaseType] =
    useState<KnowledgeBaseType>()

  const handleChangeDesc = async (newDesc: string) => {
    try {
      // if (!newDesc.trim()) {
      //   message.error(t('app.docs.detail.fileEdit.nameRequired'))
      //   return
      // }
      if (newDesc === knowledgeBaseDesc) {
        // message.success(t('app.docs.detail.fileEdit.renameSuccess'))
        return
      }
      if (forestId !== undefined && forestId !== null) {
        await updateKnowledgeBaseDesc({
          forest_id: forestId,
          description: newDesc,
        })
        setKnowledgeBaseDesc(newDesc)
        message.success(t('app.docs.detail.fileEdit.renameSuccess'))
      }
    } catch (error) {
      console.log(error)
    }
  }

  const handleChangeName = async (newName: string) => {
    try {
      if (!newName.trim()) {
        message.error(t('app.docs.kbNameRequired'))
        return
      }
      if (newName === knowledgeBaseName) {
        // message.success(t('app.docs.detail.fileEdit.renameSuccess'))
        return
      }
      if (forestId !== undefined && forestId !== null) {
        await updateKnowledgeBaseName({ id: forestId, name: newName })
        setKnowledgeBaseName(newName)
        message.success(t('app.docs.detail.fileEdit.renameSuccess'))
      }
    } catch (error) {
      console.log(error)
    }
  }

  // 获取知识库详情
  const fetchKnowledgeBaseInfo = useCallback(async () => {
    try {
      if (forestId) {
        const res = await getKnowledgeBaseDetail({ id: forestId })

        if (res && res.data) {
          setKnowledgeBaseName(res.data.name)
          setKnowledgeBaseDesc(res.data.description || '')
          setIsAdmin(res.data.is_admin)
          setKnowledgeBaseFileCount(res.data.file_count)
          // 设置创建时间和类型
          setKnowledgeBaseCreatedAt(res.data.CreatedAt)
          // 优先使用 data 类型下的具体子类型来判定（excel/db）
          const forestType = res.data.forest_type
          const dataSourceType = res.data.data_source_type
          let kbType: KnowledgeBaseType
          if (forestType === 'data') {
            kbType = dataSourceType === 'excel' ? 'excel' : 'data'
          } else {
            kbType = normalizeKnowledgeBaseType(forestType)
          }
          setKnowledgeBaseType(kbType)
        }
      }
    } catch (error) {
      console.error('获取知识库详情失败:', error)
      // 失败时保持空字符串，不显示错误的默认名称
      setKnowledgeBaseName('')
      setKnowledgeBaseDesc('')
    }
  }, [forestId])

  useEffect(() => {
    fetchKnowledgeBaseInfo()
  }, [])

  useEffect(() => {
    if (forestId) {
      fetchAdminStatus()
      fetchQAList()
    }
  }, [forestId, currentPage, pageSize, searchText])

  // 权限数据获取（打开抽屉时拉取）
  useEffect(() => {
    if (forestId && permissionDrawerVisible) {
      ;(async () => {
        try {
          const res = await getKnowledgeBaseDetail({ id: forestId })
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
      })()
    }
  }, [forestId, permissionDrawerVisible])

  const onEnableChange = async (enable: boolean, file: any) => {
    // 如果正在处理该问答对，直接返回
    if (enablingQaId === file.id) {
      return
    }

    // 设置loading状态，禁用开关
    setEnablingQaId(file.id)

    try {
      await updateResourceEnable({
        enable: enable ? 1 : -1,
        forest_id: Number(id),
        resource_ids: [String(file.id)],
        resource_type: 'qa_pair',
      })

      // 接口成功后再更新UI状态
      setQAList(
        qaList.map((item) =>
          item.id === file.id ? { ...item, enable } : item,
        ),
      )
    } catch (error) {
      console.log(error)
    } finally {
      // 清除loading状态
      setEnablingQaId(null)
    }
  }

  return (
    <div className='w-full h-full'>
      <div className='bg-white rounded-lg p-4 w-full h-full flex flex-col'>
        <Nav />
        <div className='flex flex-1 flex-col px-[48px] py-[10px] overflow-hidden gap-[20px]'>
          {/* 知识库信息区域 - 在所有层级显示 */}
          <KnowledgeBaseInfo
            knowledgeBaseName={knowledgeBaseName}
            knowledgeBaseId={forestId}
            knowledgeBaseFileCount={knowledgeBaseFileCount}
            createdAt={knowledgeBaseCreatedAt}
            isAdmin={isAdmin}
            knowledgeBaseType={knowledgeBaseType}
            disabled={!isAdmin}
            description={knowledgeBaseDesc}
            // onCreateFolder={createNewFolder}
            onChangeDesc={handleChangeDesc}
            onChangeName={handleChangeName}
          />

          <div className='flex justify-end items-center gap-[12px]'>
            <SearchInput onSearch={handleSearch} />
            <ActionButtonGroup
              forest_id={forestId}
              onAdd={handleAdd}
              isAdmin={isAdmin && !adminLoading}
              reloadData={fetchQAList}
            />
          </div>
          <div className='flex-1 overflow-hidden'>
            <QuestionsList
              loading={loading}
              dataSource={qaList}
              total={total}
              current={currentPage}
              pageSize={pageSize}
              onPageChange={handlePageChange}
              onEdit={handleEdit}
              onDelete={handleDelete}
              isAdmin={isAdmin}
              onEnableChange={onEnableChange}
              enablingQaId={enablingQaId}
            />
          </div>
        </div>

        <QuestionsModal
          open={modalState.visible}
          mode={modalState.mode}
          data={modalState.data}
          forestId={forestId}
          onCancel={handleModalClose}
          onSuccess={handleModalSuccess}
        />

        {/* 权限管理侧边栏 */}
        {permissionData && (
          <PermissionSidebar
            open={permissionDrawerVisible}
            onClose={() => setPermissionDrawerVisible(false)}
            knowledgeBaseId={forestId}
            initialData={permissionData}
            onSuccess={() => {
              // 关闭后重新拉取以便数据刷新
              setPermissionDrawerVisible(false)
            }}
          />
        )}
      </div>
    </div>
  )
}
