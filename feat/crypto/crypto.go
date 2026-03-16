package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
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

1. 选择加密算法：AES、MD5、SHA1、SHA256、SHA512
2. 对于AES算法，选择操作模式：加密、解密
3. 对于AES算法，输入加密密钥，长度必须为16、24或32字节
4. 输入要处理的文本
5. 点击"执行"按钮执行操作
6. 查看结果

支持的算法：
- AES：对称加密算法，需要密钥
- MD5：哈希算法，生成32位十六进制字符串
- SHA1：哈希算法，生成40位十六进制字符串
- SHA256：哈希算法，生成64位十六进制字符串
- SHA512：哈希算法，生成128位十六进制字符串

注意事项：
- AES加密和解密必须使用相同的密钥
- AES加密结果会进行Base64编码
- 哈希算法是单向的，无法从哈希值还原原始文本`
}

// Create 创建功能模块的UI组件
func (f *CryptoFeature) Create() fyne.CanvasObject {
	// 加密算法选择
	cryptoAlgorithm := widget.NewSelect([]string{"AES", "MD5", "SHA1", "SHA256", "SHA512"}, func(s string) {})
	cryptoAlgorithm.SetSelected("AES")

	// 操作模式选择
	cryptoMode := widget.NewSelect([]string{"加密", "解密"}, func(s string) {})
	cryptoMode.SetSelected("加密")

	// 密钥输入
	cryptoKey := widget.NewPasswordEntry()
	cryptoKey.SetPlaceHolder("输入加密密钥（AES需要16、24或32字节）")

	// 输入文本
	cryptoText := widget.NewMultiLineEntry()
	cryptoText.SetPlaceHolder("输入要加密/解密的文本")

	// 结果文本
	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	// 监听算法变化
	cryptoAlgorithm.OnChanged = func(s string) {
		if s == "AES" {
			cryptoMode.Hidden = false
			cryptoKey.Hidden = false
		} else {
			cryptoMode.Hidden = true
			cryptoKey.Hidden = true
		}
	}

	// 执行按钮
	executeButton := widget.NewButton("执行", func() {
		go func() {
			text := cryptoText.Text
			if text == "" {
				resultText.SetText("请输入要处理的文本")
				return
			}

			var result string
			var err error

			switch cryptoAlgorithm.Selected {
			case "AES":
				key := cryptoKey.Text
				if len(key) != 16 && len(key) != 24 && len(key) != 32 {
					resultText.SetText("AES密钥长度必须为16、24或32字节")
					return
				}
				if cryptoMode.Selected == "加密" {
					result, err = Encrypt(text, key)
				} else {
					result, err = Decrypt(text, key)
				}
			case "MD5":
				result = MD5Hash(text)
			case "SHA1":
				result = SHA1Hash(text)
			case "SHA256":
				result = SHA256Hash(text)
			case "SHA512":
				result = SHA512Hash(text)
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
			widget.NewLabel("算法:"),
			cryptoAlgorithm,
			layout.NewSpacer(),
		),
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

// MD5Hash 计算MD5哈希
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// SHA1Hash 计算SHA1哈希
func SHA1Hash(text string) string {
	hash := sha1.Sum([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// SHA256Hash 计算SHA256哈希
func SHA256Hash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// SHA512Hash 计算SHA512哈希
func SHA512Hash(text string) string {
	hash := sha512.Sum512([]byte(text))
	return fmt.Sprintf("%x", hash)
}
