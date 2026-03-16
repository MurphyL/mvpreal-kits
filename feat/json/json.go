package json

import (
	"bytes"
	"encoding/json"
	"fmt"

	"mvpreal/feat"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// init 函数，自动注册功能模块
func init() {
	feat.RegisterFeature(&JSONFeature{})
}

// JSONFeature JSON功能模块
type JSONFeature struct{}

// Name 返回功能模块名称
func (f *JSONFeature) Name() string {
	return "JSON工具"
}

// Help 返回帮助文档
func (f *JSONFeature) Help() string {
	return `JSON工具使用说明：

1. 选择操作类型：格式化、校验、转义、反转义
2. 输入要处理的JSON文本
3. 点击"执行"按钮执行操作
4. 查看结果

操作类型说明：
- 格式化：将紧凑的JSON格式化为缩进格式，便于阅读
- 校验：检查JSON格式是否正确
- 转义：将JSON中的特殊字符进行转义
- 反转义：将转义的JSON字符串还原为原始格式

注意事项：
- 格式化和校验操作需要输入有效的JSON
- 转义和反转义操作可以处理任意文本`
}

// Create 创建功能模块的UI组件
func (f *JSONFeature) Create() fyne.CanvasObject {
	// 操作类型选择
	operationType := widget.NewSelect([]string{"格式化", "校验", "转义", "反转义"}, func(s string) {})
	operationType.SetSelected("格式化")

	// 输入文本
	inputText := widget.NewMultiLineEntry()
	inputText.SetPlaceHolder("输入要处理的JSON文本")

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

			switch operationType.Selected {
			case "格式化":
				result, err = FormatJSON(text)
			case "校验":
				result, err = ValidateJSON(text)
			case "转义":
				result = EscapeJSON(text)
			case "反转义":
				result = UnescapeJSON(text)
			}

			if err != nil {
				resultText.SetText(fmt.Sprintf("执行失败: %v", err))
				return
			}

			resultText.SetText(result)
		}()
	})

	// 布局
	form := container.NewVBox(
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

// FormatJSON 格式化JSON
func FormatJSON(text string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		return "", err
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(text), "", "  "); err != nil {
		return "", err
	}

	return prettyJSON.String(), nil
}

// ValidateJSON 校验JSON
func ValidateJSON(text string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		return fmt.Sprintf("JSON格式错误: %v", err), nil
	}

	return "JSON格式正确", nil
}

// EscapeJSON 转义JSON
func EscapeJSON(text string) string {
	escaped, _ := json.Marshal(text)
	return string(escaped)
}

// UnescapeJSON 反转义JSON
func UnescapeJSON(text string) string {
	var unescaped string
	json.Unmarshal([]byte(text), &unescaped)
	return unescaped
}
