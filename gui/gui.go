package gui

import (
	"mvpreal/feat"
	_ "mvpreal/feat/codec"
	"mvpreal/feat/config"
	_ "mvpreal/feat/config"
	_ "mvpreal/feat/crypto"
	_ "mvpreal/feat/http"
	_ "mvpreal/feat/jwt"
	_ "mvpreal/feat/nosql"
	_ "mvpreal/feat/rdbms"
	_ "mvpreal/feat/uri"
	_ "mvpreal/feat/webhook"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// registerFeatures 注册所有功能模块
// 通过导入功能模块包触发 init 函数自动注册到全局工厂
func registerFeatures() []feat.Feature {
	return feat.GlobalFeatureFactory.GetAll()
}

// createFeatureContent 创建功能模块的 UI 内容容器
// 包含功能内容和帮助、收藏、置顶按钮
func createFeatureContent(feature feat.Feature, cfg *config.Config, myWindow fyne.Window) fyne.CanvasObject {
	// 创建功能模块的 UI 组件
	content := feature.Create()

	// 创建帮助按钮
	helpButton := widget.NewButton("帮助", func() {
		helpContent := widget.NewLabel(feature.Help())
		helpContent.Wrapping = fyne.TextWrapWord
		dialog.NewCustom(feature.Name()+"帮助", "关闭", container.NewScroll(helpContent), myWindow).Show()
	})

	// 创建收藏按钮
	favoriteButton := widget.NewButton("", nil)
	isFavorite := func(name string) bool {
		for _, fav := range cfg.Favorites {
			if fav == name {
				return true
			}
		}
		return false
	}
	if isFavorite(feature.Name()) {
		favoriteButton.SetText("★")
	} else {
		favoriteButton.SetText("☆")
	}
	favoriteButton.OnTapped = func() {
		toggleFavorite(feature.Name(), cfg)
		if isFavorite(feature.Name()) {
			favoriteButton.SetText("★")
		} else {
			favoriteButton.SetText("☆")
		}
	}

	// 创建置顶按钮
	pinnedButton := widget.NewButton("", nil)
	isPinned := func(name string) bool {
		for _, pinned := range cfg.PinnedModules {
			if pinned == name {
				return true
			}
		}
		return false
	}
	if isPinned(feature.Name()) {
		pinnedButton.SetText("↑")
	} else {
		pinnedButton.SetText("↓")
	}
	pinnedButton.OnTapped = func() {
		togglePinned(feature.Name(), cfg)
		if isPinned(feature.Name()) {
			pinnedButton.SetText("↑")
		} else {
			pinnedButton.SetText("↓")
		}
	}

	// 创建包含功能内容和按钮的容器
	featureContent := container.NewBorder(
		container.NewHBox(helpButton, favoriteButton, pinnedButton),
		nil, nil, nil,
		content,
	)

	return featureContent
}

// toggleFavorite 切换模块的收藏状态
func toggleFavorite(name string, cfg *config.Config) {
	var newFavorites []string
	found := false
	for _, fav := range cfg.Favorites {
		if fav != name {
			newFavorites = append(newFavorites, fav)
		} else {
			found = true
		}
	}
	if !found {
		newFavorites = append(newFavorites, name)
	}
	cfg.Favorites = newFavorites
	config.Save(cfg)
}

// togglePinned 切换模块的置顶状态
func togglePinned(name string, cfg *config.Config) {
	var newPinned []string
	found := false
	for _, pinned := range cfg.PinnedModules {
		if pinned != name {
			newPinned = append(newPinned, pinned)
		} else {
			found = true
		}
	}
	if !found {
		newPinned = append(newPinned, name)
	}
	cfg.PinnedModules = newPinned
	config.Save(cfg)
}

// RunApp 运行应用程序
func RunApp() {
	myApp := app.New()
	myWindow := myApp.NewWindow("多功能执行器")

	// 加载配置，失败则使用默认配置
	cfg, err := config.Load()
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	// 设置窗口大小
	myWindow.Resize(fyne.NewSize(float32(cfg.WindowSize.Width), float32(cfg.WindowSize.Height)))

	// 设置主题
	if cfg.Theme == "dark" {
		myApp.Settings().SetTheme(theme.DarkTheme())
	} else {
		myApp.Settings().SetTheme(theme.LightTheme())
	}

	// 动态加载功能模块
	features := registerFeatures()
	tabs := container.NewAppTabs()

	// 首先添加置顶的模块
	for _, feature := range features {
		if isPinned(feature.Name(), cfg) {
			featureContent := createFeatureContent(feature, cfg, myWindow)
			tabs.Append(container.NewTabItem(feature.Name(), featureContent))
		}
	}

	// 然后添加收藏的模块（非置顶）
	for _, feature := range features {
		if !isPinned(feature.Name(), cfg) && isFavorite(feature.Name(), cfg) {
			featureContent := createFeatureContent(feature, cfg, myWindow)
			tabs.Append(container.NewTabItem(feature.Name(), featureContent))
		}
	}

	// 最后添加非置顶且非收藏的模块
	for _, feature := range features {
		if !isPinned(feature.Name(), cfg) && !isFavorite(feature.Name(), cfg) {
			featureContent := createFeatureContent(feature, cfg, myWindow)
			tabs.Append(container.NewTabItem(feature.Name(), featureContent))
		}
	}

	myWindow.SetContent(tabs)

	// 监听窗口关闭事件，保存窗口大小
	myWindow.SetOnClosed(func() {
		size := myWindow.Canvas().Size()
		cfg.WindowSize.Width = int(size.Width)
		cfg.WindowSize.Height = int(size.Height)
		config.Save(cfg)
	})

	myWindow.ShowAndRun()
}

// isFavorite 检查模块是否在收藏列表中
func isFavorite(name string, cfg *config.Config) bool {
	for _, fav := range cfg.Favorites {
		if fav == name {
			return true
		}
	}
	return false
}

// isPinned 检查模块是否在置顶列表中
func isPinned(name string, cfg *config.Config) bool {
	for _, pinned := range cfg.PinnedModules {
		if pinned == name {
			return true
		}
	}
	return false
}
