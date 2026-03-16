package jwt

import (
	"encoding/json"
	"fmt"

	"mvpreal/feat"

	"github.com/golang-jwt/jwt/v5"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// init 函数，自动注册功能模块
func init() {
	feat.RegisterFeature(&JWTFeature{})
}

// JWTFeature JWT功能模块
type JWTFeature struct{}

// Name 返回功能模块名称
func (f *JWTFeature) Name() string {
	return "JWT"
}

// Help 返回帮助文档
func (f *JWTFeature) Help() string {
	return `JWT使用说明：

1. 选择模式：生成、解析
2. 输入JWT密钥
3. 输入载荷/令牌：
   - 生成模式：输入JSON格式的载荷
   - 解析模式：输入JWT令牌
4. 点击"执行"按钮执行操作
5. 查看结果

模式说明：
- 生成：根据载荷和密钥生成JWT令牌
- 解析：根据密钥解析JWT令牌，提取载荷

注意事项：
- 使用HS256算法签名
- 载荷必须是有效的JSON格式
- 解析时需要使用与生成时相同的密钥`
}

// Create 创建功能模块的UI组件
func (f *JWTFeature) Create() fyne.CanvasObject {
	jwtMode := widget.NewSelect([]string{"生成", "解析"}, func(s string) {})
	jwtMode.SetSelected("生成")

	jwtSecret := widget.NewPasswordEntry()
	jwtSecret.SetPlaceHolder("输入JWT密钥")

	jwtPayload := widget.NewMultiLineEntry()
	jwtPayload.SetPlaceHolder("输入JWT载荷JSON")
	jwtPayload.SetText(`{
  "sub": "1234567890",
  "name": "John Doe",
  "iat": 1516239022
}`)

	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	executeButton := widget.NewButton("执行", func() {
		go func() {
			secret := jwtSecret.Text
			if secret == "" {
				resultText.SetText("请输入JWT密钥")
				return
			}

			if jwtMode.Selected == "生成" {
				var payload map[string]interface{}
				if err := json.Unmarshal([]byte(jwtPayload.Text), &payload); err != nil {
					resultText.SetText(fmt.Sprintf("解析JSON失败: %v", err))
					return
				}

				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payload))
				tokenString, err := token.SignedString([]byte(secret))
				if err != nil {
					resultText.SetText(fmt.Sprintf("生成JWT失败: %v", err))
					return
				}

				resultText.SetText(tokenString)
			} else {
				tokenString := jwtPayload.Text
				if tokenString == "" {
					resultText.SetText("请输入JWT令牌")
					return
				}

				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return []byte(secret), nil
				})

				if err != nil {
					resultText.SetText(fmt.Sprintf("解析JWT失败: %v", err))
					return
				}

				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					prettyClaims, err := json.MarshalIndent(claims, "", "  ")
					if err != nil {
						resultText.SetText(fmt.Sprintf("格式化载荷失败: %v", err))
						return
					}
					resultText.SetText(string(prettyClaims))
				} else {
					resultText.SetText("无效的JWT令牌")
				}
			}
		}()
	})

	form := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("模式:"),
			jwtMode,
			layout.NewSpacer(),
		),
		container.NewHBox(
			widget.NewLabel("密钥:"),
			jwtSecret,
		),
		widget.NewLabel("载荷/令牌:"),
		jwtPayload,
		executeButton,
		widget.NewLabel("结果:"),
		resultText,
	)

	return container.NewScroll(form)
}
