package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type TableNode struct {
	container   *DraggableContainer
	rect        *canvas.Rectangle
	name        *canvas.Text
	columns     []*ColumnItem
	position    fyne.Position
	selected    bool
	columnsBtn  *widget.Button
	joinBtn     *widget.Button
	showColumns bool
}

type ColumnItem struct {
	container *fyne.Container
	name      *widget.Label
	checkbox  *widget.Check
	dataType  string
	comment   string
}

// 创建新的表节点的工厂方法
func NewTableNode(tableName string, columns []string) *TableNode {
	node := &TableNode{
		container: NewDraggableContainer(),
		rect: &canvas.Rectangle{
			FillColor:   color.NRGBA{R: 240, G: 240, B: 240, A: 255},
			StrokeColor: color.Black,
			StrokeWidth: 3,
		},
		name:        canvas.NewText(tableName, color.Black),
		showColumns: true,
	}

	// ... 初始化其他属性
	return node
}
