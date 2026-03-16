package rdbms

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"mvpreal/feat"
	"mvpreal/feat/config"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xuri/excelize/v2"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// init 函数，自动注册功能模块
func init() {
	feat.RegisterFeature(&RDBMSFeature{})
}

// RDBMSFeature RDBMS功能模块
type RDBMSFeature struct{}

// DBConnection 数据库连接管理结构体
type DBConnection struct {
	db     *sql.DB
	driver string
	dsn    string
}

// NewDBConnection 创建新的数据库连接
func NewDBConnection(driver, dsn string) (*DBConnection, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	// 设置连接池参数
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	return &DBConnection{
		db:     db,
		driver: driver,
		dsn:    dsn,
	}, nil
}

// Close 关闭数据库连接
func (c *DBConnection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Query 执行查询语句
func (c *DBConnection) Query(query string) (*sql.Rows, error) {
	return c.db.Query(query)
}

// Exec 执行非查询语句
func (c *DBConnection) Exec(query string) (sql.Result, error) {
	return c.db.Exec(query)
}

// Name 返回功能模块名称
func (f *RDBMSFeature) Name() string {
	return "RDBMS"
}

// Help 返回帮助文档
func (f *RDBMSFeature) Help() string {
	return `RDBMS执行器使用说明：

1. 选择数据库类型：MySQL、PostgreSQL、SQLite
2. 输入数据库连接信息：
   - MySQL/PostgreSQL：主机、端口、数据库名称、用户名、密码
   - SQLite：数据库名称（会生成.db文件）
3. 输入SQL语句
4. 点击"执行SQL"按钮执行语句
5. 查看执行结果

支持的SQL语句：
- SELECT 查询语句
- INSERT、UPDATE、DELETE 等非查询语句

注意事项：
- MySQL默认端口：3306
- PostgreSQL默认端口：5432`
}

// Create 创建功能模块的UI组件
func (f *RDBMSFeature) Create() fyne.CanvasObject {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	dbType := widget.NewSelect([]string{"MySQL", "PostgreSQL", "SQLite"}, func(s string) {})
	dbType.SetSelected("MySQL")

	dbHost := widget.NewEntry()
	dbHost.SetPlaceHolder("数据库主机")
	dbHost.SetText("localhost")

	dbPort := widget.NewEntry()
	dbPort.SetPlaceHolder("数据库端口")
	dbPort.SetText("3306")

	dbName := widget.NewEntry()
	dbName.SetPlaceHolder("数据库名称")

	dbUser := widget.NewEntry()
	dbUser.SetPlaceHolder("用户名")

	dbPassword := widget.NewPasswordEntry()
	dbPassword.SetPlaceHolder("密码")

	// 数据库连接实例选择
	dbConnNames := []string{"新建连接"}
	dbConnMap := make(map[string]config.DBConnection)

	// 过滤出RDBMS类型的连接
	for _, conn := range cfg.DBConnections {
		if conn.Type == "mysql" || conn.Type == "postgres" || conn.Type == "sqlite" {
			dbConnNames = append(dbConnNames, conn.Name)
			dbConnMap[conn.Name] = conn
		}
	}

	dbConnSelector := widget.NewSelect(dbConnNames, func(s string) {
		if s != "新建连接" {
			// 加载选中的连接信息
			conn := dbConnMap[s]
			switch conn.Type {
			case "mysql":
				dbType.SetSelected("MySQL")
			case "postgres":
				dbType.SetSelected("PostgreSQL")
			case "sqlite":
				dbType.SetSelected("SQLite")
			}
			dbHost.SetText(conn.Host)
			dbPort.SetText(fmt.Sprintf("%d", conn.Port))
			dbName.SetText(conn.Database)
			dbUser.SetText(conn.Username)
			dbPassword.SetText(conn.Password)
		}
	})
	dbConnSelector.SetSelected("新建连接")

	sqlEditor := widget.NewMultiLineEntry()
	sqlEditor.SetPlaceHolder("输入SQL语句")

	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	// 数据库连接管理
	var dbConn *DBConnection
	var lastColumns []string
	var lastQuery string

	// 导出为CSV
	var exportToCSV func()
	// 导出为Excel
	var exportToExcel func()

	// 连接按钮
	connectButton := widget.NewButton("连接", func() {
		go func() {
			var dsn string
			var driver string

			switch dbType.Selected {
			case "MySQL":
				driver = "mysql"
				dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
					dbUser.Text, dbPassword.Text, dbHost.Text, dbPort.Text, dbName.Text)
			case "PostgreSQL":
				driver = "postgres"
				dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
					dbHost.Text, dbPort.Text, dbUser.Text, dbPassword.Text, dbName.Text)
			case "SQLite":
				driver = "sqlite3"
				dsn = dbName.Text + ".db"
			default:
				resultText.SetText("不支持的数据库类型")
				return
			}

			// 关闭已有的连接
			if dbConn != nil {
				dbConn.Close()
			}

			// 创建新连接
			conn, err := NewDBConnection(driver, dsn)
			if err != nil {
				resultText.SetText(fmt.Sprintf("数据库连接失败: %v", err))
				return
			}

			dbConn = conn
			resultText.SetText("数据库连接成功")
		}()
	})

	// 断开连接按钮
	disconnectButton := widget.NewButton("断开连接", func() {
		if dbConn != nil {
			err := dbConn.Close()
			if err != nil {
				resultText.SetText(fmt.Sprintf("断开连接失败: %v", err))
				return
			}
			dbConn = nil
			resultText.SetText("数据库连接已断开")
		} else {
			resultText.SetText("当前没有活跃连接")
		}
	})

	executeButton := widget.NewButton("执行SQL", func() {
		go func() {
			// 检查是否已连接
			if dbConn == nil {
				resultText.SetText("请先连接数据库")
				return
			}

			// 保存查询语句
			lastQuery = sqlEditor.Text

			rows, err := dbConn.Query(sqlEditor.Text)
			if err != nil {
				// 尝试执行非查询语句
				result, err := dbConn.Exec(sqlEditor.Text)
				if err != nil {
					resultText.SetText(fmt.Sprintf("SQL执行失败: %v", err))
					return
				}
				rowsAffected, _ := result.RowsAffected()
				resultText.SetText(fmt.Sprintf("执行成功，影响行数: %d", rowsAffected))
				lastColumns = nil
				return
			}

			columns, err := rows.Columns()
			if err != nil {
				resultText.SetText(fmt.Sprintf("获取列信息失败: %v", err))
				rows.Close()
				return
			}

			lastColumns = columns

			// 构建结果
			var result string
			for _, col := range columns {
				result += col + "\t"
			}
			result += "\n"

			// 读取数据
			values := make([]sql.RawBytes, len(columns))
			scanArgs := make([]interface{}, len(values))
			for i := range values {
				scanArgs[i] = &values[i]
			}

			for rows.Next() {
				err = rows.Scan(scanArgs...)
				if err != nil {
					resultText.SetText(fmt.Sprintf("读取数据失败: %v", err))
					rows.Close()
					return
				}

				for _, v := range values {
					if v == nil {
						result += "NULL\t"
					} else {
						result += string(v) + "\t"
					}
				}
				result += "\n"
			}

			if err = rows.Err(); err != nil {
				resultText.SetText(fmt.Sprintf("遍历结果失败: %v", err))
				rows.Close()
				return
			}

			rows.Close()
			resultText.SetText(result)
		}()
	})

	exportButton := widget.NewButton("导出数据", func() {
		if dbConn == nil || lastQuery == "" || lastColumns == nil {
			resultText.SetText("请先执行查询获取数据")
			return
		}

		// 创建导出格式选择对话框
		formatSelect := widget.NewSelect([]string{"CSV", "Excel"}, func(s string) {
			if s == "CSV" {
				exportToCSV()
			} else if s == "Excel" {
				exportToExcel()
			}
		})
		formatSelect.SetSelected("CSV")

		dialog.NewCustom("选择导出格式", "确定", container.NewVBox(
			widget.NewLabel("请选择导出格式:"),
			formatSelect,
		), nil).Show()
	})

	// 导出为CSV
	exportToCSV = func() {
		// 生成文件名
		filename := fmt.Sprintf("sql_export_%d.csv", time.Now().Unix())

		// 创建文件
		file, err := os.Create(filename)
		if err != nil {
			resultText.SetText(fmt.Sprintf("创建文件失败: %v", err))
			return
		}
		defer file.Close()

		// 写入列名
		file.WriteString(strings.Join(lastColumns, ",") + "\n")

		// 使用现有连接执行查询
		rows, err := dbConn.Query(lastQuery)
		if err != nil {
			resultText.SetText(fmt.Sprintf("执行查询失败: %v", err))
			return
		}
		defer rows.Close()

		// 读取数据
		values := make([]sql.RawBytes, len(lastColumns))
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		for rows.Next() {
			err = rows.Scan(scanArgs...)
			if err != nil {
				resultText.SetText(fmt.Sprintf("读取数据失败: %v", err))
				return
			}

			// 写入数据
			var rowData []string
			for _, v := range values {
				if v == nil {
					rowData = append(rowData, "")
				} else {
					// 处理CSV中的特殊字符
					value := string(v)
					if strings.Contains(value, ",") || strings.Contains(value, "\n") || strings.Contains(value, "\r") || strings.Contains(value, "\t") {
						value = "\"" + strings.ReplaceAll(value, "\"", "\"\"") + "\""
					}
					rowData = append(rowData, value)
				}
			}
			file.WriteString(strings.Join(rowData, ",") + "\n")
		}

		if err = rows.Err(); err != nil {
			resultText.SetText(fmt.Sprintf("遍历结果失败: %v", err))
			return
		}

		resultText.SetText(fmt.Sprintf("数据已导出到: %s", filename))
	}

	// 导出为Excel
	exportToExcel = func() {
		// 生成文件名
		filename := fmt.Sprintf("sql_export_%d.xlsx", time.Now().Unix())

		// 创建Excel文件
		f := excelize.NewFile()
		sheetName := "Sheet1"

		// 写入列名
		for i, col := range lastColumns {
			cell := fmt.Sprintf("%c1", 'A'+i)
			f.SetCellValue(sheetName, cell, col)
		}

		// 使用现有连接执行查询
		rows, err := dbConn.Query(lastQuery)
		if err != nil {
			resultText.SetText(fmt.Sprintf("执行查询失败: %v", err))
			return
		}
		defer rows.Close()

		// 读取数据
		values := make([]sql.RawBytes, len(lastColumns))
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		rowNum := 2
		for rows.Next() {
			err = rows.Scan(scanArgs...)
			if err != nil {
				resultText.SetText(fmt.Sprintf("读取数据失败: %v", err))
				return
			}

			// 写入数据
			for i, v := range values {
				cell := fmt.Sprintf("%c%d", 'A'+i, rowNum)
				if v == nil {
					f.SetCellValue(sheetName, cell, nil)
				} else {
					f.SetCellValue(sheetName, cell, string(v))
				}
			}
			rowNum++
		}

		if err = rows.Err(); err != nil {
			resultText.SetText(fmt.Sprintf("遍历结果失败: %v", err))
			return
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
			layout.NewSpacer(),
		),
		container.NewHBox(
			widget.NewLabel("数据库类型:"),
			dbType,
			layout.NewSpacer(),
		),
		container.NewHBox(
			widget.NewLabel("主机:"),
			dbHost,
			widget.NewLabel("端口:"),
			dbPort,
		),
		container.NewHBox(
			widget.NewLabel("数据库:"),
			dbName,
		),
		container.NewHBox(
			widget.NewLabel("用户名:"),
			dbUser,
			widget.NewLabel("密码:"),
			dbPassword,
		),
		container.NewHBox(
			connectButton,
			disconnectButton,
			layout.NewSpacer(),
		),
		widget.NewLabel("SQL语句:"),
		sqlEditor,
		container.NewHBox(
			executeButton,
			exportButton,
			importButton,
		),
		widget.NewLabel("执行结果:"),
		resultText,
	)

	return container.NewScroll(form)
}
