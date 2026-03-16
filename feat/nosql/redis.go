package nosql

import (
	"fmt"
	"os"
	"strings"
	"time"

	"mvpreal/feat"
	"mvpreal/feat/config"

	"github.com/go-redis/redis/v8"
	"github.com/xuri/excelize/v2"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// init 函数，自动注册功能模块
func init() {
	feat.RegisterFeature(&RedisNoSQLFeature{})
}

// RedisNoSQLFeature Redis功能模块
type RedisNoSQLFeature struct{}

// RedisConnection Redis连接管理结构体
type RedisConnection struct {
	client   *redis.Client
	addr     string
	password string
	db       int
}

// NewRedisConnection 创建新的Redis连接
func NewRedisConnection(addr, password string, db int) (*RedisConnection, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx := client.Context()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		client.Close()
		return nil, err
	}

	return &RedisConnection{
		client:   client,
		addr:     addr,
		password: password,
		db:       db,
	}, nil
}

// Close 关闭Redis连接
func (c *RedisConnection) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Exec 执行Redis命令
func (c *RedisConnection) Exec(command string, args ...string) (interface{}, error) {
	ctx := c.client.Context()
	cmd := strings.ToUpper(command)

	switch cmd {
	case "GET":
		if len(args) < 1 {
			return nil, fmt.Errorf("GET命令需要指定key")
		}
		return c.client.Get(ctx, args[0]).Result()
	case "SET":
		if len(args) < 2 {
			return nil, fmt.Errorf("SET命令需要指定key和value")
		}
		return c.client.Set(ctx, args[0], args[1], 0).Result()
	case "DEL":
		if len(args) < 1 {
			return nil, fmt.Errorf("DEL命令需要指定key")
		}
		return c.client.Del(ctx, args...).Result()
	case "KEYS":
		if len(args) < 1 {
			return nil, fmt.Errorf("KEYS命令需要指定模式")
		}
		return c.client.Keys(ctx, args[0]).Result()
	case "PING":
		return c.client.Ping(ctx).Result()
	default:
		return nil, fmt.Errorf("不支持的命令")
	}
}

// Name 返回功能模块名称
func (f *RedisNoSQLFeature) Name() string {
	return "Redis"
}

// Help 返回帮助文档
func (f *RedisNoSQLFeature) Help() string {
	return `Redis客户端使用说明：

1. 输入Redis连接信息：主机、端口、密码、数据库编号
2. 输入Redis命令，格式：命令 参数1 参数2 ...
3. 点击"执行命令"按钮执行命令
4. 查看执行结果

支持的命令：
- GET key：获取键的值
- SET key value：设置键的值
- DEL key1 key2 ...：删除一个或多个键
- KEYS pattern：查找匹配模式的键
- PING：测试连接

注意事项：
- Redis默认端口：6379
- 默认数据库编号：0`
}

// Create 创建功能模块的UI组件
func (f *RedisNoSQLFeature) Create() fyne.CanvasObject {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	redisHost := widget.NewEntry()
	redisHost.SetPlaceHolder("Redis主机")
	redisHost.SetText("localhost")

	redisPort := widget.NewEntry()
	redisPort.SetPlaceHolder("Redis端口")
	redisPort.SetText("6379")

	redisPassword := widget.NewPasswordEntry()
	redisPassword.SetPlaceHolder("Redis密码")

	redisDB := widget.NewEntry()
	redisDB.SetPlaceHolder("数据库编号")
	redisDB.SetText("0")

	// 数据库连接实例选择
	dbConnNames := []string{"新建连接"}
	dbConnMap := make(map[string]config.DBConnection)

	// 过滤出Redis类型的连接
	for _, conn := range cfg.DBConnections {
		if conn.Type == "redis" {
			dbConnNames = append(dbConnNames, conn.Name)
			dbConnMap[conn.Name] = conn
		}
	}

	dbConnSelector := widget.NewSelect(dbConnNames, func(s string) {
		if s != "新建连接" {
			// 加载选中的连接信息
			conn := dbConnMap[s]
			redisHost.SetText(conn.Host)
			redisPort.SetText(fmt.Sprintf("%d", conn.Port))
			redisPassword.SetText(conn.Password)
			redisDB.SetText(conn.Database)
		}
	})
	dbConnSelector.SetSelected("新建连接")

	redisCommand := widget.NewEntry()
	redisCommand.SetPlaceHolder("输入Redis命令，格式：命令 参数1 参数2 ...")

	resultText := widget.NewLabel("")
	resultText.Wrapping = fyne.TextWrapWord

	// Redis连接管理
	var redisConn *RedisConnection
	var lastCommand string
	var lastResult string

	// 导出为TXT
	var exportToTXT func()
	// 导出为Excel
	var exportToExcel func()

	// 连接按钮
	connectButton := widget.NewButton("连接", func() {
		go func() {
			addr := redisHost.Text + ":" + redisPort.Text
			password := redisPassword.Text
			db := 0

			// 解析数据库编号
			if redisDB.Text != "" {
				_, err := fmt.Sscanf(redisDB.Text, "%d", &db)
				if err != nil {
					resultText.SetText("数据库编号必须是数字")
					return
				}
			}

			// 关闭已有的连接
			if redisConn != nil {
				redisConn.Close()
			}

			// 创建新连接
			conn, err := NewRedisConnection(addr, password, db)
			if err != nil {
				resultText.SetText(fmt.Sprintf("Redis连接失败: %v", err))
				return
			}

			redisConn = conn
			resultText.SetText("Redis连接成功")
		}()
	})

	// 断开连接按钮
	disconnectButton := widget.NewButton("断开连接", func() {
		if redisConn != nil {
			err := redisConn.Close()
			if err != nil {
				resultText.SetText(fmt.Sprintf("断开连接失败: %v", err))
				return
			}
			redisConn = nil
			resultText.SetText("Redis连接已断开")
		} else {
			resultText.SetText("当前没有活跃连接")
		}
	})

	executeButton := widget.NewButton("执行命令", func() {
		go func() {
			// 检查是否已连接
			if redisConn == nil {
				resultText.SetText("请先连接Redis")
				return
			}

			cmd := redisCommand.Text
			args := strings.Fields(cmd)

			if len(args) == 0 {
				resultText.SetText("请输入Redis命令")
				return
			}

			command := args[0]
			var cmdArgs []string
			if len(args) > 1 {
				cmdArgs = args[1:]
			}

			result, err := redisConn.Exec(command, cmdArgs...)
			if err != nil {
				resultText.SetText(fmt.Sprintf("执行失败: %v", err))
				return
			}

			lastCommand = cmd
			lastResult = fmt.Sprintf("%v", result)
			resultText.SetText(fmt.Sprintf("执行结果: %v", result))
		}()
	})

	exportButton := widget.NewButton("导出数据", func() {
		if redisConn == nil || lastResult == "" {
			resultText.SetText("请先执行命令获取数据")
			return
		}

		// 创建导出格式选择对话框
		formatSelect := widget.NewSelect([]string{"TXT", "Excel"}, func(s string) {
			if s == "TXT" {
				exportToTXT()
			} else if s == "Excel" {
				exportToExcel()
			}
		})
		formatSelect.SetSelected("TXT")

		dialog.NewCustom("选择导出格式", "确定", container.NewVBox(
			widget.NewLabel("请选择导出格式:"),
			formatSelect,
		), nil).Show()
	})

	// 导出为TXT
	exportToTXT = func() {
		// 生成文件名
		filename := fmt.Sprintf("redis_export_%d.txt", time.Now().Unix())

		// 创建文件
		file, err := os.Create(filename)
		if err != nil {
			resultText.SetText(fmt.Sprintf("创建文件失败: %v", err))
			return
		}
		defer file.Close()

		// 写入命令和结果
		file.WriteString(fmt.Sprintf("命令: %s\n", lastCommand))
		file.WriteString(fmt.Sprintf("结果: %v\n", lastResult))

		resultText.SetText(fmt.Sprintf("数据已导出到: %s", filename))
	}

	// 导出为Excel
	exportToExcel = func() {
		// 生成文件名
		filename := fmt.Sprintf("redis_export_%d.xlsx", time.Now().Unix())

		// 创建Excel文件
		f := excelize.NewFile()
		sheetName := "Sheet1"

		// 写入命令和结果
		f.SetCellValue(sheetName, "A1", "命令")
		f.SetCellValue(sheetName, "B1", lastCommand)
		f.SetCellValue(sheetName, "A2", "结果")
		f.SetCellValue(sheetName, "B2", fmt.Sprintf("%v", lastResult))

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
		container.NewHBox(
			widget.NewLabel("主机:"),
			redisHost,
			widget.NewLabel("端口:"),
			redisPort,
		),
		container.NewHBox(
			widget.NewLabel("密码:"),
			redisPassword,
			widget.NewLabel("数据库:"),
			redisDB,
		),
		container.NewHBox(
			connectButton,
			disconnectButton,
		),
		widget.NewLabel("Redis命令:"),
		redisCommand,
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
