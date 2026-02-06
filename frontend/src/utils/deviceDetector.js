import { ref } from 'vue'

// 设备状态
export const isMobile = ref(false)

/**
 * 检测移动端设备
 * @returns {boolean} 是否为移动端设备
 */
export const detectMobile = () => {
  if (typeof window !== 'undefined') {
    const isMobileWidth = window.innerWidth <= 768
    const isMobileUA = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)

    return isMobileWidth || isMobileUA
  }
  return false
}

/**
 * 监听窗口大小变化
 */
export const handleResize = () => {
  isMobile.value = detectMobile()
}

/**
 * 初始化设备检测
 */
export const initDeviceDetection = () => {
  // 初始化移动端检测
  isMobile.value = detectMobile()
  
  // 添加窗口大小变化监听器
  if (typeof window !== 'undefined') {
    window.addEventListener('resize', handleResize)
  }
}

/**
 * 清理设备检测监听器
 */
export const cleanupDeviceDetection = () => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', handleResize)
  }
}

/**
 * 获取设备信息
 * @returns {Object} 设备信息对象
 */
export const getDeviceInfo = () => {
  if (typeof window === 'undefined') {
    return {
      isMobile: false,
      isTablet: false,
      isDesktop: true,
      screenWidth: 0,
      screenHeight: 0,
      userAgent: ''
    }
  }

  const width = window.innerWidth
  const height = window.innerHeight
  const userAgent = navigator.userAgent

  const isMobileDevice = detectMobile()
  const isTablet = width >= 768 && width <= 1024
  const isDesktop = width > 1024

  return {
    isMobile: isMobileDevice,
    isTablet,
    isDesktop,
    screenWidth: width,
    screenHeight: height,
    userAgent
  }
}

/**
 * 检测是否为触摸设备
 * @returns {boolean} 是否为触摸设备
 */
export const isTouchDevice = () => {
  if (typeof window === 'undefined') return false
  
  return 'ontouchstart' in window || 
         navigator.maxTouchPoints > 0 || 
         navigator.msMaxTouchPoints > 0
}

/**
 * 检测浏览器类型
 * @returns {string} 浏览器类型
 */
export const getBrowserType = () => {
  if (typeof window === 'undefined') return 'unknown'
  
  const userAgent = navigator.userAgent.toLowerCase()
  
  if (userAgent.includes('chrome')) return 'chrome'
  if (userAgent.includes('firefox')) return 'firefox'
  if (userAgent.includes('safari') && !userAgent.includes('chrome')) return 'safari'
  if (userAgent.includes('edge')) return 'edge'
  if (userAgent.includes('opera')) return 'opera'
  
  return 'unknown'
}

/**
 * 检测操作系统
 * @returns {string} 操作系统类型
 */
export const getOperatingSystem = () => {
  if (typeof window === 'undefined') return 'unknown'
  
  const userAgent = navigator.userAgent.toLowerCase()
  
  if (userAgent.includes('windows')) return 'windows'
  if (userAgent.includes('mac')) return 'macos'
  if (userAgent.includes('linux')) return 'linux'
  if (userAgent.includes('android')) return 'android'
  if (userAgent.includes('ios') || userAgent.includes('iphone') || userAgent.includes('ipad')) return 'ios'
  
  return 'unknown'
}