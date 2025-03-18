package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

type SMS struct {
	ID            int    `json:"id"`
	Message       string `json:"message"`
	Sender        string `json:"sender"`
	LocalSentTime string `json:"local_sent_time"`
	SentTime      string `json:"sent_time"`
}

type QuerySMSResponse struct {
	Count   int   `json:"count"`
	Limit   int   `json:"limit"`
	Records []SMS `json:"records"`
}

func initDB(dbPath string) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		db, err := sql.Open("sqlite", defaultDbPath)
		if err != nil {
			log.Fatalf("无法创建数据库: %v", err)
		}
		defer db.Close()
		createTableSQL := `CREATE TABLE sms (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message TEXT,
			sender TEXT NOT NULL,
			local_sent_time TEXT NOT NULL,
			sent_time TEXT NOT NULL
		);`
		_, err = db.Exec(createTableSQL)
		if err != nil {
			log.Fatalf("无法创建表: %v", err)
		}
		log.Println("数据库和表创建成功")
	} else {
		log.Println("数据库文件已存在")
	}
}
func querySMS(db *sql.DB, page, limit int) (QuerySMSResponse, error) {
	offset := (page - 1) * limit
	rows, err := db.Query("SELECT id, message, sender, local_sent_time, sent_time FROM sms LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return QuerySMSResponse{}, err
	}
	defer rows.Close()
	var records []SMS
	for rows.Next() {
		var sms SMS
		if err := rows.Scan(&sms.ID, &sms.Message, &sms.Sender, &sms.LocalSentTime, &sms.SentTime); err != nil {
			return QuerySMSResponse{}, err
		}
		records = append(records, sms)
	}
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sms").Scan(&count)
	if err != nil {
		return QuerySMSResponse{}, err
	}
	response := QuerySMSResponse{
		Count:   count,
		Limit:   limit,
		Records: records,
	}
	return response, nil
}

func deleteSMS(db *sql.DB, ids []int) (int64, error) {
	query := "DELETE FROM sms WHERE id IN (?" + strings.Repeat(",?", len(ids)-1) + ")"
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}
func insertSMS(db *sql.DB, sender string, message string, localTime string, sentTime string) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO sms (sender, message, local_sent_time, sent_time) VALUES (?, ?, ?, ?)",
		sender, message, localTime, sentTime,
	)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}
