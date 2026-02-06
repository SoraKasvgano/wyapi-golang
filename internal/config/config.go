package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Security SecurityConfig `json:"security"`
	Cookie   CookieConfig   `json:"cookie"`
	Download DownloadConfig `json:"download"`
	CORS     CORSConfig     `json:"cors"`
	Log      LogConfig      `json:"log"`
}

type ServerConfig struct {
	Host                  string `json:"host"`
	Port                  int    `json:"port"`
	ReadTimeoutSeconds    int    `json:"read_timeout_seconds"`
	WriteTimeoutSeconds   int    `json:"write_timeout_seconds"`
	IdleTimeoutSeconds    int    `json:"idle_timeout_seconds"`
	RequestTimeoutSeconds int    `json:"request_timeout_seconds"`
}

type SecurityConfig struct {
	APIToken     string `json:"api_token"`
	RequireToken bool   `json:"require_token"`
}

type CookieConfig struct {
	File string `json:"file"`
}

type DownloadConfig struct {
	Dir           string `json:"dir"`
	InMemory      bool   `json:"in_memory"`
	MaxFileSizeMB int    `json:"max_file_size_mb"`
	MaxConcurrent int    `json:"max_concurrent"`
}

type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
}

type LogConfig struct {
	Level string `json:"level"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:                  "0.0.0.0",
			Port:                  8000,
			ReadTimeoutSeconds:    15,
			WriteTimeoutSeconds:   120,
			IdleTimeoutSeconds:    60,
			RequestTimeoutSeconds: 30,
		},
		Security: SecurityConfig{
			APIToken:     "",
			RequireToken: false,
		},
		Cookie: CookieConfig{
			File: "cookie.txt",
		},
		Download: DownloadConfig{
			Dir:           "downloads",
			InMemory:      true,
			MaxFileSizeMB: 500,
			MaxConcurrent: 3,
		},
		CORS: CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Token", "X-API-Key"},
			ExposedHeaders:   []string{"X-Download-Message", "X-Download-Filename"},
			AllowCredentials: false,
		},
		Log: LogConfig{
			Level: "info",
		},
	}
}

func LoadOrCreate(path string) (*Config, bool, error) {
	cfg, err := Load(path)
	if err == nil {
		cfg.ApplyDefaults()
		if cfg.Security.APIToken == "" {
			token, tokenErr := GenerateToken(32)
			if tokenErr != nil {
				return nil, false, tokenErr
			}
			cfg.Security.APIToken = token
			if saveErr := cfg.Save(path); saveErr != nil {
				return nil, false, saveErr
			}
		}
		return cfg, false, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return nil, false, err
	}

	cfg = DefaultConfig()
	token, tokenErr := GenerateToken(32)
	if tokenErr != nil {
		return nil, false, tokenErr
	}
	cfg.Security.APIToken = token
	if saveErr := cfg.Save(path); saveErr != nil {
		return nil, false, saveErr
	}
	return cfg, true, nil
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) ApplyDefaults() {
	defaults := DefaultConfig()

	if c.Server.Host == "" {
		c.Server.Host = defaults.Server.Host
	}
	if c.Server.Port == 0 {
		c.Server.Port = defaults.Server.Port
	}
	if c.Server.ReadTimeoutSeconds == 0 {
		c.Server.ReadTimeoutSeconds = defaults.Server.ReadTimeoutSeconds
	}
	if c.Server.WriteTimeoutSeconds == 0 {
		c.Server.WriteTimeoutSeconds = defaults.Server.WriteTimeoutSeconds
	}
	if c.Server.IdleTimeoutSeconds == 0 {
		c.Server.IdleTimeoutSeconds = defaults.Server.IdleTimeoutSeconds
	}
	if c.Server.RequestTimeoutSeconds == 0 {
		c.Server.RequestTimeoutSeconds = defaults.Server.RequestTimeoutSeconds
	}

	if c.Cookie.File == "" {
		c.Cookie.File = defaults.Cookie.File
	}

	if c.Download.Dir == "" {
		c.Download.Dir = defaults.Download.Dir
	}
	if c.Download.MaxFileSizeMB == 0 {
		c.Download.MaxFileSizeMB = defaults.Download.MaxFileSizeMB
	}
	if c.Download.MaxConcurrent == 0 {
		c.Download.MaxConcurrent = defaults.Download.MaxConcurrent
	}

	if len(c.CORS.AllowedOrigins) == 0 {
		c.CORS.AllowedOrigins = defaults.CORS.AllowedOrigins
	}
	if len(c.CORS.AllowedMethods) == 0 {
		c.CORS.AllowedMethods = defaults.CORS.AllowedMethods
	}
	if len(c.CORS.AllowedHeaders) == 0 {
		c.CORS.AllowedHeaders = defaults.CORS.AllowedHeaders
	}
	if len(c.CORS.ExposedHeaders) == 0 {
		c.CORS.ExposedHeaders = defaults.CORS.ExposedHeaders
	}

	if c.Log.Level == "" {
		c.Log.Level = defaults.Log.Level
	}
}

func GenerateToken(byteLen int) (string, error) {
	if byteLen <= 0 {
		return "", errors.New("invalid token length")
	}
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
