import { ref } from 'vue'

// 主题状态
const isDark = ref(false)

/**
 * 应用主题（带动画效果）
 * @param {boolean} dark 是否为深色主题
 */
export const applyTheme = (dark) => {
  if (typeof document !== 'undefined') {
    const body = document.body
    
    // 添加主题切换动画类
    body.classList.add('theme-switching')
    
    // 延迟应用主题，让动画效果更自然
    setTimeout(() => {
      if (dark) {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
      }
    }, 50)
    
    // 动画完成后移除动画类
    setTimeout(() => {
      body.classList.remove('theme-switching')
    }, 600)
  }
}

/**
 * 切换主题
 */
export const toggleTheme = () => {
  isDark.value = !isDark.value
  applyTheme(isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// 导出主题状态
export { isDark }

// 初始化主题（基于本地存储，不跟随系统）
export const initThemeFromLocalStorage = () => {
  const savedTheme = typeof window !== 'undefined' ? localStorage.getItem('theme') : null
  const dark = savedTheme === 'dark'
  isDark.value = dark
  applyTheme(dark)
}