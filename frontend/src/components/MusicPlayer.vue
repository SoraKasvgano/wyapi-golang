<template>
  <el-card class="result-card" shadow="hover">
    <div class="info-container">
      <h2 class="song-title">{{ musicInfo.name }}</h2>
      <p class="artist-name">{{ musicInfo.artist }}</p>
      <div class="metadata-grid">
        <span class="meta-label">专辑</span>
        <span class="meta-value">{{ musicInfo.album }}</span>
        <span class="meta-label">时长</span>
        <span class="meta-value">{{ formatDuration(musicInfo.duration) }}</span>
        <span class="meta-label">音质</span>
        <span class="meta-value">{{ musicInfo.qualityName }}</span>
        <span class="meta-label">大小</span>
        <span class="meta-value">{{ formatFileSize(musicInfo.fileSize) }}</span>
      </div>
    </div>
    <div id="aplayer-container" class="aplayer-instance"></div>
    <div class="direct-link-section">
      <el-input :model-value="musicInfo.url" readonly :size="isMobile ? 'small' : undefined">
        <template #prepend>
          <el-icon><Link /></el-icon>
        </template>
      </el-input>
    </div>
    
    <div class="action-section">
      <el-button 
        type="primary" 
        @click="downloadMusic"
        :loading="downloading"
        :disabled="!musicInfo.id"
        class="download-btn"
        :size="isMobile ? 'default' : 'large'"
      >
        <el-icon><Download /></el-icon>
        <span class="btn-text">{{ downloading ? progressText : '下载音乐包' }}</span>
      </el-button>
    </div>
    
    <!-- 进度条容器 - 移动到按钮下方 -->
    <div v-if="downloading" class="progress-container">
      <div class="progress-info">
        <span class="progress-text">{{ progressText }}</span>
        <span class="progress-percentage">{{ downloadProgress }}%</span>
      </div>
      <el-progress
        :percentage="downloadProgress"
        :status="progressStatus"
        :stroke-width="8"
        :show-text="false"
        style="width: 100%;"
      />
    </div>
  </el-card>
</template>

<script setup>
import { ref, watch, nextTick, computed } from 'vue';
import { Link, Download } from '@element-plus/icons-vue';
import { ElNotification, ElMessage } from 'element-plus';
import { embedMetadata } from '../services/metadataWriter.js';
import { settings } from '../utils/settingsManager.js'
import { lrcToSrt } from '../utils/lyricsConverter.js'
import { isDark } from '../utils/themeManager.js'
import { saveBlob, ensureBlobType, getMimeByExtension, sanitizeFilename } from '../utils/downloadHelper.js'
import musicApi from '../services/musicApi.js'

const props = defineProps({
  musicInfo: {
    type: Object,
    required: true
  }
});

const ap = ref(null);
const downloading = ref(false);
const downloadProgress = ref(0);
const progressStatus = ref('');
const progressText = ref('准备下载');

// 移动端判断（用于控制按钮与输入框尺寸、间距）
const isMobile = computed(() => typeof window !== 'undefined' && window.innerWidth <= 768);

// 已移除与杜比禁用相关的计算属性

// 获取格式化的文件名
const getFormattedFilename = (songName, artistName, extension) => {
  const sanitizedName = songName.replace(/[<>:"/\\|?*]/g, '_');
  const sanitizedArtist = artistName.replace(/[<>:"/\\|?*]/g, '_');
  
  // 从设置中读取文件命名格式
  const filenameFormat = settings.filenameFormat || 'song-artist';
  
  if (filenameFormat === 'artist-song') {
    return `${sanitizedArtist} - ${sanitizedName}${extension}`;
  } else if (filenameFormat === 'song-only') {
    // 仅歌曲名
    return `${sanitizedName}${extension}`;
  } else {
    // 默认格式：song-artist
    return `${sanitizedName} - ${sanitizedArtist}${extension}`;
  }
};

// 解析歌词时间戳和文本
const lrctrim = (lyrics) => {
  const lines = lyrics.split('\n');
  const data = [];

  lines.forEach((line, index) => {
    const matches = line.match(/\[(\d{2}):(\d{2}[.:]?\d*)\]/);
    if (matches) {
      const minutes = parseInt(matches[1], 10);
      const seconds = parseFloat(matches[2].replace('.', ':')) || 0;
      const timestamp = minutes * 60000 + seconds * 1000;

      let text = line.replace(/\[\d{2}:\d{2}[.:]?\d*\]/g, '').trim();
      text = text.replace(/\s\s+/g, ' '); // Replace multiple spaces with a single space

      data.push([timestamp, index, text]);
    }
  });

  data.sort((a, b) => a[0] - b[0]);

  return data;
};

// 合并原文歌词和翻译歌词
const lrctran = (lyric, tlyric) => {
  const lyricData = lrctrim(lyric);
  const tlyricData = lrctrim(tlyric);

  const len1 = lyricData.length;
  const len2 = tlyricData.length;
  let result = "";

  for (let i = 0, j = 0; i < len1 && j < len2; i++) {
    while (lyricData[i][0] > tlyricData[j][0] && j + 1 < len2) {
      j++;
    }

    if (lyricData[i][0] === tlyricData[j][0]) {
      tlyricData[j][2] = tlyricData[j][2].replace('/', '');
      if (tlyricData[j][2]) {
        lyricData[i][2] += ` (翻译：${tlyricData[j][2]})`;
      }
      j++;
    }
  }

  for (let i = 0; i < len1; i++) {
    const t = lyricData[i][0];
    result += `[${String(Math.floor(t / 60000)).padStart(2, '0')}:${String(Math.floor((t % 60000) / 1000)).padStart(2, '0')}.${String(t % 1000).padStart(3, '0')}]${lyricData[i][2]}\n`;
  }

  return result;
};

// 兼容旧函数名
const mergeLyrics = (originalLrc, translationLrc) => {
  if (!originalLrc) return '';
  if (!translationLrc) return originalLrc;
  return lrctran(originalLrc, translationLrc);
};

const downloadMusic = async () => {
    if (!props.musicInfo?.url) return;

    // 已移除杜比音质拦截，允许下载流程继续

    downloading.value = true;
    downloadProgress.value = 0;
    progressStatus.value = '';
    
    // 检查是否启用ZIP打包下载
    const zipDownloadEnabled = settings.zipDownload;
    
    if (!zipDownloadEnabled) {
        // 普通下载模式
        progressText.value = '准备下载';
        try {
            downloadProgress.value = 10;
            
            // 下载音频文件
            progressText.value = '下载音频文件';
            let currentUrl = props.musicInfo.url;
            let audioResponse = await fetch(currentUrl);
            if (!audioResponse.ok) {
                // 首次下载失败，尝试刷新播放链接后重试一次
                try {
                  const refreshed = await musicApi.getMusicUrl(
                    props.musicInfo.id,
                    props.musicInfo.quality || 'lossless',
                    { bypassCache: true }
                  )
                  if (refreshed && refreshed.url) {
                    currentUrl = refreshed.url
                    audioResponse = await fetch(currentUrl)
                  }
                } catch { void 0 }
            }
            if (!audioResponse.ok) {
              throw new Error(`音频下载失败: ${audioResponse.statusText}`);
            }
            
            const audioBuffer = await audioResponse.arrayBuffer();
            downloadProgress.value = 50;
            
            // 确定文件扩展名（优先 Content-Type，再回退 URL）
            const contentType = (audioResponse.headers.get('Content-Type') || '').toLowerCase();
            let fileExtension = '.mp3';
            if (contentType.includes('flac')) {
                fileExtension = '.flac';
            } else if (contentType.includes('mp4')) {
                fileExtension = '.m4a';
            } else if (contentType.includes('eac3') || contentType.includes('ec3')) {
                fileExtension = '.eac3';
            } else {
                try {
                    const url = new URL(currentUrl);
                    const pathname = url.pathname.toLowerCase();
                    if (pathname.endsWith('.flac')) {
                        fileExtension = '.flac';
                    } else if (pathname.endsWith('.m4a') || pathname.endsWith('.mp4')) {
                        fileExtension = '.m4a';
                    } else if (pathname.endsWith('.eac3') || pathname.endsWith('.ec3')) {
                        fileExtension = '.eac3';
                    }
                } catch { void 0 }
            }
            
            // 处理元数据
            progressText.value = '处理元数据';
            downloadProgress.value = 70;
            
            let finalAudioBuffer;
            const canWriteMetadata = settings.writeMetadata && (fileExtension === '.mp3' || fileExtension === '.flac');
            if (canWriteMetadata) {
                const metadata = {
                    name: props.musicInfo.name || '未知歌曲',
                    artist: props.musicInfo.artist || '未知艺术家',
                    album: props.musicInfo.album || '未知专辑',
                    year: props.musicInfo.year || '',
                    lyrics: mergeLyrics(props.musicInfo.lrc, props.musicInfo.tlyric),
                    cover: props.musicInfo.cover || null
                };
                
                try {
                    finalAudioBuffer = await embedMetadata(audioBuffer, metadata, fileExtension);
                } catch {
                    finalAudioBuffer = audioBuffer;
                }
            } else {
                finalAudioBuffer = audioBuffer;
            }
            
            downloadProgress.value = 90;
            
            // 下载文件
            progressText.value = '准备下载';
            const audioMime = audioResponse.headers.get('Content-Type') || getMimeByExtension(fileExtension);
            const audioBlob = new Blob([finalAudioBuffer], { type: audioMime });
            const fileName = sanitizeFilename(getFormattedFilename(props.musicInfo.name, props.musicInfo.artist, fileExtension));
            saveBlob(audioBlob, fileName);
            
            progressText.value = '下载完成';
            progressStatus.value = 'success';
            downloadProgress.value = 100;
            
            ElNotification({
                title: '下载成功',
                message: `成功下载音频文件：${props.musicInfo.name}`,
                type: 'success',
                duration: 3000
            });
            
        } catch (error) {
            progressText.value = '下载失败';
            progressStatus.value = 'exception';
            ElMessage.error(`下载失败：${error.message}`);
        }
    } else {
        // ZIP打包下载模式
        progressText.value = '准备打包';
        try {
            // 动态导入JSZip以减少初始包大小
            const { default: JSZip } = await import('jszip');
            const zip = new JSZip();
            
            // 第一步：下载音频文件
            progressText.value = '下载音频文件';
            downloadProgress.value = 10;
            
            let currentUrlZip = props.musicInfo.url;
            let audioResponse = await fetch(currentUrlZip);
            if (!audioResponse.ok) {
              // ZIP模式下也尝试刷新播放链接后重试一次
              try {
                const refreshed = await musicApi.getMusicUrl(
                  props.musicInfo.id,
                  props.musicInfo.quality || 'lossless',
                  { bypassCache: true }
                )
                if (refreshed && refreshed.url) {
                  currentUrlZip = refreshed.url
                  audioResponse = await fetch(currentUrlZip)
                }
              } catch { void 0 }
            }
            if (!audioResponse.ok) {
              throw new Error(`音频下载失败: ${audioResponse.statusText}`);
            }
            
            const audioBuffer = await audioResponse.arrayBuffer();
            downloadProgress.value = 30;
            
            // 确定文件扩展名（优先 Content-Type，再回退 URL）
            const zipContentType = (audioResponse.headers.get('Content-Type') || '').toLowerCase();
            let fileExtension = '.mp3';
            if (zipContentType.includes('flac')) {
                fileExtension = '.flac';
            } else if (zipContentType.includes('mp4')) {
                fileExtension = '.m4a';
            } else if (zipContentType.includes('eac3') || zipContentType.includes('ec3')) {
                fileExtension = '.eac3';
            } else {
                try {
                    const url = new URL(currentUrlZip);
                    const pathname = url.pathname.toLowerCase();
                    if (pathname.endsWith('.flac')) {
                        fileExtension = '.flac';
                    } else if (pathname.endsWith('.m4a') || pathname.endsWith('.mp4')) {
                        fileExtension = '.m4a';
                    } else if (pathname.endsWith('.eac3') || pathname.endsWith('.ec3')) {
                        fileExtension = '.eac3';
                    }
                } catch { void 0 }
            }
            
            // 第二步：处理元数据
            progressText.value = '处理元数据';
            downloadProgress.value = 40;
            
            const metadata = {
                name: props.musicInfo.name || '未知歌曲',
                artist: props.musicInfo.artist || '未知艺术家',
                album: props.musicInfo.album || '未知专辑',
                year: props.musicInfo.year || '',
                lyrics: mergeLyrics(props.musicInfo.lrc || '', props.musicInfo.tlyric || ''),
                cover: props.musicInfo.cover || null
            };
            
            // 尝试嵌入元数据
            let finalAudioBuffer;
            const zipCanWriteMetadata = settings.writeMetadata && (fileExtension === '.mp3' || fileExtension === '.flac');
            if (zipCanWriteMetadata) {
                progressText.value = '写入元数据';
                try {
                    finalAudioBuffer = await embedMetadata(audioBuffer, metadata, fileExtension);
                } catch {
                    finalAudioBuffer = audioBuffer;
                }
            } else {
                // 不写入元数据，直接使用原始文件
                finalAudioBuffer = audioBuffer;
            }
            
            downloadProgress.value = 50;
            
            // 第三步：添加音频文件到ZIP
            const audioFileName = getFormattedFilename(props.musicInfo.name, props.musicInfo.artist, fileExtension);
            
            zip.file(audioFileName, finalAudioBuffer);
            progressText.value = '添加音频文件';
            downloadProgress.value = 60;
            
            // 第四步：下载并添加封面图片
            if (props.musicInfo.cover) {
                try {
                    progressText.value = '下载封面图片';
                    const coverResponse = await fetch(props.musicInfo.cover);
                    if (coverResponse.ok) {
                        const coverBuffer = await coverResponse.arrayBuffer();
                        const coverExtension = coverResponse.headers.get('content-type')?.includes('png') ? '.png' : '.jpg';
                        const coverFileName = getFormattedFilename(props.musicInfo.name, props.musicInfo.artist, coverExtension);
                        zip.file(coverFileName, coverBuffer);
                    }
                } catch {
                  void 0
                }
            }
            downloadProgress.value = 70;
            
            // 第五步：添加歌词文件
            if (props.musicInfo.lrc) {
                // 根据设置决定生成 LRC 或 SRT
                const mergedLrc = mergeLyrics(props.musicInfo.lrc, props.musicInfo.tlyric)
                if (settings.srtLyricsDownload) {
                  progressText.value = '添加SRT歌词';
                  try {
                    const srtContent = lrcToSrt(mergedLrc)
                    if (srtContent) {
                      const srtFileName = getFormattedFilename(props.musicInfo.name, props.musicInfo.artist, '.srt')
                      zip.file(srtFileName, srtContent)
                    }
                  } catch {
                    const lrcFileName = getFormattedFilename(props.musicInfo.name, props.musicInfo.artist, '.lrc')
                    zip.file(lrcFileName, mergedLrc)
                  }
                } else {
                  progressText.value = '添加LRC歌词';
                  const lrcFileName = getFormattedFilename(props.musicInfo.name, props.musicInfo.artist, '.lrc')
                  zip.file(lrcFileName, mergedLrc)
                }
            }
            downloadProgress.value = 80;
            
            // 第六步：添加信息文件
            progressText.value = '生成信息文件';
            const infoContent = `歌曲信息
==================
歌名: ${props.musicInfo.name}
艺术家: ${props.musicInfo.artist}
专辑: ${props.musicInfo.album}
时长: ${formatDuration(props.musicInfo.duration)}
音质: ${props.musicInfo.qualityName}
大小: ${formatFileSize(props.musicInfo.fileSize)}
下载时间: ${new Date().toLocaleString('zh-CN')}
==================
`;
            
            zip.file('歌曲信息.txt', infoContent);
            downloadProgress.value = 85;
            
            // 第七步：生成ZIP文件
            progressText.value = '生成ZIP文件';
            const zipBlob = await zip.generateAsync({
                type: 'blob',
                compression: 'DEFLATE',
                compressionOptions: {
                    level: 6
                }
            }, (metadata) => {
                // 更新压缩进度
                const progress = 85 + (metadata.percent * 0.1);
                downloadProgress.value = Math.round(progress);
            });
            
            downloadProgress.value = 95;
            
            // 第八步：下载ZIP文件
            progressText.value = '准备下载';
            // Some browsers need explicit MIME type for ZIP blobs
            const zipTypedBlob = ensureBlobType(zipBlob, getMimeByExtension('.zip'));
            const zipFileName = sanitizeFilename(getFormattedFilename(props.musicInfo.name, props.musicInfo.artist, '.zip'));
            saveBlob(zipTypedBlob, zipFileName);
            
            progressText.value = '下载完成';
            progressStatus.value = 'success';
            downloadProgress.value = 100;
            
            ElNotification({
                title: '下载成功',
                message: `成功下载音乐包：${props.musicInfo.name}（包含音频、封面、歌词、信息文件）`,
                type: 'success',
                duration: 5000
            });
        } catch (error) {
            progressText.value = '下载失败';
            progressStatus.value = 'exception';
            ElMessage.error(`下载失败：${error.message}`);
        }
    }
    
    // 延迟重置下载状态
    setTimeout(() => {
        downloading.value = false;
    }, 2000);
};

const formatDuration = (milliseconds) => {
  if (!milliseconds) return '00:00';
  const totalSeconds = Math.floor(milliseconds / 1000);
  const min = Math.floor(totalSeconds / 60);
  const sec = totalSeconds % 60;
  return `${String(min).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
};

const formatFileSize = (bytes) => {
    if (!bytes) return '未知';
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
}

watch(() => props.musicInfo, (newMusicInfo) => {
  if (newMusicInfo && newMusicInfo.url) {
    if (ap.value) {
      ap.value.destroy();
    }
    nextTick(() => {
      if (typeof window !== 'undefined' && typeof window.APlayer !== 'undefined') {
        ap.value = new window.APlayer({
          container: document.getElementById('aplayer-container'),
          // 进度条主题色，随深色模式调整
          theme: isDark.value ? '#90caf9' : '#409EFF',
          audio: [{
            name: newMusicInfo.name,
            artist: newMusicInfo.artist,
            url: newMusicInfo.url,
            cover: newMusicInfo.cover,
            lrc: mergeLyrics(newMusicInfo.lrc, newMusicInfo.tlyric),
          }],
          lrcType: 1,
        });
      }
    });
  }
}, { deep: true, immediate: true });
</script>

<style scoped>
/* 结果卡片 */
.result-card {
  margin-top: var(--space-4);
}

.info-container {
  margin-bottom: var(--space-4);
}

.song-title {
  font-size: 28px;
  font-weight: 700;
  margin: 0 0 8px 0;
}

.artist-name {
  font-size: 18px;
  color: var(--el-text-color-regular);
  margin: 0 0 var(--space-4) 0;
}

.metadata-grid {
  display: grid;
  grid-template-columns: auto 1fr auto 1fr;
  gap: var(--space-2) calc(var(--space-3) + var(--space-1));
  font-size: 14px;
  padding: var(--space-2);
  border-top: 1px solid var(--el-border-color-lighter);
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.meta-label {
  color: var(--el-text-color-secondary);
}

.meta-value {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.aplayer-instance {
  margin-bottom: var(--space-3);
}

.direct-link-section {
  margin-top: var(--space-3);
}

.action-section {
  margin-top: var(--space-4);
  display: flex;
  justify-content: center;
}

.download-btn {
  width: 100%;
  max-width: 280px;
  font-weight: 500;
}

.download-btn .btn-text {
  margin-left: 8px;
}

/* 进度条样式 */
.progress-container {
  margin: var(--space-3) auto 0;
  width: 100%;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  width: 100%;
}

.progress-text {
  font-size: 12px;
  color: var(--el-text-color-regular);
  font-weight: 500;
}

.progress-percentage {
  font-size: 12px;
  color: var(--el-color-primary);
  font-weight: 600;
}

@media (max-width: 768px) {
  :deep(.el-card__body) {
    padding: var(--space-3);
  }
  .result-card {
    margin-top: var(--space-2);
  }
  .info-container {
    text-align: center;
  }
  .song-title {
    font-size: 22px;
    margin-bottom: 4px;
  }
  .artist-name {
    font-size: 14px;
    margin-bottom: var(--space-2);
  }
  .metadata-grid {
    grid-template-columns: auto 1fr;
    gap: var(--space-1) var(--space-2);
    font-size: 13px;
    padding: var(--space-1);
  }
  .download-btn {
    max-width: 100%;
  }
  .aplayer-instance {
    margin-bottom: var(--space-2);
  }
  .direct-link-section {
    margin-top: var(--space-2);
  }
  .action-section {
    margin-top: var(--space-3);
  }
  
  .progress-container {
    padding: 0 8px;
    margin-top: var(--space-2);
    max-width: 100%;
  }
  
  .progress-text,
  .progress-percentage {
    font-size: 11px;
  }
}

@media (max-width: 480px) {
  .download-btn .btn-text {
    display: none;
  }
  .download-btn .el-icon {
    margin: 0;
  }
}
</style>

<!-- 深色模式下覆盖 APlayer 样式（非 scoped，确保能作用到第三方组件） -->
<style>
/* 容器与文字颜色 */
.dark .aplayer {
  background-color: #1e1f22 !important;
  color: #e5eaf3 !important;
  border-color: var(--el-border-color) !important;
}
.dark .aplayer .aplayer-title { color: #e5eaf3 !important; }
.dark .aplayer .aplayer-author { color: #cfd3dc !important; }

.aplayer .aplayer-lrc p {
  word-break: break-all;
}

/* 歌词颜色 */
.dark .aplayer .aplayer-lrc { color: #cfd3dc !important; }
.dark .aplayer .aplayer-lrc p.aplayer-lrc-current { color: var(--el-text-color-primary) !important; }

/* 进度条与音量条背景 */
.dark .aplayer .aplayer-bar-wrap .aplayer-bar { background-color: #3a3a3a !important; }
.dark .aplayer .aplayer-played { background-color: var(--el-color-primary) !important; }
.dark .aplayer .aplayer-loading { background-color: var(--el-color-primary-light-5) !important; }
.dark .aplayer .aplayer-volume .aplayer-volume-bar { background-color: #3a3a3a !important; }

/* 播放列表样式 */
.dark .aplayer .aplayer-list { background-color: #1e1f22 !important; border-color: #2a2b2e !important; }
.dark .aplayer .aplayer-list li:hover { background-color: #262727 !important; }
.dark .aplayer .aplayer-list li .aplayer-list-author { color: #909399 !important; }

/* 深色模式下移除 APlayer 组件内的伪元素 */
.dark .aplayer *::before,
.dark .aplayer *::after {
  content: none !important;
  display: none !important;
}
</style>
