import { ref } from 'vue'

// 分页相关状态
export const currentPage = ref(1)
export const pageSize = ref(20)
export const totalTracks = ref(0)
export const allTracks = ref([]) // 存储所有歌曲数据
export const displayTracks = ref([]) // 当前页显示的歌曲

/**
 * 更新当前页显示的歌曲
 */
export const updateDisplayTracks = () => {
  const startIndex = (currentPage.value - 1) * pageSize.value
  const endIndex = startIndex + pageSize.value
  displayTracks.value = allTracks.value.slice(startIndex, endIndex)
}

/**
 * 处理分页变化
 * @param {number} page 新的页码
 */
export const handlePageChange = (page) => {
  currentPage.value = page
  updateDisplayTracks()
}

/**
 * 设置每页显示数量
 * @param {number} size 每页显示的数量
 */
export const setPageSize = (size) => {
  pageSize.value = size
  // 重新计算当前页，确保不超出范围
  const maxPage = Math.ceil(totalTracks.value / pageSize.value)
  if (currentPage.value > maxPage) {
    currentPage.value = Math.max(1, maxPage)
  }
  updateDisplayTracks()
}

/**
 * 重置分页状态
 */
export const resetPagination = () => {
  currentPage.value = 1
  totalTracks.value = 0
  allTracks.value = []
  displayTracks.value = []
}

/**
 * 设置歌曲数据
 * @param {Array} tracks 歌曲数组
 */
export const setTracks = (tracks) => {
  allTracks.value = tracks || []
  totalTracks.value = allTracks.value.length
  currentPage.value = 1
  updateDisplayTracks()
}

/**
 * 获取分页信息
 * @returns {Object} 分页信息对象
 */
export const getPaginationInfo = () => {
  const totalPages = Math.ceil(totalTracks.value / pageSize.value)
  const startIndex = (currentPage.value - 1) * pageSize.value + 1
  const endIndex = Math.min(currentPage.value * pageSize.value, totalTracks.value)
  
  return {
    currentPage: currentPage.value,
    pageSize: pageSize.value,
    totalTracks: totalTracks.value,
    totalPages,
    startIndex,
    endIndex,
    hasNextPage: currentPage.value < totalPages,
    hasPrevPage: currentPage.value > 1
  }
}

/**
 * 跳转到指定页
 * @param {number} page 目标页码
 */
export const goToPage = (page) => {
  const totalPages = Math.ceil(totalTracks.value / pageSize.value)
  const targetPage = Math.max(1, Math.min(page, totalPages))
  handlePageChange(targetPage)
}

/**
 * 跳转到下一页
 */
export const nextPage = () => {
  const totalPages = Math.ceil(totalTracks.value / pageSize.value)
  if (currentPage.value < totalPages) {
    handlePageChange(currentPage.value + 1)
  }
}

/**
 * 跳转到上一页
 */
export const prevPage = () => {
  if (currentPage.value > 1) {
    handlePageChange(currentPage.value - 1)
  }
}