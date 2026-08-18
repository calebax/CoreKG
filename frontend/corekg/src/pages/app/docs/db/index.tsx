import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Input, message } from 'antd'
import { useCounter } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import {
  getKnowledgeBaseDetail,
  updateKnowledgeBaseName,
  updateKnowledgeBaseDesc,
} from '@/api/knowledge'
import PermissionIcon from '@/assets/icons/docs/permission-icon.svg?react'
import PermissionSidebar from '@/pages/app/docs/detail/components/PermissionSidebar'
import { KnowledgeBaseType } from '../detail/components/ActionButtons/UploadButton/Uploader'
import KnowledgeBaseInfo from '../detail/components/KnowledgeBaseInfo'
import BreadcrumbNav from './components/BreadcrumbNav'
import { ConfigBDBtn } from './components/ConfigBDBtn'
import { DBs, Headers, Tables } from './components/MySQL'
import SearchInput from './components/SearchInput'
import { useIds } from './hooks/useIds'

export default function FileExplorer() {
  const id_info = useIds()
  const { type, forest_id, forest_db_name, forest_table_name } = id_info
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const [refreshKey, { inc }] = useCounter()
  const { t } = useTranslation('pages')
  const [isAdmin, setIsAdmin] = useState<boolean>(false)
  const [adminLoading, setAdminLoading] = useState(true)
  const [permissionDrawerVisible, setPermissionDrawerVisible] = useState(false)
  const [permissionData, setPermissionData] = useState<{
    manager_ids: number[]
    public_scope: 'company' | 'custom'
    scope_ids: number[]
    name: string
    description: string
  } | null>(null)
  useEffect(() => {
    if (type === 'error') {
      navigate('/docs')
    }
  }, [navigate, type])

  // 获取管理员权限
  useEffect(() => {
    if (!forest_id) return
    ;(async () => {
      try {
        setAdminLoading(true)
        const res = await getKnowledgeBaseDetail({ id: forest_id })
        if (res?.data) {
          setIsAdmin(res.data.is_admin || false)
        }
      } catch (error) {
        setIsAdmin(false)
      } finally {
        setAdminLoading(false)
      }
    })()
  }, [forest_id])

  // 权限数据获取（打开抽屉时拉取）
  useEffect(() => {
    if (forest_id && permissionDrawerVisible) {
      ;(async () => {
        try {
          const res = await getKnowledgeBaseDetail({ id: forest_id })
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
          // noop
        }
      })()
    }
  }, [forest_id, permissionDrawerVisible])

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
  const [knowledgeBaseSize, setKnowledgeBaseSize] = useState<number>(0)
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
      if (forest_id !== undefined && forest_id !== null) {
        await updateKnowledgeBaseDesc({
          forest_id,
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
      if (forest_id !== undefined && forest_id !== null) {
        await updateKnowledgeBaseName({ id: forest_id, name: newName })
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
      if (forest_id) {
        const res = await getKnowledgeBaseDetail({ id: forest_id })

        if (res && res.data) {
          setKnowledgeBaseName(res.data.name)
          setKnowledgeBaseDesc(res.data.description || '')
          setIsAdmin(res.data.is_admin)
          setKnowledgeBaseFileCount(res.data.file_count)
          setKnowledgeBaseSize(res.data.total_size)

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
  }, [forest_id])

  useEffect(() => {
    fetchKnowledgeBaseInfo()
  }, [])

  if (id_info.type === 'error') return null
  return (
    <div className='w-full h-full bg-white rounded-lg p-4 w-full h-full flex flex-col'>
      <BreadcrumbNav id_info={id_info} />
      {/* 知识库信息区域 - 在所有层级显示 */}

      <div className='flex-1 flex flex-col px-[48px] py-[10px] overflow-hidden gap-[20px]'>
        <KnowledgeBaseInfo
          knowledgeBaseName={knowledgeBaseName}
          knowledgeBaseFileCount={knowledgeBaseFileCount}
          knowledgeBaseSize={knowledgeBaseSize}
          knowledgeBaseId={forest_id}
          createdAt={knowledgeBaseCreatedAt}
          isAdmin={isAdmin}
          knowledgeBaseType={knowledgeBaseType}
          disabled={!isAdmin}
          description={knowledgeBaseDesc}
          // onCreateFolder={createNewFolder}
          onChangeDesc={handleChangeDesc}
          onChangeName={handleChangeName}
        />
        <div className='flex gap-2.5 mr-4 justify-end'>
          <SearchInput onSearch={setSearch} />
          <ConfigBDBtn
            forest_id={forest_id}
            afterConfig={() => {
              inc()
              navigate(`/docs/db/${forest_id}`)
            }}
          />
        </div>
        <div className='flex-1 flex'>
          {(() => {
            console.log(id_info)

            switch (id_info.type) {
              case 'db':
                return (
                  <DBs key={refreshKey} forest_id={forest_id} search={search} />
                )
              case 'table':
                return (
                  <Tables
                    key={refreshKey}
                    forest_id={forest_id}
                    forest_db_name={forest_db_name}
                    search={search}
                  />
                )
              case 'header':
                return (
                  <Headers
                    key={refreshKey}
                    forest_id={forest_id}
                    forest_db_name={forest_db_name}
                    forest_table_name={forest_table_name}
                    search={search}
                  />
                )
            }
          })()}
        </div>
      </div>
    </div>
  )
}
