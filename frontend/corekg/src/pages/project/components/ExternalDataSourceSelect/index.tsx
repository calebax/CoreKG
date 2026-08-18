import { FC } from 'react'
import { App, Checkbox, Popover } from 'antd'
import { useTranslation } from 'react-i18next'
import config from '@/config'
import { cn } from '@/utils'
import { getBindAccountState } from '@/api/accountBindings'
import { withCheckboxStyle } from '../CheckboxStyleProvider'
import ExternalDataSourceIcon from './images/externalDataSource.svg?react'
import styles from './index.module.scss'

export type ExternalDataSourceItem = {
  id?: number
  provider: string
  logo: string
  label: string
}

type ExternalDataSourceSelectProps = {
  list: ExternalDataSourceItem[]
  checkedList?: number[]
  onChange?: (checkedList: number[]) => void
  allowSelect?: boolean
  disabled?: boolean
}

export const ExternalDataSourceSelect: FC<
  Style & ExternalDataSourceSelectProps
> = withCheckboxStyle((props) => {
  const { t } = useTranslation('pages')
  const { message } = App.useApp()
  const {
    list,
    checkedList = [],
    onChange,
    allowSelect,
    disabled,
    className,
    style,
  } = props
  const handleAuth = async (item: ExternalDataSourceItem) => {
    try {
      const { state } = await getBindAccountState({
        provider: item.provider,
        redirectUrl: location.href,
      })
      location.href = `${config.withPrefix('account.Connect')}/${item.provider}?state=${state}`
    } catch (error) {
      console.log(error)
    }
  }

  const renderItemConnection = (item: ExternalDataSourceItem) => {
    return (
      <div
        onClick={(e) => {
          e.stopPropagation()
          handleAuth(item)
        }}
        className={styles.externalDataSourceSelectItemAuthBtn}
      >
        {t('project.deauthorize')}
      </div>
    )
  }

  const handleChange = (checked: boolean, item: ExternalDataSourceItem) => {
    const result = checked
      ? [...checkedList, item.id!]
      : checkedList.filter((selected) => selected !== item.id)
    onChange?.(result)
  }

  const renderItem = (item: ExternalDataSourceItem) => {
    return (
      <Checkbox
        key={item.provider}
        disabled={!item.id}
        onChange={(e) => {
          if (!allowSelect) return
          handleChange(e.target.checked, item)
        }}
        className={styles.externalDataSourceSelectItem}
        checked={!!item.id && checkedList.includes(item.id)}
      >
        <img
          src={item.logo}
          className={styles.externalDataSourceSelectItemImage}
        />
        <div>{item.label}</div>
        {!item.id && allowSelect ? renderItemConnection(item) : ''}
      </Checkbox>
    )
  }

  const renderContent = () => {
    return (
      <div
        className={cn(
          styles.externalDataSourceSelectWrapper,
          props.checkboxClassName,
        )}
      >
        {list
          .filter((item) => {
            return allowSelect || checkedList.includes(item.id!)
          })
          .map(renderItem)}
      </div>
    )
  }

  const handleClick = () => {
    if (disabled) {
      message.warning('功能开发中，敬请期待~')
      return
    }
  }

  return (
    <Popover
      placement='topLeft'
      arrow={false}
      content={disabled ? null : renderContent}
    >
      <div
        className={cn(
          styles.externalDataSourceSelectBtn,
          {
            [styles.externalDataSourceSelectBtnActive]: checkedList.length,
            [styles.externalDataSourceSelectBtnDisabled]: disabled,
          },
          className,
        )}
        style={style}
        onClick={handleClick}
      >
        <ExternalDataSourceIcon />
        {t('project.externalDataSource')}
      </div>
    </Popover>
  )
})
