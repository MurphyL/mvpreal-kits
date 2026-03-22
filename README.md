
- [项目介绍](#项目介绍)

- [界面原型设计](https://www.figma.com/proto/RgF4TvXGbMMOT7kbASzDeN/Chat-for-desktop-mobile-%7C-Free-to-use--Community-?node-id=8-1346&t=l1EQedW799LhKm73-1)

```
package gui

import (
	"mvpreal/feat/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

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
	contentContainer.Resize(fyne.NewSize(400, 300))

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

```