import { Link, useLocation, useNavigate } from 'react-router-dom'
import { Breadcrumb } from 'antd'
import { useTranslation } from 'react-i18next'
import NavigationIcon from '@/assets/icons/docs/navigation.svg?react'

export default function Nav() {
  const location = useLocation()
  const navigate = useNavigate()
  const { t } = useTranslation('pages')

  return (
    <div className='mb-6'>
      <Breadcrumb
        className='[&_span.ant-breadcrumb-separator]:inline-flex [&_span.ant-breadcrumb-separator]:items-center [&_span.ant-breadcrumb-separator]:align-middle'
        separator={<NavigationIcon className='inline-block' />}
        items={[
          {
            title: (
              <span
                onClick={() => {
                  navigate(`/docs`)
                }}
                className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
              >
                {t('app.docs.knowledgeBase')}
              </span>
            ),
          },
          {
            title: (
              <span className='text-sm font-medium text-[#3C4149]'>
                {t('app.docs.qaPair')}
              </span>
            ),
          },
        ]}
      />
    </div>
  )
}
