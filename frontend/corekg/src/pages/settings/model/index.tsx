import { FC } from 'react'
import { useNavigate } from 'react-router-dom'
import { Link } from 'react-router-dom'
import { App, Button, Skeleton } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useBoolean, useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { createCustomModel, listCustomModel } from '@/api'
import { cn } from '@/utils'
import { encryptPassword } from '@/utils/crypto'
import { ModelModal } from './components/ModelModal'
import { Models } from './components/Models'

/** 模型管理 */
const Model: FC = () => {
  const navigate = useNavigate()
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const { t: tC } = useTranslation('common')
  const [open, { toggle }] = useBoolean()
  const { message } = App.useApp()
  const { data, loading, refresh } = useRequest(async () => {
    const { Data } = await listCustomModel()
    const res: any[] = Data ?? []
    return res
  })
  return (
    <>
      <div className={cn('w-full h-full overflow-hidden', 'flex flex-col')}>
        <Button
          type='primary'
          className='self-start mb-2 mt-4 ml-4'
          onClick={toggle}
          icon={<PlusOutlined />}
        >
          {tC('button.add', { target: t('settings.model') })}
        </Button>
        <span className='text-title text-base font-semibold ml-4'>
          {t('settings.model', { after: tC('custom.list') })}
        </span>

        <div className='flex-1 overflow-auto pt-4 px-4'>
          {loading ? (
            <Skeleton active />
          ) : (
            <Models value={data} refresh={refresh} />
          )}
        </div>
      </div>
      {open ? (
        <ModelModal
          title={tC('button.add', { target: t('settings.model') })}
          onClose={toggle}
          onSubmit={async (val) => {
            if (val.api_key) {
              const encryptedVal = {
                ...val,
                api_key: encryptPassword(val.api_key),
              }
              await createCustomModel(encryptedVal as any)
              message.success(tM('createSuccess'))
              refresh()
            }
          }}
        />
      ) : null}
    </>
  )
}

export default Model
