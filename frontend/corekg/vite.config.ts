import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react-swc'
import { fileURLToPath, URL } from 'node:url'
import { visualizer } from 'rollup-plugin-visualizer'
import AutoImport from 'unplugin-auto-import/vite'
import { defineConfig } from 'vite'
import svgr from 'vite-plugin-svgr'

// https://vite.dev/config/
export default defineConfig((options) => {
  const rollupPlugins = []
  if (options.mode === 'test') {
    // test build时 增加一个分析插件
    rollupPlugins.push(
      visualizer({
        filename: './dist/stats.html',
        gzipSize: true,
        brotliSize: true,
      }),
    )
  }
  return {
    define: {
      // 定义全局变量，支持process.env
      'process.env.DEPLOY_MODE': JSON.stringify(process.env.DEPLOY_MODE),
    },
    plugins: [
      react(),
      svgr(),
      tailwindcss(),
      AutoImport({
        imports: ['react', 'react-router-dom'],
        // 可以自动导入antd组件
        resolvers: [
          // 如需自动导入antd组件，可添加对应resolver
        ],
        dirs: ['./api'],
        dts: './src/auto-imports.d.ts', // 生成typescript声明
      }),
      {
        // 为script标签引入查询参数
        name: 'js-query',
        transformIndexHtml: (html) => {
          const ts = Date.now()
          return html.replace(
            /<script\s+[^>]*src=["'](\/[^"']+\.js(\?[^"'])*)["'][^>]*><\/script>/g,
            (match, src, query) => {
              // query是目前的查询参数
              return match.replace(src, `${src}${query ? '&' : '?'}${ts}`)
            },
          )
        },
      },
    ],
    build: {
      rollupOptions: {
        plugins: rollupPlugins,
      },
    },
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        Graph: fileURLToPath(
          new URL('./src/pages/graph/common.ts', import.meta.url),
        ),
        Personnel: fileURLToPath(
          new URL('./src/pages/personnel/common.ts', import.meta.url),
        ),
      },
    },
    base: '/',
    server: {
      host: '127.0.0.1',
      proxy: {
        '/v2': {
          target: 'https://tapi.example.com',
          changeOrigin: true,
        },
        '/v3': {
          target: 'https://tapi.example.com',
          changeOrigin: true,
        },
        // Coze 本地服务（默认 8088），勿把具体 space 路径写进 target
        '/coze': {
          target: 'http://localhost:8088',
          changeOrigin: true,
        },
        // Coze 构建产物使用根路径绝对引用，iframe 嵌入时需转发到 Coze 服务
        '^/vendors-.*\\.js$': {
          target: 'http://localhost:8088',
          changeOrigin: true,
        },
        '^/index~\\d+\\.js$': {
          target: 'http://localhost:8088',
          changeOrigin: true,
        },
      },
      port: 3001,
      hmr: {
        overlay: true,
      },
      allowedHosts: ['corekg.com', 'example.com', 'dev.ckeyer.com'],
    },
  }
})
