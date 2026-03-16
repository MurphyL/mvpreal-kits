package codec

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"net/url"

	"mvpreal/feat"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// init 函数，自动注册功能模块
func init() {
	feat.RegisterFeature(&CodecFeature{})
}

// CodecFeature 编码功能模块
type CodecFeature struct{}

// Name 返回功能模块名称
func (f *CodecFeature) Name() string {
	return "编码"
}

// Help 返回帮助文档
func (f *CodecFeature) Help() string {
	return `编码功能使用说明：

1. 选择编码类型：Base64、URL、HTML、Hex
2. 选择操作类型：编码、解码
3. 输入要编码或解码的文本
4. 点击"执行"按钮执行操作
5. 查看结果

支持的编码类型：
- Base64：将二进制数据转换为ASCII字符串
- URL：将特殊字符转换为%XX格式
- HTML：将特殊字符转换为HTML实体
- Hex：将二进制数据转换为十六进制字符串

注意事项：
- 编码：将普通文本转换为对应编码格式
- 解码：将编码文本转换为普通文本`
}

// Create 创建功能模块的UI组件
func (f *CodecFeature) Create() fyne.CanvasObject {
	// 编码类型选择
	codecType := widget.NewSelect([]string{"Base64", "URL", "HTML", "Hex"}, func(s string) {})
	codecType.SetSelected("Base64")

	// 操作类型选择
	operationType := widget.NewSelect([]string{"编码", "解码"}, func(s string) {})
	operationType.SetSelected("编码")

	// 输入文本
	inputText := widget.NewMultiLineEntry()
	inputText.SetPlaceHolder("输入要编码/解码的文本")

	// 结果文本
	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	// 执行按钮
	executeButton := widget.NewButton("执行", func() {
		go func() {
			text := inputText.Text
			if text == "" {
				resultText.SetText("请输入要处理的文本")
				return
			}

			var result string
			var err error

			switch codecType.Selected {
			case "Base64":
				if operationType.Selected == "编码" {
					result = base64.StdEncoding.EncodeToString([]byte(text))
				} else {
					data, err := base64.StdEncoding.DecodeString(text)
					if err != nil {
						resultText.SetText(fmt.Sprintf("解码失败: %v", err))
						return
					}
					result = string(data)
				}
			case "URL":
				if operationType.Selected == "编码" {
					result = url.QueryEscape(text)
				} else {
					result, err = url.QueryUnescape(text)
					if err != nil {
						resultText.SetText(fmt.Sprintf("解码失败: %v", err))
						return
					}
				}
			case "HTML":
				if operationType.Selected == "编码" {
					result = html.EscapeString(text)
				} else {
					result = html.UnescapeString(text)
				}
			case "Hex":
				if operationType.Selected == "编码" {
					result = hex.EncodeToString([]byte(text))
				} else {
					data, err := hex.DecodeString(text)
					if err != nil {
						resultText.SetText(fmt.Sprintf("解码失败: %v", err))
						return
					}
					result = string(data)
				}
			}

			resultText.SetText(result)
		}()
	})

	// 布局
	form := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("编码类型:"),
			codecType,
			layout.NewSpacer(),
		),
		container.NewHBox(
			widget.NewLabel("操作类型:"),
			operationType,
			layout.NewSpacer(),
		),
		widget.NewLabel("输入文本:"),
		inputText,
		executeButton,
		widget.NewLabel("结果:"),
		resultText,
	)

	return container.NewScroll(form)
}
