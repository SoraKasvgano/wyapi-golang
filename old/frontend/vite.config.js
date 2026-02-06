import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import compression from 'vite-plugin-compression'
import { compression as compression2 } from 'vite-plugin-compression2'
import { visualizer } from 'rollup-plugin-visualizer'

// https://vite.dev/config/
export default defineConfig({
  // 设置基础路径，支持子目录部署
  base: '/',
  // 定义全局变量，用于外部化依赖
  define: {
    __VUE_OPTIONS_API__: true,
    __VUE_PROD_DEVTOOLS__: false
  },
  plugins: [
    vue(),
    vueDevTools(),
    compression({
      algorithm: 'gzip',
      ext: '.gz',
      threshold: 1024,
      deleteOriginFile: false
    }),
    compression2({
      algorithms: ['brotliCompress'],
      threshold: 1024,
      deleteOriginFile: false
    }),
    visualizer({
      filename: 'dist/bundle-analysis.html',
      open: true,
      gzipSize: true,
      brotliSize: true
    })
  ],
  server: {
    host: '0.0.0.0', // 允许局域网访问
    port: 5173,      // 指定端口
    open: true,       // 自动打开浏览器
    proxy: {
      '/wyapi2': {
        target: 'https://wyapi-2.toubiec.cn',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/wyapi2/, '')
      },
      '/wyapi': {
        target: 'https://wyapi-1.toubiec.cn',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/wyapi/, '')
      },
      '/netease': {
        target: 'https://netease-cloud-music-api-psi-six-56.vercel.app',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/netease/, '')
      }
    }
  },
  // 预览模式配置
  preview: {
    host: '0.0.0.0', // 允许局域网访问
    port: 4173,      // 预览端口
    open: true       // 自动打开浏览器
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  build: {
    // 启用代码压缩 - 使用 terser 进行最佳压缩
    minify: 'terser',
    // 启用CSS代码分割，优化CSS文件大小
    cssCodeSplit: true,
    // Terser 压缩选项
    terserOptions: {
      compress: {
        // 移除 console.log
        drop_console: true,
        // 移除 debugger
        drop_debugger: true,
        // 移除未使用的代码
        dead_code: true,
        // 压缩条件表达式
        conditionals: true,
        // 压缩布尔值
        booleans: true,
        // 移除未使用的变量
        unused: true,
        // 内联函数
        inline: 2,
        // 压缩对象属性
        properties: true,
        // 移除无用的分号
        sequences: true
      },
      mangle: {
        // 混淆变量名
        toplevel: true,
        // 保留函数名（可选）
        keep_fnames: false
      },
      format: {
        // 移除注释
        comments: false
      }
    },
    // 启用代码分割优化打包大小
    rollupOptions: {
      output: {
        // 添加哈希值避免缓存问题
        entryFileNames: 'assets/js/[name].[hash].js',
        chunkFileNames: 'assets/js/[name].[hash].js',
        assetFileNames: 'assets/[name].[hash].[ext]',
        // 全局变量映射
        globals: {
          'vue': 'Vue',
          'element-plus': 'ElementPlus',
          'axios': 'axios'
        },
        // 优化的代码分割策略 - 将所有主要依赖分割成独立chunk
        manualChunks: (id) => {
          // Vue核心
          if (id.includes('node_modules/vue/') || id.includes('node_modules/@vue/')) {
            return 'vue';
          }
          // Element Plus UI库
          if (id.includes('node_modules/element-plus/')) {
            return 'element-plus';
          }
          // HTTP请求库
          if (id.includes('node_modules/axios/')) {
            return 'axios';
          }
          // 音频处理库
          if (id.includes('jszip') || id.includes('node-id3')) {
            return 'audio-libs';
          }
          // 工具库
          if (id.includes('lodash-es') || id.includes('dayjs')) {
            return 'utils';
          }
          // 加密和文件处理
          if (id.includes('crypto-js') || id.includes('file-saver')) {
            return 'crypto-file';
          }
          // 其他第三方库
          if (id.includes('node_modules/')) {
            return 'vendor';
          }
        }
      },
      // 禁用外部化，保持本地打包但优化代码分割
      // external: [
      //   'vue',
      //   'element-plus', 
      //   'axios'
      // ],
      // 插件优化 - 确保所有chunk都在HTML中预加载
      plugins: []
    },
     // 启用源码映射（生产环境可关闭）
    sourcemap: false,
    // 设置目标浏览器
    target: 'es2015',
    // 报告压缩后的文件大小
    reportCompressedSize: true,
    // 设置 chunk 大小警告限制 (KB)
    chunkSizeWarningLimit: 500
  },
})
