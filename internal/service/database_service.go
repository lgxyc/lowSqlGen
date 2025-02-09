package service

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lowSqlGen/internal/model"
)

type DatabaseService interface {
	GetDatabases() ([]string, error)
	GetTables(dbName string) ([]string, error)
	GetColumns(dbName, tableName string) ([]string, error)
	Close() error
	GetColumnType(tableName, columnName string) string
	GetColumnComment(tableName, columnName string) string
	GetTableComment(dbName, tableName string) string
}

type databaseService struct {
	db     *sql.DB
	config *model.DatabaseConfig
}

func NewDatabaseService(config *model.DatabaseConfig) (DatabaseService, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("测试数据库连接失败: %v", err)
	}

	return &databaseService{
		db:     db,
		config: config,
	}, nil
}

func (s *databaseService) GetDatabases() ([]string, error) {
	rows, err := s.db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	// 需要过滤的系统数据库
	systemDBs := map[string]bool{
		"information_schema": true,
		"mysql":              true,
		"performance_schema": true,
		"sys":                true,
	}

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		// 过滤系统数据库
		if !systemDBs[dbName] {
			databases = append(databases, dbName)
		}
	}

	return databases, nil
}

func (s *databaseService) GetTables(dbName string) ([]string, error) {
	// 切换到指定数据库，使用反引号包裹数据库名
	if _, err := s.db.Exec(fmt.Sprintf("USE `%s`", dbName)); err != nil {
		return nil, err
	}

	rows, err := s.db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

func (s *databaseService) GetColumns(dbName, tableName string) ([]string, error) {
	// 切换到指定数据库，使用反引号包裹数据库名
	if _, err := s.db.Exec(fmt.Sprintf("USE `%s`", dbName)); err != nil {
		return nil, err
	}

	// 获取列信息，使用反引号包裹表名
	rows, err := s.db.Query(fmt.Sprintf("SHOW COLUMNS FROM `%s`", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var field string
		var type_, null, key string
		var default_ sql.NullString
		var extra string
		if err := rows.Scan(&field, &type_, &null, &key, &default_, &extra); err != nil {
			return nil, err
		}
		columns = append(columns, field)
	}

	return columns, nil
}

func (s *databaseService) Close() error {
	return s.db.Close()
}

func (s *databaseService) GetColumnType(tableName, columnName string) string {
	// Implementation of GetColumnType method
	return "" // Placeholder return, actual implementation needed
}

func (s *databaseService) GetColumnComment(tableName, columnName string) string {
	// Implementation of GetColumnComment method
	return "" // Placeholder return, actual implementation needed
}

// GetTableComment 获取表注释
func (s *databaseService) GetTableComment(dbName, tableName string) string {
	query := `
		SELECT table_comment 
		FROM information_schema.tables 
		WHERE table_schema = ? AND table_name = ?
	`
	var comment string
	err := s.db.QueryRow(query, dbName, tableName).Scan(&comment)
	if err != nil {
		return ""
	}
	return comment
}
