package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"wyapi-golang/internal/config"
	"wyapi-golang/internal/cookie"
	"wyapi-golang/internal/downloader"
	"wyapi-golang/internal/netease"
	"wyapi-golang/pkg/response"
)

var (
	idRegex         = regexp.MustCompile(`(?i)(?:id=)(\d+)`)
	numberRegex     = regexp.MustCompile(`\d{5,}`)
	digitsOnlyRegex = regexp.MustCompile(`^\d+$`)
	songIDPattern   = regexp.MustCompile(`(?i)(song|album|playlist)[^\d]*(\d+)`)
)

type Handler struct {
	cfg        *config.Config
	netease    *netease.Client
	cookies    *cookie.Manager
	downloader *downloader.Downloader
	openAPI    []byte
}

func NewHandler(cfg *config.Config, neteaseClient *netease.Client, cookieManager *cookie.Manager, downloader *downloader.Downloader, openAPI []byte) *Handler {
	return &Handler{
		cfg:        cfg,
		netease:    neteaseClient,
		cookies:    cookieManager,
		downloader: downloader,
		openAPI:    openAPI,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	cookieValid := false
	if h.cookies != nil {
		parsed, err := h.cookies.Parse()
		cookieValid = err == nil && len(parsed) > 0
	}

	data := map[string]interface{}{
		"service":       "running",
		"timestamp":     time.Now().Unix(),
		"cookie_status": map[bool]string{true: "valid", false: "invalid"}[cookieValid],
		"version":       "2.0.0",
	}
	response.Success(w, data, "API服务运行正常")
}

func (h *Handler) APIInfo(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"name":        "网易云音乐API服务",
		"version":     "2.0.0",
		"description": "提供网易云音乐相关API服务",
		"endpoints": map[string]string{
			"/health":             "GET - 健康检查",
			"/song":               "GET/POST - 获取歌曲信息",
			"/search":             "GET/POST - 搜索音乐",
			"/playlist":           "GET/POST - 获取歌单详情",
			"/album":              "GET/POST - 获取专辑详情",
			"/download":           "GET/POST - 下载音乐",
			"/api/info":           "GET - API信息",
			"/api/music/url":      "GET/POST - 获取歌曲链接",
			"/api/music/detail":   "GET/POST - 获取歌曲详情",
			"/api/music/lyric":    "GET/POST - 获取歌词",
			"/api/music/playlist": "GET/POST - 获取歌单详情",
			"/api/music/album":    "GET/POST - 获取专辑详情",
		},
		"supported_qualities": []string{
			"standard", "exhigh", "lossless", "hires", "sky", "jyeffect", "jymaster", "dolby",
		},
	}

	response.Success(w, data, "API信息获取成功")
}

func (h *Handler) OpenAPI(w http.ResponseWriter, r *http.Request) {
	if len(h.openAPI) == 0 {
		response.Error(w, http.StatusNotFound, "openapi not found")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(h.openAPI)
}

func (h *Handler) Song(w http.ResponseWriter, r *http.Request) {
	data := parseRequestData(r)
	idInput := firstNonEmpty(data, "ids", "id", "url")
	if idInput == "" {
		response.Error(w, http.StatusBadRequest, "必须提供 'ids'、'id' 或 'url' 参数")
		return
	}

	quality := firstNonEmpty(data, "level", "quality")
	if quality == "" {
		quality = "lossless"
	}

	infoType := firstNonEmpty(data, "type")
	if infoType == "" {
		infoType = "url"
	}

	songID, err := h.extractID(r.Context(), idInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	switch infoType {
	case "url":
		h.handleSongURL(w, r, songID, quality)
	case "name":
		h.handleSongName(w, r, songID)
	case "lyric":
		h.handleSongLyric(w, r, songID)
	case "json":
		h.handleSongJSON(w, r, songID, quality)
	default:
		response.Error(w, http.StatusBadRequest, "无效的类型参数，支持: url, name, lyric, json")
	}
}

func (h *Handler) SongURL(w http.ResponseWriter, r *http.Request) {
	data := parseRequestData(r)
	idInput := firstNonEmpty(data, "id", "ids", "url")
	if idInput == "" {
		response.Error(w, http.StatusBadRequest, "缺少歌曲ID")
		return
	}

	quality := firstNonEmpty(data, "level", "quality")
	if quality == "" {
		quality = "lossless"
	}

	songID, err := h.extractID(r.Context(), idInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	h.handleSongURL(w, r, songID, quality)
}

func (h *Handler) SongDetail(w http.ResponseWriter, r *http.Request) {
	data := parseRequestData(r)
	idInput := firstNonEmpty(data, "id", "ids", "url")
	if idInput == "" {
		response.Error(w, http.StatusBadRequest, "缺少歌曲ID")
		return
	}

	songID, err := h.extractID(r.Context(), idInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	h.handleSongDetail(w, r, songID)
}

func (h *Handler) SongLyric(w http.ResponseWriter, r *http.Request) {
	data := parseRequestData(r)
	idInput := firstNonEmpty(data, "id", "ids", "url")
	if idInput == "" {
		response.Error(w, http.StatusBadRequest, "缺少歌曲ID")
		return
	}

	songID, err := h.extractID(r.Context(), idInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	h.handleSongLyric(w, r, songID)
}

func (h *Handler) Playlist(w http.ResponseWriter, r *http.Request) {
	data := parseRequestData(r)
	idInput := firstNonEmpty(data, "id", "url")
	if idInput == "" {
		response.Error(w, http.StatusBadRequest, "缺少歌单ID")
		return
	}

	playlistID, err := h.extractID(r.Context(), idInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	cookies := h.loadCookies()
	detail, err := h.netease.GetPlaylistDetail(r.Context(), playlistID, cookies)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, map[string]interface{}{"playlist": detail}, "获取歌单详情成功")
}

func (h *Handler) PlaylistAPI(w http.ResponseWriter, r *http.Request) {
	data := parseRequestData(r)
	idInput := firstNonEmpty(data, "id", "url")
	if idInput == "" {
		response.Error(w, http.StatusBadRequest, "缺少歌单ID")
		return
	}

	playlistID, err := h.extractID(r.Context(), idInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	cookies := h.loadCookies()
	detail, err := h.netease.GetPlaylistDetail(r.Context(), playlistID, cookies)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, detail, "获取歌单详情成功")
}

func (h *Handler) Album(w http.ResponseWriter, r *http.Request) {
	data := parseRequestData(r)
	idInput := firstNonEmpty(data, "id", "url")
	if idInput == "" {
		response.Error(w, http.StatusBadRequest, "缺少专辑ID")
		return
	}

	albumID, err := h.extractID(r.Context(), idInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	cookies := h.loadCookies()
	detail, err := h.netease.GetAlbumDetail(r.Context(), albumID, cookies)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, map[string]interface{}{"album": detail}, "获取专辑详情成功")
}

func (h *Handler) AlbumAPI(w http.ResponseWriter, r *http.Request) {
	data := parseRequestData(r)
	idInput := firstNonEmpty(data, "id", "url")
	if idInput == "" {
		response.Error(w, http.StatusBadRequest, "缺少专辑ID")
		return
	}

	albumID, err := h.extractID(r.Context(), idInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	cookies := h.loadCookies()
	detail, err := h.netease.GetAlbumDetail(r.Context(), albumID, cookies)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, detail, "获取专辑详情成功")
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	data := parseRequestData(r)
	keyword := firstNonEmpty(data, "keyword", "keywords", "q")
	if keyword == "" {
		response.Error(w, http.StatusBadRequest, "关键词不能为空")
		return
	}
	limit := parseInt(firstNonEmpty(data, "limit"), 30)

	cookies := h.loadCookies()
	searchResp, err := h.netease.Search(r.Context(), keyword, limit, cookies)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]map[string]interface{}, 0, len(searchResp.Result.Songs))
	for _, song := range searchResp.Result.Songs {
		artists := make([]string, 0, len(song.Ar))
		for _, artist := range song.Ar {
			if artist.Name != "" {
				artists = append(artists, artist.Name)
			}
		}
		result = append(result, map[string]interface{}{
			"id":      song.ID,
			"name":    song.Name,
			"artists": strings.Join(artists, "/"),
			"album":   song.Al.Name,
			"picUrl":  song.Al.PicURL,
		})
	}

	response.Success(w, result, "搜索完成")
}

func (h *Handler) NeteaseSearch(w http.ResponseWriter, r *http.Request) {
	data := parseRequestData(r)
	keyword := firstNonEmpty(data, "keywords", "keyword", "q")
	if keyword == "" {
		response.Error(w, http.StatusBadRequest, "关键词不能为空")
		return
	}
	limit := parseInt(firstNonEmpty(data, "limit"), 20)

	cookies := h.loadCookies()
	searchResp, err := h.netease.Search(r.Context(), keyword, limit, cookies)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload := map[string]interface{}{
		"code":    http.StatusOK,
		"status":  http.StatusOK,
		"success": true,
		"msg":     "success",
		"message": "success",
		"result":  searchResp.Result,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	data := parseRequestData(r)
	idInput := firstNonEmpty(data, "id", "url")
	if idInput == "" {
		response.Error(w, http.StatusBadRequest, "缺少歌曲ID")
		return
	}

	quality := firstNonEmpty(data, "quality", "level")
	if quality == "" {
		quality = "lossless"
	}

	returnFormat := firstNonEmpty(data, "format")
	if returnFormat == "" {
		returnFormat = "file"
	}

	songID, err := h.extractID(r.Context(), idInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := h.downloader.GetMusicInfo(r.Context(), songID, quality)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if returnFormat == "json" {
		response.Success(w, map[string]interface{}{
			"music_id":  info.ID,
			"name":      info.Name,
			"artist":    info.Artists,
			"album":     info.Album,
			"quality":   info.Quality,
			"file_type": info.FileType,
			"file_size": info.FileSize,
			"url":       info.URL,
		}, "下载信息获取成功")
		return
	}

	h.sendDownloadFile(w, r, info)
}

func (h *Handler) handleSongURL(w http.ResponseWriter, r *http.Request, songID int64, quality string) {
	cookies := h.loadCookies()
	resp, err := h.netease.GetSongURL(r.Context(), songID, quality, cookies)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, resp.Data, "获取歌曲URL成功")
}

func (h *Handler) handleSongName(w http.ResponseWriter, r *http.Request, songID int64) {
	resp, err := h.netease.GetSongDetail(r.Context(), songID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, resp, "获取歌曲信息成功")
}

func (h *Handler) handleSongLyric(w http.ResponseWriter, r *http.Request, songID int64) {
	cookies := h.loadCookies()
	resp, err := h.netease.GetLyrics(r.Context(), songID, cookies)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	data := map[string]string{
		"lrc":     safeLyric(resp),
		"tlyric":  safeTLyric(resp),
		"romalrc": "",
		"klyric":  "",
	}
	if resp != nil {
		data["romalrc"] = resp.Romalrc.Lyric
		data["klyric"] = resp.Klyric.Lyric
	}

	response.Success(w, data, "获取歌词成功")
}

func (h *Handler) handleSongDetail(w http.ResponseWriter, r *http.Request, songID int64) {
	resp, err := h.netease.GetSongDetail(r.Context(), songID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if resp == nil || len(resp.Songs) == 0 {
		response.Error(w, http.StatusNotFound, "未找到歌曲信息")
		return
	}

	song := resp.Songs[0]
	artists := make([]string, 0, len(song.Ar))
	for _, artist := range song.Ar {
		if artist.Name != "" {
			artists = append(artists, artist.Name)
		}
	}

	data := map[string]interface{}{
		"id":       song.ID,
		"name":     song.Name,
		"singer":   strings.Join(artists, "/"),
		"album":    song.Al.Name,
		"picimg":   song.Al.PicURL,
		"duration": formatDuration(song.Dt),
	}

	response.Success(w, data, "获取歌曲信息成功")
}

func (h *Handler) handleSongJSON(w http.ResponseWriter, r *http.Request, songID int64, quality string) {
	detailResp, err := h.netease.GetSongDetail(r.Context(), songID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if detailResp == nil || len(detailResp.Songs) == 0 {
		response.Error(w, http.StatusNotFound, "未找到歌曲信息")
		return
	}

	cookies := h.loadCookies()
	urlResp, _ := h.netease.GetSongURL(r.Context(), songID, quality, cookies)
	lyricResp, _ := h.netease.GetLyrics(r.Context(), songID, cookies)

	song := detailResp.Songs[0]
	artists := make([]string, 0, len(song.Ar))
	for _, artist := range song.Ar {
		if artist.Name != "" {
			artists = append(artists, artist.Name)
		}
	}

	data := map[string]interface{}{
		"id":      songID,
		"name":    song.Name,
		"ar_name": strings.Join(artists, ", "),
		"al_name": song.Al.Name,
		"pic":     song.Al.PicURL,
		"level":   quality,
		"lyric":   safeLyric(lyricResp),
		"tlyric":  safeTLyric(lyricResp),
	}

	if urlResp != nil && len(urlResp.Data) > 0 {
		urlData := urlResp.Data[0]
		data["url"] = urlData.URL
		data["size"] = formatFileSize(urlData.Size)
		data["level"] = urlData.Level
	} else {
		data["url"] = ""
		data["size"] = "获取失败"
	}

	response.Success(w, data, "获取歌曲信息成功")
}

func (h *Handler) sendDownloadFile(w http.ResponseWriter, r *http.Request, info *downloader.MusicInfo) {
	if info == nil {
		response.Error(w, http.StatusInternalServerError, "下载信息为空")
		return
	}

	filename := h.downloader.BuildFilename(info)
	extension := info.FileType
	if extension == "" {
		extension = "mp3"
	}
	filename = fmt.Sprintf("%s.%s", filename, extension)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("X-Download-Message", "Download completed successfully")
	w.Header().Set("X-Download-Filename", url.QueryEscape(filename))

	if h.cfg != nil && !h.cfg.Download.InMemory {
		filePath, _, err := h.downloader.DownloadToFile(r.Context(), info)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		http.ServeFile(w, r, filePath)
		return
	}

	stream, headers, err := h.netease.FetchSongStream(r.Context(), info.URL)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer stream.Close()

	if contentType := headers.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "audio/"+extension)
	}

	_, _ = io.Copy(w, stream)
}

func (h *Handler) extractID(ctx context.Context, input string) (int64, error) {
	cleaned := strings.TrimSpace(input)
	if cleaned == "" {
		return 0, errors.New("ID不能为空")
	}

	if digitsOnlyRegex.MatchString(cleaned) {
		return strconv.ParseInt(cleaned, 10, 64)
	}

	if strings.Contains(cleaned, "163cn.tv") {
		resolved, err := h.netease.ResolveShortURLHead(ctx, cleaned)
		if err == nil && resolved != "" {
			cleaned = resolved
		}
	}

	if idMatch := idRegex.FindStringSubmatch(cleaned); len(idMatch) > 1 {
		return strconv.ParseInt(idMatch[1], 10, 64)
	}

	if match := songIDPattern.FindStringSubmatch(cleaned); len(match) > 2 {
		return strconv.ParseInt(match[2], 10, 64)
	}

	if numberRegex.MatchString(cleaned) {
		return strconv.ParseInt(numberRegex.FindString(cleaned), 10, 64)
	}

	return 0, errors.New("无法从输入中提取ID")
}

func (h *Handler) loadCookies() map[string]string {
	if h.cookies == nil {
		return map[string]string{}
	}
	parsed, err := h.cookies.Parse()
	if err != nil {
		return map[string]string{}
	}
	return parsed
}

func parseRequestData(r *http.Request) map[string]string {
	result := map[string]string{}
	if r == nil {
		return result
	}

	if r.Method == http.MethodGet {
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				result[key] = values[0]
			}
		}
		return result
	}

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))

		if len(body) > 0 {
			var payload map[string]interface{}
			decoder := json.NewDecoder(bytes.NewReader(body))
			decoder.UseNumber()
			if err := decoder.Decode(&payload); err == nil {
				for key, value := range payload {
					result[key] = formatJSONValue(value)
				}
				return result
			}
		}
	}

	_ = r.ParseForm()
	for key, values := range r.Form {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}

	return result
}

func firstNonEmpty(data map[string]string, keys ...string) string {
	for _, key := range keys {
		if value, ok := data[key]; ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func formatJSONValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return ""
		}
		if v == math.Trunc(v) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return ""
		}
		if float64(v) == math.Trunc(float64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case bool:
		return strconv.FormatBool(v)
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			text := formatJSONValue(item)
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, ",")
	default:
		return fmt.Sprint(v)
	}
}

func formatFileSize(sizeBytes int64) string {
	if sizeBytes <= 0 {
		return "0B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(sizeBytes)
	unitIndex := 0

	for size >= 1024.0 && unitIndex < len(units)-1 {
		size /= 1024.0
		unitIndex++
	}

	return fmt.Sprintf("%.2f%s", size, units[unitIndex])
}

func formatDuration(ms int64) string {
	if ms <= 0 {
		return "00:00"
	}
	seconds := ms / 1000
	minutes := seconds / 60
	remaining := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, remaining)
}

func safeLyric(resp *netease.LyricResponse) string {
	if resp == nil {
		return ""
	}
	return resp.Lrc.Lyric
}

func safeTLyric(resp *netease.LyricResponse) string {
	if resp == nil {
		return ""
	}
	return resp.Tlyric.Lyric
}
