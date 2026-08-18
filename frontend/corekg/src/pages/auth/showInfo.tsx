import { useState, useEffect } from 'react'
import { Input, Button, message } from 'antd'
import { useMount } from 'ahooks'
import { getLicenseInfo } from '@/api/auth'
import Notice from '@/assets/icons/auth-notice.svg?react'

interface LicenseInfo {
  id: number
  serial: string
  issuer: string
  createdAt: string
  expiredAt: string
  status: number
  validDays: number
}

function RenderImpower({ info }: { info: LicenseInfo | null }) {
  const impower = [
    {
      id: 0,
      title: '签发人',
      content: info?.issuer || '暂未授权',
    },
    {
      id: 1,
      title: '授权开始时间',
      content: info?.createdAt || '一',
    },
    {
      id: 2,
      title: '授权结束时间',
      content: info?.expiredAt || '一',
    },
  ]
  const impowerList = impower.map((impower) => (
    <div
      key={impower.id}
      className='flex justify-between items-center w-sm h-12 p-3 my-4 rounded-lg border-1 border-[#EDEDEF]'
    >
      <div className='text-[#616373]'>{impower.title}</div>
      <div className='text-[#1E1F28]'>{impower.content}</div>
    </div>
  ))
  return impowerList
}

export default function ShowInfo({
  refreshTrigger,
}: {
  refreshTrigger?: number
}) {
  const [info, setInfo] = useState<LicenseInfo | null>(null)

  const formatDate = (dateString: string) => {
    if (!dateString) return '一'
    try {
      const datePart = dateString.split('T')[0]
      return datePart
    } catch (error) {
      console.error('日期格式化失败:', error)
      return '一'
    }
  }

  const getInfo = async () => {
    try {
      const res = await getLicenseInfo({})
      const data: LicenseInfo = {
        id: res.meta?.id || 0,
        serial: res.meta?.serial || '',
        issuer: res.meta?.issuer || '',
        createdAt: res.meta?.created_at ? formatDate(res.meta.created_at) : '',
        expiredAt: res.meta?.expired_at ? formatDate(res.meta.expired_at) : '',
        status: res.status,
        validDays: res.valid_days,
      }
      setInfo(data)
      return res
    } catch (error) {
      console.error('获取license信息失败:', error)
    }
  }

  useMount(() => {
    getInfo()
  })

  // 监听 refreshTrigger 变化，重新获取许可证信息
  useEffect(() => {
    if (refreshTrigger && refreshTrigger > 0) {
      getInfo()
    }
  }, [refreshTrigger])

  // 根据status判断激活状态
  const isActivated = info?.status === 0
  const statusText = isActivated ? '已激活' : '未激活'
  const statusStyle = isActivated
    ? 'w-sm mt-3 h-36 px-5 py-6 border-0 rounded-xl text-[#E6E8F0] bg-[#00C2B3]'
    : 'w-sm mt-3 h-36 px-5 py-6 border-0 rounded-xl text-[#1E1F28] bg-[#E6E8F0]'

  return (
    <div className=' p-6 rounded-lg bg-[#FCFCFE] border-[#D7D9E5] border-1'>
      <div className='flex items-center mb-8 py-3 pl-9 bg-[#165DFF12]'>
        <Notice className='w-4 h-4 mr-1' />
        成功授权后，信息会在此实时更新。
      </div>
      <p className='font-medium text-base'>授权向导</p>
      {info && (
        <>
          <RenderImpower info={info} />
          <div className='w-sm h-36 px-5 py-6 border-0 rounded-xl text-white inset-shadow-sm shadow-[0_0_8.2px_0_rgba(255, 255, 255, 1)] bg-linear-to-r from-[#3D7FFFBD] from-40% to-[#3D7FFF]'>
            <div className='text-xl font-normal mb-8'>有效期</div>
            <div className='text-3xl font-semibold'>{info.validDays}天</div>
          </div>
          <div className={statusStyle}>
            <div className='text-xl font-normal mb-8'>激活状态</div>
            <div className='text-3xl font-semibold'>{statusText}</div>
          </div>
        </>
      )}
    </div>
  )
}
