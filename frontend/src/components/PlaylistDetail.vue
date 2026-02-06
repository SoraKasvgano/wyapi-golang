<template>
  <div class="playlist-detail">


    <!-- 歌曲列表 -->
    <div v-if="playlistData && totalTracks > 0" class="tracks-section">
      <div class="section-header">
        <div class="header-left">
          <h2>歌曲列表</h2>
          <span class="track-total">共 {{ totalTracks }} 首 (第 {{ currentPage }} 页，每页 {{ pageSize }} 首)</span>
        </div>
        <div class="header-right">
          
        </div>
      </div>
      


      <!-- 歌单信息 -->
      <div class="playlist-info-bar">
        <span class="info-item">{{ creatorLabel }}：{{ playlistData.creator }}</span>
        <span class="info-separator">•</span>
        <span class="info-item">歌曲数：{{ totalTracks }}</span>
      </div>
      
      <div class="tracks-list">
        <div 
          v-for="track in displayTracks" 
          :key="track.id"
          class="track-item"
          @click="selectTrack(track)"
          :class="{ 'selected': selectedTrack && selectedTrack.id === track.id }"
        >
          <img
            class="track-cover"
            :src="getCover(track)"
            alt="cover"
            loading="lazy"
            @error="onCoverError"
          />
          <div class="track-info">
            <div class="track-title-line">
              <span class="track-name">{{ track.name }}</span>
              <span class="track-artist" v-if="getArtist(track)">{{ getArtist(track) }}</span>
            </div>
          </div>
          <div class="track-actions">
            <el-button 
              type="primary" 
              size="small" 
              @click.stop="parseTrack(track)"
              :loading="parsingTrackId === track.id"
            >
              {{ parsingTrackId === track.id ? '解析中...' : '解析' }}
            </el-button>
          </div>
        </div>
      </div>
      
      
      
      <!-- 分页：上一页 / 下一页 按钮 -->
       <div class="pagination-section">
         <el-button size="small" @click="goPrevPage" :disabled="currentPage <= 1">上一页</el-button>
         <span class="page-info">第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
         <el-button size="small" type="primary" @click="goNextPage" :disabled="currentPage >= totalPages">下一页</el-button>
       </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>正在加载歌单信息...</span>
    </div>

    <!-- 错误状态 -->
    <div v-if="error" class="error-container">
      <el-alert
        :title="error"
        type="error"
        show-icon
        :closable="false"
      />
    </div>
  </div>
</template>

<script>
import { ref, computed } from 'vue'
import { ElMessage, ElButton, ElIcon } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'

// 移除批量下载相关依赖

export default {
  name: 'PlaylistDetail',
  components: {
    ElButton,
    ElIcon,
    Loading
  },
  props: {
    playlistInfo: {
      type: Object,
      required: true
    },
    displayTracks: {
      type: Array,
      default: () => []
    },
    currentPage: {
      type: Number,
      default: 1
    },
    pageSize: {
      type: Number,
      default: 20
    },
    totalTracks: {
      type: Number,
      default: 0
    },
    selectedQuality: {
      type: String,
      default: 'lossless'
    },
    settings: {
      type: Object,
      default: () => ({})
    }
  },
  emits: ['track-selected', 'track-parsed', 'page-change'],
  setup(props, { emit }) {
    const playlistData = computed(() => props.playlistInfo)
    const creatorLabel = computed(() => (props.playlistInfo?.isAlbum ? '作者' : '创建者'))
    const loading = ref(false)
    const error = ref('')
    const selectedTrack = ref(null)
    const parsingTrackId = ref(null)
    // 移除批量下载相关状态与函数

    // 获取封面地址（兼容多种字段）
    const getCover = (track) => {
      return (
        track?.picUrl ||
        track?.cover ||
        track?.al?.picUrl ||
        track?.album?.picUrl ||
        playlistData.value?.picUrl ||
        ''
      )
    }

    const onCoverError = (e) => {
      // 加载失败时使用一个透明占位，避免破图
      e.target.style.visibility = 'hidden'
    }

    // 获取歌手名称（兼容多种字段结构）
    const getArtist = (track) => {
      return (
        track?.artist ||
        track?.singer ||
        (typeof track?.artists === 'string' ? track.artists : null) ||
        (Array.isArray(track?.artists) && track.artists[0]?.name) ||
        (Array.isArray(track?.ar) && track.ar[0]?.name) ||
        ''
      )
    }

    // 选择歌曲
    const selectTrack = (track) => {
      selectedTrack.value = track
      emit('track-selected', track)
    }

    // 解析歌曲
    const parseTrack = async (track) => {
      parsingTrackId.value = track.id
      
      try {
        // 确保传递正确的音质参数，避免事件对象污染
        const qualityValue = typeof props.selectedQuality === 'string' ? props.selectedQuality : 'lossless'
        
        // 这里调用现有的解析逻辑，传递音质参数
        emit('track-parsed', { track, quality: qualityValue })
        ElMessage.success(`开始解析：${track.name}`)
      } catch {
        ElMessage.error('解析失败，请重试')
      } finally {
        parsingTrackId.value = null
      }
    }

    // 处理分页变化
    const handlePageChange = (page) => {
      emit('page-change', page)
    }


    // 格式化播放量
    const formatPlayCount = (count) => {
      if (count >= 100000000) {
        return Math.floor(count / 100000000) + '亿'
      } else if (count >= 10000) {
        return Math.floor(count / 10000) + '万'
      }
      return count.toString()
    }

    // 计算总页数
    const totalPages = computed(() => {
      const size = Number(props.pageSize) || 1
      const total = Number(props.totalTracks) || 0
      return Math.max(1, Math.ceil(total / size))
    })

    // 上一页 / 下一页
    const goPrevPage = () => {
      if (props.currentPage > 1) {
        handlePageChange(props.currentPage - 1)
      }
    }
    const goNextPage = () => {
      if (props.currentPage < totalPages.value) {
        handlePageChange(props.currentPage + 1)
      }
    }



    return {
      playlistData,
      creatorLabel,
      loading,
      error,
      selectedTrack,
      parsingTrackId,
      selectTrack,
      parseTrack,
      handlePageChange,
      formatPlayCount,
      getCover,
      onCoverError,
      getArtist,
      totalPages,
      goPrevPage,
      goNextPage,
    }
  }
}
</script>

<style>
/* 全局深色模式样式 - 不使用scoped以确保样式能够应用 */
.dark .tracks-section {
  background: #1a1a1a !important;
  color: #e5eaf3 !important;
}

.dark .section-header h2 {
  color: #e5eaf3 !important;
}

.dark .page-info {
  color: #cfd3dc !important;
}



.dark .track-total {
  color: #cfd3dc !important;
}

.dark .track-item {
  background: #1d1e1f !important;
  color: #e5eaf3 !important;
  border-color: #4c4d4f !important;
}

.dark .track-item:hover {
  background: #262727 !important;
}

.dark .track-item.selected {
  background: #1a1a2e !important;
  border-color: #409eff !important;
}

.dark .playlist-info-bar {
  background-color: #1d1e1f !important;
  color: #e5eaf3 !important;
}

.dark .info-item {
  color: #e5eaf3 !important;
}

.dark .info-separator {
  color: #909399 !important;
}

.dark .pagination-section {
  border-top-color: #4c4d4f !important;
}

.dark .section-header {
  border-bottom-color: #4c4d4f !important;
}

.dark .section-header h2 {
  color: #e5eaf3 !important;
}
</style>

<style scoped>
.playlist-detail {
  width: 100%;
}


.tracks-section {
  background: #fff;
  border-radius: 8px;
  padding: var(--space-4);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  /* 防止超长文本造成横向溢出影响布局 */
  overflow-x: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
  padding-bottom: 10px;
  border-bottom: 1px solid #ebeef5;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.section-header h2 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}

.track-total {
  color: #909399;
  font-size: 14px;
}

.tracks-list {
  display: flex;
  flex-direction: column;
}

.track-item {
  display: flex;
  align-items: center;
  padding: var(--space-3) var(--space-2);
  background: #fafafa;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  margin-bottom: var(--space-2);
  min-height: 60px;
  box-sizing: border-box;
}

.track-item:hover {
  background: #f0f9ff;
  transform: translateY(-1px);
}

.track-item.selected {
  background: #e1f3ff;
  border: 1px solid #409eff;
}



.track-info {
  flex: 1;
  min-width: 0;
}

.track-cover {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  object-fit: cover;
  margin-right: var(--space-2);
  background-color: #eee;
}

.track-title-line {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}


.track-name, .track-artist {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.track-name {
  flex: 1 1 auto;
  min-width: 0;
}

.track-artist {
  color: #606266;
  /* 限制歌手区域占比，避免撑开主列导致侧栏移位 */
  flex: 0 0 auto;
  max-width: 45%;
  min-width: 0;
}

.playlist-info-bar {
  background-color: #f8f9fa;
  padding: var(--space-2) var(--space-3);
  margin-bottom: var(--space-3);
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: 14px;
  flex-wrap: wrap;
}

.info-item {
  color: #606266;
  white-space: nowrap;
}

.info-separator {
  color: #909399;
  font-weight: bold;
}

.track-name {
  font-weight: 600;
  font-size: 16px;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.track-actions {
  margin-left: var(--space-2);
  flex-shrink: 0;
}

.track-actions .el-button {
  min-width: 60px;
  height: 32px;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #606266;
}

.loading-container .el-icon {
  font-size: 32px;
  margin-bottom: 15px;
}

.error-container {
  padding: 20px;
}

.pagination-section {
  margin-top: var(--space-4);
  padding: var(--space-3) 0;
  border-top: 1px solid #ebeef5;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: var(--space-2);
}

.page-info {
  color: #606266;
}

/* 循环解析模式样式 */

.dark .track-cover {
  background-color: #2a2a2a;
}

.dark .track-artist {
  color: #cfd3dc;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-2);
  }
  
  .header-left {
    width: 100%;
  }
  
  .header-right {
    width: 100%;
    justify-content: flex-start;
  }
  
  .track-item {
    padding: var(--space-2) var(--space-1);
    margin-bottom: var(--space-2);
  }
  .tracks-section {
    padding: var(--space-3);
  }
  .section-header {
    margin-bottom: var(--space-2);
  }
  .track-cover {
    margin-right: var(--space-1);
  }
  
  .playlist-info-bar {
    padding: var(--space-1) var(--space-2);
    margin-bottom: var(--space-2);
    font-size: 12px;
    gap: 6px;
  }
  
  .track-name {
    font-size: 14px;
  }
  
  .pagination-section {
    margin-top: var(--space-3);
    padding: var(--space-2) 0;
  }
  
  .pagination-section .el-pagination {
    font-size: 12px;
  }
}


</style>
