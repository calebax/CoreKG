// NodeJS类型扩展
declare namespace NodeJS {
  interface ProcessEnv {
    NODE_ENV: 'development' | 'production' | 'test'
    VITE_APP_BASE_URL: string
    [key: string]: string | undefined
  }

  // 为Timeout类型提供声明
  type Timeout = ReturnType<typeof setTimeout>
}
