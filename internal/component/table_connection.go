package gui

import (
	"fyne.io/fyne/v2/canvas"
)

type TableConnection struct {
	sourceTable     *TableNode
	targetTable     *TableNode
	sourceLine      *canvas.Line
	targetLine      *canvas.Line
	sourceColumn    string
	targetColumn    string
	connectionLabel *canvas.Text
}

// 使用建造者模式创建连接
type TableConnectionBuilder struct {
	connection *TableConnection
}

func NewTableConnectionBuilder() *TableConnectionBuilder {
	return &TableConnectionBuilder{
		connection: &TableConnection{},
	}
}

func (b *TableConnectionBuilder) SetSource(table *TableNode, column string) *TableConnectionBuilder {
	b.connection.sourceTable = table
	b.connection.sourceColumn = column
	return b
}

func (b *TableConnectionBuilder) SetTarget(table *TableNode, column string) *TableConnectionBuilder {
	b.connection.targetTable = table
	b.connection.targetColumn = column
	return b
}

func (b *TableConnectionBuilder) Build() *TableConnection {
	// 创建连接线和标签
	return b.connection
}
