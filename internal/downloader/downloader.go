package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"wyapi-golang/internal/cookie"
	"wyapi-golang/internal/netease"
)

var illegalFilenameChars = regexp.MustCompile(`[<>:"/\\|?*]`)

// MusicInfo represents normalized music metadata for download.
type MusicInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Artists  string `json:"artists"`
	Album    string `json:"album"`
	PicURL   string `json:"pic_url"`
	Duration int64  `json:"duration"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
	Quality  string `json:"quality"`
	URL      string `json:"url"`
	Lyric    string `json:"lyric"`
	TLyric   string `json:"tlyric"`
}

type Downloader struct {
	client        *netease.Client
	cookieManager *cookie.Manager
	downloadDir   string
}

func NewDownloader(client *netease.Client, cookieManager *cookie.Manager, downloadDir string) *Downloader {
	if downloadDir == "" {
		downloadDir = "downloads"
	}
	return &Downloader{
		client:        client,
		cookieManager: cookieManager,
		downloadDir:   downloadDir,
	}
}

func (d *Downloader) GetMusicInfo(ctx context.Context, songID int64, quality string) (*MusicInfo, error) {
	if d.client == nil {
		return nil, errors.New("netease client is nil")
	}
	cookies := map[string]string{}
	if d.cookieManager != nil {
		if parsed, err := d.cookieManager.Parse(); err == nil {
			cookies = parsed
		}
	}

	urlResp, err := d.client.GetSongURL(ctx, songID, quality, cookies)
	if err != nil {
		return nil, err
	}
	if urlResp == nil || len(urlResp.Data) == 0 || urlResp.Data[0].URL == "" {
		return nil, errors.New("music url not available")
	}
	urlData := urlResp.Data[0]

	detailResp, err := d.client.GetSongDetail(ctx, songID)
	if err != nil {
		return nil, err
	}
	if detailResp == nil || len(detailResp.Songs) == 0 {
		return nil, errors.New("music detail not available")
	}
	song := detailResp.Songs[0]

	lyricResp, _ := d.client.GetLyrics(ctx, songID, cookies)
	lyricText := ""
	tlyricText := ""
	if lyricResp != nil {
		lyricText = lyricResp.Lrc.Lyric
		tlyricText = lyricResp.Tlyric.Lyric
	}

	artists := make([]string, 0, len(song.Ar))
	for _, artist := range song.Ar {
		if artist.Name != "" {
			artists = append(artists, artist.Name)
		}
	}

	fileType := strings.ToLower(urlData.Type)
	if fileType == "" {
		fileType = detectExtension(urlData.URL)
	}

	info := &MusicInfo{
		ID:       songID,
		Name:     song.Name,
		Artists:  strings.Join(artists, "/"),
		Album:    song.Al.Name,
		PicURL:   song.Al.PicURL,
		Duration: song.Dt,
		FileType: fileType,
		FileSize: urlData.Size,
		Quality:  quality,
		URL:      urlData.URL,
		Lyric:    lyricText,
		TLyric:   tlyricText,
	}

	return info, nil
}

func (d *Downloader) DownloadToFile(ctx context.Context, info *MusicInfo) (string, int64, error) {
	if info == nil {
		return "", 0, errors.New("music info is nil")
	}
	if info.URL == "" {
		return "", 0, errors.New("download url empty")
	}

	if err := os.MkdirAll(d.downloadDir, 0755); err != nil {
		return "", 0, err
	}

	filename := d.BuildFilename(info)
	if !strings.HasSuffix(filename, "."+info.FileType) {
		filename = filename + "." + info.FileType
	}

	filePath := filepath.Join(d.downloadDir, filename)
	if stat, err := os.Stat(filePath); err == nil && stat.Size() > 0 {
		return filePath, stat.Size(), nil
	}

	stream, _, err := d.client.FetchSongStream(ctx, info.URL)
	if err != nil {
		return "", 0, err
	}
	defer stream.Close()

	file, err := os.Create(filePath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	written, err := io.Copy(file, stream)
	if err != nil {
		return "", 0, err
	}
	return filePath, written, nil
}

func (d *Downloader) BuildFilename(info *MusicInfo) string {
	base := fmt.Sprintf("%s - %s", info.Artists, info.Name)
	base = illegalFilenameChars.ReplaceAllString(base, "_")
	base = strings.TrimSpace(base)
	if len(base) > 200 {
		base = base[:200]
	}
	if base == "" {
		base = "unknown"
	}
	return base
}

func detectExtension(rawURL string) string {
	if rawURL == "" {
		return "mp3"
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "mp3"
	}
	lower := strings.ToLower(parsed.Path)
	switch {
	case strings.HasSuffix(lower, ".flac"):
		return "flac"
	case strings.HasSuffix(lower, ".m4a") || strings.HasSuffix(lower, ".mp4"):
		return "m4a"
	case strings.HasSuffix(lower, ".mp3"):
		return "mp3"
	}
	return "mp3"
}
