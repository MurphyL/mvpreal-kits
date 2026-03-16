package datetime

import (
	"fmt"
	"strconv"
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
	feat.RegisterFeature(&DateTimeFeature{})
}

// DateTimeFeature 日期时间功能模块
type DateTimeFeature struct{}

// Name 返回功能模块名称
func (f *DateTimeFeature) Name() string {
	return "日期时间工具"
}

// Help 返回帮助文档
func (f *DateTimeFeature) Help() string {
	return `日期时间工具使用说明：

1. 选择操作类型：格式化、Unix时间戳转换、日期加减
2. 输入日期时间或Unix时间戳
3. 对于日期加减操作，输入要加减的天数
4. 点击"执行"按钮执行操作
5. 查看结果

操作类型说明：
- 格式化：将日期时间格式化为指定格式
- Unix时间戳转换：将Unix时间戳转换为日期时间
- 日期加减：在指定日期时间上加减天数

注意事项：
- 日期时间格式：2006-01-02 15:04:05
- Unix时间戳：秒级时间戳`
}

// Create 创建功能模块的UI组件
func (f *DateTimeFeature) Create() fyne.CanvasObject {
	// 操作类型选择
	operationType := widget.NewSelect([]string{"格式化", "Unix时间戳转换", "日期加减", "日期时间校验", "Cron 表达式"}, func(s string) {})
	operationType.SetSelected("格式化")

	// 输入文本
	inputText := widget.NewEntry()
	inputText.SetPlaceHolder("输入日期时间(2006-01-02 15:04:05)或Unix时间戳")

	// 天数输入（用于日期加减）
	daysEntry := widget.NewEntry()
	daysEntry.SetPlaceHolder("输入要加减的天数")
	daysEntry.Hidden = true

	// 结果文本
	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	// 监听操作类型变化
	operationType.OnChanged = func(s string) {
		if s == "日期加减" {
			daysEntry.Hidden = false
		} else {
			daysEntry.Hidden = true
		}
		if s == "日期时间校验" {
			inputText.SetPlaceHolder("输入要校验的日期时间(2006-01-02 15:04:05)")
		} else if s == "Cron 表达式" {
			inputText.SetPlaceHolder("输入 Cron 表达式")
		} else {
			inputText.SetPlaceHolder("输入日期时间(2006-01-02 15:04:05)或Unix时间戳")
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

			switch operationType.Selected {
			case "格式化":
				result, err = FormatDateTime(text)
			case "Unix时间戳转换":
				result, err = UnixTimestampToDateTime(text)
			case "日期加减":
				daysStr := daysEntry.Text
				if daysStr == "" {
					resultText.SetText("请输入要加减的天数")
					return
				}
				days, err := strconv.Atoi(daysStr)
				if err != nil {
					resultText.SetText("天数必须是整数")
					return
				}
				result, err = AddDaysToDateTime(text, days)
			case "日期时间校验":
				result, err = ValidateDateTime(text)
			case "Cron 表达式":
				result, err = ParseCronExpression(text)
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
		container.NewHBox(
			widget.NewLabel("输入:"),
			inputText,
		),
		container.NewHBox(
			widget.NewLabel("天数:"),
			daysEntry,
		),
		executeButton,
		widget.NewLabel("结果:"),
		resultText,
	)

	return container.NewScroll(form)
}

// FormatDateTime 格式化日期时间
func FormatDateTime(text string) (string, error) {
	// 尝试解析日期时间
	t, err := time.Parse("2006-01-02 15:04:05", text)
	if err != nil {
		return "", err
	}

	// 多种格式输出
	formats := []struct {
		name   string
		format string
	}{
		{"标准格式", "2006-01-02 15:04:05"},
		{"日期", "2006-01-02"},
		{"时间", "15:04:05"},
		{"Unix时间戳", "Unix"},
		{"RFC3339", time.RFC3339},
	}

	result := ""
	for _, f := range formats {
		if f.format == "Unix" {
			result += fmt.Sprintf("%s: %d\n", f.name, t.Unix())
		} else {
			result += fmt.Sprintf("%s: %s\n", f.name, t.Format(f.format))
		}
	}

	return result, nil
}

// UnixTimestampToDateTime Unix时间戳转换为日期时间
func UnixTimestampToDateTime(text string) (string, error) {
	// 解析Unix时间戳
	timestamp, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return "", err
	}

	// 转换为时间
	t := time.Unix(timestamp, 0)

	// 多种格式输出
	formats := []struct {
		name   string
		format string
	}{
		{"标准格式", "2006-01-02 15:04:05"},
		{"日期", "2006-01-02"},
		{"时间", "15:04:05"},
		{"RFC3339", time.RFC3339},
	}

	result := ""
	for _, f := range formats {
		result += fmt.Sprintf("%s: %s\n", f.name, t.Format(f.format))
	}

	return result, nil
}

// AddDaysToDateTime 日期加减
func AddDaysToDateTime(text string, days int) (string, error) {
	// 解析日期时间
	t, err := time.Parse("2006-01-02 15:04:05", text)
	if err != nil {
		return "", err
	}

	// 加减天数
	newT := t.AddDate(0, 0, days)

	// 输出结果
	result := fmt.Sprintf("原始日期: %s\n", t.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("加减天数: %d\n", days)
	result += fmt.Sprintf("结果日期: %s\n", newT.Format("2006-01-02 15:04:05"))

	return result, nil
}

// ValidateDateTime 日期时间校验
func ValidateDateTime(text string) (string, error) {
	// 尝试解析日期时间
	_, err := time.Parse("2006-01-02 15:04:05", text)
	if err != nil {
		return fmt.Sprintf("日期时间格式错误: %v", err), nil
	}

	return "日期时间格式正确", nil
}

// ParseCronExpression 解析 Cron 表达式
func ParseCronExpression(cron string) (string, error) {
	// 简化实现，实际项目中可以使用更完善的 Cron 表达式解析库
	fields := strings.Fields(cron)
	if len(fields) != 5 {
		return "Cron 表达式格式错误：应该包含 5 个字段", nil
	}

	// 解析每个字段
	result := "Cron 表达式解析结果：\n"
	result += fmt.Sprintf("分钟: %s\n", fields[0])
	result += fmt.Sprintf("小时: %s\n", fields[1])
	result += fmt.Sprintf("日: %s\n", fields[2])
	result += fmt.Sprintf("月: %s\n", fields[3])
	result += fmt.Sprintf("星期: %s\n", fields[4])

	return result, nil
}
