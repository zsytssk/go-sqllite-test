package dbt

import (
	"database/sql"
	"fmt"
	"quick-cmd/utils"
	"sort"
	"strings"
)

func InitHistoryTable(db *sql.DB, lineMap map[string]int) {
	exist := checkTableExist(db, "history")
	if !exist {
		createTable(db, genTableStmt("history"))
	}

	stmtSelect := `SELECT priority FROM history WHERE name = ?`
	stmtInsert := `INSERT INTO history (name, priority) VALUES (?, ?)`
	stmtUpdate := `UPDATE history SET priority = ? WHERE name = ?`

	for k, v := range lineMap {

		if v <= 2 {
			continue // 只处理出现次数大于 2 的
		}

		var currentPriority int
		err := db.QueryRow(stmtSelect, k).Scan(&currentPriority)

		switch {
		case err == sql.ErrNoRows:
			// 不存在，插入新记录
			_, _ = db.Exec(stmtInsert, k, v)

		case err == nil && currentPriority < v:
			// 已存在，且 priority 更小，更新
			_, _ = db.Exec(stmtUpdate, v, k)

		// err == nil 且 currentPriority >= v，不处理
		case err != nil:
			// 其他查询错误
			fmt.Println("查询失败:", err)
		}
	}
}

func GetHistory(db *sql.DB) (items []Item, err error) {
	lineMap, err := utils.ReadFile("~/.bash_history")

	if err != nil {
		return
	}

	for key := range lineMap {
		if strings.HasPrefix(key, "cd") && !strings.Contains(key, "&&") {
			delete(lineMap, key)
			continue
		}
	}

	exist := checkTableExist(db, "history")
	if !exist {
		InitHistoryTable(db, lineMap)
	}

	items, err = GetItems(db, "history")
	if err != nil {
		return
	}
	for key, count := range lineMap {
		index := utils.ArrFindIndex(items, func(item Item, _ int) bool {
			return item.Name == key
		})
		if index != -1 {
			continue
		}
		// fmt.Println("test:>", key, count)
		item := Item{-1, key, count}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Priority > items[j].Priority
	})

	// index := utils.FindItemIndex(items, func(item Item, _ int) bool {
	// 	return item.Name == "rm ./go-test"
	// })
	// fmt.Println("test:>", items)

	return
}

func UpdateHistoryPriority(db *sql.DB, item Item) (err error) {
	if item.ID != -1 {
		return UpdateItemPriority(db, "history", item.ID, item.Priority+1)
	}
	if item.Priority >= 2 {
		return InsertItemPriority(db, "history", item.Name, item.Priority+1)
	}
	return
}
