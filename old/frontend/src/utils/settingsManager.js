import { reactive } from 'vue'

// 默认设置
export const defaultSettings = {
  filenameFormat: 'song-artist', // song-artist, artist-song, song
  writeMetadata: false, // 是否写入元数据
  zipDownload: false, // 是否压缩下载
  srtLyricsDownload: false, // 是否下载SRT歌词（否则下载LRC）
  layoutMode: 'single-column', // 布局模式: dual-column, single-column
  // 播放链接缓存设置
  enableUrlCache: false, // 是否启用播放链接缓存
  urlCacheTTLMinutes: 15, // 缓存时间（分钟） 
  // 极验验证码设置已移除
}

// 当前设置
export const settings = reactive({ ...defaultSettings })

/**
 * 加载设置
 */
export const loadSettings = () => {
  try {
    const savedSettings = localStorage.getItem('app-settings')
    if (savedSettings) {
      const parsed = JSON.parse(savedSettings)
      Object.assign(settings, { ...defaultSettings, ...parsed })
    }
  } catch {
    Object.assign(settings, defaultSettings)
  }
}

/**
 * 保存设置
 */
export const saveSettings = () => {
  try {
    localStorage.setItem('app-settings', JSON.stringify(settings))
  } catch {
    void 0
  }
}

/**
 * 更新设置
 * @param {Object} newSettings 新的设置对象
 */
export const updateSettings = (newSettings) => {
  Object.assign(settings, newSettings)
  saveSettings()
}

/**
 * 重置设置为默认值
 */
export const resetSettings = () => {
  Object.assign(settings, defaultSettings)
  saveSettings()
}

/**
 * 获取设置值
 * @param {string} key 设置键名
 * @returns {any} 设置值
 */
export const getSetting = (key) => {
  return settings[key]
}

/**
 * 设置单个配置项
 * @param {string} key 设置键名
 * @param {any} value 设置值
 */
export const setSetting = (key, value) => {
  settings[key] = value
  saveSettings()
}
