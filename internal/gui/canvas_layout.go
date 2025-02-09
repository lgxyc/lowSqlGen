package gui

import (
	"fyne.io/fyne/v2"
)

type CanvasLayout struct {
	tableDepths map[string]int
	tableRows   map[string]int
}

func NewCanvasLayout() *CanvasLayout {
	return &CanvasLayout{
		tableDepths: make(map[string]int),
		tableRows:   make(map[string]int),
	}
}

func (l *CanvasLayout) updateTablePosition(node *TableNode, canvas *Canvas) {
	// Temporary implementation
	node.container.Move(fyne.NewPos(100, 100))
}

func (l *CanvasLayout) findNonCollidingPosition(node *TableNode, baseX, baseY float32, canvas *Canvas) (float32, float32) {
	// Temporary implementation
	return baseX, baseY
}
