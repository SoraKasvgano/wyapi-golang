import { ref} from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import musicApi from '../services/musicApi.js'
import { allTracks, totalTracks, currentPage, displayTracks } from './paginationManager.js'
import { settings } from './settingsManager.js'

const getApiVersionLabel = () => {
  const v = settings?.apiVersion
  return v === 'API_V2' ? '接口2' : '接口1'
}
// 解析状态
export const musicUrl = ref('')
export const loading = ref(false)
export const musicInfo = ref(null)
export const playlistUrl = ref('')
export const playlistLoading = ref(false)
export const playlistInfo = ref(null)
export const albumUrl = ref('')
export const albumLoading = ref(false)
export const albumInfo = ref(null)

// 移除循环解析模式相关状态

// 计时器相关
export const parseStartTime = ref(0)
export const elapsedTime = ref(0)
let timerInterval = null

/**
 * 启动计时器
 */
const startTimer = () => {
  parseStartTime.value = Date.now()
  elapsedTime.value = 0
  
  timerInterval = setInterval(() => {
    elapsedTime.value = Math.floor((Date.now() - parseStartTime.value) / 1000)
  }, 100)
}

/**
 * 停止计时器
 */
const stopTimer = () => {
  if (timerInterval) {
    clearInterval(timerInterval)
    timerInterval = null
  }
}

/**
 * 清理计时器
 */
export const cleanupTimer = () => {
  stopTimer()
}

/**
 * 解析音乐信息
 * @param {string} selectedQuality 选择的音质
 */
export const parseMusic = async (selectedQuality) => {
  if (!musicUrl.value.trim()) {
    ElMessage.warning('请输入音乐链接')
    return
  }

  // 验证链接格式
  if (!musicApi.validateMusicUrl(musicUrl.value)) {
    ElMessage.error('请输入有效的网易云音乐链接')
    return
  }

  loading.value = true

  try {
    const result = await musicApi.parseMusicInfo(musicUrl.value, selectedQuality)
    musicInfo.value = result

    const getApiVersionLabel = () => {
      const v = settings?.apiVersion
      switch (v) {
        case 'API_V2':
          return '接口2'
        case 'API_V1':
        default:
          return '接口1'
      }
    }

    ElNotification({
      title: '解析成功',
      message: `成功解析歌曲：${result.name} (${result.qualityName}) ${getApiVersionLabel()}`,
      type: 'success'
    })
  } catch (error) {
    // 根据错误类型显示不同的提示信息
    let errorMessage = '解析失败，请检查链接是否正确'
    
    if (error.message && error.message.includes('API服务器暂时不可用')) {
      errorMessage = 'API服务器暂时不可用，请稍后重试。这可能是由于服务器维护或网络问题导致的。'
      ElMessage({
        message: errorMessage,
        type: 'warning',
        duration: 8000,
        showClose: true
      })
    } else if (error.message && (error.message.includes('付费') || error.message.includes('版权限制'))) {
      // detail 返回 free=4 且无法获取 url 的情况
      errorMessage = '该歌曲为付费或受版权限制，暂无法获取播放链接'
      ElMessage({
        message: errorMessage,
        type: 'error',
        duration: 6000,
        showClose: true
      })
    } else if (error.message && (error.message.includes('下架') || error.message.includes('无法获取'))) {
      // detail 成功但 url 未返回的统一提示
      errorMessage = '该歌曲已下架或者无法获取'
      ElMessage({
        message: errorMessage,
        type: 'error',
        duration: 6000,
        showClose: true
      })
    } else if (error.message && error.message.includes('网络连接失败')) {
      errorMessage = error.message
      ElMessage({
        message: errorMessage,
        type: 'error',
        duration: 6000,
        showClose: true
      })
  } else {
    ElMessage.error(error.message || errorMessage)
  }
  
} finally {
  loading.value = false
}
}

// 当前解析状态
export const currentParsingTrack = ref(null)
export const parsingProgress = ref({
  currentIndex: 0,
  totalTracks: 0,
  successfulTracks: []
})

/**
 * 解析歌单
 * @param {Function} updateDisplayTracks 更新显示歌曲的回调函数
 */
export const parsePlaylist = async (updateDisplayTracks) => {
  // 直接使用导入的ref对象，避免参数传递问题
  
  if (!playlistUrl.value.trim()) {
    ElMessage.warning('请输入歌单链接')
    return
  }

  // 验证歌单链接格式
  if (!musicApi.validatePlaylistUrl(playlistUrl.value)) {
    ElMessage.error('请输入有效的网易云音乐歌单链接')
    return
  }

  // 开始计时
  startTimer()

  playlistLoading.value = true
  currentParsingTrack.value = null
  parsingProgress.value = { currentIndex: 0, totalTracks: 0, successfulTracks: [] }
  
  try {
    // 调用 API 获取歌单详情
    const result = await musicApi.getPlaylistDetail(playlistUrl.value)
    
    if (result.success) {
      // 设置数据
      playlistInfo.value = result.data
      // 为每个track初始化parsed字段为false，表示尚未解析
      const tracks = (result.data.tracks || []).map(track => ({
        ...track,
        parsed: false
      }))
      allTracks.value = tracks
      
      totalTracks.value = allTracks.value.length
      
      // 重置分页并显示第一页
      currentPage.value = 1
      
      if (updateDisplayTracks) {
        updateDisplayTracks()
      }
      
      const totalTime = Math.floor((Date.now() - parseStartTime.value) / 1000)
      // 获取歌单信息时，成功解析数量应为 0（实际歌曲解析后才会统计）
      
      ElNotification({
        title: '歌单解析完成！',
        message: `共 ${totalTracks.value} 首歌曲，用时 ${totalTime} 秒 ${getApiVersionLabel()}`,
        type: 'success',
        duration: 5000
      })
    } else {
      throw new Error(result.error)
    }
  } catch (error) {
    // 根据错误类型显示不同的提示信息
    let errorMessage = '歌单解析失败，请检查链接是否正确'
    
    if (error.message && error.message.includes('API服务器暂时不可用')) {
      errorMessage = 'API服务器暂时不可用，请稍后重试。这可能是由于服务器维护或网络问题导致的。'
      ElMessage({
        message: errorMessage,
        type: 'warning',
        duration: 8000,
        showClose: true
      })
    } else if (error.message && error.message.includes('网络请求失败')) {
      errorMessage = error.message
      ElMessage({
        message: errorMessage,
        type: 'error',
        duration: 6000,
        showClose: true
      })
  } else {
    ElMessage.error(error.message || errorMessage)
  }
  
} finally {
  // 停止计时器
  stopTimer()
  playlistLoading.value = false
  currentParsingTrack.value = null
}
}

/**
 * 解析专辑
 * @param {Function} updateDisplayTracks 更新显示歌曲的回调函数
 */
export const parseAlbum = async (updateDisplayTracks) => {
  if (!albumUrl.value.trim()) {
    ElMessage.warning('请输入专辑链接')
    return
  }

  if (!musicApi.validateAlbumUrl(albumUrl.value)) {
    ElMessage.error('请输入有效的网易云音乐专辑链接')
    return
  }

  albumLoading.value = true
  startTimer()

  try {
    const result = await musicApi.getAlbumDetail(albumUrl.value)
    if (result.success) {
      albumInfo.value = result.data
      allTracks.value = result.data.tracks.map(track => ({ ...track, parsed: false }))
      totalTracks.value = allTracks.value.length
      currentPage.value = 1

      // 解析成功后立即更新显示
      if (updateDisplayTracks) {
        updateDisplayTracks()
      }
      
      ElNotification({
        title: '专辑解析成功！',
        message: `成功解析专辑：${result.data.name} ${getApiVersionLabel()}`,
        type: 'success'
      })
    } else {
      throw new Error(result.error)
    }
  } catch (error) {
    ElMessage.error(error.message || '专辑解析失败，请检查链接是否正确')
  } finally {
    stopTimer()
    albumLoading.value = false
  }
}

/**
 * 清空音乐解析结果
 */
export const clearMusicResult = () => {
  musicInfo.value = null
  musicUrl.value = ''
}

/**
 * 清空歌单解析结果
 */
export const clearPlaylistResult = () => {
  playlistInfo.value = null
  playlistUrl.value = ''
  allTracks.value = []
  displayTracks.value = []
  totalTracks.value = 0
  currentPage.value = 1
}

/**
 * 清空专辑解析结果
 */
export const clearAlbumResult = () => {
  albumInfo.value = null
  albumUrl.value = ''
  allTracks.value = []
  displayTracks.value = []
  totalTracks.value = 0
  currentPage.value = 1
}

/**
 * 继续解析下一页（循环解析模式）
 * @param {Function} updateDisplayTracks 更新显示歌曲的回调函数
 */
// 移除循环解析相关导出与逻辑

/**
 * 设置示例链接
 * @param {string} url 示例链接
 * @param {string} type 链接类型 ('music' 或 'playlist' 或 'album')
 */
export const setExampleUrl = (url, type) => {
  if (type === 'playlist') {
    playlistUrl.value = url
  } else if (type === 'album') {
    albumUrl.value = url
  } else {
    musicUrl.value = url
  }
}
