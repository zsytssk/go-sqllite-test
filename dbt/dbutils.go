package dbt

import (
	"log"
	"quick-cmd/utils"
	"regexp"
	"strings"

	"database/sql"
	"fmt"
)

func Init(dbPath string) (db *sql.DB, err error) {
	filePath, err := utils.GetCurDirFilePath(dbPath)
	if err != nil {
		return
	}
	db, err = sql.Open("sqlite3", filePath)
	if err != nil {
		return
	}
	return
}

type Item struct {
	ID       int
	Name     string
	Priority int
}

// UpdatePriority 根据名称更新优先级
func UpdateItemPriority(db *sql.DB, tableName string, id int, priority int) error {
	stmt := fmt.Sprintf(`UPDATE %s
         SET priority = ?
         WHERE id = ?
         AND priority <> ?`, tableName)

	result, err := db.Exec(
		stmt,
		priority, id, priority,
	)
	if err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}

	// 简化版影响检查
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("记录不存在或值未变化")
	}

	return nil
}

func UpdateItem(db *sql.DB, tableName string, id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("至少需要提供一个更新字段")
	}

	// 构建SQL语句
	var setClause strings.Builder
	params := make([]interface{}, 0, len(updates)+1)

	for field, value := range updates {
		if !isValidFieldName(field) {
			return fmt.Errorf("非法字段名: %s", field)
		}
		if setClause.Len() > 0 {
			setClause.WriteString(", ")
		}
		setClause.WriteString(fmt.Sprintf("%s = ?", field))
		params = append(params, value)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", tableName, setClause.String())
	params = append(params, id)

	_, err := db.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}

	return nil
}

// 校验字段名合法性（防止 SQL 注入）
func isValidFieldName(field string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", field)
	return matched
}

// UpdatePriority 根据名称更新优先级
func InsertItemPriority(db *sql.DB, tableName string, name string, priority int) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	stmt := fmt.Sprintf(`INSERT OR REPLACE INTO %s (name, priority) VALUES (?, ?)`, tableName)
	_, err := db.Exec(stmt, name, priority)
	if err != nil {
		return fmt.Errorf("failed to insert/update: %w", err)
	}
	return nil
}

func GetItems(db *sql.DB, tableName string) (items []Item, err error) {
	queryStr := fmt.Sprintf("SELECT id, name, priority FROM %s ORDER BY priority desc", tableName)
	rows, err := db.Query(queryStr)
	if err != nil {
		return
	}
	defer rows.Close()

	// 读取数据到结构体切片
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.Name, &it.Priority); err != nil {
			log.Fatal(err)
		}
		items = append(items, it)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("迭代记录时出错: %w", err)
	}

	return
}

func checkTableExist(db *sql.DB, tableName string) (exist bool) {
	// 执行查询，检查 sqlite_master 表中是否存在指定名称的表
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return false
	}

	// 如果查询结果大于 0，则表示表存在
	return count > 0
}

func createTable(db *sql.DB, sqlStmt string) (sus bool, err error) {
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return false, err
	}
	return true, nil
}

func genTableStmt(tableName string) string {
	return fmt.Sprintf(`
	CREATE TABLE %s (
    	id INTEGER PRIMARY KEY,
    	name TEXT NOT NULL,
    	priority INTEGER DEFAULT 0
	);`, tableName)
}
