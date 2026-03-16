package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"mvpreal/feat"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// init 函数，自动注册功能模块
func init() {
	feat.RegisterFeature(&CryptoFeature{})
}

// CryptoFeature 加密功能模块
type CryptoFeature struct{}

// Name 返回功能模块名称
func (f *CryptoFeature) Name() string {
	return "文本加解密"
}

// Help 返回帮助文档
func (f *CryptoFeature) Help() string {
	return `文本加解密使用说明：

1. 选择模式：加密、解密
2. 输入加密密钥，长度必须为16、24或32字节
3. 输入要加密或解密的文本
4. 点击"执行"按钮执行操作
5. 查看结果

注意事项：
- 加密和解密必须使用相同的密钥
- 加密结果会进行Base64编码
- 使用AES-CFB加密算法`
}

// Create 创建功能模块的UI组件
func (f *CryptoFeature) Create() fyne.CanvasObject {
	cryptoMode := widget.NewSelect([]string{"加密", "解密"}, func(s string) {})
	cryptoMode.SetSelected("加密")

	cryptoKey := widget.NewPasswordEntry()
	cryptoKey.SetPlaceHolder("输入加密密钥（16、24或32字节）")

	cryptoText := widget.NewMultiLineEntry()
	cryptoText.SetPlaceHolder("输入要加密/解密的文本")

	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	executeButton := widget.NewButton("执行", func() {
		go func() {
			key := cryptoKey.Text
			if len(key) != 16 && len(key) != 24 && len(key) != 32 {
				resultText.SetText("密钥长度必须为16、24或32字节")
				return
			}

			text := cryptoText.Text
			if text == "" {
				resultText.SetText("请输入要处理的文本")
				return
			}

			var result string
			var err error

			if cryptoMode.Selected == "加密" {
				result, err = Encrypt(text, key)
			} else {
				result, err = Decrypt(text, key)
			}

			if err != nil {
				resultText.SetText(fmt.Sprintf("执行失败: %v", err))
				return
			}

			resultText.SetText(result)
		}()
	})

	form := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("模式:"),
			cryptoMode,
			layout.NewSpacer(),
		),
		container.NewHBox(
			widget.NewLabel("密钥:"),
			cryptoKey,
		),
		widget.NewLabel("文本:"),
		cryptoText,
		executeButton,
		widget.NewLabel("结果:"),
		resultText,
	)

	return container.NewScroll(form)
}

// 加密函数
func Encrypt(plaintext, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 解密函数
func Decrypt(ciphertext, key string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(data) < aes.BlockSize {
		return "", fmt.Errorf("密文太短")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(data, data)

	return string(data), nil
}
