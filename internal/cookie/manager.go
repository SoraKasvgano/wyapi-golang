package cookie

import (
	"errors"
	"os"
	"strings"
)

type Manager struct {
	FilePath string
}

func NewManager(path string) *Manager {
	return &Manager{FilePath: path}
}

func (m *Manager) EnsureFile() error {
	if m.FilePath == "" {
		return errors.New("cookie file path empty")
	}
	if _, err := os.Stat(m.FilePath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	f, err := os.Create(m.FilePath)
	if err != nil {
		return err
	}
	return f.Close()
}

func (m *Manager) Read() (string, error) {
	if m.FilePath == "" {
		return "", errors.New("cookie file path empty")
	}
	data, err := os.ReadFile(m.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (m *Manager) Write(content string) error {
	if m.FilePath == "" {
		return errors.New("cookie file path empty")
	}
	if strings.TrimSpace(content) == "" {
		return errors.New("cookie content empty")
	}
	return os.WriteFile(m.FilePath, []byte(strings.TrimSpace(content)), 0644)
}

func (m *Manager) Parse() (map[string]string, error) {
	content, err := m.Read()
	if err != nil {
		return nil, err
	}
	return ParseCookieString(content), nil
}

func ParseCookieString(cookieString string) map[string]string {
	result := map[string]string{}
	cookieString = strings.TrimSpace(cookieString)
	cookieString = strings.TrimPrefix(cookieString, "\ufeff")
	if strings.HasPrefix(strings.ToLower(cookieString), "cookie:") {
		cookieString = strings.TrimSpace(cookieString[7:])
	}
	if cookieString == "" {
		return result
	}

	var parts []string
	if strings.Contains(cookieString, ";") {
		parts = strings.Split(cookieString, ";")
	} else if strings.Contains(cookieString, "\n") {
		parts = strings.Split(cookieString, "\n")
	} else {
		parts = []string{cookieString}
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" || value == "" {
			continue
		}
		result[key] = value
	}

	return result
}
