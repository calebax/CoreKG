import { FC } from 'react'
import { App, Button } from 'antd'
import { useBoolean, useRequest } from 'ahooks'
import { t } from 'i18next'
import { useTranslation } from 'react-i18next'
import { editEmployee } from '@/api/apiKey'
import {
  deleteEmployee,
  deleteEmployeePrivate,
  modifyChatPermSet,
  modifyForestPermSet,
} from '@/api/perm'
import useLocalStore from '@/stores/local'
import { encryptPassword } from '@/utils/crypto'
import { useAdmin } from '@/utils/useAdmin'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useVersion } from '@/utils/useVersion'
import { EmployeeInfo, EmployeeInfoModal } from '../EmployeeInfoModal'
import { PermModal } from '../PermModal'

export type Operator = {
  employeeInfo: EmployeeInfo & { uin: number }
  /** 用户uin */
  uin: number
  /** 员工id */
  id: number
  /** 重新加载表格数据 */
  reload: () => void
}
export const Operator: FC<Operator> = (props) => {
  const { version } = useDeployConfig()
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const { t: tM } = useTranslation('messages')
  const { uin, id, reload, employeeInfo } = props
  const { refresh } = useVersion()
  const [infoOpen, { toggle: toggleInfo }] = useBoolean(false)
  const [permOpen, { toggle: togglePerm }] = useBoolean(false)
  const { message } = App.useApp()
  const rmHandler = useRequest(
    async () => {
      if (version === 'custom') {
        await deleteEmployeePrivate({ employee_id: id })
      } else {
        await deleteEmployee({ employee_id: id })
      }
      reload()
      refresh()
      message.success(tM('operationSuccess'))
    },
    { manual: true },
  )
  const { adminIds } = useAdmin()
  const userInfo = useLocalStore((state) => state.userInfo)
  const disabled = useMemo(() => {
    // 管理员不能操作自己
    if (adminIds.includes(uin)) return true
    // 如果管理员信息未能成功获取的兜底
    // 只有管理员才能进入此页面
    if ((userInfo.uinId as any) === uin) return true
    return false
  }, [adminIds, uin, userInfo.uinId])
  return (
    <>
      <span className='flex gap-2'>
        {version === 'custom' ? (
          <Button type='link' onClick={toggleInfo} disabled={disabled}>
            编辑
          </Button>
        ) : null}
        <Button type='link' onClick={togglePerm} disabled={disabled}>
          {t('settings.permission')}
        </Button>
        <Button
          type='link'
          loading={rmHandler.loading}
          onClick={rmHandler.run}
          disabled={disabled}
        >
          {t('settings.moveOut')}
        </Button>
      </span>
      {infoOpen ? (
        <EmployeeInfoModal
          title='编辑成员'
          value={employeeInfo}
          open={infoOpen}
          onClose={toggleInfo}
          onOk={async (val) => {
            await editEmployee({ ...val, uin: employeeInfo.uin })
            message.success('操作成功')
            refresh()
            reload()
          }}
        />
      ) : null}
      <PermModal
        uin={uin}
        open={permOpen}
        close={togglePerm}
        okText={tC('button.confirm')}
        onSubmit={async (val) => {
          const { chatPs, forestPs } = val
          const forestPermHandle =
            forestPs.length === 0
              ? null
              : modifyForestPermSet({
                  uin,
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
                  uin,
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
          message.success(tM('permissionUpdateSuccess'))
          reload()
        }}
      />
    </>
  )
}
