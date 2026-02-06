// Unified blob download helper with cross-browser fallbacks
// Ensures proper MIME type, defers URL revocation, and supports legacy Edge/IE
export function saveBlob(blob, filename) {
  if (!blob) return;

  const safeName = sanitizeFilename(filename || 'download');

  // Legacy Edge / IE fallback
  try {
    if (typeof window !== 'undefined' && window.navigator && typeof window.navigator.msSaveOrOpenBlob === 'function') {
      window.navigator.msSaveOrOpenBlob(blob, safeName);
      return;
    }
  } catch {
    // ignore
  }

  const isIOS = /iP(ad|hone|od)/.test(navigator.userAgent) || /Macintosh/.test(navigator.platform) && 'ontouchend' in document;
  const supportsDownloadAttr = 'download' in HTMLAnchorElement.prototype;

  // iOS Safari 对 blob URL 下载支持不一致，采用 dataURL 兜底
  if (isIOS && !supportsDownloadAttr) {
    try {
      const reader = new FileReader();
      reader.onload = () => {
        const url = reader.result;
        const a = document.createElement('a');
        a.href = url;
        a.download = safeName;
        a.rel = 'noopener';
        a.target = '_blank';
        a.style.display = 'none';
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
      };
      reader.readAsDataURL(blob);
      return;
    } catch {
      // 继续走常规流程
    }
  }

  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = safeName;
  a.rel = 'noopener';
  a.target = '_self';
  a.style.display = 'none';
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);

  // Safari 等浏览器需要延迟撤销 URL，避免下载被中断
  setTimeout(() => {
    try { URL.revokeObjectURL(url); } catch { void 0 }
  }, 2000);
}

// Ensure blob has a desired MIME type if missing
export function ensureBlobType(blob, desiredType) {
  if (!blob) return blob;
  const type = blob.type || '';
  if (type && type === desiredType) return blob;
  if (!type && desiredType) {
    return new Blob([blob], { type: desiredType });
  }
  return blob;
}

// Map extension to MIME type
export function getMimeByExtension(ext) {
  switch ((ext || '').toLowerCase()) {
    case '.mp3': return 'audio/mpeg';
    case '.flac': return 'audio/flac';
    case '.m4a': return 'audio/mp4';
    case '.mp4': return 'audio/mp4';
    case '.eac3': return 'audio/eac3';
    case '.ec3': return 'audio/eac3';
    case '.lrc': return 'text/plain';
    case '.srt': return 'application/x-subrip';
    case '.zip': return 'application/zip';
    case '.jpg':
    case '.jpeg': return 'image/jpeg';
    case '.png': return 'image/png';
    default: return 'application/octet-stream';
  }
}

// Sanitize filename for cross-browser safety
export function sanitizeFilename(name) {
  return String(name)
    .replace(/[\\/:*?"<>|]/g, '-') // Windows 禁止字符
    .replace(/[\r\n]/g, ' ') // 移除换行
    .trim()
    .slice(0, 180); // 限制长度避免异常
}
