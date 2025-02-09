package service

import (
	"fmt"
	"strings"
)

type JoinInfo struct {
	SourceTable  string
	TargetTable  string
	SourceColumn string
	TargetColumn string
}

type SQLGenerator struct {
	selectedColumns map[string][]string // 表名 -> 选中的列
	joins           []JoinInfo
	mainTable       string            // 主表（第一个表）
	tableAliases    map[string]string // 表名 -> 别名
}

func NewSQLGenerator() *SQLGenerator {
	return &SQLGenerator{
		selectedColumns: make(map[string][]string),
		tableAliases:    make(map[string]string),
	}
}

func (g *SQLGenerator) SetMainTable(tableName string) {
	g.mainTable = tableName
	g.tableAliases[tableName] = fmt.Sprintf("t1")
}

func (g *SQLGenerator) AddSelectedColumns(tableName string, columns []string) {
	g.selectedColumns[tableName] = columns
	if _, exists := g.tableAliases[tableName]; !exists {
		g.tableAliases[tableName] = fmt.Sprintf("t%d", len(g.tableAliases)+1)
	}
}

func (g *SQLGenerator) AddJoin(sourceTable, targetTable, sourceColumn, targetColumn string) {
	g.joins = append(g.joins, JoinInfo{
		SourceTable:  sourceTable,
		TargetTable:  targetTable,
		SourceColumn: sourceColumn,
		TargetColumn: targetColumn,
	})

	// 确保两个表都有别名
	if _, exists := g.tableAliases[targetTable]; !exists {
		g.tableAliases[targetTable] = fmt.Sprintf("t%d", len(g.tableAliases)+1)
	}
}

func (g *SQLGenerator) GenerateSQL() (string, error) {
	if g.mainTable == "" {
		return "", fmt.Errorf("未设置主表")
	}

	// 构建SELECT子句
	var selectClauses []string
	for tableName, columns := range g.selectedColumns {
		alias := g.tableAliases[tableName]
		for _, col := range columns {
			selectClauses = append(selectClauses,
				fmt.Sprintf("%s.%s", alias, col))
		}
	}

	if len(selectClauses) == 0 {
		return "", fmt.Errorf("未选择任何列")
	}

	// 构建JOIN子句
	var joinClauses []string
	for _, join := range g.joins {
		sourceAlias := g.tableAliases[join.SourceTable]
		targetAlias := g.tableAliases[join.TargetTable]

		joinClauses = append(joinClauses, fmt.Sprintf(
			"LEFT JOIN %s %s ON %s.%s = %s.%s",
			join.TargetTable, targetAlias,
			sourceAlias, join.SourceColumn,
			targetAlias, join.TargetColumn,
		))
	}

	// 组装SQL语句
	sql := fmt.Sprintf("SELECT %s FROM %s %s",
		strings.Join(selectClauses, ", "),
		g.mainTable,
		g.tableAliases[g.mainTable],
	)

	if len(joinClauses) > 0 {
		sql += " " + strings.Join(joinClauses, " ")
	}

	return sql + ";", nil
}
