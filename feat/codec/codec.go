package codec

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"net"
	"net/url"
	"strconv"
	"strings"

	"mvpreal/feat"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/golang-jwt/jwt/v5"
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

1. 选择编码类型：Base64、URL、HTML、Hex、JWT
2. 选择操作类型：编码、解码
3. 输入要编码或解码的文本
4. 对于JWT类型，还需要输入JWT密钥
5. 点击"执行"按钮执行操作
6. 查看结果

支持的编码类型：
- Base64：将二进制数据转换为ASCII字符串
- URL：将特殊字符转换为%XX格式
- HTML：将特殊字符转换为HTML实体
- Hex：将二进制数据转换为十六进制字符串
- JWT：JSON Web Token的生成和解析

注意事项：
- 编码：将普通文本转换为对应编码格式
- 解码：将编码文本转换为普通文本
- JWT编码：输入JSON格式的载荷，生成JWT令牌
- JWT解码：输入JWT令牌，解析出载荷`
}

// Create 创建功能模块的UI组件
func (f *CodecFeature) Create() fyne.CanvasObject {
	// 编码类型选择
	codecType := widget.NewSelect([]string{"Base64", "URL", "HTML", "Hex", "JWT", "JSON", "XML", "ASCII", "人民币大写", "IPv4/IPv6", "单位换算", "键盘 KeyCode"}, func(s string) {})
	codecType.SetSelected("Base64")

	// 操作类型选择
	operationType := widget.NewSelect([]string{"编码", "解码"}, func(s string) {})
	operationType.SetSelected("编码")

	// 输入文本
	inputText := widget.NewMultiLineEntry()
	inputText.SetPlaceHolder("输入要编码/解码的文本")

	// JWT密钥输入
	jwtSecret := widget.NewPasswordEntry()
	jwtSecret.SetPlaceHolder("输入JWT密钥")
	jwtSecret.Hidden = true

	// 结果文本
	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	// 监听编码类型变化
	codecType.OnChanged = func(s string) {
		if s == "JWT" {
			jwtSecret.Hidden = false
		} else {
			jwtSecret.Hidden = true
		}
	}

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
			case "JWT":
				secret := jwtSecret.Text
				if secret == "" {
					resultText.SetText("请输入JWT密钥")
					return
				}
				if operationType.Selected == "编码" {
					var payload map[string]interface{}
					if err := json.Unmarshal([]byte(text), &payload); err != nil {
						resultText.SetText(fmt.Sprintf("解析JSON失败: %v", err))
						return
					}
					token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payload))
					tokenString, err := token.SignedString([]byte(secret))
					if err != nil {
						resultText.SetText(fmt.Sprintf("生成JWT失败: %v", err))
						return
					}
					result = tokenString
				} else {
					tokenString := text
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
						result = string(prettyClaims)
					} else {
						resultText.SetText("无效的JWT令牌")
						return
					}
				}
			case "JSON":
				if operationType.Selected == "编码" {
					jsonData, err := json.Marshal(text)
					if err != nil {
						resultText.SetText(fmt.Sprintf("编码失败: %v", err))
						return
					}
					result = string(jsonData)
				} else {
					var decoded string
					err := json.Unmarshal([]byte(text), &decoded)
					if err != nil {
						resultText.SetText(fmt.Sprintf("解码失败: %v", err))
						return
					}
					result = decoded
				}
			case "XML":
				if operationType.Selected == "编码" {
					xmlData, err := xml.Marshal(text)
					if err != nil {
						resultText.SetText(fmt.Sprintf("编码失败: %v", err))
						return
					}
					result = string(xmlData)
				} else {
					var decoded string
					err := xml.Unmarshal([]byte(text), &decoded)
					if err != nil {
						resultText.SetText(fmt.Sprintf("解码失败: %v", err))
						return
					}
					result = decoded
				}
			case "ASCII":
				if operationType.Selected == "编码" {
					// 文本转ASCII码
					var asciiCodes []string
					for _, r := range text {
						asciiCodes = append(asciiCodes, strconv.Itoa(int(r)))
					}
					result = strings.Join(asciiCodes, " ")
				} else {
					// ASCII码转文本
					codes := strings.Split(text, " ")
					var chars []rune
					for _, code := range codes {
						code = strings.TrimSpace(code)
						if code == "" {
							continue
						}
						n, err := strconv.Atoi(code)
						if err != nil {
							resultText.SetText(fmt.Sprintf("解码失败: %v", err))
							return
						}
						chars = append(chars, rune(n))
					}
					result = string(chars)
				}
			case "人民币大写":
				result = ChineseNumber(text)
			case "IPv4/IPv6":
				if operationType.Selected == "编码" {
					// IPv4 转 IPv6
					result, err = IPv4ToIPv6(text)
				} else {
					// IPv6 转 IPv4
					result, err = IPv6ToIPv4(text)
				}
			case "单位换算":
				result = "单位换算功能开发中..."
			case "键盘 KeyCode":
				result = "键盘 KeyCode 值查询功能开发中..."
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
		container.NewHBox(
			widget.NewLabel("JWT密钥:"),
			jwtSecret,
		),
		widget.NewLabel("输入文本:"),
		inputText,
		executeButton,
		widget.NewLabel("结果:"),
		resultText,
	)

	return container.NewScroll(form)
}

// ChineseNumber 将数字转换为人民币大写
func ChineseNumber(num string) string {
	// 简化实现，实际项目中可以使用更完善的实现
	chinese := []string{"零", "壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖"}
	units := []string{"", "拾", "佰", "仟"}
	bigUnits := []string{"", "万", "亿"}

	// 处理整数部分
	integerPart := num
	decimalPart := ""

	// 分离整数和小数部分
	if dotIndex := strings.Index(num, "."); dotIndex != -1 {
		integerPart = num[:dotIndex]
		decimalPart = num[dotIndex+1:]
	}

	// 处理整数部分
	var result strings.Builder
	length := len(integerPart)
	for i := 0; i < length; i++ {
		digit := integerPart[i] - '0'
		if digit == 0 {
			// 处理连续的零
			if i > 0 && integerPart[i-1] != '0' {
				result.WriteString(chinese[0])
			}
		} else {
			result.WriteString(chinese[digit])
			result.WriteString(units[(length-i-1)%4])
		}
		// 处理大单位
		if (length-i-1)%4 == 0 && length-i-1 > 0 {
			result.WriteString(bigUnits[(length-i-1)/4])
		}
	}

	// 处理小数部分
	if decimalPart != "" {
		result.WriteString("点")
		for _, d := range decimalPart {
			digit := d - '0'
			result.WriteString(chinese[digit])
		}
	}

	return result.String()
}

// IPv4ToIPv6 IPv4 转 IPv6
func IPv4ToIPv6(ipv4 string) (string, error) {
	// 解析 IPv4 地址
	ip := net.ParseIP(ipv4)
	if ip == nil {
		return "", fmt.Errorf("无效的 IPv4 地址")
	}

	// 确保是 IPv4 地址
	ipv4Addr := ip.To4()
	if ipv4Addr == nil {
		return "", fmt.Errorf("不是 IPv4 地址")
	}

	// 转换为 IPv6 地址
	ipv6Addr := net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, ipv4Addr[0], ipv4Addr[1], ipv4Addr[2], ipv4Addr[3]}

	return ipv6Addr.String(), nil
}

// IPv6ToIPv4 IPv6 转 IPv4
func IPv6ToIPv4(ipv6 string) (string, error) {
	// 解析 IPv6 地址
	ip := net.ParseIP(ipv6)
	if ip == nil {
		return "", fmt.Errorf("无效的 IPv6 地址")
	}

	// 转换为 IPv4 地址
	ipv4Addr := ip.To4()
	if ipv4Addr == nil {
		return "", fmt.Errorf("不是 IPv4 映射的 IPv6 地址")
	}

	return ipv4Addr.String(), nil
}
