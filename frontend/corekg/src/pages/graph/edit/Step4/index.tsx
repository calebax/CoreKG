import { FC } from 'react'
import { Button, Result, Spin } from 'antd'
import { useCountDown, useRequest } from 'ahooks'
import { parseGraph } from '@/api/graph'
import type { StepComponent } from '..'
import { useGraphInfo } from '../../GraphProvider'

export const Step4: StepComponent = () => {
  const { data } = useGraphInfo()
  const { loading, error } = useRequest(async () => {
    await parseGraph({ graph_id: data!.id })
  })

  if (loading) {
    return (
      <div className='w-full h-full flex items-center justify-center min-h-[400px]'>
        <div className='flex items-center gap-2'>
          <Spin />
          <span className='text-base'>正在生成图谱...</span>
        </div>
      </div>
    )
  }
  if (error) {
    return (
      <Result status={'error'} title='图谱创建失败' subTitle='请稍后再试' />
    )
  }
  return <SussessResult />
}

const SussessResult: FC = () => {
  const [count] = useCountDown({ leftTime: 3000 })
  const navigate = useNavigate()
  useEffect(() => {
    if (count === 0) {
      navigate('/graph')
    }
  }, [count, navigate])

  return (
    <Result
      status={'success'}
      title='创建成功'
      subTitle={`${Math.ceil(count / 1000)}秒后自动返回图谱主页`}
      extra={
        <Link to='/graph'>
          <Button type='primary'>返回图谱主页</Button>
        </Link>
      }
    />
  )
}
