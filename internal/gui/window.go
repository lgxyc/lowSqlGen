package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/lowSqlGen/internal/model"
	"github.com/lowSqlGen/internal/service"
)

type MainWindow struct {
	window    fyne.Window
	canvas    *Canvas
	leftBar   *widget.Tree
	rightBar  *widget.Entry
	dbConfig  *model.DatabaseConfig
	dbService service.DatabaseService
	dbTables  map[string][]string
	firstTable       bool
	currentAddedTable string
}

func InitMainWindow(window fyne.Window) *MainWindow {
	mainWindow := &MainWindow{
		window:           window,
		rightBar:         widget.NewEntry(),
		dbTables:        make(map[string][]string),
		firstTable:      true,  // Initialize state
		currentAddedTable: "",
	}

	// Remove local state variables and use struct fields instead
	mainWindow.canvas = NewCanvas(nil, nil, mainWindow) // Pass mainWindow for state access

	// 修改树形结构的创建
	mainWindow.leftBar = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				// root节点返回所有数据库
				return mainWindow.getDatabases()
			}
			// 数据库节点返回其包含的表
			if tables, ok := mainWindow.dbTables[id]; ok {
				var nodeIDs []widget.TreeNodeID
				for _, table := range tables {
					nodeIDs = append(nodeIDs, id+"/"+table)
				}
				return nodeIDs
			}
			return nil
		},
		func(id widget.TreeNodeID) bool {
			// 如果是数据库节点则返回true
			return !strings.Contains(id, "/")
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return widget.NewLabel("Template")
			}
			// 为表节点创建一个容器，包含标签和按钮
			label := widget.NewLabel("Template")
			btn := widget.NewButton("Add", nil)
			return container.NewBorder(nil, nil, nil, btn, label)
		},
		func(id widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			if branch {
				label := node.(*widget.Label)
				label.SetText(id)
				return
			}

			// 表节点
			cont := node.(*fyne.Container)
			label := cont.Objects[0].(*widget.Label)
			btn := cont.Objects[1].(*widget.Button)

			parts := strings.Split(id, "/")
			dbName := parts[0]
			tableName := parts[1]

			// 设置表名和注释
			if mainWindow.dbService != nil {
				comment := mainWindow.dbService.GetTableComment(dbName, tableName)
				if comment != "" {
					label.SetText(fmt.Sprintf("%s // %s", tableName, comment))
				} else {
					label.SetText(tableName)
				}
			} else {
				label.SetText(tableName)
			}

			// 根据状态设置按钮
			if mainWindow.currentAddedTable == tableName {
				btn.SetText("Cancel")
				btn.OnTapped = func() {
					mainWindow.canvas.Clear()
					mainWindow.currentAddedTable = ""
					mainWindow.firstTable = true
					mainWindow.leftBar.Refresh()
				}
			} else {
				btn.SetText("Add")
				if mainWindow.firstTable || mainWindow.currentAddedTable == "" {
					btn.Show()
				} else {
					btn.Hide()
				}
				btn.OnTapped = func() {
					columns, err := mainWindow.dbService.GetColumns(dbName, tableName)
					if err != nil {
						dialog.ShowError(err, mainWindow.window)
						return
					}
					mainWindow.canvas.AddTable(tableName, columns)
					mainWindow.window.Canvas().Refresh(mainWindow.canvas.container)

					mainWindow.currentAddedTable = tableName
					mainWindow.firstTable = false
					mainWindow.leftBar.Refresh()
				}
			}
		},
	)

	// 删除原来的 OnSelected 事件处理
	mainWindow.leftBar.OnSelected = nil

	// 创建数据库连接按钮
	connectBtn := widget.NewButton("Click here to Connect to Database", func() {
		dialog := NewDBConfigDialog(window)
		dialog.SetOnSubmit(func(config *model.DatabaseConfig) {
			mainWindow.dbConfig = config
			mainWindow.connectToDatabase()
		})
		dialog.Show()
	})

	// 创建生成SQL按钮
	generateBtn := widget.NewButton("Generate SQL", func() {
		mainWindow.generateSQL()
	})

	// 创建左侧面板
	leftContainer := container.NewVBox(
		connectBtn,
		widget.NewLabel("Databases & Tables"),
	)

	// 创建一个滚动容器来包装树形结构
	treeScroll := container.NewVScroll(mainWindow.leftBar)
	treeScroll.SetMinSize(fyne.NewSize(200, 600)) // 设置最小高度

	leftContainer.Add(treeScroll)

	// 创建中间画布
	centerContainer := container.NewVBox(
		widget.NewLabel("Design Area"),
		container.NewPadded(mainWindow.canvas.container), // 添加内边距
	)

	// 设置中间容器的最小大小
	centerContainer.Resize(fyne.NewSize(800, 600))

	// 创建右侧SQL预览
	mainWindow.rightBar.MultiLine = true             // 启用多行模式
	mainWindow.rightBar.Wrapping = fyne.TextWrapWord // 启用自动换行

	// 创建一个滚动容器来包装SQL预览
	sqlScroll := container.NewVScroll(mainWindow.rightBar)
	sqlScroll.SetMinSize(fyne.NewSize(200, 600)) // 设置最小高度

	rightContainer := container.NewVBox(
		widget.NewLabel("SQL Preview"),
		generateBtn,
		sqlScroll, // 使用滚动容器替代直接的文本框
	)

	// 创建主布局，调整比例
	split := container.NewHSplit(
		leftContainer, // 移除额外的 HBox 容器
		container.NewHSplit(
			centerContainer,
			rightContainer,
		),
	)
	split.SetOffset(0.2) // 左侧面板占20%

	// 设置右侧分割比例
	rightSplit := split.Trailing.(*container.Split)
	rightSplit.SetOffset(0.7) // 设计区域占70%，SQL预览占30%

	window.SetContent(split)
	window.Resize(fyne.NewSize(1200, 800)) // 设置更大的默认窗口大小

	return mainWindow
}

// 获取所有数据库名称
func (m *MainWindow) getDatabases() []widget.TreeNodeID {
	var databases []widget.TreeNodeID
	for db := range m.dbTables {
		databases = append(databases, widget.TreeNodeID(db))
	}
	return databases
}

func (m *MainWindow) connectToDatabase() {
	dbService, err := service.NewDatabaseService(m.dbConfig)
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}

	m.dbService = dbService
	// Reset state when connecting to new database
	m.firstTable = true
	m.currentAddedTable = ""
	
	// Create new canvas with proper initialization
	m.canvas = NewCanvas(dbService, m.dbConfig, m)
	m.canvas.container.Resize(fyne.NewSize(800, 600))

	// 刷新中间容器
	if centerContainer, ok := m.window.Content().(*container.Split).Trailing.(*container.Split).Leading.(*fyne.Container); ok {
		centerContainer.Objects[1] = container.NewPadded(m.canvas.container)
		centerContainer.Refresh()
	}

	// 获取数据库列表
	databases, err := m.dbService.GetDatabases()
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}

	// 获取每个数据库的表
	for _, dbName := range databases {
		tables, err := m.dbService.GetTables(dbName)
		if err != nil {
			dialog.ShowError(err, m.window)
			continue
		}
		m.dbTables[dbName] = tables
		m.dbConfig.CurrentDB = dbName
	}

	// 刷新界面
	m.leftBar.Refresh()
}

func (m *MainWindow) generateSQL() {
	// 创建SQL生成器
	generator := service.NewSQLGenerator()

	// 设置主表
	mainTable := m.canvas.GetMainTable()
	if mainTable == "" {
		dialog.ShowError(fmt.Errorf("Please add a table first"), m.window)
		return
	}
	generator.SetMainTable(mainTable)

	// 添加选中的列
	selectedColumns := m.canvas.GetAllSelectedColumns()
	if len(selectedColumns) == 0 {
		dialog.ShowError(fmt.Errorf("Please select the columns to query"), m.window)
		return
	}
	for table, columns := range selectedColumns {
		generator.AddSelectedColumns(table, columns)
	}

	// 添加连接信息
	joins := m.canvas.GetAllJoins()
	for _, join := range joins {
		generator.AddJoin(
			join.SourceTable,
			join.TargetTable,
			join.SourceColumn,
			join.TargetColumn,
		)
	}

	// 生成SQL
	sql, err := generator.GenerateSQL()
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}

	// 显示生成的SQL
	m.rightBar.SetText(sql)
}
