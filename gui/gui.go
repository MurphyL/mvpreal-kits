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

// 注册所有功能模块
func registerFeatures() []feat.Feature {
	// 导入所有功能模块包，触发init函数自动注册
	// 这里的导入已经在文件顶部完成

	// 从全局工厂获取所有功能模块
	return feat.GlobalFeatureFactory.GetAll()
}

func RunApp() {
	myApp := app.New()
	myWindow := myApp.NewWindow("多功能执行器")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		// 加载失败，使用默认配置
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

	// 主题切换按钮
	themeButton := widget.NewButton("切换主题", func() {
		// 简单的主题切换实现
		// 注意：在不同版本的 Fyne 中，主题切换的方法可能不同
		// 这里使用一种兼容的方法
		if myApp.Settings().Theme() == theme.LightTheme() {
			myApp.Settings().SetTheme(theme.DarkTheme())
			cfg.Theme = "dark"
		} else {
			myApp.Settings().SetTheme(theme.LightTheme())
			cfg.Theme = "light"
		}
		// 保存配置
		config.Save(cfg)
	})

	// 检查模块是否在收藏列表中
	isFavorite := func(name string) bool {
		for _, fav := range cfg.Favorites {
			if fav == name {
				return true
			}
		}
		return false
	}

	// 检查模块是否在置顶列表中
	isPinned := func(name string) bool {
		for _, pinned := range cfg.PinnedModules {
			if pinned == name {
				return true
			}
		}
		return false
	}

	// 切换收藏状态
	toggleFavorite := func(name string) {
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

	// 切换置顶状态
	togglePinned := func(name string) {
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

	// 首先添加置顶的模块
	for _, feature := range features {
		if isPinned(feature.Name()) {
			// 创建功能模块的UI组件
			content := feature.Create()

			// 创建帮助按钮
			helpButton := widget.NewButton("帮助", func() {
				// 创建帮助文档对话框
				helpContent := widget.NewLabel(feature.Help())
				helpContent.Wrapping = fyne.TextWrapWord
				dialog.NewCustom(feature.Name()+"帮助", "关闭", container.NewScroll(helpContent), myWindow).Show()
			})

			// 创建收藏按钮
			favoriteButton := widget.NewButton("", nil)
			if isFavorite(feature.Name()) {
				favoriteButton.SetText("★")
			} else {
				favoriteButton.SetText("☆")
			}
			favoriteButton.OnTapped = func() {
				toggleFavorite(feature.Name())
				if isFavorite(feature.Name()) {
					favoriteButton.SetText("★")
				} else {
					favoriteButton.SetText("☆")
				}
			}

			// 创建置顶按钮
			pinnedButton := widget.NewButton("", nil)
			if isPinned(feature.Name()) {
				pinnedButton.SetText("↑")
			} else {
				pinnedButton.SetText("↓")
			}
			pinnedButton.OnTapped = func() {
				togglePinned(feature.Name())
				if isPinned(feature.Name()) {
					pinnedButton.SetText("↑")
				} else {
					pinnedButton.SetText("↓")
				}
			}

			// 创建包含功能内容和按钮的容器
			featureContent := container.NewBorder(
				container.NewHBox(helpButton, favoriteButton, pinnedButton, themeButton),
				nil, nil, nil,
				content,
			)

			tabs.Append(container.NewTabItem(feature.Name(), featureContent))
		}
	}

	// 然后添加非置顶的模块
	for _, feature := range features {
		if !isPinned(feature.Name()) {
			// 创建功能模块的UI组件
			content := feature.Create()

			// 创建帮助按钮
			helpButton := widget.NewButton("帮助", func() {
				// 创建帮助文档对话框
				helpContent := widget.NewLabel(feature.Help())
				helpContent.Wrapping = fyne.TextWrapWord
				dialog.NewCustom(feature.Name()+"帮助", "关闭", container.NewScroll(helpContent), myWindow).Show()
			})

			// 创建收藏按钮
			favoriteButton := widget.NewButton("", nil)
			if isFavorite(feature.Name()) {
				favoriteButton.SetText("★")
			} else {
				favoriteButton.SetText("☆")
			}
			favoriteButton.OnTapped = func() {
				toggleFavorite(feature.Name())
				if isFavorite(feature.Name()) {
					favoriteButton.SetText("★")
				} else {
					favoriteButton.SetText("☆")
				}
			}

			// 创建置顶按钮
			pinnedButton := widget.NewButton("", nil)
			if isPinned(feature.Name()) {
				pinnedButton.SetText("↑")
			} else {
				pinnedButton.SetText("↓")
			}
			pinnedButton.OnTapped = func() {
				togglePinned(feature.Name())
				if isPinned(feature.Name()) {
					pinnedButton.SetText("↑")
				} else {
					pinnedButton.SetText("↓")
				}
			}

			// 创建包含功能内容和按钮的容器
			featureContent := container.NewBorder(
				container.NewHBox(helpButton, favoriteButton, pinnedButton, themeButton),
				nil, nil, nil,
				content,
			)

			tabs.Append(container.NewTabItem(feature.Name(), featureContent))
		}
	}

	myWindow.SetContent(tabs)

	// 监听窗口大小变化
	myWindow.SetOnClosed(func() {
		// 保存窗口大小
		size := myWindow.Canvas().Size()
		cfg.WindowSize.Width = int(size.Width)
		cfg.WindowSize.Height = int(size.Height)
		// 保存配置
		config.Save(cfg)
	})

	myWindow.ShowAndRun()
}
