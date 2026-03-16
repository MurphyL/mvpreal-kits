package nosql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"mvpreal/feat"
	"mvpreal/feat/config"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/xuri/excelize/v2"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// init 函数，自动注册功能模块
func init() {
	feat.RegisterFeature(&ElasticsearchNoSQLFeature{})
}

// ElasticsearchNoSQLFeature Elasticsearch功能模块
type ElasticsearchNoSQLFeature struct{}

// ElasticsearchConnection Elasticsearch连接管理结构体
type ElasticsearchConnection struct {
	client *elasticsearch.Client
	hosts  []string
}

// NewElasticsearchConnection 创建新的Elasticsearch连接
func NewElasticsearchConnection(hosts []string) (*ElasticsearchConnection, error) {
	cfg := elasticsearch.Config{
		Addresses: hosts,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// 测试连接
	_, err = client.Info()
	if err != nil {
		return nil, err
	}

	return &ElasticsearchConnection{
		client: client,
		hosts:  hosts,
	}, nil
}

// Close 关闭Elasticsearch连接
func (c *ElasticsearchConnection) Close() error {
	// Elasticsearch客户端没有显式的Close方法
	// 连接会在后台自动管理
	return nil
}

// Search 执行Elasticsearch搜索
func (c *ElasticsearchConnection) Search(index string, query string) ([]byte, error) {
	res, err := c.client.Search(
		c.client.Search.WithIndex(index),
		c.client.Search.WithBody(strings.NewReader(query)),
		c.client.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Name 返回功能模块名称
func (f *ElasticsearchNoSQLFeature) Name() string {
	return "Elasticsearch"
}

// Help 返回帮助文档
func (f *ElasticsearchNoSQLFeature) Help() string {
	return `Elasticsearch客户端使用说明：

1. 输入Elasticsearch主机地址（例如：http://localhost:9200）
2. 输入索引名称
3. 输入JSON格式的查询语句
4. 点击"执行查询"按钮执行查询
5. 查看查询结果

查询语句示例：
{
  "query": {
    "match_all": {}
  }
}

注意事项：
- Elasticsearch默认端口：9200
- 确保Elasticsearch服务已启动`
}

// Create 创建功能模块的UI组件
func (f *ElasticsearchNoSQLFeature) Create() fyne.CanvasObject {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	esHost := widget.NewEntry()
	esHost.SetPlaceHolder("Elasticsearch主机地址")
	esHost.SetText("http://localhost:9200")

	// 数据库连接实例选择
	dbConnNames := []string{"新建连接"}
	dbConnMap := make(map[string]config.DBConnection)

	// 过滤出Elasticsearch类型的连接
	for _, conn := range cfg.DBConnections {
		if conn.Type == "elasticsearch" {
			dbConnNames = append(dbConnNames, conn.Name)
			dbConnMap[conn.Name] = conn
		}
	}

	dbConnSelector := widget.NewSelect(dbConnNames, func(s string) {
		if s != "新建连接" {
			// 加载选中的连接信息
			conn := dbConnMap[s]
			esHost.SetText(fmt.Sprintf("http://%s:%d", conn.Host, conn.Port))
		}
	})
	dbConnSelector.SetSelected("新建连接")

	esIndex := widget.NewEntry()
	esIndex.SetPlaceHolder("索引名称")
	esIndex.SetText("index")

	esQuery := widget.NewMultiLineEntry()
	esQuery.SetPlaceHolder("输入JSON格式的查询语句")
	esQuery.SetText(`{
  "query": {
    "match_all": {}
  }
}`)

	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	// Elasticsearch连接管理
	var esConn *ElasticsearchConnection
	var lastResponse []byte

	// 导出为JSON
	var exportToJSON func()
	// 导出为Excel
	var exportToExcel func()

	// 连接按钮
	connectButton := widget.NewButton("连接", func() {
		go func() {
			hosts := []string{esHost.Text}

			// 关闭已有的连接
			if esConn != nil {
				esConn.Close()
			}

			// 创建新连接
			conn, err := NewElasticsearchConnection(hosts)
			if err != nil {
				resultText.SetText(fmt.Sprintf("Elasticsearch连接失败: %v", err))
				return
			}

			esConn = conn
			resultText.SetText("Elasticsearch连接成功")
		}()
	})

	// 断开连接按钮
	disconnectButton := widget.NewButton("断开连接", func() {
		if esConn != nil {
			err := esConn.Close()
			if err != nil {
				resultText.SetText(fmt.Sprintf("断开连接失败: %v", err))
				return
			}
			esConn = nil
			resultText.SetText("Elasticsearch连接已断开")
		} else {
			resultText.SetText("当前没有活跃连接")
		}
	})

	executeButton := widget.NewButton("执行查询", func() {
		go func() {
			// 检查是否已连接
			if esConn == nil {
				resultText.SetText("请先连接Elasticsearch")
				return
			}

			// 执行查询
			body, err := esConn.Search(esIndex.Text, esQuery.Text)
			if err != nil {
				resultText.SetText(fmt.Sprintf("执行查询失败: %v", err))
				return
			}

			// 格式化JSON
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
				resultText.SetText(fmt.Sprintf("格式化JSON失败: %v", err))
				return
			}

			lastResponse = body
			resultText.SetText(prettyJSON.String())
		}()
	})

	exportButton := widget.NewButton("导出数据", func() {
		if esConn == nil || len(lastResponse) == 0 {
			resultText.SetText("请先执行查询获取数据")
			return
		}

		// 创建导出格式选择对话框
		formatSelect := widget.NewSelect([]string{"JSON", "Excel"}, func(s string) {
			if s == "JSON" {
				exportToJSON()
			} else if s == "Excel" {
				exportToExcel()
			}
		})
		formatSelect.SetSelected("JSON")

		dialog.NewCustom("选择导出格式", "确定", container.NewVBox(
			widget.NewLabel("请选择导出格式:"),
			formatSelect,
		), nil).Show()
	})

	// 导出为JSON
	exportToJSON = func() {
		// 生成文件名
		filename := fmt.Sprintf("elasticsearch_export_%d.json", time.Now().Unix())

		// 写入文件
		err := os.WriteFile(filename, lastResponse, 0644)
		if err != nil {
			resultText.SetText(fmt.Sprintf("导出失败: %v", err))
			return
		}

		resultText.SetText(fmt.Sprintf("数据已导出到: %s", filename))
	}

	// 导出为Excel
	exportToExcel = func() {
		// 生成文件名
		filename := fmt.Sprintf("elasticsearch_export_%d.xlsx", time.Now().Unix())

		// 解析JSON响应
		var response map[string]interface{}
		if err := json.Unmarshal(lastResponse, &response); err != nil {
			resultText.SetText(fmt.Sprintf("解析JSON失败: %v", err))
			return
		}

		// 创建Excel文件
		f := excelize.NewFile()
		sheetName := "Sheet1"

		// 写入基本信息
		f.SetCellValue(sheetName, "A1", "Elasticsearch 查询结果")
		f.SetCellValue(sheetName, "A3", "查询URL")
		f.SetCellValue(sheetName, "B3", esHost.Text+"/"+esIndex.Text+"/_search")
		f.SetCellValue(sheetName, "A4", "查询时间")
		f.SetCellValue(sheetName, "B4", time.Now().Format("2006-01-02 15:04:05"))

		// 处理命中数据
		if hits, ok := response["hits"].(map[string]interface{}); ok {
			if hitsArray, ok := hits["hits"].([]interface{}); ok {
				// 写入表头
				f.SetCellValue(sheetName, "A6", "序号")
				f.SetCellValue(sheetName, "B6", "ID")
				f.SetCellValue(sheetName, "C6", "得分")
				f.SetCellValue(sheetName, "D6", "源数据")

				// 写入数据
				for i, hit := range hitsArray {
					if hitMap, ok := hit.(map[string]interface{}); ok {
						rowNum := i + 7
						f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), i+1)
						if id, ok := hitMap["_id"].(string); ok {
							f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), id)
						}
						if score, ok := hitMap["_score"].(float64); ok {
							f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), score)
						}
						if source, ok := hitMap["_source"].(map[string]interface{}); ok {
							sourceJSON, _ := json.MarshalIndent(source, "", "  ")
							f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), string(sourceJSON))
						}
					}
				}
			}
		}

		// 保存Excel文件
		if err := f.SaveAs(filename); err != nil {
			resultText.SetText(fmt.Sprintf("保存Excel文件失败: %v", err))
			return
		}

		resultText.SetText(fmt.Sprintf("数据已导出到: %s", filename))
	}

	importButton := widget.NewButton("导入数据", func() {
		resultText.SetText("导入功能开发中...")
	})

	form := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("连接实例:"),
			dbConnSelector,
		),
		widget.NewLabel("Elasticsearch主机地址:"),
		esHost,
		widget.NewLabel("索引名称:"),
		esIndex,
		container.NewHBox(
			connectButton,
			disconnectButton,
		),
		widget.NewLabel("查询语句:"),
		esQuery,
		container.NewHBox(
			executeButton,
			exportButton,
			importButton,
		),
		widget.NewLabel("查询结果:"),
		resultText,
	)

	return container.NewScroll(form)
}
