import { FC, useState } from 'react'
import { Button, Drawer, DrawerProps, Skeleton } from 'antd'
import { useRequest } from 'ahooks'
import { cn, formatFileSize } from '@/utils'
import { getAllPackage } from '@/api/pay.ts'
import PriceCard from './PriceCard'
import { PurchaseModal } from './PurchaseModal'
import Return from './images/return.svg?react'
import styles from './styles.module.scss'

/** 获取联系售前页面完整 URL，支持带 base 的部署路径 */
const getContactSalesUrl = () => {
  const base = (import.meta.env.BASE_URL || '/').replace(/\/$/, '') || ''
  return `${window.location.origin}${base}/version`
}

export const UpgradeDrawer: FC<{ open?: boolean; onClose?: () => void }> = (
  props,
) => {
  const { open, onClose } = props
  const [package_id, setId] = useState<number>()
  const { data: packages } = useRequest(async () => {
    const res = await getAllPackage()
    const list = res.list ?? []
    // 如果所有套餐都未购买 将免费套餐设置为已购买
    const freePackage = list.find((item) => item.edition === 'free_trail')
    if (freePackage && list.every((item) => !item.is_purchased)) {
      freePackage.is_purchased = true
    }
    return list
  })
  return (
    <>
      <FullScreenDrawer open={open}>
        <ReturnBtn onClick={onClose} />
        <div
          className={cn(
            'w-full min-h-full py-25 flex flex-col items-center',
            styles.bg,
          )}
        >
          {!packages ? (
            <Skeleton active className='p-4 pt-25' />
          ) : (
            <>
              <span className='text-[50px] font-medium'>版本订阅</span>
              <span className='text-[22px] mt-2.5'>
                免费使用或升级更高的版本
              </span>
              <div className='mt-15 px-50 flex gap-6'>
                {packages.map((p) => (
                  <PriceCard
                    key={p.package_id}
                    inUse={p.is_purchased}
                    title={p.name}
                    desc={p.description}
                    price={
                      p.sale_price
                        ? `￥${(p.sale_price / 100).toFixed(2)}/月`
                        : `￥0.00`
                    }
                    discount={p.price !== p.sale_price}
                    underlinePrice={
                      p.price !== p.sale_price
                        ? `￥${(p.price / 100).toFixed(2)}/月`
                        : ''
                    }
                    btn={{
                      text: p.price ? '升级版本' : '免费',
                      onClick: () => {
                        setId(p.package_id)
                      },
                      disabled: !p.price || p.is_purchased,
                    }}
                    features={[
                      `${p.qa_quota}次/日问答`,
                      `${p.agent_quota}个智能体`,
                      `不限知识库数量`,
                      `${formatFileSize(p.disk_quota)}知识库容量`,
                      `${p.employee_quota}个团队成员`,
                    ].concat(p.additional_notes)}
                  />
                ))}
                <PriceCard
                  title='定制版本'
                  price='待议价格'
                  desc={'根据实际场景灵活打造解决方案'}
                  btn={{
                    text: '联系售前',
                    // 跳转独立 /version 路由，刷新时不会触发个人资料页接口
                    onClick: () => window.open(getContactSalesUrl(), '_blank'),
                  }}
                  features={[
                    '私有部署',
                    '功能深度定制',
                    '快速集成协同',
                    '适配国产数据库',
                    '支持接入自有模型',
                    '完整的RBAC',
                    '专属客户支持',
                    'API响应',
                  ]}
                />
              </div>
            </>
          )}
        </div>
      </FullScreenDrawer>
      <PurchaseModal
        package_id={package_id}
        amount={
          packages?.find((item) => item.package_id === package_id)?.sale_price
        }
        onClose={() => setId(undefined)}
      />
    </>
  )
}

const FullScreenDrawer: FC<Omit<DrawerProps, 'classNames'>> = (props) => {
  return (
    <Drawer
      classNames={{
        wrapper: 'w-full!',
        header: 'hidden',
        footer: 'hidden',
        body: 'p-0 relative overflow-auto',
      }}
      {...props}
    ></Drawer>
  )
}

const ReturnBtn: FC<Style & { onClick?: () => void }> = (props) => {
  const { onClick, className, style } = props
  return (
    <Button
      type='link'
      onClick={onClick}
      icon={<Return />}
      className={cn('p-0 absolute left-16 top-14 ', className)}
      style={style}
    >
      返回
    </Button>
  )
}
