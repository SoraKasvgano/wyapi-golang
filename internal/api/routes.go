package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"wyapi-golang/internal/config"
)

func NewRouter(handler *Handler, cfg *config.Config, staticHandler http.Handler, swaggerHandler http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	if cfg != nil {
		r.Use(CORSMiddleware(cfg.CORS))
	}

	r.Get("/health", handler.Health)
	r.Get("/openapi.json", handler.OpenAPI)
	if swaggerHandler != nil {
		r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
			swaggerHandler.ServeHTTP(w, r)
		})
	}

	r.Group(func(api chi.Router) {
		api.Use(AuthMiddleware(cfg))

		api.Get("/api/info", handler.APIInfo)

		api.MethodFunc(http.MethodGet, "/song", handler.Song)
		api.MethodFunc(http.MethodPost, "/song", handler.Song)
		api.MethodFunc(http.MethodGet, "/Song_V1", handler.Song)
		api.MethodFunc(http.MethodPost, "/Song_V1", handler.Song)

		api.MethodFunc(http.MethodGet, "/search", handler.Search)
		api.MethodFunc(http.MethodPost, "/search", handler.Search)
		api.MethodFunc(http.MethodGet, "/Search", handler.Search)
		api.MethodFunc(http.MethodPost, "/Search", handler.Search)

		api.MethodFunc(http.MethodGet, "/playlist", handler.Playlist)
		api.MethodFunc(http.MethodPost, "/playlist", handler.Playlist)
		api.MethodFunc(http.MethodGet, "/Playlist", handler.Playlist)
		api.MethodFunc(http.MethodPost, "/Playlist", handler.Playlist)

		api.MethodFunc(http.MethodGet, "/album", handler.Album)
		api.MethodFunc(http.MethodPost, "/album", handler.Album)
		api.MethodFunc(http.MethodGet, "/Album", handler.Album)
		api.MethodFunc(http.MethodPost, "/Album", handler.Album)

		api.MethodFunc(http.MethodGet, "/download", handler.Download)
		api.MethodFunc(http.MethodPost, "/download", handler.Download)
		api.MethodFunc(http.MethodGet, "/Download", handler.Download)
		api.MethodFunc(http.MethodPost, "/Download", handler.Download)

		api.MethodFunc(http.MethodGet, "/api/music/url", handler.SongURL)
		api.MethodFunc(http.MethodPost, "/api/music/url", handler.SongURL)
		api.MethodFunc(http.MethodGet, "/api/music/detail", handler.SongDetail)
		api.MethodFunc(http.MethodPost, "/api/music/detail", handler.SongDetail)
		api.MethodFunc(http.MethodGet, "/api/getMusicInfo", handler.SongDetail)
		api.MethodFunc(http.MethodPost, "/api/getMusicInfo", handler.SongDetail)
		api.MethodFunc(http.MethodGet, "/api/music/lyric", handler.SongLyric)
		api.MethodFunc(http.MethodPost, "/api/music/lyric", handler.SongLyric)
		api.MethodFunc(http.MethodGet, "/api/music/playlist", handler.PlaylistAPI)
		api.MethodFunc(http.MethodPost, "/api/music/playlist", handler.PlaylistAPI)
		api.MethodFunc(http.MethodGet, "/api/music/album", handler.AlbumAPI)
		api.MethodFunc(http.MethodPost, "/api/music/album", handler.AlbumAPI)

		api.MethodFunc(http.MethodGet, "/netease/search", handler.NeteaseSearch)
		api.MethodFunc(http.MethodPost, "/netease/search", handler.NeteaseSearch)
	})

	if staticHandler != nil {
		r.Method(http.MethodGet, "/*", staticHandler)
		r.Method(http.MethodHead, "/*", staticHandler)
	}

	return r
}
