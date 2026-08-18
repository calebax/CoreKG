import { Skeleton } from 'antd'
import { createRoot } from 'react-dom/client'
import { initializeI18n } from '@/locales'
import '@/styles/index.scss'
import '@/utils/polyfills'
import App from './App.tsx'
import './tailwind.css'

const root = createRoot(document.getElementById('root')!)
root.render(<Skeleton active className='p-4' />)
initializeI18n()
  .then(() => {
    root.render(<App />)
  })
  .catch(() => {
    console.error('I18n load error')
    root.render(<App />)
  })
