import { FC, memo, useMemo } from 'react'
import { Table, TableColumnType, Tooltip, Modal, message } from 'antd'
import { Employee, usePersonnelData } from 'Personnel'
import dayjs from 'dayjs'
import { cn, useOpen } from '@/utils'
import { resetPassword } from '@/api/account'
import { modifyChatPermSet, modifyForestPermSet } from '@/api/perm'
import useLocalStore from '@/stores/local'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { EmployeeDrawer } from '../EmployeeDrawer'
import { PermModal } from '../PermModal'
import EditIcon from './images/edit.svg?react'
import EmpolyeeIcon from './images/employee.svg?react'
import PermIcon from './images/perm.svg?react'
import ResetPasswordIcon from './images/reset-password.svg?react'
import styles from './styles.module.css'

export const EmployeeTable: FC<Style & { dataSource?: Employee[] }> = memo(
  (props) => {
    const { dataSource, className, style } = props
    // 按照创建时间降序排序（最新的在前）
    const sortedDataSource = useMemo(() => {
      if (!dataSource) return undefined
      return [...dataSource].sort((a, b) => {
        const timeA = dayjs(a.created_at).valueOf()
        const timeB = dayjs(b.created_at).valueOf()
        return timeB - timeA // 降序排序
      })
    }, [dataSource])
    const columns = useMemo(() => {
      const col: TableColumnType<Employee>[] = [
        {
          title: '用户名',
          render: (_, record) => (
            <span className='flex items-center'>
              <EmpolyeeIcon />
              {record.user_name}
            </span>
          ),
        },
        {
          title: '组织昵称',
          dataIndex: 'name',
        },
        {
          title: '所属组织角色',
          render: (_, record) => <EmployeeRole role={record.role} />,
        },
        {
          title: '手机号码',
          render: (_, record) => record.phone || '-',
        },
        {
          title: '邮箱',
          render: (_, record) => record.email || '-',
        },
        {
          title: '创建时间',
          render: (_, record) =>
            dayjs(record.created_at).format('YYYY-MM-DD HH:mm:ss'),
        },
        {
          title: '操作',
          render: (_, record) => <EmployeeOperator value={record} />,
        },
      ]
      return col
    }, [])
    return (
      <Table
        className={cn(className, styles.table)}
        style={style}
        dataSource={sortedDataSource}
        rowKey={'id'}
        columns={columns}
        pagination={{
          defaultPageSize: 10,
          showSizeChanger: true,
        }}
        scroll={{ x: 'max-content' }}
      ></Table>
    )
  },
)

const EmployeeRole: FC<Style & Pick<Employee, 'role'>> = (props) => {
  const { role, className, style } = props
  return (
    <span
      className={cn(
        'inline-flex items-center justify-center font-medium px-4 py-0.5 rounded-full w-[75px]',
        {
          'bg-[#F2F4F6] text-[#576275]': role === 'sys_employee',
          'bg-[#0C99FF]/20 text-[#0C99FF]': role === 'sys_admin',
        },
        className,
      )}
      style={style}
    >
      {role === 'sys_admin' ? '管理员' : '成员'}
    </span>
  )
}

const EmployeeOperator: FC<Style & { value: Employee }> = (props) => {
  const { value, className, style } = props
  const { version } = useDeployConfig()
  const [permOpen, { toggle: togPerm }, permKey] = useOpen()
  const [editOpen, { toggle: togEdit }, editKey] = useOpen()
  const { dispatchEmployeeAction } = usePersonnelData()
  const { userInfo, uinList, setLogin } = useLocalStore()
  const handleResetPassword = () => {
    Modal.confirm({
      title: '确认重置密码',
      content: `确定要重置 ${value.name} 的密码吗？`,
      okText: '确认',
      cancelText: '取消',
      onOk: async () => {
        try {
          await resetPassword({ uin: value.uin })
          message.success('密码重置成功')
        } catch (error) {
          console.log('重置密码失败:', error)
        }
      },
    })
  }

  return (
    <>
      <div className={cn('flex items-center gap-3', className)} style={style}>
        <Tooltip title='编辑权限' placement='top'>
          <PermIcon className='cursor-pointer' onClick={togPerm} />
        </Tooltip>
        <Tooltip title='编辑成员' placement='top'>
          <EditIcon className='cursor-pointer' onClick={togEdit} />
        </Tooltip>
        <Tooltip title='重置密码' placement='top'>
          <ResetPasswordIcon
            className='cursor-pointer'
            onClick={handleResetPassword}
          />
        </Tooltip>
      </div>
      <PermModal
        key={permKey}
        uin={value.uin}
        title='编辑权限'
        open={permOpen}
        onCancel={togPerm}
        onSubmit={async (val) => {
          const { chatPs, forestPs } = val
          const forestPermHandle =
            forestPs.length === 0
              ? null
              : modifyForestPermSet({
                  uin: value.uin,
                  perm_set: forestPs.map((item) => {
                    const {
                      manage_perm,
                      use_perm,
                      forest: { ID },
                    } = item
                    return {
                      manage_perm,
                      use_perm,
                      forest: { ID },
                      act_option: 'update',
                    }
                  }),
                })
          const chatPermHandle =
            chatPs.length === 0
              ? null
              : modifyChatPermSet({
                  uin: value.uin,
                  perm_set: chatPs.map((item) => {
                    const {
                      manage_perm,
                      use_perm,
                      agent: { ID },
                    } = item
                    return {
                      manage_perm,
                      use_perm,
                      agent: { ID },
                      act_option: 'update',
                    }
                  }),
                })
          await Promise.all([forestPermHandle, chatPermHandle])
        }}
      />
      <EmployeeDrawer
        key={editKey}
        title='编辑成员'
        open={editOpen}
        onClose={togEdit}
        defaultValue={value}
        onSubmit={async (val) => {
          await dispatchEmployeeAction({
            type: 'edit',
            ...val,
            id: value.id,
            version,
          })
          const newUinList = uinList.map((item) => {
            if (String(item.id) !== String(userInfo.uinId)) return item
            return { ...item, uinName: val.name }
          })
          setLogin({
            uinList: newUinList,
          })
        }}
      />
    </>
  )
}
