import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Spin } from 'antd'
import useLocalStore from '@/stores/local'
import { useDeployConfig } from '@/utils/useDeployConfig'

/**
 * DotpenWeb 跳转页面
 * 用于从其他项目跳转到本项目，实现账号体系打通
 *
 * 仅在以下环境可用：
 * - 本地环境（development）
 * - 测试环境（test）
 * - custom 版本且 mode 为 cimc 或 h3c
 *
 * URL 参数：
 * - token: 用户认证 token
 * - path: 目标页面路径（global/docs/agents）
 */
const DotpenWeb = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setToken } = useLocalStore()
  const { version, mode } = useDeployConfig()
  const [status, setStatus] = useState<'loading' | 'error'>('loading')
  const [errorMessage, setErrorMessage] = useState('')

  // 判断是否在允许的环境中
  const isDevEnv = import.meta.env.MODE === 'development'
  const isTestEnv = import.meta.env.MODE === 'test'
  const isCimcOrH3cMode =
    version === 'custom' && (mode === 'cimc' || mode === 'h3c')
  const isAllowedEnv = isDevEnv || isTestEnv || isCimcOrH3cMode

  useEffect(() => {
    // 环境校验：仅在本地、测试或 custom+cimc/h3c 环境下可用
    if (!isAllowedEnv) {
      setStatus('error')
      setErrorMessage('当前环境不支持此功能')
      return
    }

    // 获取 URL 参数
    const token = searchParams.get('token')
    const path = searchParams.get('path')

    // 校验参数
    if (!token) {
      setStatus('error')
      setErrorMessage('缺少 token 参数')
      return
    }

    if (!path) {
      setStatus('error')
      setErrorMessage('缺少 path 参数')
      return
    }

    // 校验 path 是否为有效值
    const validPaths = ['global', 'docs', 'agents']
    if (!validPaths.includes(path)) {
      setStatus('error')
      setErrorMessage(
        `无效的 path 参数：${path}，有效值为：${validPaths.join('、')}`,
      )
      return
    }

    // 替换本地存储的 token
    setToken(token)

    // 根据 path 跳转到对应页面
    const targetPath = `/${path}`
    navigate(targetPath, { replace: true })
  }, [searchParams, setToken, navigate, isAllowedEnv])

  // 加载中状态
  if (status === 'loading') {
    return (
      <div className='flex h-screen w-full items-center justify-center'>
        <Spin size='large' tip='正在跳转...' />
      </div>
    )
  }

  // 错误状态
  if (status === 'error') {
    return (
      <div className='flex h-screen w-full flex-col items-center justify-center gap-4'>
        <div className='text-lg text-red-500'>跳转失败</div>
        <div className='text-gray-500'>{errorMessage}</div>
      </div>
    )
  }

  return null
}

export default DotpenWeb
