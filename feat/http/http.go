package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mvpreal/feat"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// init 函数，自动注册功能模块
func init() {
	feat.RegisterFeature(&HTTPFeature{})
}

// HTTPFeature HTTP功能模块
type HTTPFeature struct{}

// Name 返回功能模块名称
func (f *HTTPFeature) Name() string {
	return "HTTP请求"
}

// Help 返回帮助文档
func (f *HTTPFeature) Help() string {
	return `HTTP请求执行器使用说明：

1. 选择请求方法：GET、POST、PUT、DELETE、PATCH
2. 输入请求URL
3. 输入请求头，格式：Key: Value，每行一个
4. 输入请求体
5. 点击"发送请求"按钮执行请求
6. 查看响应结果，包括状态码、响应头和响应体

注意事项：
- POST、PUT、PATCH 方法需要设置请求体
- 默认Content-Type为application/json
- 请求超时时间为30秒`
}

// Create 创建功能模块的UI组件
func (f *HTTPFeature) Create() fyne.CanvasObject {
	method := widget.NewSelect([]string{"GET", "POST", "PUT", "DELETE", "PATCH"}, func(s string) {})
	method.SetSelected("GET")

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("输入请求URL")

	headerEditor := widget.NewMultiLineEntry()
	headerEditor.SetPlaceHolder("输入请求头，格式：Key: Value")

	bodyEditor := widget.NewMultiLineEntry()
	bodyEditor.SetPlaceHolder("输入请求体")

	responseText := widget.NewLabel("")
	responseText.Wrapping = fyne.TextWrapWord

	sendButton := widget.NewButton("发送请求", func() {
		go func() {
			client := &http.Client{
				Timeout: 30 * time.Second,
			}

			req, err := http.NewRequest(method.Selected, urlEntry.Text, strings.NewReader(bodyEditor.Text))
			if err != nil {
				responseText.SetText(fmt.Sprintf("错误: %v", err))
				return
			}

			// 解析请求头
			headers := strings.Split(headerEditor.Text, "\n")
			for _, header := range headers {
				header = strings.TrimSpace(header)
				if header == "" {
					continue
				}
				parts := strings.SplitN(header, ":", 2)
				if len(parts) == 2 {
					req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}

			// 设置默认Content-Type
			if req.Method != "GET" && req.Header.Get("Content-Type") == "" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := client.Do(req)
			if err != nil {
				responseText.SetText(fmt.Sprintf("错误: %v", err))
				return
			}
			defer resp.Body.Close()

			// 读取响应
			responseBody := make([]byte, 1024*1024) // 1MB 缓冲区
			n, _ := resp.Body.Read(responseBody)
			responseBody = responseBody[:n]

			// 格式化响应
			var formattedBody string
			if resp.Header.Get("Content-Type") == "application/json" {
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, responseBody, "", "  "); err == nil {
					formattedBody = prettyJSON.String()
				} else {
					formattedBody = string(responseBody)
				}
			} else {
				formattedBody = string(responseBody)
			}

			responseText.SetText(fmt.Sprintf("状态码: %d\n\n响应头:\n%v\n\n响应体:\n%s",
				resp.StatusCode, resp.Header, formattedBody))
		}()
	})

	form := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("方法:"),
			method,
			layout.NewSpacer(),
		),
		container.NewHBox(
			widget.NewLabel("URL:"),
			urlEntry,
		),
		widget.NewLabel("请求头:"),
		headerEditor,
		widget.NewLabel("请求体:"),
		bodyEditor,
		sendButton,
		widget.NewLabel("响应:"),
		responseText,
	)

	return container.NewScroll(form)
}
