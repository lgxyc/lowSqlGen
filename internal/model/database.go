package model

// DatabaseConfig 数据库连接配置
type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

// Table 表结构
type Table struct {
	Name    string
	Columns []Column
}

// Column 列结构
type Column struct {
	Name     string
	Type     string
	Selected bool
} 