package gui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/lowSqlGen/internal/component"
	"github.com/lowSqlGen/internal/model"
	"github.com/lowSqlGen/internal/service"
)

type Canvas struct {
	container        *DraggableContainer
	tables           map[string]*TableNode
	connections      []*TableConnection
	connecting       *TableNode              // 当前正在建立连接的表
	connectingColumn string                  // 当前选中的连接列
	dbService        service.DatabaseService // Change from *service.DatabaseService to service.DatabaseService
	dbConfig         *model.DatabaseConfig   // Added dbConfig to the Canvas struct
	content          *fyne.Container         // 添加一个主内容容器
	layout           *CanvasLayout           // 使用组合而不是继承
	tempConnection   *TableConnection
	mainWindow       *MainWindow // Add reference to main window for state access
}

// 创建一个可拖动的容器
type DraggableContainer struct {
	widget.BaseWidget
	content      *fyne.Container // 改名为 content 以避免混淆
	isDragging   bool
	dragStartPos fyne.Position
	dragOffset   fyne.Position
}

// 创建一个可拖动的容器
func NewDraggableContainer() *DraggableContainer {
	d := &DraggableContainer{
		content: container.NewWithoutLayout(),
	}
	d.ExtendBaseWidget(d)

	// 确保容器被正确初始化
	d.Resize(fyne.NewSize(800, 600))
	d.content.Resize(fyne.NewSize(800, 600))

	return d
}

// 鼠标按下事件
func (d *DraggableContainer) MouseDown(e *fyne.PointEvent) {
	d.isDragging = true
	d.dragStartPos = e.Position
	d.dragOffset = d.content.Position()
}

// 鼠标释放事件
func (d *DraggableContainer) MouseUp(e *fyne.PointEvent) {
	d.isDragging = false
}

// 鼠标移动事件
func (d *DraggableContainer) MouseMoved(e *fyne.PointEvent) {
	if !d.isDragging {
		return
	}

	deltaX := e.Position.X - d.dragStartPos.X
	deltaY := e.Position.Y - d.dragStartPos.Y

	newPos := fyne.NewPos(
		d.dragOffset.X+deltaX,
		d.dragOffset.Y+deltaY,
	)

	d.content.Move(newPos)
	d.Refresh()
}

// 创建渲染器
func (d *DraggableContainer) CreateRenderer() fyne.WidgetRenderer {
	if d.content == nil {
		d.content = container.NewWithoutLayout()
	}
	return widget.NewSimpleRenderer(d.content)
}

// 获取当前窗口
func (d *DraggableContainer) Window() fyne.Window {
	return fyne.CurrentApp().Driver().AllWindows()[0]
}

const (
	tableWidth     = 200
	tableMinHeight = 100
	columnHeight   = 40
	padding        = 20
	innerPadding   = 15
	headerHeight   = 40
	lineColor      = 0x666666ff
	horizontalGap  = 400 // 增加表之间的水平间距
	verticalGap    = 100 // 增加垂直间距以避免重叠
	labelWidth     = 200 // 添加标签宽度常量
)

// 添加一个结构来表示矩形区域
type Rect struct {
	x, y, width, height float32
}

// 检查两个矩形是否重叠
func (r1 Rect) intersects(r2 Rect) bool {
	return !(r1.x+r1.width <= r2.x ||
		r2.x+r2.width <= r1.x ||
		r1.y+r1.height <= r2.y ||
		r2.y+r2.height <= r1.y)
}

// 获取表格的边界矩形
func (c *Canvas) getTableBounds(node *TableNode) Rect {
	pos := node.container.Position()
	height := float32(tableMinHeight)
	if node.showColumns {
		height = float32(headerHeight + len(node.columns)*columnHeight + 2*padding)
		if height < tableMinHeight {
			height = tableMinHeight
		}
	}
	return Rect{
		x:      pos.X,
		y:      pos.Y,
		width:  tableWidth,
		height: height,
	}
}

// 检查位置是否有碰撞
func (c *Canvas) hasCollision(node *TableNode, x, y float32) bool {
	proposedRect := Rect{
		x:      x,
		y:      y,
		width:  tableWidth,
		height: float32(tableMinHeight),
	}

	// 检查与所有其他表的碰撞
	for name, otherNode := range c.tables {
		if name == node.name.Text {
			continue
		}
		otherBounds := c.getTableBounds(otherNode)
		if proposedRect.intersects(otherBounds) {
			return true
		}
	}
	return false
}

// 找到一个无碰撞的位置
func (c *Canvas) findNonCollidingPosition(node *TableNode, baseX, baseY float32) (float32, float32) {
	fmt.Println("findNonCollidingPosition", baseX, baseY)
	// 定义搜索范围
	maxAttempts := 10
	spiralStep := float32(50) // 每次尝试移动的距离

	x, y := baseX, baseY

	// 使用螺旋形搜索模式
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// 检查当前位置
		if !c.hasCollision(node, x, y) {
			return x, y
		}

		// 螺旋形搜索下一个位置
		switch attempt % 4 {
		case 0: // 右
			x += spiralStep
		case 1: // 下
			y += spiralStep
		case 2: // 左
			x -= spiralStep
		case 3: // 上
			y -= spiralStep
			spiralStep += 50 // 增加搜索范围
		}
	}

	// 如果找不到无碰撞位置，返回原始位置加上偏移
	return baseX + float32(len(c.tables))*50, baseY + float32(len(c.tables))*50
}

// 更新表的位置
func (c *Canvas) updateTablePosition(node *TableNode) {
	tableName := node.name.Text

	// 如果这是第一个表，放在起始位置
	if len(c.tables) == 1 {
		c.layout.tableDepths[tableName] = 0
		c.layout.tableRows[tableName] = 0
		node.container.Move(fyne.NewPos(100, 100))
		return
	}

	// 查找所有与当前表相关的连接
	var relatedConnections []*TableConnection
	for _, conn := range c.connections {
		if conn.targetTable == node {
			relatedConnections = append(relatedConnections, conn)
		}
	}

	// 如果是作为目标表被连接
	if len(relatedConnections) > 0 {
		sourceTable := relatedConnections[0].sourceTable
		sourcePos := sourceTable.container.Position()

		// 计算同源表的其他目标表数量
		sameSourceTargets := 0
		for _, conn := range c.connections {
			if conn.sourceTable == sourceTable && conn.targetTable != node {
				sameSourceTargets++
			}
		}

		// 基础位置计算
		baseX := sourcePos.X + horizontalGap
		baseY := sourcePos.Y // 默认与源表在同一水平线

		// 如果有其他同源的目标表，错开Y轴位置
		if sameSourceTargets > 0 {
			// 找到同源的最后一个已放置的表
			var lastTableY float32
			lastTableHeight := float32(0)
			found := false

			for _, conn := range c.connections {
				if conn.sourceTable == sourceTable && conn.targetTable != node {
					targetPos := conn.targetTable.container.Position()
					targetBounds := c.getTableBounds(conn.targetTable)
					if targetPos.Y+targetBounds.height > lastTableY {
						lastTableY = targetPos.Y
						lastTableHeight = targetBounds.height
						found = true
					}
				}
			}

			// 如果找到了前一个表，在其下方放置新表
			if found {
				baseY = lastTableY + lastTableHeight + verticalGap
			}
		}

		// 寻找无碰撞位置
		finalX, finalY := c.findNonCollidingPosition(node, baseX, baseY)
		node.container.Move(fyne.NewPos(finalX, finalY))

		// 更新深度信息
		c.layout.tableDepths[tableName] = c.layout.tableDepths[sourceTable.name.Text] + 1
	} else {
		// 独立表的位置计算保持不变
		baseX := float32(100 + len(c.tables)*50)
		baseY := float32(100 + len(c.tables)*50)
		finalX, finalY := c.findNonCollidingPosition(node, baseX, baseY)
		node.container.Move(fyne.NewPos(finalX, finalY))
	}

	// 更新所有连接线
	c.updateAllConnections()
}

// 添加新的辅助方法来更新所有连接
func (c *Canvas) updateAllConnections() {
	for _, conn := range c.connections {
		c.updateConnectionPosition(conn)
	}
}

func (c *Canvas) Clear() {
	// 先清除所有连接
	if c.connections != nil {
		for _, conn := range c.connections {
			if conn != nil && c.content != nil && c.content.Objects != nil {
				// 安全地移除连接线和标签
				if conn.sourceLine != nil {
					c.content.Remove(conn.sourceLine)
				}
				if conn.targetLine != nil {
					c.content.Remove(conn.targetLine)
				}
				if conn.connectionLabel != nil {
					c.content.Remove(conn.connectionLabel)
				}
			}
		}
		c.connections = nil
	}

	// 取消当前正在进行的连接
	c.CancelConnection()
	c.connecting = nil
	c.connectingColumn = ""

	// 清除表
	if c.tables != nil {
		for _, node := range c.tables {
			if node != nil && c.content != nil {
				c.content.Remove(node.container)
			}
		}
		c.tables = make(map[string]*TableNode)
	}

	// 刷新容器
	if c.content != nil {
		c.content.Refresh()
	}
	if c.container != nil {
		c.container.Refresh()
	}

	// 强制刷新窗口画布
	if app := fyne.CurrentApp(); app != nil {
		if window := app.Driver().AllWindows()[0]; window != nil {
			window.Canvas().Refresh(c.container)
		}
	}
}

// GetSelectedColumns 获取选中的列
func (c *Canvas) GetSelectedColumns(tableName string) []string {
	var selected []string
	if node, ok := c.tables[tableName]; ok {
		for _, col := range node.columns {
			if col.checkbox.Checked {
				selected = append(selected, col.name.Text)
			}
		}
	}
	return selected
}

func (c *Canvas) StartConnection(tableName, columnName string) {
	if node, ok := c.tables[tableName]; ok {
		c.connecting = node
		c.connectingColumn = columnName

		c.tempConnection = &TableConnection{
			sourceTable:     node,
			sourceColumn:    columnName,
			sourceLine:      canvas.NewLine(color.NRGBA{R: 0, G: 0, B: 0, A: 255}),
			targetLine:      canvas.NewLine(color.NRGBA{R: 0, G: 0, B: 0, A: 255}),
			connectionLabel: canvas.NewText("", color.Black),
		}
		c.tempConnection.sourceLine.StrokeWidth = 3
		c.tempConnection.targetLine.StrokeWidth = 3

		c.content.Add(c.tempConnection.sourceLine)
		c.content.Add(c.tempConnection.targetLine)
		c.content.Add(c.tempConnection.connectionLabel)
	}
}

func (c *Canvas) CancelConnection() {
	if c.tempConnection != nil {
		// 从画布移除临时连接线和文本
		c.content.Remove(c.tempConnection.sourceLine)
		c.content.Remove(c.tempConnection.targetLine)
		c.content.Remove(c.tempConnection.connectionLabel)
		c.tempConnection = nil
	}
	c.connecting = nil
	c.connectingColumn = ""
}

func (c *Canvas) CompleteConnection(targetTableName, targetColumnName string) {
	// 检查画布是否正确初始化
	if c == nil || c.content == nil || c.content.Objects == nil {
		return
	}

	// 检查连接参数
	if c.connecting == nil || targetTableName == "" || targetColumnName == "" {
		c.CancelConnection()
		return
	}

	// Add this block back
	targetNode, ok := c.tables[targetTableName]
	if !ok || targetNode == nil || targetNode == c.connecting {
		c.CancelConnection()
		return
	}

	// 检查连接是否已存在
	for _, conn := range c.connections {
		if conn.sourceTable == c.connecting && conn.targetTable == targetNode &&
			conn.sourceColumn == c.connectingColumn && conn.targetColumn == targetColumnName {
			c.CancelConnection()
			return
		}
	}

	// 初始化连接数组（如果为nil）
	if c.connections == nil {
		c.connections = make([]*TableConnection, 0)
	}

	// 创建新的永久连接
	connection := &TableConnection{
		sourceTable:  c.connecting,
		targetTable:  targetNode,
		sourceColumn: c.connectingColumn,
		targetColumn: targetColumnName,
	}

	// 创建连接线的视觉元素
	lineStyle := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	connection.sourceLine = canvas.NewLine(lineStyle)
	connection.targetLine = canvas.NewLine(lineStyle)

	// 检查连接线是否创建成功
	if connection.sourceLine == nil || connection.targetLine == nil {
		return
	}

	connection.sourceLine.StrokeWidth = 3
	connection.targetLine.StrokeWidth = 3

	// 创建连接说明文本
	labelText := fmt.Sprintf("%s.%s = %s.%s",
		c.connecting.name.Text, c.connectingColumn,
		targetNode.name.Text, targetColumnName)

	connection.connectionLabel = canvas.NewText(labelText, lineStyle)
	if connection.connectionLabel == nil {
		return
	}

	// 添加到画布前进行检查
	if c.content != nil && c.content.Objects != nil {
		c.content.Add(connection.sourceLine)
		c.content.Add(connection.targetLine)
		c.content.Add(connection.connectionLabel)
		c.connections = append(c.connections, connection)

		// 更新连接线位置
		c.updateConnectionPosition(connection)
	}
}

func (c *Canvas) updateConnectionPosition(conn *TableConnection) {
	sourcePos := conn.sourceTable.container.Position()
	targetPos := conn.targetTable.container.Position()

	// 计算连接线的起点和终点（连接到表头位置）
	startX := sourcePos.X + tableWidth     // 源表右边
	startY := sourcePos.Y + headerHeight/2 // 表头中间位置
	endX := targetPos.X                    // 目标表左边
	endY := targetPos.Y + headerHeight/2   // 表头中间位置

	// 确保连接线的宽度至少和标签一样宽
	minWidth := float32(labelWidth)
	actualWidth := endX - startX
	if actualWidth < minWidth {
		// 调整目标表的位置以确保最小宽度
		endX = startX + minWidth
		targetPos.X = endX
		conn.targetTable.container.Move(targetPos)
	}

	// 计算中间点的X坐标
	midX := startX + (endX-startX)/2

	// 设置连接线位置
	conn.sourceLine.Position1 = fyne.NewPos(startX, startY)
	conn.sourceLine.Position2 = fyne.NewPos(midX, startY)

	conn.targetLine.Position1 = fyne.NewPos(midX, endY)
	conn.targetLine.Position2 = fyne.NewPos(endX, endY)

	// 设置连接说明文本位置，确保文本居中显示
	labelX := midX - labelWidth/2
	labelY := startY
	if startY != endY {
		labelY = startY + (endY-startY)/2
	}
	conn.connectionLabel.Move(fyne.NewPos(labelX, labelY-10))
	conn.connectionLabel.Resize(fyne.NewSize(labelWidth, 20))

	// 刷新连接线和标签
	conn.sourceLine.Refresh()
	conn.targetLine.Refresh()
	conn.connectionLabel.Refresh()
}

func (c *Canvas) updateConnectionsForTable(node *TableNode) {
	for _, conn := range c.connections {
		if conn.sourceTable == node || conn.targetTable == node {
			c.updateConnectionPosition(conn)
		}
	}
}

// GetMainTable 获取主表（第一个添加的表）
func (c *Canvas) GetMainTable() string {
	if len(c.tables) == 0 {
		return ""
	}
	// 返回第一个添加的表
	for name := range c.tables {
		return name
	}
	return ""
}

// GetAllSelectedColumns 获取所有表的选中列
func (c *Canvas) GetAllSelectedColumns() map[string][]string {
	result := make(map[string][]string)
	for tableName, node := range c.tables {
		var selectedColumns []string
		for _, col := range node.columns {
			if col.checkbox.Checked {
				selectedColumns = append(selectedColumns, col.name.Text)
			}
		}
		if len(selectedColumns) > 0 {
			result[tableName] = selectedColumns
		}
	}
	return result
}

// GetAllJoins 获取所有表连接信息
func (c *Canvas) GetAllJoins() []service.JoinInfo {
	var joins []service.JoinInfo
	for _, conn := range c.connections {
		joins = append(joins, service.JoinInfo{
			SourceTable:  conn.sourceTable.name.Text,
			TargetTable:  conn.targetTable.name.Text,
			SourceColumn: conn.sourceColumn,
			TargetColumn: conn.targetColumn,
		})
	}
	return joins
}

// 修改 updateTableDisplay 中的刷新调用
func (c *Canvas) updateTableDisplay(node *TableNode) {
	if node == nil || node.container == nil || node.container.content == nil {
		return
	}

	// 获取主容器（VBox）
	mainContainer := node.container.content.Objects[0].(*fyne.Container)

	// 获取包含矩形和列的容器（Stack）
	stackContainer := mainContainer.Objects[1].(*fyne.Container)

	// 获取列容器（Padded container）
	columnsPadded := stackContainer.Objects[1].(*fyne.Container)

	// 显示或隐藏列
	if node.showColumns {
		columnsPadded.Show()
		// 调整矩形高度以适应列内容
		totalHeight := headerHeight + float32(len(node.columns))*columnHeight + 2*padding
		if totalHeight < tableMinHeight {
			totalHeight = tableMinHeight
		}
		node.rect.Resize(fyne.NewSize(tableWidth, totalHeight))
	} else {
		columnsPadded.Hide()
		// 恢复最小高度
		node.rect.Resize(fyne.NewSize(tableWidth, tableMinHeight))
	}

	// 刷新显示
	node.rect.Refresh()
	node.container.Refresh()
	c.updateConnectionsForTable(node) // 更新连接线位置
}

// AddTable 添加一个新的表到画布
func (c *Canvas) AddTable(tableName string, columns []string) {
	// Validate required services are available
	if c.dbService == nil {
		dialog.ShowError(fmt.Errorf("Database service not initialized"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	// Check if table already exists
	if _, exists := c.tables[tableName]; exists {
		return
	}

	// Create table node with proper service references
	node := &TableNode{
		container: NewDraggableContainer(),
		rect: &canvas.Rectangle{
			FillColor:   color.NRGBA{R: 240, G: 240, B: 240, A: 255},
			StrokeColor: color.Black,
			StrokeWidth: 3,
		},
		name:        canvas.NewText(tableName, color.Black),
		showColumns: true,
		dbService:   c.dbService, // Pass database service reference
	}

	// 设置表格容器大小
	totalHeight := headerHeight + float32(len(columns))*columnHeight + 2*padding
	if totalHeight < tableMinHeight {
		totalHeight = tableMinHeight
	}

	// 设置矩形大小
	node.rect.Resize(fyne.NewSize(tableWidth, totalHeight))

	// 设置表名文本位置
	node.name.Move(fyne.NewPos(padding, padding))
	node.name.Resize(fyne.NewSize(tableWidth-2*padding, headerHeight))

	// 创建列项
	for _, colName := range columns {
		columnItem := createColumnItem(colName, c, tableName)
		node.columns = append(node.columns, columnItem)
	}

	// 创建 "Fields" 和 "Join" 按钮
	node.columnsBtn = widget.NewButton("Fields", func() {
		node.showColumns = !node.showColumns
		c.updateTableDisplay(node)
	})

	node.joinBtn = widget.NewButton("+", func() {
		// 创建并显示连接对话框
		joinDialog := NewJoinDialog(
			fyne.CurrentApp().Driver().AllWindows()[0],
			tableName,
			c.dbService,
			c.dbConfig.CurrentDB,
		)

		joinDialog.SetOnConfirm(func(targetTable, sourceColumn, targetColumn string) {
			// 获取目标表的列
			columns, err := c.dbService.GetColumns(c.dbConfig.CurrentDB, targetTable)
			if err != nil {
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}

			// 创建连接
			c.StartConnection(tableName, sourceColumn)
			// 添加目标表
			c.AddTable(targetTable, columns)
			c.CompleteConnection(targetTable, targetColumn)
		})

		joinDialog.Show()
	})

	// 创建按钮容器
	buttonsContainer := container.NewHBox(node.columnsBtn, node.joinBtn)

	// 创建表头容器（包含表名和按钮）
	headerContainer := container.NewHBox(
		widget.NewLabel(tableName),
		buttonsContainer,
	)

	// 创建列容器
	columnsContainer := container.NewVBox()
	for _, col := range node.columns {
		columnsContainer.Add(col.container)
	}
	columnsPadded := container.NewPadded(columnsContainer)

	// 创建堆叠容器（矩形和列）
	stackContainer = container.NewStack(
		node.rect,
		container.NewPadded(
			container.NewVBox(
				node.name,
				columnsPadded,
			),
		),
	)

	// 创建主容器
	mainContainer = container.NewVBox(
		headerContainer,
		stackContainer,
	)

	// 设置主容器大小
	mainContainer.Resize(fyne.NewSize(tableWidth, totalHeight))

	// 将主容器添加到节点的容器中
	node.container.content.Add(mainContainer)
	node.container.Resize(fyne.NewSize(tableWidth, totalHeight))

	// 添加到画布中
	c.content.Add(node.container)
	c.tables[tableName] = node

	// 更新表的位置
	c.updateTablePosition(node)

	// 刷新所有容器
	node.container.Refresh()
	c.content.Refresh()
	c.container.Refresh()

	// 使用安全的方式刷新窗口
	if app := fyne.CurrentApp(); app != nil {
		if window := app.Driver().AllWindows()[0]; window != nil {
			window.Canvas().Refresh(c.container)
		}
	}
}

func createColumnItem(name string, canvas *Canvas, tableName string) *ColumnItem {
	// 创建基本标签
	fullInfo := name
	var dataType, comment string

	// 安全地获取列信息
	if canvas != nil && canvas.dbService != nil {
		dataType = canvas.dbService.GetColumnType(tableName, name)
		if dataType != "" {
			fullInfo += " " + dataType
		}

		comment = canvas.dbService.GetColumnComment(tableName, name)
		if comment != "" {
			fullInfo += fmt.Sprintf(" // %s", comment)
		}
	}

	label := widget.NewLabel(fullInfo)
	checkbox := widget.NewCheck("", nil)

	container := container.NewPadded( // 添加内边距
		container.NewHBox(
			checkbox,
			label,
		),
	)

	return &ColumnItem{
		container: container,
		name:      label,
		checkbox:  checkbox,
		dataType:  dataType,
		comment:   comment,
	}
}

func NewCanvas(dbService service.DatabaseService, dbConfig *model.DatabaseConfig, mainWindow *MainWindow) *Canvas {
	if mainWindow == nil {
		panic("MainWindow reference cannot be nil")
	}

	c := &Canvas{
		tables:      make(map[string]*TableNode),
		dbService:   dbService,
		dbConfig:    dbConfig,
		connections: make([]*TableConnection, 0),
		content:     container.NewWithoutLayout(),
		layout:      &CanvasLayout{
			tableDepths: make(map[string]int),
			tableRows:   make(map[string]int),
		},
		mainWindow: mainWindow,
	}

	// Create draggable container with proper initialization
	c.container = NewDraggableContainer()
	if c.container == nil {
		panic("Failed to create draggable container")
	}

	// Create and setup background
	background := &canvas.Rectangle{
		FillColor: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
	}
	background.Resize(fyne.NewSize(800, 600))

	// Initialize containers with proper sizes
	c.container.Resize(fyne.NewSize(800, 600))
	c.content.Resize(fyne.NewSize(800, 600))

	c.content.Add(background)
	c.container.content.Add(c.content)

	return c
}

// 使用观察者模式处理连接状态变化
type ConnectionObserver interface {
	OnConnectionChanged(conn *TableConnection)
}

func (c *Canvas) AddConnectionObserver(observer ConnectionObserver) {
	// ... 添加观察者
}
