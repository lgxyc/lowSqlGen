package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/lowSqlGen/internal/service"
)

type Canvas struct {
	container *fyne.Container
	tables    map[string]*TableNode
	connections  []*TableConnection
	connecting   *TableNode    // 当前正在建立连接的表
	connectingColumn string    // 当前选中的连接列
}

type TableNode struct {
	container *fyne.Container
	rect      *canvas.Rectangle
	name      *canvas.Text
	columns   []*ColumnItem
	position  fyne.Position
	selected  bool
}

type ColumnItem struct {
	container *fyne.Container
	name      *widget.Label
	checkbox  *widget.Check
	joinBtn   *widget.Button
}

type TableConnection struct {
	sourceTable      *TableNode
	targetTable      *TableNode
	sourceLine       *canvas.Line
	targetLine       *canvas.Line
	sourceColumn     string
	targetColumn     string
	connectionLabel  *canvas.Text
}

const (
	tableWidth     = 200
	tableMinHeight = 100
	columnHeight   = 30
	padding        = 10
	lineColor     = 0x666666ff // 连接线的颜色
)

func NewCanvas() *Canvas {
	return &Canvas{
		container: container.NewWithoutLayout(),
		tables:    make(map[string]*TableNode),
	}
}

func (c *Canvas) AddTable(name string, columns []string) *TableNode {
	// 创建表节点
	node := &TableNode{
		rect: canvas.NewRectangle(fyne.NewColor(0.9, 0.9, 0.9, 1)), // 浅灰色背景
		name: canvas.NewText(name, fyne.NewColor(0, 0, 0, 1)),      // 黑色文字
	}
	node.name.TextStyle.Bold = true
	
	// 创建列项
	columnsContainer := container.NewVBox()
	for _, col := range columns {
		columnItem := createColumnItem(col, c, name)
		node.columns = append(node.columns, columnItem)
		columnsContainer.Add(columnItem.container)
	}
	
	// 计算表格高度
	tableHeight := float32(len(columns))*columnHeight + 40 // 40是表头的高度
	if tableHeight < tableMinHeight {
		tableHeight = tableMinHeight
	}
	
	// 设置矩形大小
	node.rect.Resize(fyne.NewSize(tableWidth, tableHeight))
	
	// 创建表格容器
	node.container = container.NewWithoutLayout()
	node.container.Add(node.rect)
	node.container.Add(container.NewPadded(
		container.NewVBox(
			node.name,
			columnsContainer,
		),
	))
	
	// 添加拖拽功能
	node.container.OnMouseDown = func(e *fyne.PointEvent) {
		node.selected = true
		startPos := e.Position
		startNodePos := node.container.Position()
		
		// 添加鼠标移动处理
		node.container.OnMouseMove = func(e *fyne.PointEvent) {
			if !node.selected {
				return
			}
			
			// 计算新位置
			deltaX := e.Position.X - startPos.X
			deltaY := e.Position.Y - startPos.Y
			newPos := fyne.NewPos(
				startNodePos.X + deltaX,
				startNodePos.Y + deltaY,
			)
			
			// 移动表格
			node.container.Move(newPos)
			
			// 更新连接线
			c.updateConnectionsForTable(node)
		}
	}
	
	// 添加鼠标释放处理
	node.container.OnMouseUp = func(e *fyne.PointEvent) {
		node.selected = false
		node.container.OnMouseMove = nil
	}
	
	c.tables[name] = node
	c.container.Add(node.container)
	
	// 初始位置
	c.updateTablePosition(node)
	
	return node
}

func createColumnItem(name string, canvas *Canvas, tableName string) *ColumnItem {
	label := widget.NewLabel(name)
	checkbox := widget.NewCheck("", nil)
	
	// 添加连接按钮
	joinBtn := widget.NewButton("Join", func() {
		if canvas.connecting == nil {
			canvas.StartConnection(tableName, name)
		} else {
			canvas.CompleteConnection(tableName, name)
		}
	})
	
	container := container.NewHBox(
		checkbox,
		label,
		joinBtn,
	)
	
	return &ColumnItem{
		container: container,
		name:      label,
		checkbox:  checkbox,
		joinBtn:   joinBtn,
	}
}

func (c *Canvas) updateTablePosition(node *TableNode) {
	// 计算新表的位置
	x := float32(len(c.tables)-1) * (tableWidth + padding)
	y := float32(padding)
	
	node.container.Move(fyne.NewPos(x, y))
}

func (c *Canvas) Clear() {
	// 清除表
	for _, node := range c.tables {
		c.container.Remove(node.container)
	}
	c.tables = make(map[string]*TableNode)

	// 清除连接
	for _, conn := range c.connections {
		c.container.Remove(conn.sourceLine)
		c.container.Remove(conn.targetLine)
		c.container.Remove(conn.connectionLabel)
	}
	c.connections = nil
	c.connecting = nil
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
	}
}

func (c *Canvas) CompleteConnection(targetTableName, targetColumnName string) {
	if c.connecting == nil {
		return
	}

	targetNode, ok := c.tables[targetTableName]
	if !ok || targetNode == c.connecting {
		c.connecting = nil
		return
	}

	// 创建连接线
	connection := &TableConnection{
		sourceTable:  c.connecting,
		targetTable:  targetNode,
		sourceColumn: c.connectingColumn,
		targetColumn: targetColumnName,
	}

	// 创建连接线的视觉元素
	connection.sourceLine = canvas.NewLine(fyne.NewColor(
		float32((lineColor>>24)&0xFF)/255,
		float32((lineColor>>16)&0xFF)/255,
		float32((lineColor>>8)&0xFF)/255,
		float32(lineColor&0xFF)/255,
	))
	connection.targetLine = canvas.NewLine(connection.sourceLine.StrokeColor)

	// 创建连接说明文本
	connection.connectionLabel = canvas.NewText(
		fmt.Sprintf("%s.%s = %s.%s", 
			c.connecting.name.Text, c.connectingColumn,
			targetNode.name.Text, targetColumnName,
		),
		connection.sourceLine.StrokeColor,
	)

	// 添加到画布
	c.container.Add(connection.sourceLine)
	c.container.Add(connection.targetLine)
	c.container.Add(connection.connectionLabel)
	c.connections = append(c.connections, connection)

	// 更新连接线位置
	c.updateConnectionPosition(connection)

	c.connecting = nil
}

func (c *Canvas) updateConnectionPosition(conn *TableConnection) {
	// 获取源表和目标表的位置
	sourcePos := conn.sourceTable.container.Position()
	targetPos := conn.targetTable.container.Position()

	// 计算连接线的起点和终点
	startX := sourcePos.X + tableWidth
	startY := sourcePos.Y + tableMinHeight/2
	endX := targetPos.X
	endY := targetPos.Y + tableMinHeight/2

	// 设置连接线位置
	midX := (startX + endX) / 2
	conn.sourceLine.Position1 = fyne.NewPos(startX, startY)
	conn.sourceLine.Position2 = fyne.NewPos(midX, startY)
	conn.targetLine.Position1 = fyne.NewPos(midX, endY)
	conn.targetLine.Position2 = fyne.NewPos(endX, endY)

	// 设置连接说明文本位置
	conn.connectionLabel.Move(fyne.NewPos(midX-50, (startY+endY)/2-10))

	// 刷新画布
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