import { FC } from 'react'
import { Outlet, useNavigate, useLocation, useParams } from 'react-router-dom'
import { Layout, Menu } from 'antd'
import {
  Setting1Icon,
  UserLockedIcon,
  ArrowLeftIcon,
} from 'tdesign-icons-react'

const { Sider, Content } = Layout

const AgentSettings: FC = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { id } = useParams<{ id: string }>()

  // 根据当前路径确定选中的菜单项
  const currentPath = location.pathname.split('/').pop()
  const selectedKey = currentPath === 'access' ? 'access' : 'edit'

  const menuItems = [
    {
      key: 'my-apps',
      icon: <ArrowLeftIcon />,
      label: '我的智能体',
    },
    {
      key: 'edit',
      icon: <Setting1Icon />,
      label: '智能体配置',
    },
    {
      key: 'access',
      icon: <UserLockedIcon />,
      label: '访问设置',
    },
  ]

  const handleMenuClick = ({ key }: { key: string }) => {
    if (key === 'my-apps') {
      navigate('/agents')
    } else {
      navigate(`/agents/${id}/${key}`)
    }
  }

  return (
    <Layout className='h-full'>
      <Sider
        width={150}
        className='bg-gray-50 border-r border-gray-200'
        theme='light'
      >
        <Menu
          mode='inline'
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={handleMenuClick}
          className='border-none h-full [&_.ant-menu-item-selected]:!bg-transparent'
        />
      </Sider>
      <Layout className='bg-[#F8F9FB]'>
        <Content className='p-2 overflow-auto'>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}

export default AgentSettings
