import { FC } from 'react'
import { useSearchParams } from 'react-router-dom'
import { App, Button } from 'antd'
import { useBoolean } from 'ahooks'
import { cn } from '@/utils'
import useLocalStore from '@/stores/local'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useVersion } from '@/utils/useVersion'
import { UpgradeDrawer } from './UpgradeDrawer'

interface UsageItemProps {
  title: string
  current?: string | number
  limit: string | number
  rate: number
}

const UsageItem: FC<UsageItemProps> = ({ title, current, limit, rate }) => {
  const percentage = Math.min(rate * 100, 100)
  return (
    <div className='flex flex-col gap-[6px] w-full'>
      <div className='flex items-center justify-between w-full'>
        <span className='text-base text-[#4E5969] font-normal leading-[24.7px]'>
          {title}
        </span>
        <span className='text-base font-normal leading-[24.7px]'>
          {current !== undefined ? (
            <>
              <span className='text-[#86909C]'>{current}</span>
              <span className='text-[#86909C]'>/</span>
              <span className='text-[#86909C]'>{limit}</span>
            </>
          ) : (
            <span className='text-[#86909C]'>{limit}</span>
          )}
        </span>
      </div>
      <div className='relative bg-[#eaf2ff] rounded-[50px] h-1.5 w-full'>
        <div
          className='absolute left-0 top-0 bottom-0 rounded-[50px] bg-[#3d7fff]'
          style={{ width: `${percentage}%` }}
        ></div>
      </div>
    </div>
  )
}

export const UsageManagement: FC = () => {
  const { version: deployVersion } = useDeployConfig()
  const { version } = useVersion()
  const isAdmin = useLocalStore((state) => {
    const { uinList, userInfo } = state
    const uin = uinList.find((item) => userInfo.uinId === item.id)
    return uin?.role === 'sys_admin'
  })
  const [searchParams] = useSearchParams()
  const upgradeParam = searchParams.get('upgrade')
  const [open, { toggle }] = useBoolean(Boolean(upgradeParam && isAdmin))
  const { message } = App.useApp()
  // 仅在 saas 版本显示且有版本数据时显示
  if (deployVersion !== 'saas' || !version) {
    return null
  }

  const { qa, agent, disk: knowledge, employee: team } = version

  return (
    <>
      <div
        className='border border-[#E3E6ED] rounded-[10px] px-6 py-5 w-full h-fit'
        style={{
          background:
            'linear-gradient(181.017deg, rgba(152, 199, 255, 0.48) 1.5241%, rgba(255, 255, 255, 0.424) 27.629%), linear-gradient(90deg, rgb(255, 255, 255) 0%, rgb(255, 255, 255) 100%)',
          boxShadow: '0px 4px 14px 0px rgba(0, 0, 0, 0.1)',
        }}
      >
        <div className='flex '>
          <div className='flex flex-col gap-1 leading-[19.759px] mb-3'>
            <span className='text-[#0C99FF] text-lg font-medium'>
              {version.name}
            </span>
            <span className='text-[#3C4149] text-lg font-normal'>
              今日可用问答共{qa.quota}次，当前剩余{qa.quota - qa.used}次
            </span>
          </div>
          <Button
            className={cn(
              'ml-auto',
              'border-[#0C99FF] border-2 text-[#0C99FF] font-medium bg-transparent',
              {
                'opacity-50': !isAdmin,
              },
            )}
            onClick={() => {
              if (!isAdmin) {
                message.warning('若需提升额度请联系管理员办理升级')
              } else {
                toggle()
              }
            }}
          >
            升级版本
          </Button>
        </div>

        {/* 用量列表 */}
        <div className='flex flex-col gap-4 mb-5'>
          <UsageItem
            title='智能体数量'
            current={agent.used}
            limit={agent.quota + '个'}
            rate={agent.quota ? agent.used / agent.quota : 0}
          />
          <UsageItem
            title='知识库容量'
            limit={knowledge.quota}
            rate={Number(knowledge.use_ratio)}
          />
          <UsageItem
            title='团队成员数量'
            current={team.used}
            limit={team.quota}
            rate={team.used / team.quota}
          />
        </div>
      </div>
      <UpgradeDrawer open={open} onClose={toggle} />
    </>
  )
}
