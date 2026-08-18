import { FC } from 'react'
import { cn } from '@/utils'
import { scroll } from '@/styles/scroll'
import { VersionForm } from './VersionForm'
import Header from './images/header.png'
import styles from './styles.module.scss'

const Page: FC = () => {
  return (
    <div className={cn('h-full w-full overflow-auto', styles.bg, scroll)}>
      <div className={cn('my-25 mx-auto w-[700px]', 'flex flex-col')}>
        <img src={Header} className='w-full rounded-t-4xl' />
        <div className='rounded-b-4xl bg-white'>
          <VersionForm />
        </div>
      </div>
    </div>
  )
}

export default Page
