package config

import (
	"encoding/json"
	"fmt"
	"mvpreal/feat"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Config 用户配置结构体
type Config struct {
	Theme      string `json:"theme"` // 主题：light 或 dark
	WindowSize struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"window_size"`
	RecentConnections []ConnectionInfo `json:"recent_connections"` // 最近的连接信息
	DBConnections     []DBConnection   `json:"db_connections"`     // 数据库连接实例
	Favorites         []string         `json:"favorites"`          // 收藏的功能模块
	PinnedModules     []string         `json:"pinned_modules"`     // 置顶的功能模块
}

// DBConnection 数据库连接实例结构体
type DBConnection struct {
	ID       string `json:"id"`       // 唯一标识符
	Name     string `json:"name"`     // 连接名称
	Type     string `json:"type"`     // 连接类型：mysql, postgres, sqlite, redis, elasticsearch
	Host     string `json:"host"`     // 主机地址
	Port     int    `json:"port"`     // 端口
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
	Database string `json:"database"` // 数据库名称
	DSN      string `json:"dsn"`      // 数据源名称（DSN）
}

// ConnectionInfo 连接信息结构体
type ConnectionInfo struct {
	Type     string `json:"type"`     // 连接类型：rdbms, redis, elasticsearch, webhook
	Name     string `json:"name"`     // 连接名称
	URL      string `json:"url"`      // 连接 URL 或地址
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
	Database string `json:"database"` // 数据库名称
}

// 配置文件路径
var configPath string

// init 初始化配置文件路径
// 根据操作系统设置配置文件存储位置
func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("获取用户主目录失败:", err)
		return
	}

	// 根据操作系统设置配置文件路径
	switch runtime.GOOS {
	case "windows":
		configPath = filepath.Join(homeDir, "AppData", "Roaming", "mvpreal", "config.json")
	case "darwin":
		configPath = filepath.Join(homeDir, "Library", "Application Support", "mvpreal", "config.json")
	default:
		configPath = filepath.Join(homeDir, ".config", "mvpreal", "config.json")
	}

	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Println("创建配置目录失败:", err)
	}
}

// Load 加载配置
// 从配置文件读取用户配置，如果文件不存在则返回默认配置
func Load() (*Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return GetDefaultConfig(), nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置文件
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// Save 保存配置
// 将配置序列化并写入配置文件
func Save(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// GetDefaultConfig 获取默认配置
// 返回应用程序的默认配置参数
func GetDefaultConfig() *Config {
	return &Config{
		Theme: "light",
		WindowSize: struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		}{
			Width:  800,
			Height: 600,
		},
		RecentConnections: []ConnectionInfo{},
		DBConnections:     []DBConnection{},
		Favorites:         []string{},
		PinnedModules:     []string{},
	}
}

// ConfigFeature 配置管理功能模块
type ConfigFeature struct{}

// Name 返回功能模块名称
func (f *ConfigFeature) Name() string {
	return "配置管理"
}

// Help 返回帮助文档
func (f *ConfigFeature) Help() string {
	return `配置管理使用说明：

1. 查看当前配置
2. 编辑配置（暂未实现）
3. 重置为默认配置

配置文件存储在系统用户目录下：
- Windows: %APPDATA%\mvpreal\config.json
- macOS: ~/Library/Application Support/mvpreal/config.json
- Linux: ~/.config/mvpreal/config.json`
}

// Create 创建功能模块的 UI 组件
func (f *ConfigFeature) Create() fyne.CanvasObject {
	cfg, err := Load()
	if err != nil {
		cfg = GetDefaultConfig()
	}

	// 显示配置信息
	configInfo := widget.NewLabel(fmt.Sprintf("主题: %s\n窗口大小: %dx%d\n最近连接数: %d\n数据库连接数: %d\n收藏模块数: %d\n置顶模块数: %d",
		cfg.Theme, cfg.WindowSize.Width, cfg.WindowSize.Height, len(cfg.RecentConnections), len(cfg.DBConnections), len(cfg.Favorites), len(cfg.PinnedModules)))
	configInfo.Wrapping = fyne.TextWrapWord

	// 数据库连接列表
	var selectedConnIndex int = -1
	dbConnList := widget.NewList(
		func() int {
			return len(cfg.DBConnections)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Connection")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			conn := cfg.DBConnections[id]
			label := item.(*widget.Label)
			label.SetText(fmt.Sprintf("%s (%s)", conn.Name, conn.Type))
		},
	)

	dbConnList.OnSelected = func(id widget.ListItemID) {
		selectedConnIndex = id
	}

	// 收藏模块列表
	favoritesList := widget.NewList(
		func() int {
			return len(cfg.Favorites)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Favorite")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			favorite := cfg.Favorites[id]
			label := item.(*widget.Label)
			label.SetText(favorite)
		},
	)

	// 置顶模块列表
	pinnedList := widget.NewList(
		func() int {
			return len(cfg.PinnedModules)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Pinned")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			pinned := cfg.PinnedModules[id]
			label := item.(*widget.Label)
			label.SetText(pinned)
		},
	)

	// 重置按钮
	resetButton := widget.NewButton("重置为默认配置", func() {
		defaultCfg := GetDefaultConfig()
		Save(defaultCfg)
		configInfo.SetText(fmt.Sprintf("主题: %s\n窗口大小: %dx%d\n最近连接数: %d\n数据库连接数: %d\n收藏模块数: %d\n置顶模块数: %d\n\n配置已重置为默认值",
			defaultCfg.Theme, defaultCfg.WindowSize.Width, defaultCfg.WindowSize.Height, len(defaultCfg.RecentConnections), len(defaultCfg.DBConnections), len(defaultCfg.Favorites), len(defaultCfg.PinnedModules)))
		selectedConnIndex = -1
		dbConnList.Refresh()
		favoritesList.Refresh()
		pinnedList.Refresh()
	})

	// 添加连接按钮
	addConnButton := widget.NewButton("添加数据库连接", func() {
		dialog.NewInformation("提示", "添加数据库连接功能开发中...", nil).Show()
	})

	// 编辑连接按钮
	editConnButton := widget.NewButton("编辑数据库连接", func() {
		if selectedConnIndex >= 0 && selectedConnIndex < len(cfg.DBConnections) {
			dialog.NewInformation("提示", "编辑数据库连接功能开发中...", nil).Show()
		} else {
			dialog.NewInformation("提示", "请先选择一个数据库连接", nil).Show()
		}
	})

	// 删除连接按钮
	deleteConnButton := widget.NewButton("删除数据库连接", func() {
		if selectedConnIndex >= 0 && selectedConnIndex < len(cfg.DBConnections) {
			cfg.DBConnections = append(cfg.DBConnections[:selectedConnIndex], cfg.DBConnections[selectedConnIndex+1:]...)
			Save(cfg)
			configInfo.SetText(fmt.Sprintf("主题: %s\n窗口大小: %dx%d\n最近连接数: %d\n数据库连接数: %d\n收藏模块数: %d\n置顶模块数: %d",
				cfg.Theme, cfg.WindowSize.Width, cfg.WindowSize.Height, len(cfg.RecentConnections), len(cfg.DBConnections), len(cfg.Favorites), len(cfg.PinnedModules)))
			selectedConnIndex = -1
			dbConnList.Refresh()
		} else {
			dialog.NewInformation("提示", "请先选择一个数据库连接", nil).Show()
		}
	})

	// 布局
	form := container.NewVBox(
		widget.NewLabel("当前配置:"),
		configInfo,
		resetButton,
		widget.NewSeparator(),
		widget.NewLabel("数据库连接管理:"),
		dbConnList,
		container.NewHBox(
			addConnButton,
			editConnButton,
			deleteConnButton,
		),
		widget.NewSeparator(),
		widget.NewLabel("收藏的功能模块:"),
		favoritesList,
		widget.NewSeparator(),
		widget.NewLabel("置顶的功能模块:"),
		pinnedList,
	)

	return container.NewScroll(form)
}

// init 自动注册配置管理功能模块
func init() {
	feat.RegisterFeature(&ConfigFeature{})
}
