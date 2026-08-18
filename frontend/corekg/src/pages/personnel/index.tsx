import { FC, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Breadcrumb, Button, Modal, Select, Skeleton, Typography } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { cn, useOpen } from '@/utils'
import { getBindCompanyKeyWithPermSet } from '@/api/perm'
import SeparatorIcon from '@/assets/separator.svg?react'
import { PersonnelTree } from '@/components/PersonnelTree'
import { useQuotaLimitModal } from '@/hooks/useQuotaLimitModal'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useVersion } from '@/utils/useVersion'
import { Department, Employee } from './common'
import { EmployeeDrawer } from './components/EmployeeDrawer'
import { EmployeeTable } from './components/EmployeeTable'
import { InviteMemberStepModal } from './components/InviteMemberStepModal'
import { PermModal } from './components/PermModal'
import EmployeeIcon from './images/employee.svg?react'
import styles from './styles.module.scss'
import { usePersonnelData } from './usePersonnelData'

const Personnel: FC = () => {
  const {
    loadData,
    data: { employee },
  } = usePersonnelData()
  useEffect(() => {
    if (!employee) {
      loadData('employee')
    }
  }, [employee, loadData])
  const navigate = useNavigate()
  if (!employee) return <Skeleton active className='p-4'></Skeleton>
  return (
    <div className='w-full h-full overflow-hidden flex flex-col'>
      <div className='border-b border-[#EFF1F4] pt-[14px] pl-5 pb-3'>
        <Breadcrumb
          className={styles.layoutHeader}
          separator={<SeparatorIcon />}
          items={[
            {
              title: (
                <span
                  className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
                  onClick={() => {
                    navigate(`/`)
                  }}
                >
                  问答
                </span>
              ),
            },
            {
              title: (
                <span className='cursor-pointer text-sm font-medium text-[#3C4149]'>
                  通讯录
                </span>
              ),
            },
          ]}
        />
      </div>
      <PersonnelInner className='flex-1 overflow-hidden' />
    </div>
  )
}

export default Personnel

const PersonnelInner: FC<Style> = (props) => {
  const { className, style } = props
  const { version: deployVersion } = useDeployConfig()
  const { version, refresh: refreshVersion } = useVersion()
  const { show: showQuotaLimitModal } = useQuotaLimitModal()
  const isQuotaLimited =
    version && version.employee.used >= version.employee.quota
  const {
    data: { department, employee },
    dispatchEmployeeAction,
  } = usePersonnelData()
  const [selectedIds, setSelectedIds] = useState<PersonnelTree['selectedIds']>(
    () => {
      const topDepartment = department!.find((item) => !item.parentId)
      return [
        {
          id: topDepartment?.id,
          type: 'department',
        },
      ]
    },
  )
  const [role, setRole] = useState<'all' | Employee['role']>('all')
  const currentDepartment = useMemo(() => {
    const departmentId = selectedIds?.[0]?.id
    const currentDepartment = department?.find(
      (item) => item.id === departmentId,
    )
    return currentDepartment
  }, [department, selectedIds])
  const dataSource = useMemo(() => {
    if (!currentDepartment) return
    return employee?.filter((item) => {
      if (role && role !== 'all' && role !== item.role) return false
      if (!currentDepartment?.parentId) return true
      return item.departmentIds?.includes(currentDepartment.id)
    })
  }, [currentDepartment, employee, role])

  const [employeeDrawerOpen, { toggle: togDrawer }, drawerKey] = useOpen()
  // 邀请成员步骤弹窗
  const [inviteStepOpen, { toggle: togInviteStep }, inviteStepKey] = useOpen()
  // 权限选择弹窗
  const [permOpen, { toggle: togPerm }, permKey] = useOpen()
  const [inviteUrl, setUrl] = useState<string>()
  // 存储邀请成员步骤中选择的数据
  const [inviteStepData, setInviteStepData] = useState<{
    departmentIds: Department['id'][]
    role: 'sys_admin' | 'sys_employee'
  }>()

  return (
    <>
      <div className={cn('overflow-hidden flex', className)} style={style}>
        <PersonnelTree
          className='p-2.5 w-70 h-full border-r border-[#EFF1F4]'
          placeholder='搜索'
          showDepartmentOperators
          selectable
          selectedIds={selectedIds}
          onSelect={(ids) => {
            if (!ids?.length) return
            setSelectedIds(ids)
          }}
        />
        <div className='flex-1 overflow-hidden bg-[#FAFAFA] p-4 flex flex-col  gap-4'>
          <span className='flex items-center gap-2'>
            <span className='font-medium text-base text-black'>
              {currentDepartment?.name}
            </span>
            <span className='font-medium text-[#919497]'>
              当前{currentDepartment?.parentId ? '部门' : '组织'}有
              {dataSource?.length}人
              {currentDepartment?.parentId ? '（不包含子部门）' : null}
            </span>
          </span>
          <span className='flex items-center gap-3'>
            <span className='font-medium flex items-center gap-4 mr-auto'>
              角色
              <Select
                className='w-60'
                value={role}
                onChange={setRole}
                options={[
                  { label: '全部', value: 'all' },
                  { label: '管理员', value: 'sys_admin' },
                  {
                    label: '成员',
                    value: 'sys_employee',
                  },
                ]}
              />
            </span>
            {deployVersion !== 'custom' ? (
              <Button
                onClick={() => {
                  if (isQuotaLimited) {
                    showQuotaLimitModal({ type: 'member' })
                    return
                  }
                  togInviteStep()
                }}
                icon={<EmployeeIcon />}
                className={cn('hover:text-[#0C99FF]', {
                  'opacity-50': isQuotaLimited,
                })}
              >
                邀请成员
              </Button>
            ) : null}
            <Button
              onClick={() => {
                if (isQuotaLimited) {
                  showQuotaLimitModal({ type: 'member' })
                  return
                }
                togDrawer()
              }}
              icon={<PlusOutlined />}
              className={cn('text-[#0C99FF] border-[#0C99FF]', {
                'opacity-50': isQuotaLimited,
              })}
            >
              创建成员
            </Button>
          </span>
          <EmployeeTable dataSource={dataSource} className='flex-1' />
        </div>
      </div>
      {/* 邀请成员步骤弹窗 */}
      <InviteMemberStepModal
        key={inviteStepKey}
        open={inviteStepOpen}
        onCancel={() => {
          setInviteStepData(undefined)
          togInviteStep()
        }}
        onNext={(data) => {
          setInviteStepData(data)
          togInviteStep()
          togPerm()
        }}
        initialValues={inviteStepData}
      />
      {/* 权限选择弹窗 */}
      <PermModal
        key={permKey}
        open={permOpen}
        title='邀请成员'
        cancelText='上一步'
        onCancel={() => {
          togPerm()
          togInviteStep()
        }}
        onSuccess={() => {
          // 提交成功后只关闭权限选择弹窗，不打开第一个弹窗
          togPerm()
        }}
        onSubmit={async (val) => {
          if (!inviteStepData) return
          const { key } = await getBindCompanyKeyWithPermSet({
            perm_set: {
              chatPs: val.chatPs.map((item) => {
                const {
                  use_perm,
                  manage_perm,
                  agent: { ID },
                } = item
                return {
                  use_perm,
                  manage_perm,
                  agent: { ID },
                  act_option: 'update',
                }
              }),
              forestPs: val.forestPs.map((item) => {
                const {
                  use_perm,
                  manage_perm,
                  forest: { ID },
                } = item
                return {
                  use_perm,
                  manage_perm,
                  forest: { ID },
                  act_option: 'update',
                }
              }),
            },
            count: 1,
            invitation_role: inviteStepData.role,
            department_ids: inviteStepData.departmentIds.map((id) =>
              Number(id),
            ),
            issuer: 'yygu',
          })
          // 刷新版本信息
          refreshVersion()
          const url = `${location.origin}/invite?key=${encodeURIComponent(key)}`
          setUrl(url)
          setInviteStepData(undefined)
        }}
      />
      <InviteUrlModal url={inviteUrl} onClose={() => setUrl('')} />
      <EmployeeDrawer
        key={drawerKey}
        title='添加成员'
        open={employeeDrawerOpen}
        onClose={togDrawer}
        defaultValue={{
          departmentIds: selectedIds?.slice(0, 1).map((item) => item.id),
        }}
        onSubmit={async (val) => {
          return dispatchEmployeeAction({
            type: 'add',
            version: deployVersion,
            ...val,
          })
        }}
      />
    </>
  )
}

const InviteUrlModal: FC<{ url?: string; onClose: () => void }> = (props) => {
  const { url, onClose } = props
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  return (
    <Modal
      open={Boolean(url)}
      onCancel={onClose}
      title={t('settings.copyLink')}
      maskClosable={false}
      keyboard={false}
      width={'50vw'}
      footer={
        <div className='flex gap-2 justify-end'>
          <Button onClick={() => onClose()}>{tC('button.cancel')}</Button>
          <Typography.Text
            copyable={{
              icon: [
                <Button type='primary'>{t('settings.copyLink')}</Button>,
                <Button type='primary'>{t('settings.copyLink')}</Button>,
              ],
              text: url,
            }}
          ></Typography.Text>
        </div>
      }
    >
      <div className='flex flex-col'>
        <span className='text-[#4E5969] text-base'>
          {t('settings.invitationLinkGeneratedCopySend')}
        </span>
        <div className='bg-[#F0F2F7] text-[#0E42D2] text-base rounded p-6'>
          {url}
        </div>
      </div>
    </Modal>
  )
}
