package netease

type DeviceHeader struct {
	OS        string `json:"os"`
	AppVer    string `json:"appver"`
	OSVer     string `json:"osver"`
	DeviceID  string `json:"deviceId"`
	RequestID string `json:"requestId,omitempty"`
}

type SongURLPayload struct {
	IDs         []int64 `json:"ids"`
	Level       string  `json:"level"`
	EncodeType  string  `json:"encodeType"`
	Header      string  `json:"header"`
	ImmerseType string  `json:"immerseType,omitempty"`
}

type SongDetailPayload struct {
	ID int64 `json:"id"`
	V  int   `json:"v"`
}

type SongURLResponse struct {
	Code int           `json:"code"`
	Data []SongURLData `json:"data"`
}

type SongURLData struct {
	ID    int64  `json:"id"`
	URL   string `json:"url"`
	Level string `json:"level"`
	Size  int64  `json:"size"`
	Type  string `json:"type"`
	Br    int64  `json:"br"`
}

type SongDetailResponse struct {
	Code  int              `json:"code"`
	Songs []SongDetailSong `json:"songs"`
}

type SongDetailSong struct {
	ID   int64    `json:"id"`
	Name string   `json:"name"`
	Ar   []Artist `json:"ar"`
	Al   Album    `json:"al"`
	Dt   int64    `json:"dt"`
	No   int      `json:"no"`
}

type Artist struct {
	Name string `json:"name"`
}

type Album struct {
	Name   string `json:"name"`
	PicURL string `json:"picUrl"`
	Pic    int64  `json:"pic"`
}

type LyricResponse struct {
	Code    int       `json:"code"`
	Lrc     LyricLine `json:"lrc"`
	Tlyric  LyricLine `json:"tlyric"`
	Romalrc LyricLine `json:"romalrc"`
	Klyric  LyricLine `json:"klyric"`
}

type LyricLine struct {
	Lyric string `json:"lyric"`
}

type SearchResponse struct {
	Code   int          `json:"code"`
	Result SearchResult `json:"result"`
}

type SearchResult struct {
	Songs []SearchSong `json:"songs"`
}

type SearchSong struct {
	ID   int64    `json:"id"`
	Name string   `json:"name"`
	Ar   []Artist `json:"ar"`
	Al   Album    `json:"al"`
	Dt   int64    `json:"dt"`
}

type PlaylistDetailResponse struct {
	Code     int      `json:"code"`
	Playlist Playlist `json:"playlist"`
}

type Playlist struct {
	ID          int64             `json:"id"`
	Name        string            `json:"name"`
	CoverImgURL string            `json:"coverImgUrl"`
	Creator     PlaylistCreator   `json:"creator"`
	TrackCount  int               `json:"trackCount"`
	Description string            `json:"description"`
	TrackIds    []PlaylistTrackID `json:"trackIds"`
}

type PlaylistCreator struct {
	Nickname string `json:"nickname"`
}

type PlaylistTrackID struct {
	ID int64 `json:"id"`
}

type AlbumDetailResponse struct {
	Code  int              `json:"code"`
	Album AlbumDetail      `json:"album"`
	Songs []SongDetailSong `json:"songs"`
}

type AlbumDetail struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Pic         int64  `json:"pic"`
	Artist      Artist `json:"artist"`
	PublishTime int64  `json:"publishTime"`
	Description string `json:"description"`
}

type TrackInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Artists  string `json:"artists"`
	Album    string `json:"album"`
	PicURL   string `json:"picUrl"`
	Duration int64  `json:"duration"`
}

type PlaylistInfo struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	CoverImgURL string      `json:"coverImgUrl"`
	PicURL      string      `json:"picUrl"`
	Creator     string      `json:"creator"`
	TrackCount  int         `json:"trackCount"`
	Description string      `json:"description"`
	Tracks      []TrackInfo `json:"tracks"`
}

type AlbumInfo struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	CoverImgURL string      `json:"coverImgUrl"`
	PicURL      string      `json:"picUrl"`
	Artist      string      `json:"artist"`
	PublishTime int64       `json:"publishTime"`
	Description string      `json:"description"`
	TrackCount  int         `json:"trackCount"`
	Tracks      []TrackInfo `json:"tracks"`
}
