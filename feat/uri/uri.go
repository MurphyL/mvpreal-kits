package uri

import (
	"fmt"
	"net/url"
	"strings"

	"mvpreal/feat"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// init 函数，自动注册功能模块
func init() {
	feat.RegisterFeature(&URIFeature{})
}

// URIFeature URI功能模块
type URIFeature struct{}

// Name 返回功能模块名称
func (f *URIFeature) Name() string {
	return "URI"
}

// Help 返回帮助文档
func (f *URIFeature) Help() string {
	return `URL编码使用说明：

1. 选择模式：编码、解码、解析QueryParams
2. 输入要处理的URL
3. 点击"执行"按钮执行操作
4. 查看结果

模式说明：
- 编码：将普通文本转换为URL编码格式
- 解码：将URL编码格式转换为普通文本
- 解析QueryParams：解析URL中的查询参数，格式为key=value&key=value

注意事项：
- URL编码会将特殊字符转换为%XX格式
- 解析QueryParams时，会提取URL中?后面的部分`
}

// Create 创建功能模块的UI组件
func (f *URIFeature) Create() fyne.CanvasObject {
	urlMode := widget.NewSelect([]string{"编码", "解码", "解析QueryParams"}, func(s string) {})
	urlMode.SetSelected("编码")

	urlText := widget.NewMultiLineEntry()
	urlText.SetPlaceHolder("输入要编码/解码的URL或包含QueryParams的URL")

	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	executeButton := widget.NewButton("执行", func() {
		go func() {
			text := urlText.Text
			if text == "" {
				resultText.SetText("请输入要处理的URL")
				return
			}

			var result string

			switch urlMode.Selected {
			case "编码":
				result = url.QueryEscape(text)
			case "解码":
				data, err := url.QueryUnescape(text)
				if err != nil {
					resultText.SetText(fmt.Sprintf("解码失败: %v", err))
					return
				}
				result = data
			case "解析QueryParams":
				// 解析URL中的QueryParams
				u, err := url.Parse(text)
				if err != nil {
					resultText.SetText(fmt.Sprintf("解析URL失败: %v", err))
					return
				}

				// 获取QueryParams
				params := u.Query()
				if len(params) == 0 {
					resultText.SetText("URL中没有QueryParams")
					return
				}

				// 格式化QueryParams
				var sb strings.Builder
				for key, values := range params {
					sb.WriteString(fmt.Sprintf("%s: %v\n", key, values))
				}
				result = sb.String()
			}

			resultText.SetText(result)
		}()
	})

	form := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("模式:"),
			urlMode,
			layout.NewSpacer(),
		),
		widget.NewLabel("URL:"),
		urlText,
		executeButton,
		widget.NewLabel("结果:"),
		resultText,
	)

	return container.NewScroll(form)
}
