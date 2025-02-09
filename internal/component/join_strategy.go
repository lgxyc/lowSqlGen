package gui

import (
	"fmt"
)

type JoinStrategy interface {
	CreateJoin(source, target *TableNode, sourceCol, targetCol string) string
}

type LeftJoinStrategy struct{}

func (s *LeftJoinStrategy) CreateJoin(source, target *TableNode, sourceCol, targetCol string) string {
	return fmt.Sprintf("LEFT JOIN %s ON %s.%s = %s.%s",
		target.name.Text, source.name.Text, sourceCol, target.name.Text, targetCol)
} 