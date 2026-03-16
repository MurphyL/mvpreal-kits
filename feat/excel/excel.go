package excel

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"mvpreal/feat"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// init 函数，自动注册功能模块
func init() {
	feat.RegisterFeature(&ExcelFeature{})
}

// ExcelFeature Excel功能模块
type ExcelFeature struct{}

// Name 返回功能模块名称
func (f *ExcelFeature) Name() string {
	return "Excel工具"
}

// Help 返回帮助文档
func (f *ExcelFeature) Help() string {
	return `Excel工具使用说明：

1. 选择操作类型：Excel/JSON 互转、Excel/CSV 互转、Excel 转置、Excel 合并、Excel 转 Markdown、Excel 生成 SQL
2. 根据选择的操作类型输入相应的内容
3. 点击"执行"按钮执行操作
4. 查看结果

操作类型说明：
- Excel/JSON 互转：将 Excel 数据转换为 JSON，或 JSON 数据转换为 Excel
- Excel/CSV 互转：将 Excel 数据转换为 CSV，或 CSV 数据转换为 Excel
- Excel 转置：将 Excel 数据进行行列转置
- Excel 合并：将多个 Excel worksheet 合并为一个
- Excel 转 Markdown：将 Excel 数据转换为 Markdown 表格
- Excel 生成 SQL：根据 Excel 数据生成 SQL 查询语句

注意事项：
- 输入的数据格式要正确
- 对于文件操作，需要确保文件路径正确`
}

// Create 创建功能模块的UI组件
func (f *ExcelFeature) Create() fyne.CanvasObject {
	// 操作类型选择
	operationType := widget.NewSelect([]string{
		"Excel/JSON 互转",
		"Excel/CSV 互转",
		"Excel 转置",
		"Excel 合并",
		"Excel 转 Markdown",
		"Excel 生成 SQL",
		"Excel/JSONLines 互转",
	}, func(s string) {})
	operationType.SetSelected("Excel/JSON 互转")

	// 子操作类型选择
	subOperationType := widget.NewSelect([]string{"Excel 转 JSON", "JSON 转 Excel"}, func(s string) {})
	subOperationType.SetSelected("Excel 转 JSON")

	// 输入文本
	inputText := widget.NewMultiLineEntry()
	inputText.SetPlaceHolder("输入要处理的数据")

	// 结果文本
	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	// 监听操作类型变化
	operationType.OnChanged = func(s string) {
		switch s {
		case "Excel/JSON 互转":
			subOperationType.Options = []string{"Excel 转 JSON", "JSON 转 Excel"}
			subOperationType.SetSelected("Excel 转 JSON")
			inputText.SetPlaceHolder("输入 Excel 或 JSON 数据")
		case "Excel/CSV 互转":
			subOperationType.Options = []string{"Excel 转 CSV", "CSV 转 Excel"}
			subOperationType.SetSelected("Excel 转 CSV")
			inputText.SetPlaceHolder("输入 Excel 或 CSV 数据")
		case "Excel 转置":
			subOperationType.Options = []string{"转置"}
			subOperationType.SetSelected("转置")
			inputText.SetPlaceHolder("输入 Excel 数据")
		case "Excel 合并":
			subOperationType.Options = []string{"合并"}
			subOperationType.SetSelected("合并")
			inputText.SetPlaceHolder("输入多个 Excel 数据，用 --- 分隔")
		case "Excel 转 Markdown":
			subOperationType.Options = []string{"转 Markdown"}
			subOperationType.SetSelected("转 Markdown")
			inputText.SetPlaceHolder("输入 Excel 数据")
		case "Excel 生成 SQL":
			subOperationType.Options = []string{"生成 SQL"}
			subOperationType.SetSelected("生成 SQL")
			inputText.SetPlaceHolder("输入 Excel 数据")
		case "Excel/JSONLines 互转":
			subOperationType.Options = []string{"Excel 转 JSONLines", "JSONLines 转 Excel"}
			subOperationType.SetSelected("Excel 转 JSONLines")
			inputText.SetPlaceHolder("输入 Excel 或 JSONLines 数据")
		}
	}

	// 执行按钮
	executeButton := widget.NewButton("执行", func() {
		go func() {
			text := inputText.Text
			if text == "" {
				resultText.SetText("请输入要处理的数据")
				return
			}

			var result string
			var err error

			switch operationType.Selected {
			case "Excel/JSON 互转":
				switch subOperationType.Selected {
				case "Excel 转 JSON":
					result, err = ExcelToJSON(text)
				case "JSON 转 Excel":
					result, err = JSONToExcel(text)
				}
			case "Excel/CSV 互转":
				switch subOperationType.Selected {
				case "Excel 转 CSV":
					result, err = ExcelToCSV(text)
				case "CSV 转 Excel":
					result, err = CSVToExcel(text)
				}
			case "Excel 转置":
				result, err = TransposeExcel(text)
			case "Excel 合并":
				result, err = MergeExcel(text)
			case "Excel 转 Markdown":
				result, err = ExcelToMarkdown(text)
			case "Excel 生成 SQL":
				result, err = ExcelToSQL(text)
			case "Excel/JSONLines 互转":
				switch subOperationType.Selected {
				case "Excel 转 JSONLines":
					result, err = ExcelToJSONLines(text)
				case "JSONLines 转 Excel":
					result, err = JSONLinesToExcel(text)
				}
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
			widget.NewLabel("子操作:"),
			subOperationType,
			layout.NewSpacer(),
		),
		widget.NewLabel("输入数据:"),
		inputText,
		executeButton,
		widget.NewLabel("结果:"),
		resultText,
	)

	return container.NewScroll(form)
}

// ExcelToJSON Excel 转 JSON
func ExcelToJSON(text string) (string, error) {
	// 这里使用简化的 Excel 格式，实际项目中可以使用 excelize 库处理真实的 Excel 文件
	// 简化格式：每行是一个单元格，用制表符分隔，第一行是表头
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("Excel 数据至少需要两行（表头和数据）")
	}

	// 解析表头
	headers := strings.Split(lines[0], "\t")

	// 解析数据
	var data []map[string]string
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		cells := strings.Split(line, "\t")
		row := make(map[string]string)
		for j, cell := range cells {
			if j < len(headers) {
				row[headers[j]] = cell
			}
		}
		data = append(data, row)
	}

	// 转换为 JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// JSONToExcel JSON 转 Excel
func JSONToExcel(text string) (string, error) {
	// 解析 JSON
	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		return "", err
	}

	if len(data) == 0 {
		return "", fmt.Errorf("JSON 数据为空")
	}

	// 提取表头
	headers := make([]string, 0)
	for key := range data[0] {
		headers = append(headers, key)
	}

	// 构建 Excel 数据
	var excelData strings.Builder
	// 写入表头
	excelData.WriteString(strings.Join(headers, "\t"))
	excelData.WriteString("\n")

	// 写入数据
	for _, row := range data {
		var cells []string
		for _, header := range headers {
			if value, ok := row[header]; ok {
				cells = append(cells, fmt.Sprintf("%v", value))
			} else {
				cells = append(cells, "")
			}
		}
		excelData.WriteString(strings.Join(cells, "\t"))
		excelData.WriteString("\n")
	}

	return excelData.String(), nil
}

// ExcelToCSV Excel 转 CSV
func ExcelToCSV(text string) (string, error) {
	// 这里使用简化的 Excel 格式，实际项目中可以使用 excelize 库处理真实的 Excel 文件
	lines := strings.Split(text, "\n")

	// 构建 CSV 数据
	var csvData strings.Builder
	writer := csv.NewWriter(&csvData)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		cells := strings.Split(line, "\t")
		if err := writer.Write(cells); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return csvData.String(), nil
}

// CSVToExcel CSV 转 Excel
func CSVToExcel(text string) (string, error) {
	// 解析 CSV
	reader := csv.NewReader(strings.NewReader(text))
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}

	// 构建 Excel 数据
	var excelData strings.Builder
	for _, record := range records {
		excelData.WriteString(strings.Join(record, "\t"))
		excelData.WriteString("\n")
	}

	return excelData.String(), nil
}

// TransposeExcel Excel 转置
func TransposeExcel(text string) (string, error) {
	// 解析 Excel 数据
	lines := strings.Split(text, "\n")
	var rows [][]string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		cells := strings.Split(line, "\t")
		rows = append(rows, cells)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("Excel 数据为空")
	}

	// 计算转置后的行列数
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	// 执行转置
	transposed := make([][]string, maxCols)
	for i := 0; i < maxCols; i++ {
		transposed[i] = make([]string, len(rows))
		for j := 0; j < len(rows); j++ {
			if i < len(rows[j]) {
				transposed[i][j] = rows[j][i]
			} else {
				transposed[i][j] = ""
			}
		}
	}

	// 构建转置后的 Excel 数据
	var result strings.Builder
	for _, row := range transposed {
		result.WriteString(strings.Join(row, "\t"))
		result.WriteString("\n")
	}

	return result.String(), nil
}

// MergeExcel Excel 合并
func MergeExcel(text string) (string, error) {
	// 解析多个 Excel 数据，用 --- 分隔
	excels := strings.Split(text, "---")
	if len(excels) < 2 {
		return "", fmt.Errorf("至少需要两个 Excel 数据进行合并")
	}

	// 解析第一个 Excel 作为基础
	baseLines := strings.Split(strings.TrimSpace(excels[0]), "\n")
	var merged [][]string

	for _, line := range baseLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		cells := strings.Split(line, "\t")
		merged = append(merged, cells)
	}

	// 合并其他 Excel 数据
	for i := 1; i < len(excels); i++ {
		lines := strings.Split(strings.TrimSpace(excels[i]), "\n")
		for j, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// 跳过表头（如果有的话）
			if j == 0 && len(merged) > 0 {
				continue
			}

			cells := strings.Split(line, "\t")
			merged = append(merged, cells)
		}
	}

	// 构建合并后的 Excel 数据
	var result strings.Builder
	for _, row := range merged {
		result.WriteString(strings.Join(row, "\t"))
		result.WriteString("\n")
	}

	return result.String(), nil
}

// ExcelToMarkdown Excel 转 Markdown
func ExcelToMarkdown(text string) (string, error) {
	// 解析 Excel 数据
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("Excel 数据至少需要两行（表头和数据）")
	}

	// 解析表头
	headers := strings.Split(strings.TrimSpace(lines[0]), "\t")

	// 构建 Markdown 表格
	var markdown strings.Builder

	// 写入表头
	markdown.WriteString("|")
	for _, header := range headers {
		markdown.WriteString(" " + header + " |")
	}
	markdown.WriteString("\n")

	// 写入分隔线
	markdown.WriteString("|")
	for range headers {
		markdown.WriteString(" --- |")
	}
	markdown.WriteString("\n")

	// 写入数据
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		cells := strings.Split(line, "\t")
		markdown.WriteString("|")
		for j, cell := range cells {
			if j < len(headers) {
				markdown.WriteString(" " + cell + " |")
			}
		}
		markdown.WriteString("\n")
	}

	return markdown.String(), nil
}

// ExcelToSQL Excel 生成 SQL
func ExcelToSQL(text string) (string, error) {
	// 解析 Excel 数据
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("Excel 数据至少需要两行（表头和数据）")
	}

	// 解析表头
	headers := strings.Split(strings.TrimSpace(lines[0]), "\t")

	// 构建 SQL 语句
	var sql strings.Builder
	sql.WriteString("INSERT INTO table_name (\n")

	// 写入列名
	sql.WriteString("  " + strings.Join(headers, ",\n  "))
	sql.WriteString("\n")
	sql.WriteString(") VALUES\n")

	// 写入数据
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		cells := strings.Split(line, "\t")
		sql.WriteString("  (")

		for j, cell := range cells {
			if j > 0 {
				sql.WriteString(", ")
			}
			// 简单处理，实际项目中需要根据数据类型进行适当的转义
			sql.WriteString("'" + strings.ReplaceAll(cell, "'", "''") + "'")
		}

		sql.WriteString(")")
		if i < len(lines)-1 {
			sql.WriteString(",")
		}
		sql.WriteString("\n")
	}

	return sql.String(), nil
}

// ExcelToJSONLines Excel 转 JSONLines
func ExcelToJSONLines(text string) (string, error) {
	// 解析 Excel 数据
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("Excel 数据至少需要两行（表头和数据）")
	}

	// 解析表头
	headers := strings.Split(strings.TrimSpace(lines[0]), "\t")

	// 构建 JSONLines 数据
	var jsonLines strings.Builder

	// 写入数据
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		cells := strings.Split(line, "\t")
		row := make(map[string]string)

		for j, cell := range cells {
			if j < len(headers) {
				row[headers[j]] = cell
			}
		}

		// 转换为 JSON
		jsonData, err := json.Marshal(row)
		if err != nil {
			return "", err
		}

		jsonLines.WriteString(string(jsonData))
		jsonLines.WriteString("\n")
	}

	return jsonLines.String(), nil
}

// JSONLinesToExcel JSONLines 转 Excel
func JSONLinesToExcel(text string) (string, error) {
	// 解析 JSONLines 数据
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("JSONLines 数据为空")
	}

	// 解析第一行，提取表头
	var headers []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		for key := range data {
			headers = append(headers, key)
		}
		break
	}

	if len(headers) == 0 {
		return "", fmt.Errorf("无法提取表头")
	}

	// 构建 Excel 数据
	var excelData strings.Builder

	// 写入表头
	excelData.WriteString(strings.Join(headers, "\t"))
	excelData.WriteString("\n")

	// 写入数据
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		var cells []string
		for _, header := range headers {
			if value, ok := data[header]; ok {
				cells = append(cells, fmt.Sprintf("%v", value))
			} else {
				cells = append(cells, "")
			}
		}

		excelData.WriteString(strings.Join(cells, "\t"))
		excelData.WriteString("\n")
	}

	return excelData.String(), nil
}
