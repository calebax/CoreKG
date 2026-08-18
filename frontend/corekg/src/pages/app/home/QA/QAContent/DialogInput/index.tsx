import { FC, useState, useCallback, useMemo } from 'react'
import { Button, List, Popover, Tooltip } from 'antd'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import ArrowDown from '@/assets/icons/arrow-down.svg?react'
import InfoWarningIcon from '@/assets/icons/info-warning.svg?react'
import { DialogInput as GlobalDialogInput } from '@/components/dialog/DialogInput'
import { scroll } from '@/styles/scroll'
import { useQaInputMaxLength } from '@/utils/useQaInputMaxLength'
import { ModelSelect } from '../../../components/ModelSelect'
import SendIcon from '../../../images/send.svg'
import { QAData, SessionInfo } from '../../type'

type DialogInput = {
  className?: string
  /** 其他原因导致的loading */
  loading: boolean
  /** 正在回答 */
  isAnswering: boolean
  onSend: (data: QAData) => void
  onStop: () => void
  sessionInfo?: SessionInfo
}
export const DialogInput: FC<DialogInput> = (props) => {
  const {
    className,
    loading,
    isAnswering,
    onSend: _onSend,
    onStop,
    sessionInfo,
  } = props
  const { t: tC } = useTranslation('common')
  const { t } = useTranslation('pages')
  const qaInputMaxLength = useQaInputMaxLength()

  const [value, setValue] = useState('')
  const onSend = useCallback(() => {
    setValue('')
    _onSend({ text: value })
  }, [_onSend, value])
  const actionBtn = useMemo(() => {
    if (loading) {
      return (
        <div className='w-[24px] h-[24px] rounded flex items-center justify-center bg-[#dfe0eb] cursor-not-allowed'>
          <div className='relative w-4 h-4 flex items-center justify-center'>
            <img src={SendIcon} alt={tC('button.send')} />
          </div>
        </div>
      )
    }
    if (isAnswering) {
      return (
        <Tooltip title={tC('button.stopGenerating')}>
          <div
            className='w-[32px] h-[32px] rounded flex items-center justify-center bg-[#1e1f28] cursor-pointer rounded-full transition-colors hover:bg-[#2a2b36]'
            onClick={onStop}
          >
            <div className='relative w-4 h-4 flex items-center justify-center'>
              <div className='w-3 h-3 bg-white rounded-sm'></div>
            </div>
          </div>
        </Tooltip>
      )
    }
    if (!value) {
      return (
        <div className='w-[24px] h-[24px] rounded flex items-center justify-center bg-[#dfe0eb] cursor-not-allowed'>
          <div className='relative w-4 h-4 flex items-center justify-center'>
            <img src={SendIcon} alt={tC('button.send')} />
          </div>
        </div>
      )
    }
    return (
      <div
        className='w-[24px] h-[24px] rounded flex items-center justify-center bg-[#1e1f28] cursor-pointer transition-colors hover:bg-[#2a2b36]'
        onClick={onSend}
      >
        <div className='relative w-4 h-4 flex items-center justify-center'>
          <img src={SendIcon} alt={tC('button.send')} />
        </div>
      </div>
    )
  }, [isAnswering, loading, onSend, onStop, value, tC])

  const knowledgeWithLimit = useMemo((): SessionInfo['knowledge'] => {
    if (!sessionInfo) return []
    if (sessionInfo.knowledge.length <= 200) return sessionInfo.knowledge
    return sessionInfo.knowledge
      .slice(0, 200)
      .concat({ id: -1, name: t('app.home.maximumDisplay', { target: 200 }) })
  }, [sessionInfo, t])
  const knowledgeLabel = useMemo(() => {
    switch (sessionInfo?.type) {
      case 'forest':
        return t('app.home.knowledgeBase')
      case 'file_list':
        return t('app.home.file')
      case 'excel_list':
        return t('app.home.excel')
      case 'react_excel_list':
        return t('app.home.excelSheet')
      case 'db_list':
        return t('app.home.mySQLDatabase')
      case 'db_table_list':
        return t('app.home.mysqlTable')
    }
  }, [sessionInfo?.type, t])
  return (
    <GlobalDialogInput
      value={value}
      onChange={setValue}
      onSubmit={onSend}
      maxLength={qaInputMaxLength}
      className={className}
    >
      {/* 左侧控件组 */}
      <div className='flex items-center gap-2'>
        {sessionInfo ? (
          <>
            <Popover
              arrow={false}
              trigger={['hover']}
              placement='bottomLeft'
              styles={{ root: { padding: 0 }, body: { padding: 0 } }}
              content={
                <div className='bg-white rounded-[10px] shadow-[0px_2px_12px_0px_rgba(0,0,0,0.1)] p-[10px] w-[266px] flex flex-col gap-2'>
                  <div className='flex flex-col gap-1'>
                    <div className='flex items-center justify-between h-10 px-2 border-b border-[#eeeeee]'>
                      <div className='flex items-center gap-1'>
                        <span className='text-sm font-medium text-[#1e1f28]'>
                          {knowledgeLabel}
                          {t('app.home.list', {
                            target: sessionInfo.knowledge.length,
                          })}
                        </span>
                        <Tooltip
                          title={t('app.home.maximumDisplay', { target: 200 })}
                          placement='top'
                        >
                          <InfoWarningIcon className='w-4 h-4 cursor-pointer' />
                        </Tooltip>
                      </div>
                    </div>
                    <div className={cn('max-h-[28vh] overflow-auto', scroll)}>
                      <List
                        split={false}
                        dataSource={knowledgeWithLimit}
                        renderItem={(item, index) => (
                          <List.Item key={item.id} className='!py-0 !px-0'>
                            <div
                              className={cn(
                                'h-8 w-full text-sm text-[#1e1f28] px-2 cursor-default flex items-center',
                                index === 0
                                  ? 'bg-[#f8f9fd]'
                                  : 'hover:bg-[#f8f9fd] hover:cursor-pointer',
                              )}
                              style={{
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                                whiteSpace: 'nowrap',
                                minWidth: 0,
                              }}
                            >
                              <span className='truncate w-full'>
                                {item.name}
                              </span>
                            </div>
                          </List.Item>
                        )}
                      />
                    </div>
                  </div>
                </div>
              }
            >
              <Button
                // icon={<ArrowDown />}
                iconPosition='end'
                className={cn(
                  'w-auto rounded-[28px] px-2.5',
                  ' text-[#653ec4] border-none bg-[#F2F0FF]',
                  'hover:bg-[#c0b3ff] hover:text-[#341f68]',
                  'focus:bg-[#c0b3ff] focus:text-[#341f68]',
                )}
              >
                {t('app.home.linked', { target: knowledgeLabel })}
              </Button>
            </Popover>
            <ModelSelect value={sessionInfo.modelName ?? sessionInfo.model} />
          </>
        ) : null}
      </div>

      {/* 右侧发送按钮 */}
      <div className='flex items-center'>{actionBtn}</div>
    </GlobalDialogInput>
  )
}
