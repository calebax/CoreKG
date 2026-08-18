import { FC, useRef, useEffect } from 'react'
import { Modal, Button, Skeleton, QRCode } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import { useCountDown, useRequest } from 'ahooks'
import dayjs from 'dayjs'
import { match, P } from 'ts-pattern'
import { cn } from '@/utils'
import { createOrder, queryOrderStatus } from '@/api/pay'
import { useVersion } from '@/utils/useVersion'
import success from './images/success.png'
import wechat from './images/wechat.png'

type PurchaseModal = {
  package_id?: number
  amount?: number
  onClose?: () => void
}

export const PurchaseModal: FC<PurchaseModal> = (props) => {
  const { package_id } = props
  const keyRef = useRef(1)
  useEffect(() => {
    // 使弹窗平滑消失
    if (!package_id) {
      ++keyRef.current
    }
  }, [package_id])
  return <PurchaseModalInner key={`${keyRef.current}`} {...props} />
}

const PurchaseModalInner: FC<PurchaseModal> = (props) => {
  const { package_id, amount, onClose } = props
  const { refresh } = useVersion()
  const { data, run, loading } = useRequest(
    async () => {
      if (!package_id) return null
      const orderRes = await createOrder({
        package_id,
      })
      return orderRes
    },
    {
      ready: Boolean(package_id),
    },
  )

  // 状态只有支付中/成功/超时 不考虑失败的情形
  const {
    data: status,
    mutate,
    cancel,
  } = useRequest(
    async () => {
      if (!data) return
      const { expire_time, order_sn } = data
      // 超时
      if (dayjs().isAfter(dayjs(expire_time))) {
        return 'closed'
      }
      const res = await queryOrderStatus({ order_sn })
      const { status } = res
      return status
    },
    {
      ready: Boolean(data?.order_sn && package_id),
      pollingInterval: 1000,
      onSuccess: (result) => {
        // 如果状态为 closed 或 success，停止轮询
        if (result === 'closed' || result === 'success') {
          cancel()
        }
        if (result === 'success') {
          refresh()
        }
      },
    },
  )

  const retry = () => {
    run()
    mutate(undefined)
  }
  return (
    <Modal
      open={Boolean(package_id)}
      title={match(status)
        .with(
          P.union(
            // 'failed',
            'success',
          ),
          () => '支付结果',
        )
        .otherwise(() => '支付')}
      onCancel={onClose}
      footer={match(status)
        // .with('failed', () => (
        //   <Button type='primary' onClick={retry}>
        //     重新支付
        //   </Button>
        // ))
        .with('success', () => null)
        .otherwise(() => (
          <Button className='bg-[#F5F5F5]' onClick={onClose}>
            取消支付
          </Button>
        ))}
    >
      {match({ data, status, loading })
        // .with({ status: 'failed' }, () => (
        //   <div className='flex flex-col items-center justify-center text-[#999999]'>
        //     <img src={failed} className='mb-5' />
        //     <span className='text-lg font-medium mb-2'>失败</span>
        //   </div>
        // ))
        .with({ status: 'success' }, () => <PurchaseSuccess />)
        .with({ data: P.nonNullable, loading: false }, ({ data, status }) => (
          <div className='mx-auto flex flex-col gap-1 justify-center items-center text-[#6E757F]'>
            <div className='mb-6 px-2.5 py-1 bg-[#0C99FF1A] text-[#0C99FF]'>
              您购买的套餐等级高于当前套餐，该套餐将即时生效，当前套餐将延后生效。您可在「个人中心」中，查看套餐使用情况。
            </div>
            {typeof amount === 'number' && amount >= 0 ? (
              <span className='text-base'>
                应付金额
                <span className='text-lg ml-1 text-[#0C99FF]'>
                  ￥{(amount / 100).toFixed(2)}
                </span>
              </span>
            ) : null}
            <div className='w-54 h-54 relative'>
              <QRCode value={data.pay_url} className='w-full! h-full! m-0' />
              {status === 'closed' ? (
                <div
                  className={cn(
                    ' absolute inset-0 z-10 ',
                    'bg-[#FFFFFFF2]',
                    'flex flex-col gap-1 items-center justify-center',
                  )}
                >
                  <span>二维码失效</span>
                  <Button onClick={retry} icon={<ReloadOutlined />} type='link'>
                    点击刷新
                  </Button>
                </div>
              ) : null}
            </div>
            {status !== 'closed' ? (
              <>
                <span className='flex items-center gap-1'>
                  请使用
                  <img src={wechat} />
                  微信支付
                </span>
                <span className='text-xs'>支付完成后，请等待系统自动更新</span>
              </>
            ) : null}
          </div>
        ))
        .otherwise(() => (
          <Skeleton className='p-4' active />
        ))}
    </Modal>
  )
}

const PurchaseSuccess: FC = () => {
  const [, formattedRes] = useCountDown({
    leftTime: 3000,
    interval: 1000,
    onEnd: () => {
      window.location.reload()
    },
  })
  const { seconds } = formattedRes
  return (
    <div className='mx-auto flex flex-col items-center justify-center text-[#0C99FF]'>
      <img src={success} className='mb-5' />
      <span className='text-lg font-medium mb-2'>支付成功</span>
      <span>{seconds}s后将自动返回「个人中心」页</span>
    </div>
  )
}
