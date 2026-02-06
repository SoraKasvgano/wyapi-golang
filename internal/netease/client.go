package netease

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	appcrypto "wyapi-golang/internal/crypto"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Safari/537.36 Chrome/91.0.4472.164 NeteaseMusicDesktop/2.10.2.200154"
	referer   = "https://music.163.com/"

	songURLV1API      = "https://interface3.music.163.com/eapi/song/enhance/player/url/v1"
	songDetailV3API   = "https://interface3.music.163.com/api/v3/song/detail"
	lyricAPI          = "https://interface3.music.163.com/api/song/lyric"
	searchAPI         = "https://music.163.com/api/cloudsearch/pc"
	playlistDetailAPI = "https://music.163.com/api/v6/playlist/detail"
	albumDetailAPI    = "https://music.163.com/api/v1/album/"
)

var defaultCookies = map[string]string{
	"os":       "pc",
	"appver":   "",
	"osver":    "",
	"deviceId": "pyncm!",
}

type Client struct {
	httpClient *http.Client
	rng        *rand.Rand
}

func NewClient(timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (c *Client) GetSongURL(ctx context.Context, songID int64, quality string, cookies map[string]string) (*SongURLResponse, error) {
	headerJSON, err := c.buildHeaderJSON()
	if err != nil {
		return nil, err
	}

	payload := SongURLPayload{
		IDs:        []int64{songID},
		Level:      quality,
		EncodeType: "flac",
		Header:     headerJSON,
	}
	if quality == "sky" {
		payload.ImmerseType = "c51"
	}

	params, err := c.buildEAPIParams(songURLV1API, payload)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("params", params)

	body, err := c.postForm(ctx, songURLV1API, form, cookies)
	if err != nil {
		return nil, err
	}

	var resp SongURLResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, errors.New("netease: song url request failed")
	}
	return &resp, nil
}

func (c *Client) GetSongDetail(ctx context.Context, songID int64) (*SongDetailResponse, error) {
	data := url.Values{}
	payload := []SongDetailPayload{{ID: songID, V: 0}}
	payloadJSON, err := marshalNoEscape(payload)
	if err != nil {
		return nil, err
	}
	data.Set("c", string(payloadJSON))

	body, err := c.postForm(ctx, songDetailV3API, data, nil)
	if err != nil {
		return nil, err
	}

	var resp SongDetailResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, errors.New("netease: song detail request failed")
	}
	return &resp, nil
}

func (c *Client) GetLyrics(ctx context.Context, songID int64, cookies map[string]string) (*LyricResponse, error) {
	data := url.Values{}
	data.Set("id", strconv.FormatInt(songID, 10))
	data.Set("cp", "false")
	data.Set("tv", "0")
	data.Set("lv", "0")
	data.Set("rv", "0")
	data.Set("kv", "0")
	data.Set("yv", "0")
	data.Set("ytv", "0")
	data.Set("yrv", "0")

	body, err := c.postForm(ctx, lyricAPI, data, cookies)
	if err != nil {
		return nil, err
	}

	var resp LyricResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, errors.New("netease: lyric request failed")
	}
	return &resp, nil
}

func (c *Client) Search(ctx context.Context, keywords string, limit int, cookies map[string]string) (*SearchResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	data := url.Values{}
	data.Set("s", keywords)
	data.Set("type", "1")
	data.Set("limit", strconv.Itoa(limit))

	body, err := c.postForm(ctx, searchAPI, data, cookies)
	if err != nil {
		return nil, err
	}

	var resp SearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, errors.New("netease: search request failed")
	}
	return &resp, nil
}

func (c *Client) GetPlaylistDetail(ctx context.Context, playlistID int64, cookies map[string]string) (*PlaylistInfo, error) {
	data := url.Values{}
	data.Set("id", strconv.FormatInt(playlistID, 10))

	body, err := c.postForm(ctx, playlistDetailAPI, data, cookies)
	if err != nil {
		return nil, err
	}

	var resp PlaylistDetailResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, errors.New("netease: playlist request failed")
	}

	playlist := resp.Playlist
	info := &PlaylistInfo{
		ID:          playlist.ID,
		Name:        playlist.Name,
		CoverImgURL: playlist.CoverImgURL,
		PicURL:      playlist.CoverImgURL,
		Creator:     playlist.Creator.Nickname,
		TrackCount:  playlist.TrackCount,
		Description: playlist.Description,
		Tracks:      []TrackInfo{},
	}

	trackIDs := make([]int64, 0, len(playlist.TrackIds))
	for _, item := range playlist.TrackIds {
		trackIDs = append(trackIDs, item.ID)
	}

	if len(trackIDs) == 0 {
		return info, nil
	}

	for i := 0; i < len(trackIDs); i += 100 {
		end := i + 100
		if end > len(trackIDs) {
			end = len(trackIDs)
		}

		batch := trackIDs[i:end]
		songs, err := c.getSongDetailBatch(ctx, batch, cookies)
		if err != nil {
			return nil, err
		}
		for _, song := range songs {
			info.Tracks = append(info.Tracks, trackInfoFromSong(song))
		}
	}

	return info, nil
}

func (c *Client) GetAlbumDetail(ctx context.Context, albumID int64, cookies map[string]string) (*AlbumInfo, error) {
	url := albumDetailAPI + strconv.FormatInt(albumID, 10)

	body, err := c.get(ctx, url, cookies)
	if err != nil {
		return nil, err
	}

	var resp AlbumDetailResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, errors.New("netease: album request failed")
	}

	album := resp.Album
	info := &AlbumInfo{
		ID:          album.ID,
		Name:        album.Name,
		CoverImgURL: c.GetPicURL(album.Pic, 300),
		PicURL:      c.GetPicURL(album.Pic, 300),
		Artist:      album.Artist.Name,
		PublishTime: album.PublishTime,
		Description: album.Description,
		Tracks:      []TrackInfo{},
	}

	for _, song := range resp.Songs {
		info.Tracks = append(info.Tracks, trackInfoFromSong(song))
	}
	info.TrackCount = len(info.Tracks)

	return info, nil
}

func (c *Client) GetPicURL(picID int64, size int) string {
	if picID == 0 {
		return ""
	}
	if size <= 0 {
		size = 300
	}

	magic := []byte("3go8&$8*3*3h0k(2)2")
	songID := []byte(strconv.FormatInt(picID, 10))

	for i := range songID {
		songID[i] = songID[i] ^ magic[i%len(magic)]
	}

	sum := appcrypto.MD5Bytes(songID)
	encoded := appcrypto.Base64Encode(sum)
	encoded = strings.ReplaceAll(encoded, "/", "_")
	encoded = strings.ReplaceAll(encoded, "+", "-")

	return "https://p3.music.126.net/" + encoded + "/" + strconv.FormatInt(picID, 10) + ".jpg?param=" + strconv.Itoa(size) + "y" + strconv.Itoa(size)
}

func (c *Client) buildHeaderJSON() (string, error) {
	header := DeviceHeader{
		OS:        "pc",
		AppVer:    "",
		OSVer:     "",
		DeviceID:  "pyncm!",
		RequestID: c.randomRequestID(),
	}

	data, err := marshalNoEscape(header)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Client) randomRequestID() string {
	if c.rng == nil {
		return ""
	}
	return strconv.Itoa(c.rng.Intn(10000000) + 20000000)
}

func (c *Client) buildEAPIParams(rawURL string, payload interface{}) (string, error) {
	payloadJSON, err := marshalNoEscape(payload)
	if err != nil {
		return "", err
	}
	return appcrypto.EncryptEAPIParams(rawURL, payloadJSON)
}

func (c *Client) postForm(ctx context.Context, target string, data url.Values, cookies map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", referer)

	applyCookies(req, cookies)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("netease: http status " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) get(ctx context.Context, target string, cookies map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", referer)

	applyCookies(req, cookies)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("netease: http status " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) getSongDetailBatch(ctx context.Context, ids []int64, cookies map[string]string) ([]SongDetailSong, error) {
	payload := make([]SongDetailPayload, 0, len(ids))
	for _, id := range ids {
		payload = append(payload, SongDetailPayload{ID: id, V: 0})
	}

	payloadJSON, err := marshalNoEscape(payload)
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("c", string(payloadJSON))

	body, err := c.postForm(ctx, songDetailV3API, data, cookies)
	if err != nil {
		return nil, err
	}

	var resp SongDetailResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, errors.New("netease: song detail batch failed")
	}
	return resp.Songs, nil
}

func applyCookies(req *http.Request, cookies map[string]string) {
	merged := make(map[string]string, len(defaultCookies)+len(cookies))
	for k, v := range defaultCookies {
		merged[k] = v
	}
	for k, v := range cookies {
		if k == "" || v == "" {
			continue
		}
		merged[k] = v
	}

	if len(merged) == 0 {
		return
	}

	pairs := make([]string, 0, len(merged))
	for k, v := range merged {
		pairs = append(pairs, k+"="+v)
	}

	req.Header.Set("Cookie", strings.Join(pairs, "; "))
}

func marshalNoEscape(value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}

	out := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
	return out, nil
}

func trackInfoFromSong(song SongDetailSong) TrackInfo {
	artists := make([]string, 0, len(song.Ar))
	for _, artist := range song.Ar {
		if artist.Name != "" {
			artists = append(artists, artist.Name)
		}
	}

	return TrackInfo{
		ID:       song.ID,
		Name:     song.Name,
		Artists:  strings.Join(artists, "/"),
		Album:    song.Al.Name,
		PicURL:   song.Al.PicURL,
		Duration: song.Dt,
	}
}

// Helper methods used by downloader
func (c *Client) FetchSongURLDirect(ctx context.Context, url string) ([]byte, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, errors.New("netease: download status " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return body, resp.Header, nil
}

func (c *Client) FetchSongStream(ctx context.Context, url string) (io.ReadCloser, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, nil, errors.New("netease: download status " + resp.Status)
	}

	return resp.Body, resp.Header, nil
}

func (c *Client) ResolveShortURL(ctx context.Context, shortURL string) (string, error) {
	if shortURL == "" {
		return "", errors.New("empty url")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, shortURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		if location != "" {
			return location, nil
		}
	}

	return resp.Request.URL.String(), nil
}

func (c *Client) ResolveShortURLHead(ctx context.Context, shortURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, shortURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		if location := resp.Header.Get("Location"); location != "" {
			return location, nil
		}
	}

	if resp.Request != nil && resp.Request.URL != nil {
		return resp.Request.URL.String(), nil
	}
	return shortURL, nil
}
