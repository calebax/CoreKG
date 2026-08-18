import { Breadcrumb, Button, Skeleton } from 'antd'
import { SyncOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import BackIcon from '@/assets/icons/backIcon.svg?react'
import { useAccountBindings } from './hooks'
import SeparatorIcon from './images/separator.svg?react'
import styles from './index.module.scss'
import type { BindableItem } from './types'

export default function AccountBindings() {
  const { t } = useTranslation('pages')
  const {
    bindableList,
    bindInfos,
    handleBind,
    handleUnBind,
    goBack,
    loadedRef,
    loadingList,
  } = useAccountBindings()

  const renderHeader = () => {
    return (
      <>
        <div className='h-[50px] bg-[#fff] border-b border-b-[#EFF1F4] px-[16px] flex items-center'>
          <Breadcrumb
            className={styles.accountBindingsHeader}
            separator={<SeparatorIcon />}
            items={[
              {
                title: (
                  <span
                    className='cursor-pointer'
                    onClick={() => {
                      goBack()
                    }}
                  >
                    {t('app.sidebar.project')}
                  </span>
                ),
              },
              {
                title: (
                  <span className='cursor-pointer'>
                    {t('app.sidebar.externalDataSource')}
                  </span>
                ),
              },
            ]}
          />
        </div>
        <div className='flex flex-col gap-[16px] mt-[50px] mx-[50px] mb-[34px]'>
          <div className='text-[32px] text-[#0C1F17] font-[600] leading-[1]'>
            {t('settings.welcomeExtDataSource')}
          </div>
          <div className='text-[16px] text-[#0C1F17] font-[600] leading-[1]'>
            {t('settings.centralizeKnowledgeResourcesMultiFormatSupport')}
          </div>
        </div>
      </>
    )
  }

  const handleItemClick = (item: BindableItem) => {
    const bindInfo = bindInfos[item.provider]
    if (bindInfo) {
      handleUnBind(bindInfo.id, bindInfo.provider)
    } else {
      handleBind(item.provider)
    }
  }

  const renderItem = (item: BindableItem) => {
    return (
      <div key={item.provider} className={styles.accountBindingsItem}>
        <div className={styles.accountBindingsItemTitle}>
          <div className={styles.accountBindingsItemLogoWrap}>
            <img className={styles.accountBindingsItemLogo} src={item.logo} />
          </div>
          <div className={styles.accountBindingsItemText}>{item.name}</div>
        </div>
        <Button
          onClick={() => handleItemClick(item)}
          loading={
            loadingList.includes(item.provider) && {
              icon: <SyncOutlined spin />,
            }
          }
          className={`${styles.accountBindingsItemBind} ${bindInfos[item.provider] && styles.accountBindingsItemUnBind}`}
        >
          {bindInfos[item.provider]
            ? t('settings.deauthorize')
            : t('settings.deauthorize')}
        </Button>
      </div>
    )
  }

  const style = {
    width: '100%',
    height: 92,
    borderRadius: 10,
  }
  const renderSkeleton = () => {
    return (
      <>
        <Skeleton.Node active style={style} />
        <Skeleton.Node active style={style} />
        <Skeleton.Node active style={style} />
        <Skeleton.Node active style={style} />
        <Skeleton.Node active style={style} />
        <Skeleton.Node active style={style} />
        <Skeleton.Node active style={style} />
      </>
    )
  }

  return (
    <div className={styles.accountBindings}>
      {renderHeader()}
      <div className={styles.accountBindingsContent}>
        {loadedRef.current ? bindableList.map(renderItem) : renderSkeleton()}
      </div>
    </div>
  )
}
