// 示例链接 - 单曲
export const musicExampleLinks = [
  { url: 'https://music.163.com/song?id=1454730043', name: '赤伶', artist: '李玉刚' },
  { url: 'https://music.163.com/song?id=1335942780', name: '九万字', artist: '黄诗扶' },
  { url: 'https://music.163.com/song?id=110191', name: '暗里着迷', artist: '刘德华' }
]

// 示例链接 - 歌单
export const playlistExampleLinks = [
  { url: 'https://music.163.com/playlist?id=7358627467', name: '华语经典流行', creator: '网易云音乐' },
  { url: 'https://music.163.com/playlist?id=2884035', name: '那些回不去的年少时光', creator: '网易云音乐' },
  { url: 'https://music.163.com/playlist?id=19723756', name: '云音乐飙升榜', creator: '网易云音乐' }
]

// 示例链接 - 专辑
export const albumExampleLinks = [
  { url: 'https://music.163.com/#/album?id=10804', name: '第二天堂', artist: '林俊杰' },
  { url: 'https://music.163.com/#/album?id=3056951', name: '新地球 - 人 (Special Edition)', artist: '林俊杰' },
  { url: 'https://music.163.com/#/album?id=74956171', name: '进阶', artist: '林俊杰' }
]

/**
 * 根据当前视图获取示例链接
 * @param {string} currentView 当前视图类型 ('music', 'playlist', or 'album')
 * @returns {Array} 示例链接数组
 */
export const getExampleLinks = (currentView) => {
  if (currentView === 'playlist') {
    return playlistExampleLinks
  }
  if (currentView === 'album') {
    return albumExampleLinks
  }
  return musicExampleLinks
}

/**
 * 获取当前示例链接（需要在组件中传入currentView）
 * @param {string} currentView 当前视图类型
 * @returns {Array} 示例链接数组
 */
export const getCurrentExampleLinks = (currentView) => {
  return getExampleLinks(currentView)
}

/**
 * 获取随机示例链接
 * @param {string} type 链接类型 ('music', 'playlist', or 'album')
 * @returns {Object} 随机示例链接对象
 */
export const getRandomExampleLink = (type) => {
  let links
  if (type === 'playlist') {
    links = playlistExampleLinks
  } else if (type === 'album') {
    links = albumExampleLinks
  } else {
    links = musicExampleLinks
  }
  const randomIndex = Math.floor(Math.random() * links.length)
  return links[randomIndex]
}

/**
 * 获取所有示例链接
 * @returns {Object} 包含音乐、歌单和专辑示例链接的对象
 */
export const getAllExampleLinks = () => {
  return {
    music: musicExampleLinks,
    playlist: playlistExampleLinks,
    album: albumExampleLinks
  }
}

/**
 * 添加新的示例链接
 * @param {string} type 链接类型 ('music', 'playlist', or 'album')
 * @param {Object} linkData 链接数据对象
 */
export const addExampleLink = (type, linkData) => {
  if (type === 'playlist') {
    playlistExampleLinks.push(linkData)
  } else if (type === 'album') {
    albumExampleLinks.push(linkData)
  } else {
    musicExampleLinks.push(linkData)
  }
}

/**
 * 移除示例链接
 * @param {string} type 链接类型 ('music', 'playlist', or 'album')
 * @param {number} index 要移除的链接索引
 */
export const removeExampleLink = (type, index) => {
  if (type === 'playlist' && index >= 0 && index < playlistExampleLinks.length) {
    playlistExampleLinks.splice(index, 1)
  } else if (type === 'album' && index >= 0 && index < albumExampleLinks.length) {
    albumExampleLinks.splice(index, 1)
  } else if (type === 'music' && index >= 0 && index < musicExampleLinks.length) {
    musicExampleLinks.splice(index, 1)
  }
}

/**
 * 验证示例链接格式
 * @param {Object} linkData 链接数据对象
 * @param {string} type 链接类型 ('music', 'playlist', or 'album')
 * @returns {boolean} 是否为有效格式
 */
export const validateExampleLink = (linkData, type) => {
  if (!linkData || typeof linkData !== 'object') return false
  
  const hasUrl = typeof linkData.url === 'string' && linkData.url.trim() !== ''
  const hasName = typeof linkData.name === 'string' && linkData.name.trim() !== ''
  
  if (type === 'playlist') {
    const hasCreator = typeof linkData.creator === 'string' && linkData.creator.trim() !== ''
    return hasUrl && hasName && hasCreator
  } else { // 'music' or 'album'
    const hasArtist = typeof linkData.artist === 'string' && linkData.artist.trim() !== ''
    return hasUrl && hasName && hasArtist
  }
}