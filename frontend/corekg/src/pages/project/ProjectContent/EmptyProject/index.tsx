import { FC, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useProject } from '../..'
import { ProjectInput } from '../ProjectInput'
import { RecommendQuestions } from './RecommendQuestions'
import { SearchResult } from './SearchResult'
import Icon from './images/icon.svg?react'
import styles from './styles.module.scss'

export const EmptyProject: FC = () => {
  const {
    data: { project_name },
    isOtherPage,
    defaultKnowBase,
  } = useProject()
  const { t } = useTranslation('pages')
  const { version, globalGreeting } = useDeployConfig()
  const inputRef = useRef<{ startQA: () => void }>(null)

  // 根据部署配置获取应用名称：custom 版本显示 TuringQuery，其他版本显示 CoreKG
  const app_name = version === 'custom' ? 'TuringQuery' : 'CoreKG'

  const [search, setSearch] = useState<string>()
  const [hidden, setHidden] = useState<boolean>(false)

  const handleQuestionClick = (question: string) => {
    if (inputRef?.current) {
      setSearch(question)
      setTimeout(() => {
        inputRef.current?.startQA()
      }, 0)
    }
  }

  return (
    <div
      className={cn('w-full h-full overflow-hidden flex flex-col', {
        'justify-between': isOtherPage,
      })}
    >
      {isOtherPage && (
        <RecommendQuestions
          file_id={defaultKnowBase!}
          hidden={hidden}
          setHidden={setHidden}
          onSelectQusetion={handleQuestionClick}
        />
      )}
      {!isOtherPage && (
        <>
          {/* 顶部空白区域，用于自适应分配上下间距 */}
          <div className='flex-1' />
          <Icon className='mx-auto' />
          <div
            className={cn(
              styles.title,
              'text-center whitespace-pre-line mx-auto mt-2.5 font-semibold',
            )}
          >
            {globalGreeting ??
              t('project.greeting', { project_name, app_name })}
          </div>
        </>
      )}
      <ProjectInput
        ref={inputRef}
        value={search}
        onChange={setSearch}
        className={cn('w-[850px] mx-auto', {
          'w-[90%]': isOtherPage,
          'mt-18': !isOtherPage,
        })}
      />
      {!isOtherPage && (
        <>
          {/* <div className={'mx-auto text-[#BBBBBB] text-xs mt-5'}>
            {t('project.warning')}
          </div> */}
          {/* 底部空白区域，用于自适应分配上下间距 */}
          <div className='flex-1' />
        </>
      )}
      {/* {!isOtherPage && (
        <SearchResult
          className='w-[850px]  flex-1 mx-auto my-2'
          search={search}
        />
      )} */}
    </div>
  )
}
