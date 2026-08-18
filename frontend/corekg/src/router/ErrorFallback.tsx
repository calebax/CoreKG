import React from 'react'

interface ErrorFallbackProps {
  error?: Error | null
  resetErrorBoundary?: () => void
}

const ErrorFallback: React.FC<ErrorFallbackProps> = ({
  error,
  resetErrorBoundary = () => window.location.reload(),
}) => {
  return (
    <div className='flex h-full w-full flex-col items-center justify-center'>
      <h1 className='mb-4 text-2xl'>出了点小问题(O_O)...</h1>

      {error && (
        <p className='mb-4'>
          <i>错误信息：{error.message}</i>
        </p>
      )}

      <button
        onClick={resetErrorBoundary}
        className='mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors'
      >
        刷新页面
      </button>
    </div>
  )
}

export default ErrorFallback
