import React from 'react'
import { Dropdown, Modal, App, Button } from 'antd'
import type { MenuProps } from 'antd'
import { ExclamationCircleOutlined, CheckOutlined } from '@ant-design/icons'
import { cn } from '@/utils'
import { getAllUin, switchLogin } from '@/api/account'
import useLocalStore from '@/stores/local'
import { Help } from './Help'
import { VersionBtn } from './Version'
import Logout from './images/logout.svg?react'
import Settings from './images/setting.svg?react'
import Team from './images/team.svg?react'
import User from './images/user.svg?react'
import styles from './styles.module.scss'

const Header: React.FC = () => {
  const { message } = App.useApp()
  const { uinList, userInfo, setLogin, setLogout, token } = useLocalStore()
  const currentUin = useMemo(() => {
    return uinList.find((item) => item.id === userInfo.uinId)
  }, [uinList, userInfo.uinId])

  const items: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <User />,
      label: <NavLink to='/profile'>个人中心</NavLink>,
    },
    // {
    //   key: 'team',
    //   icon: <Team />,
    //   label: '切换团队',
    //   children: uinList.map((item) => {
    //     const active = item.id === userInfo.uinId
    //     const label = (
    //       <div
    //         className={cn(
    //           'my-1 h-13 w-64 py-1 pl-5 relative',
    //           'flex flex-col gap-0.5',
    //           styles.teamItem,
    //           {
    //             [styles.active]: active,
    //           },
    //         )}
    //         onClick={async () => {
    //           if (active) return
    //           const { jwt_token } = await switchLogin({
    //             login_way: userInfo.loginWay!,
    //             uin: item.id as any as number,
    //           })
    //           setLogin({
    //             token: jwt_token,
    //             uinList,
    //             userInfo: {
    //               ...userInfo,
    //               uinId: item.id,
    //             },
    //           })
    //           history.go(0)
    //         }}
    //       >
    //         <span className='flex gap-1.5 items-center'>
    //           <span className='text-sm text-[#303133]'>{item.uinName}</span>
    //           {item.role === 'sys_admin' ? (
    //             <span className=' text-[12px] text-[#165DFF]'>管理员</span>
    //           ) : (
    //             <span className='text-[#86909C] text-[12px] '>普通成员</span>
    //           )}
    //         </span>
    //         <span className='text-[#4E5969] text-sm'>{item.companyName}</span>
    //         <CheckOutlined
    //           className={cn(
    //             ' absolute top-4 bottom-4 right-5',
    //             ' text-[#165DFF]!',
    //             { 'hidden!': !active },
    //           )}
    //         />
    //       </div>
    //     )
    //     return {
    //       label,
    //       key: item.id,
    //     }
    //   }),
    //   popupClassName: styles.teamPopup,
    // },
    {
      key: 'logout',
      label: '退出登录',
      danger: true,
      icon: <Logout />,
      onClick: () => {
        Modal.confirm({
          title: '退出登录',
          icon: <ExclamationCircleOutlined />,
          content: '确定退出登录吗？',
          okText: '确认',
          cancelText: '取消',
          onOk: () => {
            setLogout()
          },
        })
      },
    },
  ]
  if (currentUin?.role === 'sys_admin') {
    items.splice(2, 0, {
      key: 'settings',
      icon: <Settings />,
      label: <NavLink to={'/settings'}>系统设置</NavLink>,
    })
  }
  return (
    <div
      className={cn(
        'w-full h-13 flex-none',
        'bg-[#FFFFFF]',
        'flex items-center justify-end',
      )}
    >
      <VersionBtn />
      <Button
        className={cn(
          'ml-4 w-21 h-8',
          'border-2 border-[#165DFF] text-[#165DFF] font-bold!',
          'rounded',
        )}
        style={{ fontFamily: 'Inter ' }}
        onClick={() => {
          message.info('API市场正在开发中，敬请期待')
        }}
      >
        API市场
      </Button>
      <Help className='ml-4' />
      <Dropdown menu={{ items }} rootClassName={styles.menu}>
        <img
          src={userInfo.avatar}
          alt='avatar'
          className='w-8 h-8 mx-8 rounded-full'
        />
      </Dropdown>
    </div>
  )
}

export default Header
