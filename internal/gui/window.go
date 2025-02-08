package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/dialog"
	"github.com/lowSqlGen/internal/model"
	"github.com/lowSqlGen/internal/service"
)

type MainWindow struct {
	window     fyne.Window
	canvas     *Canvas
	leftBar    *widget.List
	rightBar   *widget.TextArea
	dbConfig   *model.DatabaseConfig
	dbService  *service.DatabaseService
}

func InitMainWindow(window fyne.Window) *MainWindow {
	mainWindow := &MainWindow{
		window:   window,
		canvas:   NewCanvas(),
		leftBar:  widget.NewList(nil, nil, nil),
		rightBar: widget.NewTextArea(),
	}

	// 创建数据库连接按钮
	connectBtn := widget.NewButton("连接数据库", func() {
		dialog := NewDBConfigDialog(window)
		dialog.SetOnSubmit(func(config *model.DatabaseConfig) {
			mainWindow.dbConfig = config
			mainWindow.connectToDatabase()
		})
		dialog.Show()
	})

	// 创建生成SQL按钮
	generateBtn := widget.NewButton("生成SQL", func() {
		mainWindow.generateSQL()
	})

	// 创建左侧数据库列表
	leftContainer := container.NewVBox(
		connectBtn,
		widget.NewLabel("数据库列表"),
		mainWindow.leftBar,
	)

	// 创建中间画布
	centerContainer := container.NewVBox(
		widget.NewLabel("设计区域"),
		mainWindow.canvas.container,
	)

	// 创建右侧SQL预览
	rightContainer := container.NewVBox(
		widget.NewLabel("SQL预览"),
		generateBtn,
		mainWindow.rightBar,
	)

	// 创建主布局
	split := container.NewHSplit(
		leftContainer,
		container.NewHSplit(
			centerContainer,
			rightContainer,
		),
	)

	window.SetContent(split)
	
	return mainWindow
}

func (m *MainWindow) connectToDatabase() {
	// 创建数据库服务
	dbService, err := service.NewDatabaseService(m.dbConfig)
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}
	
	m.dbService = dbService
	
	// 获取数据库列表
	databases, err := dbService.GetDatabases()
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}
	
	// 更新左侧数据库列表
	m.leftBar.Length = func() int {
		return len(databases)
	}
	
	m.leftBar.CreateItem = func() fyne.CanvasObject {
		return widget.NewLabel("Template")
	}
	
	m.leftBar.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
		item.(*widget.Label).SetText(databases[id])
	}
	
	m.leftBar.OnSelected = func(id widget.ListItemID) {
		m.loadDatabaseTables(databases[id])
	}
	
	m.leftBar.Refresh()
}

func (m *MainWindow) loadDatabaseTables(dbName string) {
	// 清空现有的表显示
	m.canvas.Clear()
	
	// 获取表列表
	tables, err := m.dbService.GetTables(dbName)
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}
	
	// 为每个表获取列信息并显示
	for _, tableName := range tables {
		columns, err := m.dbService.GetColumns(dbName, tableName)
		if err != nil {
			dialog.ShowError(err, m.window)
			continue
		}
		m.canvas.AddTable(tableName, columns)
	}
}

func (m *MainWindow) generateSQL() {
	// 创建SQL生成器
	generator := service.NewSQLGenerator()

	// 设置主表
	mainTable := m.canvas.GetMainTable()
	if mainTable == "" {
		dialog.ShowError(fmt.Errorf("请先添加表"), m.window)
		return
	}
	generator.SetMainTable(mainTable)

	// 添加选中的列
	selectedColumns := m.canvas.GetAllSelectedColumns()
	if len(selectedColumns) == 0 {
		dialog.ShowError(fmt.Errorf("请选择要查询的列"), m.window)
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