// 将 LRC 歌词转换为 SRT 字幕格式
// 说明：
// - 支持标准 LRC 时间戳格式：[mm:ss.xx] 或 [mm:ss.xxx]
// - 忽略元信息标签，如 [ti:], [ar:], [al:], [by:]
// - 当缺少下一行时间戳用于结束时间时，默认持续 3 秒
// - 同一行多个时间戳将拆分为多条字幕项（共享同一文本）

/**
 * 将毫秒格式化为 SRT 时间戳（HH:MM:SS,mmm）
 * @param {number} ms - 毫秒
 * @returns {string}
 */
export const formatSrtTimestamp = (ms, omitMs = false) => {
  if (!Number.isFinite(ms) || ms < 0) ms = 0
  const hours = Math.floor(ms / 3600000)
  const minutes = Math.floor((ms % 3600000) / 60000)
  const seconds = Math.floor((ms % 60000) / 1000)
  const millis = ms % 1000
  const pad = (n, w = 2) => String(n).padStart(w, '0')
  if (omitMs) {
    return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`
  }
  return `${pad(hours)}:${pad(minutes)}:${pad(seconds)},${String(millis).padStart(3, '0')}`
}

/**
 * 解析 LRC 文本为时间戳与文本条目
 * @param {string} lrcText
 * @returns {Array<{time:number, text:string}>}
 */
export const parseLrc = (lrcText) => {
  if (!lrcText || typeof lrcText !== 'string') return []

  const lines = lrcText.split(/\r?\n/)
  // 预扫描 offset 元信息（毫秒），用于整体时间校准
  let offsetMs = 0
  for (const raw of lines) {
    const m = raw.trim().match(/^\[offset:\s*([+-]?\d+)\s*\]/i)
    if (m) {
      const val = parseInt(m[1], 10)
      if (Number.isFinite(val)) offsetMs = val
      break
    }
  }
  const entries = []

  // 匹配 [mm:ss.xx] 或 [mm:ss.xxx]，小数分隔符允许 . 或 :
  const timeTagRegex = /\[(\d{1,2}):(\d{2})(?:[.:](\d{1,3}))?\]/g

  for (const rawLine of lines) {
    const line = rawLine.trim()
    if (!line) continue

    // 跳过元信息行（offset 已在预扫描处理）
    if (/^\[(ti|ar|al|by|offset):/i.test(line)) continue

    const timeTags = [...line.matchAll(timeTagRegex)]
    if (timeTags.length === 0) continue

    // 文本：去除所有时间戳标签后剩余内容
    const text = line.replace(timeTagRegex, '').trim()
    if (!text) continue

    for (const match of timeTags) {
      const minutes = parseInt(match[1], 10) || 0
      const seconds = parseInt(match[2], 10) || 0
      const fractionRaw = match[3] || '0'
      // 将 1 位/2 位/3 位小数统一为毫秒
      let fractionMs = 0
      if (fractionRaw) {
        const frac = fractionRaw.padEnd(3, '0').slice(0, 3)
        fractionMs = parseInt(frac, 10) || 0
      }
      let timeMs = minutes * 60000 + seconds * 1000 + fractionMs
      timeMs += offsetMs
      if (timeMs < 0) timeMs = 0
      entries.push({ time: timeMs, text })
    }
  }

  // 按时间排序
  entries.sort((a, b) => a.time - b.time)

  // 去重（若同一时间戳重复出现，保留后出现的文本）
  const compact = []
  for (const item of entries) {
    if (compact.length && compact[compact.length - 1].time === item.time) {
      compact[compact.length - 1] = item
    } else {
      compact.push(item)
    }
  }

  return compact
}

/**
 * 将 LRC 文本转换为 SRT 格式
 * @param {string} lrcText - 原始 LRC 文本
 * @param {object} options - 可选项
 * @param {number} [options.fallbackDurationMs=3000] - 最后一条或相邻相同时间的默认持续时间（毫秒）
 * @returns {string} - SRT 字幕内容
 */
export const lrcToSrt = (lrcText, options = {}) => {
  const {
    fallbackDurationMs = 3000,
    minDurationMs = 500,
    maxDurationMs = 5000,
    windowsLineEndings = true,
    omitMilliseconds = true
  } = options

  const items = parseLrc(lrcText)
  if (!items.length) return ''

  const lines = []
  for (let i = 0; i < items.length; i++) {
    const startMs = items[i].time
    let endMs

    if (i < items.length - 1) {
      const nextStart = items[i + 1].time
      // 与下一条严格对齐，确保精确消失时间；同时至少 1ms，避免零时长
      endMs = Math.max(nextStart, startMs + 1)
      // 若下一条时间异常（远大于兜底/最大），仍可限制最大显示时长
      endMs = Math.min(endMs, startMs + maxDurationMs)
    } else {
      // 最后一条：无下一条，使用兜底与最大时长限制
      endMs = Math.min(startMs + fallbackDurationMs, startMs + maxDurationMs)
    }

    // 额外保护，确保结束时间至少晚于开始时间
    if (!Number.isFinite(endMs) || endMs <= startMs) {
      endMs = startMs + Math.max(minDurationMs, fallbackDurationMs)
    }

    lines.push(String(i + 1))
    lines.push(`${formatSrtTimestamp(startMs, omitMilliseconds)} --> ${formatSrtTimestamp(endMs, omitMilliseconds)}`)
    lines.push(items[i].text)
    lines.push("")
  }

  let srt = lines.join("\n").trim()
  if (windowsLineEndings) {
    // 转换为 CRLF 以提升在部分播放器上的兼容性（如 PotPlayer、MPC-HC）
    srt = srt.replace(/\n/g, "\r\n")
  }
  return srt + (windowsLineEndings ? "\r\n" : "\n")
}

export default {
  formatSrtTimestamp,
  parseLrc,
  lrcToSrt
}