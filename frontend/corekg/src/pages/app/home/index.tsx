import { FC, useState, useEffect } from 'react'
import { Typography } from 'antd'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { DataMode } from './components/DataMode'
import { KnowledgeProvider } from './components/KnowledgeContext'
import { Mode, ModeSelect } from './components/ModeSelect'
import { QAMode } from './components/QAMode'
import { SearchMode } from './components/SearchMode'
import HelloIcon from './images/hello.svg'
import Bg from '/images/bg.jpg'

const Home: FC = () => {
  // 知识问答/知识库搜索
  const [mode, setMode] = useState<Mode>('QA')
  const [greeting, setGreeting] = useState('')
  const { t } = useTranslation('pages')

  // 获取动态问候语
  const getGreeting = useCallback((): string => {
    const hour = new Date().getHours()
    if (hour < 12) {
      return t('app.home.morningGreeting')
    } else if (hour < 14) {
      return t('app.home.noonGreeting')
    } else if (hour < 18) {
      return t('app.home.afternoonGreeting')
    } else {
      return t('app.home.eveningGreeting')
    }
  }, [t])

  useEffect(() => {
    setGreeting(getGreeting())
  }, [getGreeting])

  return (
    <div
      className='w-full h-full bg-transparent overflow-hidden relative'
      style={{
        backgroundImage: `url(${Bg})`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
        borderRadius: '4px',
      }}
    >
      {/* 欢迎区域 - 居中显示 */}
      <div className='absolute top-[8vh] left-1/2 -translate-x-1/2 flex flex-col gap-12 items-center justify-center'>
        {/* 星星图标 */}
        <div className='w-9 h-[37.828px] relative'>
          <img
            alt={t('app.home.starIcon')}
            className='block max-w-none size-full'
            src={HelloIcon}
          />
        </div>

        {/* 动态问候语 */}
        <Typography.Title
          level={2}
          className='text-center text-[#160211]'
          style={{
            fontFamily: 'Manrope, "Noto Sans JP", sans-serif',
            fontSize: '24px',
            margin: 0,
          }}
        >
          {greeting}
        </Typography.Title>
      </div>

      {/* 主要内容区域 */}
      <div
        className={cn(
          'absolute top-[28vh] left-1/2 -translate-x-1/2',
          'w-[50vw] flex flex-col gap-3 items-start',
        )}
      >
        {/* 功能模式选择按钮 - 输入框上方靠左 */}
        <div className='flex gap-2'>
          <ModeSelect value={mode} onChange={setMode} />
        </div>

        {/* 输入框区域 - 使用原有组件 */}
        <KnowledgeProvider>
          <QAMode hidden={mode !== 'QA'} />
        </KnowledgeProvider>
        <SearchMode hidden={mode !== 'search'} />
        <DataMode hidden={mode !== 'data'} />
      </div>
    </div>
  )
}

export default Home
