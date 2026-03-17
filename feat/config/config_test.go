package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()
	if cfg == nil {
		t.Fatal("GetDefaultConfig() 返回 nil")
	}
	if cfg.Theme != "light" {
		t.Errorf("默认主题应为 light, 实际为 %s", cfg.Theme)
	}
	if cfg.WindowSize.Width != 800 {
		t.Errorf("默认窗口宽度应为 800, 实际为 %d", cfg.WindowSize.Width)
	}
	if cfg.WindowSize.Height != 600 {
		t.Errorf("默认窗口高度应为 600, 实际为 %d", cfg.WindowSize.Height)
	}
	if len(cfg.RecentConnections) != 0 {
		t.Errorf("默认最近连接数应为 0, 实际为 %d", len(cfg.RecentConnections))
	}
	if len(cfg.DBConnections) != 0 {
		t.Errorf("默认数据库连接数应为 0, 实际为 %d", len(cfg.DBConnections))
	}
	if len(cfg.Favorites) != 0 {
		t.Errorf("默认收藏模块数应为 0, 实际为 %d", len(cfg.Favorites))
	}
	if len(cfg.PinnedModules) != 0 {
		t.Errorf("默认置顶模块数应为 0, 实际为 %d", len(cfg.PinnedModules))
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := &Config{
		Theme: "dark",
		WindowSize: struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		}{
			Width:  1024,
			Height: 768,
		},
		RecentConnections: []ConnectionInfo{
			{
				Type:     "rdbms",
				Name:     "Test DB",
				URL:      "localhost:3306",
				Username: "test",
				Password: "test123",
				Database: "testdb",
			},
		},
		DBConnections: []DBConnection{
			{
				ID:       "1",
				Name:     "MySQL Test",
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "test",
			},
		},
		Favorites:     []string{"rdbms", "crypto"},
		PinnedModules: []string{"http"},
	}

	if cfg.Theme != "dark" {
		t.Errorf("主题应为 dark, 实际为 %s", cfg.Theme)
	}
	if cfg.WindowSize.Width != 1024 {
		t.Errorf("窗口宽度应为 1024, 实际为 %d", cfg.WindowSize.Width)
	}
	if cfg.WindowSize.Height != 768 {
		t.Errorf("窗口高度应为 768, 实际为 %d", cfg.WindowSize.Height)
	}
	if len(cfg.RecentConnections) != 1 {
		t.Errorf("最近连接数应为 1, 实际为 %d", len(cfg.RecentConnections))
	}
	if len(cfg.DBConnections) != 1 {
		t.Errorf("数据库连接数应为 1, 实际为 %d", len(cfg.DBConnections))
	}
	if len(cfg.Favorites) != 2 {
		t.Errorf("收藏模块数应为 2, 实际为 %d", len(cfg.Favorites))
	}
	if len(cfg.PinnedModules) != 1 {
		t.Errorf("置顶模块数应为 1, 实际为 %d", len(cfg.PinnedModules))
	}
}

func TestConfigFeature(t *testing.T) {
	f := &ConfigFeature{}
	if f.Name() != "配置管理" {
		t.Errorf("功能名称应为 '配置管理', 实际为 '%s'", f.Name())
	}
	help := f.Help()
	if help == "" {
		t.Error("Help() 返回空字符串")
	}
	if len(help) < 10 {
		t.Error("Help() 返回的帮助文档过短")
	}
}

func TestConfigFeatureCreate(t *testing.T) {
	f := &ConfigFeature{}
	obj := f.Create()
	if obj == nil {
		t.Error("Create() 返回 nil")
	}
}

func TestConfigFeatureHelp(t *testing.T) {
	f := &ConfigFeature{}
	help := f.Help()
	if help == "" {
		t.Error("Help() 返回空字符串")
	}
	if len(help) < 10 {
		t.Error("Help() 返回的帮助文档过短")
	}
}

func TestConfigFeatureHelpContainsInfo(t *testing.T) {
	f := &ConfigFeature{}
	help := f.Help()
	expectedSubstrings := []string{
		"配置管理",
		"使用说明",
		"Windows",
		"macOS",
		"Linux",
	}
	for _, substr := range expectedSubstrings {
		if !contains(help, substr) {
			t.Errorf("Help() 应包含 '%s'", substr)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestConfigPathInitialization(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("获取用户主目录失败: %v", err)
	}
	expectedPath := filepath.Join(homeDir, "AppData", "Roaming", "mvpreal", "config.json")
	if configPath != expectedPath {
		t.Errorf("配置路径应为 '%s', 实际为 '%s'", expectedPath, configPath)
	}
}

func TestConfigPathIsNotEmpty(t *testing.T) {
	if configPath == "" {
		t.Error("configPath 为空字符串")
	}
}

func TestConfigPathContainsFileName(t *testing.T) {
	if !contains(configPath, "config.json") {
		t.Error("configPath 应包含 'config.json'")
	}
}

func TestConfigPathContainsMvpreal(t *testing.T) {
	if !contains(configPath, "mvpreal") {
		t.Error("configPath 应包含 'mvpreal'")
	}
}
