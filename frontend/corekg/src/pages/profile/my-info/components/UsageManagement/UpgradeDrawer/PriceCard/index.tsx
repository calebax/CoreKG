import type { FC } from 'react'
import { Button } from 'antd'
import { cn } from '@/utils'
import Ok from './images/ok.svg'
import Spark from './images/sparkles-soft.svg'
import styles from './styles.module.scss'

export type PricingCardProps = Style & {
  title?: string
  /** 限时特惠 */
  discount?: boolean
  /** 正在使用 */
  inUse?: boolean
  price?: string
  underlinePrice?: string
  desc?: string
  btn?: {
    text?: string
    onClick?: () => void
    disabled?: boolean
  }
  features?: string[]
}

export const PricingCard: FC<PricingCardProps> = (props) => {
  const {
    title,
    discount,
    inUse,
    price,
    underlinePrice,
    desc,
    btn,
    features,
    className,
    style,
  } = props

  return (
    <div
      className={cn(
        'w-78 pt-9 px-8 pb-12 relative flex flex-col gap-1',
        'rounded-md bg-white',
        {
          'border-[#0C99FFCC] border-2': inUse,
        },
        styles.card,
        className,
      )}
      style={{
        backgroundImage:
          'linear-gradient(180.68deg, rgba(222, 234, 250, 0.48) 1.52%, rgba(255, 255, 255, 0.4224) 27.63%)',
        ...style,
      }}
    >
      <div
        className={cn(
          'px-2 py-1',
          'absolute right-0 top-0',
          'rounded-bl bg-[#0C99FF0D] font-medium text-base text-[#0C99FF]',
          { hidden: !inUse },
        )}
      >
        正在使用
      </div>
      <div className='text-xl font-medium text-[#0C99FF] flex items-center gap-1'>
        {title}
        {discount ? (
          <div
            className='text-xs px-1 py-0.5 text-white rounded'
            style={{
              backgroundImage:
                'linear-gradient(113.73deg, #FFB866 0%, #FF6A1F 95.84%)',
            }}
          >
            限时特惠
          </div>
        ) : null}
      </div>
      <div className={cn('flex gap-2 items-end', 'text-2xl font-medium')}>
        {price}
        {underlinePrice ? (
          <div className='text-sm text-[#ABAFB2] line-through'>
            {underlinePrice}
          </div>
        ) : null}
      </div>
      <div className=''>{desc}</div>
      {btn ? (
        <Button
          {...btn}
          className={cn({
            'border-2 border-[#0C99FF] text-[#0C99FF] bg-[#0C99FF0D]':
              !btn.disabled,
          })}
          block
        >
          {btn.text}
        </Button>
      ) : null}
      <div
        className={cn(
          'mx-auto mt-4 mb-3 flex items-center gap-1',
          'font-medium',
        )}
      >
        <img src={Spark} />
        用户权益
      </div>
      <div className='flex flex-col gap-3'>
        {features?.map((f) => (
          <div key={f} className='flex items-center gap-1.5 text-[#4E5969]'>
            <img src={Ok} />
            {f}
          </div>
        ))}
      </div>
    </div>
  )
}

export default PricingCard
