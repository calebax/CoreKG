import { useNavigate, Outlet } from 'react-router-dom'
import { Breadcrumb } from 'antd'
import { useTranslation } from 'react-i18next'
import SeparatorIcon from './images/separator.svg?react'
import styles from './styles.module.scss'

export default function ProfileLayout() {
  const navigate = useNavigate()
  const { t } = useTranslation('pages')

  return (
    <div className='w-full h-full'>
      {/* 页面面包屑 */}
      <div className='w-full h-12 bg-white flex items-center px-5 pt-[2px] border-b border-[#EFF1F4]'>
        <Breadcrumb
          className={styles.layoutHeader}
          separator={<SeparatorIcon />}
          items={[
            {
              title: (
                <span
                  className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
                  onClick={() => {
                    navigate(`/`)
                  }}
                >
                  {t('app.sidebar.project')}
                </span>
              ),
            },
            {
              title: (
                <span className='cursor-pointer text-sm font-medium text-[#3C4149]'>
                  {t('app.sidebar.personalCenter')}
                </span>
              ),
            },
          ]}
        />
      </div>
      <Outlet />
    </div>
  )
}
