package webhook

import (
	"bytes"
	"fmt"
	"net/http"

	"mvpreal/feat"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// init 函数，自动注册功能模块
func init() {
	feat.RegisterFeature(&WebhookFeature{})
}

// WebhookFeature Webhook功能模块
type WebhookFeature struct{}

// Name 返回功能模块名称
func (f *WebhookFeature) Name() string {
	return "Webhook"
}

// Help 返回帮助文档
func (f *WebhookFeature) Help() string {
	return `Webhook执行器使用说明：

1. 选择Webhook类型：通用Webhook、企业微信群机器人
2. 输入Webhook URL
3. 输入请求体（JSON格式）
4. 点击"执行"按钮发送请求
5. 查看执行结果

企业微信群机器人说明：
- 支持文本、markdown、图片、文件等消息类型
- 文本消息格式：{"content": "消息内容"}
- Markdown消息格式：{"msgtype": "markdown", "markdown": {"content": "Markdown内容"}}`
}

// Create 创建功能模块的UI组件
func (f *WebhookFeature) Create() fyne.CanvasObject {
	// Webhook类型选择
	webhookType := widget.NewSelect([]string{"通用Webhook", "企业微信群机器人"}, func(s string) {})
	webhookType.SetSelected("通用Webhook")

	// Webhook URL
	webhookURL := widget.NewEntry()
	webhookURL.SetPlaceHolder("Webhook URL")

	// 请求体
	requestBody := widget.NewMultiLineEntry()
	requestBody.SetPlaceHolder("输入请求体（JSON格式）")
	requestBody.SetText(`{
  "content": "Hello Webhook!"
}`)

	// 结果文本
	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	// 执行按钮
	executeButton := widget.NewButton("执行", func() {
		go func() {
			url := webhookURL.Text
			body := requestBody.Text

			if url == "" {
				resultText.SetText("请输入Webhook URL")
				return
			}

			if body == "" {
				resultText.SetText("请输入请求体")
				return
			}

			// 发送HTTP POST请求
			resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
			if err != nil {
				resultText.SetText(fmt.Sprintf("发送请求失败: %v", err))
				return
			}
			defer resp.Body.Close()

			// 读取响应
			var responseBody bytes.Buffer
			_, err = responseBody.ReadFrom(resp.Body)
			if err != nil {
				resultText.SetText(fmt.Sprintf("读取响应失败: %v", err))
				return
			}

			// 构建结果
			result := fmt.Sprintf("状态码: %d\n响应: %s", resp.StatusCode, responseBody.String())
			resultText.SetText(result)
		}()
	})

	// 企业微信群机器人快捷操作
	wecomButton := widget.NewButton("企业微信群机器人快捷操作", func() {
		// 创建快捷操作对话框
		dialogContent := container.NewVBox(
			widget.NewLabel("企业微信群机器人快捷操作"),
			widget.NewLabel("选择消息类型:"),
		)

		// 消息类型选择
		msgType := widget.NewSelect([]string{"文本", "Markdown"}, func(s string) {})
		msgType.SetSelected("文本")
		dialogContent.Add(msgType)

		// 消息内容
		msgContent := widget.NewMultiLineEntry()
		msgContent.SetPlaceHolder("输入消息内容")
		dialogContent.Add(widget.NewLabel("消息内容:"))
		dialogContent.Add(msgContent)

		// 确定按钮
		confirmButton := widget.NewButton("生成请求体", func() {
			var body string
			switch msgType.Selected {
			case "文本":
				content := msgContent.Text
				body = fmt.Sprintf(`{"content": "%s"}`, content)
			case "Markdown":
				content := msgContent.Text
				body = fmt.Sprintf(`{"msgtype": "markdown", "markdown": {"content": "%s"}}`, content)
			}
			requestBody.SetText(body)
			dialog.NewInformation("成功", "请求体已生成", nil).Show()
		})

		dialogContent.Add(confirmButton)

		dialog.NewCustom("企业微信群机器人快捷操作", "关闭", dialogContent, nil).Show()
	})

	// 布局
	form := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Webhook类型:"),
			webhookType,
			layout.NewSpacer(),
		),
		widget.NewLabel("Webhook URL:"),
		webhookURL,
		widget.NewLabel("请求体:"),
		requestBody,
		container.NewHBox(
			executeButton,
			wecomButton,
		),
		widget.NewLabel("执行结果:"),
		resultText,
	)

	return container.NewScroll(form)
}
