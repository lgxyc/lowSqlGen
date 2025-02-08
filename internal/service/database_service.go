package service

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/yourusername/LowSqlGen/internal/model"
)

type DatabaseService struct {
	db     *sql.DB
	config *model.DatabaseConfig
}

func NewDatabaseService(config *model.DatabaseConfig) (*DatabaseService, error) {
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
	
	return &DatabaseService{
		db:     db,
		config: config,
	}, nil
}

func (s *DatabaseService) GetDatabases() ([]string, error) {
	rows, err := s.db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		databases = append(databases, dbName)
	}
	
	return databases, nil
}

func (s *DatabaseService) GetTables(dbName string) ([]string, error) {
	// 切换到指定数据库
	if _, err := s.db.Exec("USE " + dbName); err != nil {
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

func (s *DatabaseService) GetColumns(dbName, tableName string) ([]string, error) {
	// 切换到指定数据库
	if _, err := s.db.Exec("USE " + dbName); err != nil {
		return nil, err
	}
	
	// 获取列信息
	rows, err := s.db.Query(fmt.Sprintf("SHOW COLUMNS FROM %s", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var columns []string
	for rows.Next() {
		var field, type_, null, key, default_, extra string
		if err := rows.Scan(&field, &type_, &null, &key, &default_, &extra); err != nil {
			return nil, err
		}
		columns = append(columns, field)
	}
	
	return columns, nil
}

func (s *DatabaseService) Close() error {
	return s.db.Close()
} 