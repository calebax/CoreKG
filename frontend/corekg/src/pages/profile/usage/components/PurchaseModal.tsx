import React, { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import usageAddQuantityImg from '@/assets/icons/usage-addQuantity.svg'
import popUpBgImg from '@/assets/icons/usage-popUpBg.svg'
import usageWxImg from '@/assets/icons/usage-wx.svg'

interface PackageItem {
  id: string
  amount: number // 用量
  price: number // 价格
}

interface UsageCard {
  id: string
  title: string
  currentQuota: number
  historicalUsage: number
}

interface PurchaseModalProps {
  open: boolean
  onClose: () => void
  card?: UsageCard | null
}

const mockBalance = 2000
const mockPackages: PackageItem[] = [
  { id: '1', amount: 100, price: 10 },
  { id: '2', amount: 500, price: 50 },
  { id: '3', amount: 1000, price: 100 },
  { id: '4', amount: 2000, price: 200 },
  { id: '5', amount: 5000, price: 500 },
  { id: '6', amount: 10000, price: 1000 },
  { id: '7', amount: 20000, price: 2000 },
  { id: '8', amount: 50000, price: 5000 },
]
const mockQr =
  'https://api.qrserver.com/v1/create-qr-code/?size=120x120&data=weixinpay'

const PurchaseModal: React.FC<PurchaseModalProps> = ({
  open,
  onClose,
  card,
}) => {
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const [balance, setBalance] = useState<number>(mockBalance)
  const [packages, setPackages] = useState<PackageItem[]>(mockPackages)
  const [selected, setSelected] = useState<PackageItem>(mockPackages[0])
  const [qr, setQr] = useState<string>(mockQr)

  // 预留：获取余额、套餐、二维码接口
  useEffect(() => {
    if (open) {
      // fetchBalance(card)
      // fetchPackages(card)
      // fetchQr(selected, card)
    }
  }, [open, card])

  // 切换套餐时二维码和金额变化
  const handleSelect = (item: PackageItem) => {
    setSelected(item)
    // setQr(fetchQr(item, card)) // 预留接口
  }

  const backgroundImage = {
    backgroundImage: `url(${popUpBgImg})`,
    backgroundSize: 'contain',
    backgroundPosition: 'top',
    backgroundRepeat: 'no-repeat',
  }

  if (!open) return null

  return (
    <div
      className='fixed inset-0 z-50 flex items-center justify-center bg-[rgba(0,0,0,0.4)]'
      onClick={onClose}
    >
      {' '}
      <div
        className='relative bg-white rounded-2xl shadow-xl w-[900px] max-w-[95vw] p-10 flex flex-col'
        style={{ minHeight: 480, ...backgroundImage }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* 关闭按钮 */}
        <button
          className='absolute right-6 top-6 w-8 h-8 flex items-center justify-center text-gray-400 text-3xl cursor-pointer'
          aria-label={tC('button.close')}
          tabIndex={0}
          onClick={onClose}
        >
          ×
        </button>
        {/* 余额和卡片信息 */}
        <div className='mb-3 text-base font-medium text-[#1D2129] flex items-center bg-no-repeat bg-center bg-cover'>
          {t('profile.balance', { target: ':' })}
          <span className='text-[#165DFF] text-2xl font-bold align-middle mr-2'>
            {balance}
          </span>{' '}
          <span className='text-base font-normal'>{t('profile.usage')}</span>
          {card && <span className='ml-8 text-base'>{card.title}</span>}
        </div>
        {/* 套餐列表 */}
        <div className='grid grid-cols-4 gap-3 mb-8 px-25'>
          {packages.map((item) => (
            <div
              key={item.id}
              className={`flex flex-col items-center justify-center border-2 rounded-xl px-0 py-6 cursor-pointer transition-all select-none box-border ${selected.id === item.id ? 'border-[#4495FF] bg-white' : 'border-[#F9F9F9] bg-[#F9F9F9]'}`}
              onClick={() => handleSelect(item)}
              tabIndex={0}
              aria-label={t('profile.selectCountUsage', { count: item.amount })}
              onKeyDown={(e) =>
                (e.key === 'Enter' || e.key === ' ') && handleSelect(item)
              }
            >
              <div className='flex items-center gap-1 mb-2'>
                <img
                  src={usageAddQuantityImg}
                  alt={t('profile.balance')}
                  className='w-5 h-5'
                />
                <span className='text-[#165DFF] text-base font-normal'>
                  {item.amount} {t('profile.balance')}
                </span>
              </div>
              <div className='text-[#165DFF] text-3xl font-bold'>
                ￥{item.price}
              </div>
            </div>
          ))}
        </div>
        {/* 支付区域 */}
        <div className='flex items-center gap-6 mb-2'>
          {/* 二维码 */}
          <img
            src={qr}
            alt={t('profile.wechatPaymentQrCode')}
            className='w-28 h-28 rounded bg-white border border-[#E5E6EB]'
          />
          {/* 支付信息 */}
          <div className='flex flex-col justify-center gap-2'>
            <div className='text-lg font-medium text-[#165DFF] mb-2'>
              <span className='text-base font-medium mr-4'>
                {t('profile.payment', { target: ':' })}
              </span>
              <span className='text-xl font-bold'>￥</span>
              <span className='text-2xl font-bold'>{selected.price}</span>
            </div>
            <div className='flex items-center gap-4 mb-3'>
              <img
                src={usageWxImg}
                alt={t('profile.wechatPayment')}
                className='w-8 h-8 cursor-pointer'
              />
              <span className='text-[#4E5969] text-base font-normal'>
                {t('profile.pleaseUseWechatPayment')}
              </span>
            </div>
            <div className='text-sm text-[#4E5969] font-normal'>
              {t('profile.activateImpliesAgreement')}{' '}
              <a href='#' className='text-[#1D2129] font-normal text-base'>
                {t('profile.knowledgeEngineMembershipAgreement')}
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default PurchaseModal
