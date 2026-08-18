import { FC, PropsWithChildren } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { useDeployConfig } from '@/utils/useDeployConfig'

/** 用户协议 */
export const Agreement: FC<Style> = (props) => {
  const { version } = useDeployConfig()
  const { t } = useTranslation('common')
  const serviceAgreement = useMemo(() => {
    return new URL(
      t('agreement.serviceAgreementLink') + `?${Date.now()}`,
      location.origin + import.meta.env.BASE_URL,
    ).href
  }, [t])
  const privacyAgreement = useMemo(() => {
    return new URL(
      t('agreement.privacyAgreementLink') + `?${Date.now()}`,
      location.origin + import.meta.env.BASE_URL,
    ).href
  }, [t])
  if (version === 'custom') return null
  return (
    <div
      className={cn(
        'flex items-center font-medium whitespace-pre',
        props.className,
      )}
      style={props.style}
    >
      {t('agreement.youAgreement')}
      <BaseAgreement url={serviceAgreement}>
        {t('agreement.serviceAgreement')}
      </BaseAgreement>
      {t('agreement.and')}
      <BaseAgreement url={privacyAgreement}>
        {t('agreement.privacyAgreement')}
      </BaseAgreement>
    </div>
  )
}

const BaseAgreement: FC<Style & PropsWithChildren & { url: string }> = (
  props,
) => {
  const { url } = props

  return (
    <Link
      to={url}
      target={'_blank'}
      type='link'
      className={cn('m-0 font-medium', props.className)}
      style={props.style}
    >
      {props.children}
    </Link>
  )
}
