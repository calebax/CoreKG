import { FC, PropsWithChildren, ReactNode } from 'react'
import { App, Avatar, Button, Empty, Popover } from 'antd'
import { useBoolean } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { deleteCustomModel, updateCustomModel } from '@/api'
import { cn } from '@/utils'
import { encryptPassword } from '@/utils/crypto'
import { ModelModal } from '../ModelModal'
import { ModelShowInfo } from '../type'
import styles from './styles.module.scss'

export const Models: FC<{ value?: ModelShowInfo[]; refresh: () => void }> = (
  props,
) => {
  const { value, refresh } = props
  if (!value || value.length === 0) return <Empty />
  return (
    <div
      className='grid gap-4'
      style={{
        gridTemplateColumns: 'repeat(auto-fill, minmax(330px, 1fr))',
      }}
    >
      {value.map((item) => {
        return <ModelItem key={item.ID} value={item} refresh={refresh} />
      })}
    </div>
  )
}

const ModelItem: FC<{ value: ModelShowInfo; refresh: () => void }> = (
  props,
) => {
  const { t: tM } = useTranslation('messages')
  const { t: tC } = useTranslation('common')
  const { t } = useTranslation('pages')
  const { value, refresh } = props
  const { ID, head_url, public_type, user_name, show_name, model_name } = value
  const { modal, message } = App.useApp()
  const [open, { toggle }] = useBoolean()
  const delModel = () => {
    modal.confirm({
      title: tM('confirmDeleteModel'),
      className: styles.confirmModal,
      onOk: async () => {
        await deleteCustomModel({ id: ID })
        message.success(tM('operationSuccess'))
        refresh()
      },
    })
  }
  return (
    <>
      <div
        className={cn('p-4 flex flex-col gap-2 rounded-xl', styles.modelItem)}
      >
        <div className='flex gap-2 items-center'>
          <Avatar src={head_url} size={'small'} />
          <span className='text-title'>{show_name}</span>
          {public_type === 'system' ? (
            <span className='text-primary ml-auto'>
              {t('settings.systemPreset')}
            </span>
          ) : null}
        </div>
        <div className='flex gap-2'>
          <div className='flex-1 flex flex-col'>
            <ModelBaseInfoItem
              label={t('settings.model', { after: t('settings.type') })}
            >
              {t('settings.model', { before: t('settings.largeLanguage') })}
            </ModelBaseInfoItem>
            <ModelBaseInfoItem
              label={t('settings.model', { before: t('settings.base') })}
            >
              {model_name}
            </ModelBaseInfoItem>
            <ModelBaseInfoItem label={t('settings.creator')}>
              {user_name}
            </ModelBaseInfoItem>
          </div>
          <Popover
            placement='bottom'
            content={
              <div className='flex flex-col p-2'>
                <Button
                  type='text'
                  onClick={toggle}
                  className={styles.operatorBtn}
                >
                  {tC('button.edit', { target: t('settings.model') })}
                </Button>
                <Button
                  type='text'
                  onClick={delModel}
                  className={styles.operatorBtn}
                >
                  {tC('button.delete', { target: t('settings.model') })}
                </Button>
              </div>
            }
          >
            <span
              className={cn('text-[#616373] cursor-pointer self-end', {
                hidden: public_type === 'system',
              })}
            >
              . . .
            </span>
          </Popover>
        </div>
      </div>
      {open ? (
        <ModelModal
          title={tC('button.edit', { target: t('settings.model') })}
          id={ID}
          onClose={toggle}
          onSubmit={async (val) => {
            if (val.api_key) {
              await updateCustomModel({
                id: ID,
                ...val,
                api_key: encryptPassword(val.api_key),
              })
              message.success(tM('operationSuccess'))
              refresh()
            }
          }}
        />
      ) : null}
    </>
  )
}

const ModelBaseInfoItem: FC<
  Style & { label: ReactNode } & PropsWithChildren
> = (props) => {
  const { label } = props
  return (
    <div
      className={cn('flex gap-2 items-center', props.className)}
      style={props.style}
    >
      <div className='text-[#616373] text-base min-w-16 flex-none'>{label}</div>
      <div className='flex-1 line-clamp-1'>{props.children}</div>
    </div>
  )
}
