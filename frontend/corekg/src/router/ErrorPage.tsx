import { useRouteError } from 'react-router-dom'

export const ErrorPage: React.FC = () => {
  // 这个组件也会被用在'*'路由
  const error: any = useRouteError()
  const message = error?.message
  if (
    typeof message === 'string' &&
    message.includes('Failed to fetch dynamically imported module')
  ) {
    // eslint-disable-next-line no-self-assign
    window.location.href = window.location.href
  }
  return (
    <div className='flex h-full w-full flex-col items-center justify-center'>
      <h1 className='mb-4 text-2xl'>出了点小问题(O_O)...</h1>

      <p>
        <i>错误信息：{error?.statusText || error?.message}</i>
      </p>

      <button
        onClick={() => (window.location.href = '/')}
        className='mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors'
      >
        返回首页
      </button>
    </div>
  )
}

export default ErrorPage
