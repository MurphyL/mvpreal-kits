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

// createSidebarItem 创建侧边栏导航项
func createSidebarItem(feature feat.Feature, isActive bool, callback func()) fyne.CanvasObject {
	button := widget.NewButton(feature.Name(), callback)
	if isActive {
		button.Importance = widget.HighImportance
	}
	return button
}

// createFeatureContent 创建功能模块的 UI 内容
func createFeatureContent(feature feat.Feature, cfg *config.Config, myWindow fyne.Window) fyne.CanvasObject {
	content := feature.Create()

	helpButton := widget.NewButton("帮助", func() {
		helpContent := widget.NewLabel(feature.Help())
		helpContent.Wrapping = fyne.TextWrapWord
		dialog.NewCustom(feature.Name()+"帮助", "关闭", container.NewScroll(helpContent), myWindow).Show()
	})

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

	cfg, err := config.Load()
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	myWindow.Resize(fyne.NewSize(float32(cfg.WindowSize.Width), float32(cfg.WindowSize.Height)))

	if cfg.Theme == "dark" {
		myApp.Settings().SetTheme(theme.DarkTheme())
	} else {
		myApp.Settings().SetTheme(theme.LightTheme())
	}

	features := registerFeatures()

	contentContainer := container.NewVBox(widget.NewLabel("选择一个功能模块"))
	// contentContainer.SetMinSize(fyne.NewSize(400, 300))

	var sidebarItems []fyne.CanvasObject
	for _, feature := range features {
		if isPinned(feature.Name(), cfg) {
			sidebarItems = append(sidebarItems, createSidebarItem(feature, false, func() {
				content := createFeatureContent(feature, cfg, myWindow)
				contentContainer.RemoveAll()
				contentContainer.Add(content)
			}))
		}
	}
	for _, feature := range features {
		if !isPinned(feature.Name(), cfg) && isFavorite(feature.Name(), cfg) {
			sidebarItems = append(sidebarItems, createSidebarItem(feature, true, func() {
				content := createFeatureContent(feature, cfg, myWindow)
				contentContainer.RemoveAll()
				contentContainer.Add(content)
			}))
		}
	}
	for _, feature := range features {
		if !isPinned(feature.Name(), cfg) && !isFavorite(feature.Name(), cfg) {
			sidebarItems = append(sidebarItems, createSidebarItem(feature, false, func() {
				content := createFeatureContent(feature, cfg, myWindow)
				contentContainer.RemoveAll()
				contentContainer.Add(content)
			}))
		}
	}

	sidebar := container.NewVBox(sidebarItems...)
	// sidebar.SetMinSize(fyne.NewSize(150, 0))

	mainContent := container.NewBorder(
		container.NewScroll(sidebar),
		nil, nil, nil,
		contentContainer,
	)

	myWindow.SetContent(mainContent)

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
