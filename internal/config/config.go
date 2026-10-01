package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	DefaultConfigDir  = "/etc/speedtunnel"
	DefaultConfigFile = "/etc/speedtunnel/config.json"
	Version           = "1.0.0"
)

type TunnelConfig struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`      // iran | kharej
	Transport  string    `json:"transport"` // speedtls | speedhttp | speedicmp | speedreverse
	ListenPort int       `json:"listen_port"`
	RemotePort int       `json:"remote_port"`
	RemoteAddr string    `json:"remote_addr"` // kharej IP for iran side
	ControlPort int      `json:"control_port"`
	Secret     string    `json:"secret"`
	SNI        string    `json:"sni"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
}

type GlobalConfig struct {
	Version    string         `json:"version"`
	BotToken   string         `json:"bot_token"`
	BotAdminID int64          `json:"bot_admin_id"`
	Tunnels    []TunnelConfig `json:"tunnels"`
}

var (
	mu         sync.RWMutex
	cachedCfg  *GlobalConfig
	configPath string
)

func GetConfigPath() string {
	if configPath != "" {
		return configPath
	}
	if p := os.Getenv("SPEEDTUNNEL_CONFIG"); p != "" {
		return p
	}
	return DefaultConfigFile
}

func SetConfigPath(p string) { configPath = p }

func DefaultSNI() string { return "www.digikala.com" }

func Load() (*GlobalConfig, error) {
	mu.Lock()
	defer mu.Unlock()
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := &GlobalConfig{Version: Version, Tunnels: []TunnelConfig{}}
			if err := save(path, cfg); err != nil {
				return nil, err
			}
			cachedCfg = cfg
			return cfg, nil
		}
		return nil, err
	}
	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Version == "" {
		cfg.Version = Version
	}
	if cfg.Tunnels == nil {
		cfg.Tunnels = []TunnelConfig{}
	}
	cachedCfg = &cfg
	return &cfg, nil
}

func Save(cfg *GlobalConfig) error {
	mu.Lock()
	defer mu.Unlock()
	return save(GetConfigPath(), cfg)
}

func save(path string, cfg *GlobalConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func AddTunnel(t TunnelConfig) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.Tunnels = append(cfg.Tunnels, t)
	return Save(cfg)
}

func UpdateTunnel(id string, fn func(*TunnelConfig)) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	for i := range cfg.Tunnels {
		if cfg.Tunnels[i].ID == id {
			fn(&cfg.Tunnels[i])
			return Save(cfg)
		}
	}
	return fmt.Errorf("tunnel %s not found", id)
}

func DeleteTunnel(id string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	newList := cfg.Tunnels[:0]
	found := false
	for _, t := range cfg.Tunnels {
		if t.ID == id {
			found = true
			continue
		}
		newList = append(newList, t)
	}
	if !found {
		return fmt.Errorf("tunnel %s not found", id)
	}
	cfg.Tunnels = newList
	return Save(cfg)
}

func GetTunnel(id string) (*TunnelConfig, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	for _, t := range cfg.Tunnels {
		if t.ID == id {
			cp := t
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("tunnel %s not found", id)
}

func ListTunnels() ([]TunnelConfig, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	return cfg.Tunnels, nil
}
