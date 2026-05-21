import { embedMetadata } from './metadataWriter.js'
import { saveBlob, ensureBlobType, getMimeByExtension, sanitizeFilename } from '../utils/downloadHelper.js'
import { settings } from '../utils/settingsManager.js'

// 音乐信息接口 (JSDoc 类型定义)
/**
 * @typedef {Object} MusicInfo
 * @property {string} id - 歌曲ID
 * @property {string} name - 歌曲名称
 * @property {string} artist - 歌手名称
 * @property {string} album - 专辑名称
 * @property {string} cover - 封面图片URL
 * @property {number} duration - 歌曲时长(毫秒)
 * @property {string} url - 音频文件URL
 * @property {string} [lrc] - 歌词内容
 */

// 创建 fetch 请求的通用配置（不再使用超时中断，避免长耗时请求被取消）
const createFetchOptions = (data = null) => {
  const options = {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    }
  }

  const token = settings?.apiToken
  if (token) {
    options.headers['X-API-Token'] = token
  }

  if (data) {
    options.body = JSON.stringify(data)
  }

  return options
}

const normalizeId = (value) => {
  if (value === null || value === undefined) return ''
  return String(value).trim()
}

const isOk = (payload) => payload && Number(payload.code) === 200

const getErrorMessage = (payload, fallback) => {
  return payload?.msg || payload?.message || fallback
}

const parseDurationToMs = (value) => {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }
  const text = String(value || '').trim()
  if (!text) return 0
  if (/^\d+$/.test(text)) {
    return Number(text)
  }
  const parts = text.split(':').map((item) => Number(item))
  if (parts.some((item) => Number.isNaN(item))) return 0
  if (parts.length === 2) {
    return (parts[0] * 60 + parts[1]) * 1000
  }
  if (parts.length === 3) {
    return ((parts[0] * 60 + parts[1]) * 60 + parts[2]) * 1000
  }
  return 0
}

const normalizeArtistNames = (value) => {
  if (Array.isArray(value)) {
    return value
      .map((item) => {
        if (typeof item === 'string') return item
        return item?.name || item?.nickname || ''
      })
      .filter(Boolean)
  }
  if (typeof value === 'string') {
    return value.split(/[\/,，]/).map((item) => item.trim()).filter(Boolean)
  }
  return []
}

const normalizeCreator = (value, fallback = '') => {
  if (typeof value === 'string') return value
  if (value && typeof value === 'object') {
    return value.nickname || value.name || value.nick || fallback
  }
  return fallback
}

const normalizeTrack = (track = {}) => {
  const artistNames = normalizeArtistNames(track.artists || track.artist || track.singer || track.ar)
  const albumName = typeof track.album === 'string' ? track.album : track.album?.name || track.al?.name || ''
  const picUrl = track.picUrl || track.picimg || track.cover || track.al?.picUrl || track.album?.picUrl || ''
  return {
    ...track,
    id: track.id,
    name: track.name || '',
    artist: track.artist || track.singer || artistNames.join('/'),
    singer: track.singer || track.artist || artistNames.join('/'),
    artists: typeof track.artists === 'string' ? track.artists : artistNames.join('/'),
    album: albumName,
    picUrl,
    duration: parseDurationToMs(track.duration || track.dt)
  }
}

// API 基础 URL：默认使用当前站点
const getApiBase = () => {
  try {
    if (typeof window !== 'undefined' && window.location && window.location.origin) {
      return window.location.origin
    }
  } catch {
    void 0
  }
  return ''
}

// 可用性检测与故障切换支持
const getPreferredBaseList = () => {
  const base = getApiBase()
  return base ? [base] : []
}

let resolvedBase = null
let lastResolveAt = 0
const baseResolveTTL = 5 * 60 * 1000 // 5分钟缓存

// 以 HEAD + no-cors 快速探测域名可达性（仅检测网络连通，不读取状态码）
const isReachable = async (base, timeoutMs = 2500) => {
  try {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), timeoutMs)
    // 使用 HEAD 并且 no-cors，避免跨域错误；成功返回即视为可达
    await fetch(base, { method: 'HEAD', mode: 'no-cors', signal: controller.signal })
    clearTimeout(timer)
    return true
  } catch {
    return false
  }
}

const resolveApiBaseAsync = async () => {
  if (resolvedBase && (Date.now() - lastResolveAt) < baseResolveTTL) {
    return resolvedBase
  }
  for (const base of getPreferredBaseList()) {
    if (await isReachable(base)) {
      resolvedBase = base
      lastResolveAt = Date.now()
      return resolvedBase
    }
  }
  // 都不可达时，仍返回首选，以便上层错误处理
  resolvedBase = getApiBase()
  lastResolveAt = Date.now()
  return resolvedBase
}

// 构建 API URL（拼接为绝对地址）
const buildApiUrl = (path) => {
  const clean = String(path).replace(/^\/+/, '')
  // 优先使用已解析的可用域，否则用当前设置域
  const base = resolvedBase || getApiBase()
  return `${base}/${clean}`
}

// 通用的 fetch 请求函数（含域名故障切换）
const fetchApi = async (url, data = null) => {
  const options = createFetchOptions(data)
  try {
    // 确保已解析可用域（异步懒加载，不阻塞过久）
    // 若解析失败，不影响首次请求，失败后会走下面的故障切换
    resolveApiBaseAsync().catch(() => {})

    const response = await fetch(url, options)

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()
    return { data: result }
  } catch (error) {
    // 若为绝对地址，则尝试其他可用域进行故障切换
    try {
      const isAbsolute = /^https?:\/\//i.test(url)
      const cleanPath = isAbsolute ? String(new URL(url).pathname).replace(/^\/+/, '') : String(url).replace(/^\/+/, '')
      const bases = getPreferredBaseList()
      for (const base of bases) {
        const candidate = `${base}/${cleanPath}`
        if (candidate === url) continue
        try {
          const resp = await fetch(candidate, options)
          if (resp.ok) {
            const result = await resp.json()
            // 更新已解析域，提升后续请求命中率
            resolvedBase = base
            lastResolveAt = Date.now()
            return { data: result }
          }
        } catch {
          // 忽略单次失败，尝试下一个候选域
        }
      }
    } catch {
      // 忽略解析候选域时的错误
    }
    // 所有候选域均失败，抛出原始错误
    throw error
  }
}

// 主动探测所有候选域的可用性（供设置页或调试使用）
 

// 从文本中提取URL
const extractUrlFromText = (text) => {
  if (!text) return text
  // 确保text是字符串类型
  if (typeof text !== 'string') {
    text = String(text)
  }
  const urlRegex = /(https?:\/\/[^\s"<>]+)/
  const match = text.match(urlRegex)
  return match ? match[0] : text
}

// 统一的ID提取函数 - 支持音乐和歌单链接
export const extractIdFromUrl = async (text) => {
  try {
    // 确保输入不为空且为字符串
    if (!text) return null
    if (typeof text !== 'string') {
      text = String(text)
    }
    
    const url = extractUrlFromText(text)
    if (!url) return null
    
    // 短链接模式 - 直接返回原链接让后端处理
    if (/https?:\/\/163cn\.tv\/([a-zA-Z0-9]+)/.test(url)) {
      return url
    }

    // 音乐链接模式
    const musicPatterns = [
      /https?:\/\/music\.163\.com\/song\?id=(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/song\/(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/song\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/#\/song\?id=(\d+)/
    ]
    
    // 歌单链接模式
    const playlistPatterns = [
      /https?:\/\/music\.163\.com\/playlist\?id=(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/playlist\?id=(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/playlist\/(\d+)/,
      /https?:\/\/music\.163\.com\/#\/playlist\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/discover\/toplist\?id=(\d+)/
    ]

    // 专辑链接模式
    const albumPatterns = [
      /https?:\/\/music\.163\.com\/album\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/album\/(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/album\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/#\/album\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/#\/album\/(\d+)/
    ]
    
    // 尝试从URL提取ID
    for (const pattern of [...musicPatterns, ...playlistPatterns, ...albumPatterns]) {
      const match = url.match(pattern)
      if (match) {
        return match[1]
      }
    }
    
    return null
  } catch {
    return null
  }
}

// 验证链接格式 - 统一处理音乐和歌单链接
export const validateUrl = (url) => {
  try {
    // 确保输入为字符串
    if (!url || typeof url !== 'string') {
      return false
    }
    
    const patterns = [
      // 音乐链接
      /https?:\/\/music\.163\.com\/song\?id=(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/song\/(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/song\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/#\/song\?id=(\d+)/,
      // 歌单链接
      /https?:\/\/music\.163\.com\/playlist\?id=(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/playlist\?id=(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/playlist\/(\d+)/,
      /https?:\/\/music\.163\.com\/#\/playlist\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/discover\/toplist\?id=(\d+)/,
      // 专辑链接
      /https?:\/\/music\.163\.com\/album\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/album\/(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/album\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/#\/album\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/#\/album\/(\d+)/,
      // 短链接 - 直接验证格式，不解析
      /https?:\/\/163cn\.tv\/([a-zA-Z0-9]+)/
    ]
    
    return patterns.some(pattern => pattern.test(url))
  } catch {
    return false
  }
}

// 严格类型校验：分别限定歌曲 / 歌单 / 专辑链接
export const validateMusicUrl = (url) => {
  try {
    const patterns = [
      /https?:\/\/music\.163\.com\/song\?id=(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/song\/(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/song\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/#\/song\?id=(\d+)/,
      // 支持网易云短链接（交由后端解析指向的具体类型）
      /https?:\/\/163cn\.tv\/([a-zA-Z0-9]+)/
    ]
    return patterns.some(p => p.test(url))
  } catch {
    return false
  }
}

export const validatePlaylistUrl = (url) => {
  try {
    const patterns = [
      /https?:\/\/music\.163\.com\/playlist\?id=(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/playlist\?id=(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/playlist\/(\d+)/,
      /https?:\/\/music\.163\.com\/#\/playlist\?id=(\d+)/,
      /https?:\/\/163cn\.tv\/([a-zA-Z0-9]+)/
    ]
    return patterns.some(p => p.test(url))
  } catch {
    return false
  }
}

export const validateAlbumUrl = (url) => {
  try {
    const patterns = [
      /https?:\/\/music\.163\.com\/album\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/album\/(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/album\?id=(\d+)/,
      /https?:\/\/y\.music\.163\.com\/m\/album\/(\d+)/,
      /https?:\/\/music\.163\.com\/#\/album\?id=(\d+)/,
      /https?:\/\/music\.163\.com\/#\/album\/(\d+)/,
      /https?:\/\/163cn\.tv\/([a-zA-Z0-9]+)/
    ]
    return patterns.some(p => p.test(url))
  } catch {
    return false
  }
}
export const getMusicIdFromUrl = extractIdFromUrl
export const extractPlaylistId = extractIdFromUrl
export const extractAlbumId = extractIdFromUrl

// 音质等级映射
export const QUALITY_LEVELS = {
  'jymaster': '超清母带(Master)',
  'dolby': '杜比全景声(Dolby Atmos)',
  'sky': '沉浸环绕声(Surround Audio)',
  'jyeffect': '高清臻音(Spatial Audio)',
  'hires': '高清晰度无损(Hi-Res)',
  'lossless': '无损(SQ)',
  'exhigh': '极高(HQ)',
  'standard': '标准(128k)'
}

const QUALITY_FALLBACK_ORDER = [
  'jymaster',
  'dolby',
  'sky',
  'jyeffect',
  'hires',
  'lossless',
  'exhigh',
  'standard'
]

const normalizeQuality = (value) => {
  if (typeof value !== 'string') return 'lossless'
  const trimmed = value.trim()
  return trimmed || 'lossless'
}

const buildQualityFallback = (preferred) => {
  const normalized = normalizeQuality(preferred)
  const startIndex = QUALITY_FALLBACK_ORDER.indexOf(normalized)
  if (startIndex === -1) {
    const defaults = ['lossless', 'exhigh', 'standard']
    return [normalized, ...defaults.filter((item) => item !== normalized)]
  }
  return QUALITY_FALLBACK_ORDER.slice(startIndex)
}

// 播放链接内存缓存（基于歌曲ID与音质），减少重复接口请求
const urlCache = new Map()
const getUrlCacheKey = (id, quality) => `${id}|${quality}`
const getCachedUrlData = (id, quality) => {
  try {
    const key = getUrlCacheKey(id, quality)
    const entry = urlCache.get(key)
    if (!entry) return null
    const ttlMin = Number(settings?.urlCacheTTLMinutes) || 15
    const ttlMs = ttlMin * 60 * 1000
    if (Date.now() - entry.fetchedAt > ttlMs) {
      urlCache.delete(key)
      return null
    }
    return entry.data
  } catch {
    return null
  }
}
const setCachedUrlData = (id, quality, data) => {
  urlCache.set(getUrlCacheKey(id, quality), { data, fetchedAt: Date.now() })
}

// 获取音乐播放链接
export const getMusicUrl = async (musicId, quality = 'lossless', options = {}) => {
  const { bypassCache = false, updateCache = true } = options
  // 先查缓存，避免每次都请求接口（可通过 bypassCache 跳过）
  if (settings?.enableUrlCache && !bypassCache) {
    const cached = getCachedUrlData(musicId, quality)
    if (cached && cached.url) {
      return cached
    }
  }

  const response = await fetchApi(buildApiUrl(`api/getSongUrl`), {
    id: normalizeId(musicId),
    level: quality
  })

  if (!isOk(response.data)) {
    throw new Error(getErrorMessage(response.data, '获取音乐链接失败'))
  }

  const urlData = Array.isArray(response.data.data) ? response.data.data[0] : response.data.data
  if (!urlData || !urlData.url) {
    throw new Error('该音质的音乐链接不可用')
  }
  // 写入缓存（可通过 updateCache 控制是否更新缓存）
  if (settings?.enableUrlCache && updateCache) {
    setCachedUrlData(musicId, quality, urlData)
  }

  return urlData
}

// 解析音乐信息
export const parseMusicInfo = async (url, quality = 'lossless') => {
  try {
    const musicId = await extractIdFromUrl(url)
    if (!musicId) {
      throw new Error('无法从链接中提取歌曲ID')
    }

    // 获取歌曲基本信息
    const detailResponse = await fetchApi(buildApiUrl(`api/getSongInfo`), {
      id: normalizeId(musicId)
    })
    
    if (!isOk(detailResponse.data)) {
      throw new Error(getErrorMessage(detailResponse.data, '获取歌曲信息失败'))
    }

    const songData = detailResponse.data.data
    
    // 转换时长格式
    const durationMs = parseDurationToMs(songData.duration)

    const preferredQuality = normalizeQuality(quality)

    // 获取音乐播放链接（若高音质不可用，则按顺序回退到可用音质）
    let musicUrl = null
    let actualQuality = 'standard'
    let fileSize = 0
    let bitRate = 0
    let fileType = 'mp3' // Default file type

    try {
      const fallbackLevels = buildQualityFallback(preferredQuality)
      let lastError = null
      let urlData = null
      let resolvedLevel = preferredQuality

      for (const level of fallbackLevels) {
        try {
          const candidate = await getMusicUrl(songData.id, level)
          if (candidate && candidate.url) {
            urlData = candidate
            resolvedLevel = level
            break
          }
        } catch (error) {
          lastError = error
        }
      }

      if (!urlData || !urlData.url) {
        if (lastError) {
          throw lastError
        }
        throw new Error('该歌曲已下架或者无法获取')
      }

      musicUrl = urlData.url
      actualQuality = urlData.level || resolvedLevel
      fileSize = urlData.size || 0
      bitRate = urlData.br || 0
      
      const flacQualities = ['lossless', 'hires', 'jymaster', 'sky', 'jyeffect']
      const returnedLevel = urlData.level || 'standard'

      if (urlData.type) {
        fileType = urlData.type.toLowerCase()
      } else {
        const urlExtensionMatch = musicUrl.match(/\.([a-zA-Z0-9]+)(?=\?|$)/)
        if (urlExtensionMatch && urlExtensionMatch[1]) {
          fileType = urlExtensionMatch[1].toLowerCase()
        } else if (flacQualities.includes(returnedLevel)) {
          fileType = 'flac'
        } else {
          fileType = 'mp3'
        }
      }
    } catch {
      throw new Error('该歌曲已下架或者无法获取')
    }

    // 构造返回的音乐信息
    const musicInfo = {
      id: songData.id.toString(),
      name: songData.name,
      artist: songData.singer,
      album: songData.album,
      cover: songData.picimg,
      duration: durationMs,
      url: musicUrl,
      quality: actualQuality,
      qualityName: QUALITY_LEVELS[actualQuality] || actualQuality,
      fileSize: fileSize,
      bitRate: bitRate,
      lrc: '',
      fileExtension: `.${fileType.toLowerCase()}`
    }

    // 尝试获取歌词
    try {
      const lyricsData = await getLyrics(songData.id)
      if (lyricsData && lyricsData.lrc) {
        musicInfo.lrc = lyricsData.lrc
        musicInfo.tlyric = lyricsData.tlyric || ''
        musicInfo.romalrc = lyricsData.romalrc || ''
        musicInfo.klyric = lyricsData.klyric || ''
      }
    } catch {
      void 0
    }

    // 尽量补充发行日期，供下载时写入元数据
    try {
      const wikiData = await getSongWiki(songData.id)
      if (wikiData && wikiData.publishTime) {
        const date = new Date(Number(wikiData.publishTime))
        if (!Number.isNaN(date.getTime())) {
          musicInfo.publishTime = wikiData.publishTime
          musicInfo.year = String(date.getFullYear())
        }
      }
    } catch {
      void 0
    }

    return musicInfo

  } catch (error) {
    
    if (error.response) {
      const status = error.response.status
      const data = error.response.data
      
      if (status === 404) {
        throw new Error('歌曲不存在或已被删除')
      } else if (status >= 500) {
        throw new Error('服务器暂时不可用，请稍后重试')
      } else if (data && data.msg) {
        throw new Error(data.msg)
      }
    } else if (error.request) {
      throw new Error('网络连接失败，请检查网络连接')
    }
    
    throw new Error(error.message || '解析失败，请检查链接是否正确或稍后重试')
  }
}

// 获取歌词
export const getLyrics = async (musicId) => {
  try {
    if (typeof musicId !== 'string') {
      if (musicId === null || musicId === undefined) {
        throw new Error('歌曲ID不能为空')
      }
      musicId = String(musicId)
    }
    
    const response = await fetchApi(buildApiUrl(`api/getSongLyric`), {
      id: normalizeId(musicId)
    })
    
    if (!isOk(response.data)) {
      throw new Error(getErrorMessage(response.data, '获取歌词失败'))
    }

    const lyricsData = response.data.data
    
    return {
      lrc: lyricsData.lrc || '',
      tlyric: lyricsData.tlyric || '',
      romalrc: lyricsData.romalrc || '',
      klyric: lyricsData.klyric || ''
    }
  } catch {
    return { 
      lrc: '',
      tlyric: '',
      romalrc: '',
      klyric: ''
    }
  }
}

export const getSongWiki = async (musicId) => {
  try {
    const response = await fetchApi(buildApiUrl(`api/song/wiki`), {
      id: normalizeId(musicId)
    })

    if (!isOk(response.data)) {
      throw new Error(getErrorMessage(response.data, '获取歌曲发行信息失败'))
    }

    return response.data.data || {}
  } catch {
    return {}
  }
}

// 下载音乐
export const downloadMusic = async (musicInfo, settings = {}) => {
  const {
    filenameFormat = 'song-artist',
    writeMetadata = false
  } = settings

  // 1. 确定文件名
  const extension = musicInfo.fileExtension || '.mp3'
  let filename
  if (filenameFormat === 'artist-song') {
    filename = `${musicInfo.artist} - ${musicInfo.name}${extension}`
  } else {
    filename = `${musicInfo.name} - ${musicInfo.artist}${extension}`
  }

  try {
    // 2. 使用 fetch 下载音频数据
    const response = await fetch(musicInfo.url, { cache: 'no-store', mode: 'cors' })
    if (!response.ok) {
      throw new Error(`下载音频文件失败: ${response.statusText}`)
    }
    let audioBuffer = await response.arrayBuffer()

    // 3. 如果启用，则嵌入元数据
    if (writeMetadata && (extension === '.mp3' || extension === '.flac')) {
      try {
        const metadata = {
          name: musicInfo.name,
          artist: musicInfo.artist,
          album: musicInfo.album,
          year: new Date().getFullYear().toString(),
          lyrics: musicInfo.lrc,
          cover: musicInfo.cover
        }
        audioBuffer = await embedMetadata(audioBuffer, metadata, extension)
      } catch {
        // 可选：通知用户元数据写入失败，但下载将继续
      }
    }

    // 4. 触发下载
    const mime = response.headers.get('Content-Type') || getMimeByExtension(extension)
    const typedBlob = ensureBlobType(new Blob([audioBuffer], { type: mime }), mime)
    saveBlob(typedBlob, sanitizeFilename(filename))

    return true
  } catch (error) {
    throw new Error(`下载失败: ${error.message}`)
  }
}

// 获取歌单详情
export const getPlaylistDetail = async (url) => {
  try {
    const musicId = await extractIdFromUrl(url)
    if (!musicId) {
      throw new Error('无法从链接中提取歌单ID')
    }

    const pageSize = 500
    let offset = 0
    let firstPage = null
    const allTracks = []

    for (;;) {
      const response = await fetchApi(buildApiUrl(`api/playlist_trackall`), {
        id: normalizeId(musicId),
        limit: pageSize,
        offset
      })

      if (!isOk(response.data)) {
        return {
          success: false,
          error: getErrorMessage(response.data, '获取歌单信息失败')
        }
      }

      const data = response.data.data || {}
      if (!firstPage) {
        firstPage = data
      }
      const tracks = Array.isArray(data.tracks)
        ? data.tracks
        : Array.isArray(data.songs)
          ? data.songs
          : Array.isArray(data.list)
            ? data.list
            : []
      allTracks.push(...tracks)

      if (!data.more || tracks.length === 0) {
        break
      }
      offset += tracks.length
    }

    const data = firstPage || {}
    return {
      success: true,
      data: {
        ...data,
        creator: normalizeCreator(data.creator, data.creatorName),
        picUrl: data.picUrl || data.coverImgUrl || '',
        coverImgUrl: data.coverImgUrl || data.picUrl || '',
        trackCount: data.trackCount || allTracks.length,
        tracks: allTracks.map(normalizeTrack)
      }
    }
  } catch (error) {
    
    if (error.message && error.message.includes('ERR_HTTP2_PROTOCOL_ERROR')) {
      return {
        success: false,
        error: 'API服务器暂时不可用，请稍后重试。这可能是由于服务器维护或网络问题导致的。'
      }
    }
    
    if (error.code === 'ENOTFOUND' || error.code === 'ECONNREFUSED') {
      return {
        success: false,
        error: 'API服务器无法连接，请检查网络连接或稍后重试'
      }
    }

    return {
      success: false,
      error: `网络请求失败: ${error.message || '未知错误'}，请稍后重试`
    }
  }
}

export const getToplists = async () => {
  try {
    const response = await fetchApi(buildApiUrl(`api/toplist`), {})
    if (isOk(response.data)) {
      return {
        success: true,
        data: Array.isArray(response.data.data) ? response.data.data : []
      }
    }
    return {
      success: false,
      error: getErrorMessage(response.data, '获取榜单列表失败')
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || '获取榜单列表失败'
    }
  }
}

// 获取专辑详情
export const getAlbumDetail = async (url) => {
  try {
    const albumId = await extractIdFromUrl(url)
    if (!albumId) {
      throw new Error('无法从链接中提取专辑ID')
    }
    
    const requestData = { id: normalizeId(albumId) }
    
    const response = await fetchApi(buildApiUrl(`api/getAlbum`), requestData)
    
    if (response.data && response.data.code === 200) {
      const data = response.data.data || {}
      const tracks = Array.isArray(data.tracks)
        ? data.tracks
        : Array.isArray(data.songs)
          ? data.songs
          : Array.isArray(data.list)
            ? data.list
            : []
      return {
        success: true,
        data: {
          ...data,
          artist: normalizeCreator(data.artist),
          picUrl: data.picUrl || data.coverImgUrl || '',
          coverImgUrl: data.coverImgUrl || data.picUrl || '',
          trackCount: data.trackCount || tracks.length,
          tracks: tracks.map(normalizeTrack)
        }
      }
    } else {
      return {
        success: false,
        error: getErrorMessage(response.data, '获取专辑信息失败')
      }
    }
  } catch (error) {
    return {
      success: false,
      error: `网络请求失败: ${error.message || '未知错误'}，请稍后重试`
    }
  }
}

// 获取单首歌曲信息
export const getMusicInfo = async (musicId) => {
  try {
    const endpoint = buildApiUrl(`api/getSongInfo`)
    const response = await fetchApi(endpoint, {
      id: normalizeId(musicId)
    })
    
    if (response.data && response.data.code === 200) {
      return {
        success: true,
        data: response.data.data
      }
    } else {
      return {
        success: false,
        error: getErrorMessage(response.data, '获取歌曲信息失败')
      }
    }
  } catch {
    return {
      success: false,
      error: '网络请求失败'
    }
  }
}

// 搜索音乐
export const searchMusic = async (keyword) => {
  try {
    const response = await fetchApi(buildApiUrl(`api/search`), {
      keyword,
      type: 1,
      limit: 20,
      offset: 0
    })
    
    if (isOk(response.data) && response.data.data) {
      const songs = Array.isArray(response.data.data)
        ? response.data.data
        : response.data.data.songs || response.data.data.list || []
      return {
        success: true,
        data: {
          songs: songs.map(normalizeTrack)
        }
      }
    } else {
      return {
        success: false,
        error: '搜索结果为空'
      }
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || '搜索失败'
    }
  }
}

export default {
  validateUrl,
  validateMusicUrl,
  validatePlaylistUrl,
  validateAlbumUrl,
  extractIdFromUrl,
  getMusicIdFromUrl,
  extractPlaylistId,
  extractAlbumId,
  parseMusicInfo,
  getLyrics,
  getSongWiki,
  downloadMusic,
  getPlaylistDetail,
  getToplists,
  getAlbumDetail,
  getMusicInfo,
  searchMusic
}
