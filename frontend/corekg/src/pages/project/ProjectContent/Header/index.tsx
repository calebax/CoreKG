import { FC } from 'react'
import { Popover } from 'antd'
import { useTranslation } from 'react-i18next'
import { match } from 'ts-pattern'
import { cn } from '@/utils'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useProject } from '../..'
import { KnowledgeList, KnowledgeStatus } from '../Knowledge'
import Opo from './opo.svg?react'
import styles from './styles.module.scss'

export type Header = Style & {
  toggleCol: () => void
}
export const Header: FC<Header> = (props) => {
  const { toggleCol, className, style } = props
  const {
    data: { knowledge, project_name },
  } = useProject()
  const { t } = useTranslation('pages')
  const { version } = useDeployConfig()
  return (
    <div className={cn('h-15 bg-white flex', className)} style={style}>
      {/* 左侧按钮 */}
      <div
        className={cn(
          'cursor-pointer rounded-lg',
          'p-3 ml-6 my-3',
          'flex items-center justify-center',
          styles.opo_btn,
        )}
        onClick={toggleCol}
      >
        <Opo />
        <span className={styles.text}>
          {match(version)
            .with('saas', () => 'CoreKG AI')
            .with('custom', () => 'AI')
            .with('international', () => t('project.opoAI'))
            .exhaustive()}
        </span>
      </div>
      {/* <Popover
        arrow={false}
        content={
          <KnowledgeList
            items={knowledge}
            title={t('project.selectedKnowledge', {
              project_name,
              num: knowledge.length,
            })}
          />
        }
      >
        {knowledge.length > 0 ? (
          <KnowledgeStatus
            num={knowledge.length}
            firstKnowledgeName={knowledge?.[0]?.name}
            className='self-center ml-auto mr-2.5'
          />
        ) : null}
      </Popover> */}
    </div>
  )
}
