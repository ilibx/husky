package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// 从环境变量读取数据库连接信息
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "husky"
	}

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to database!")

	// 读取迁移脚本目录
	migrationDir := "./scripts/migrations"
	files, err := ioutil.ReadDir(migrationDir)
	if err != nil {
		log.Fatalf("Failed to read migration directory: %v", err)
	}

	// 按文件名排序
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, filepath.Join(migrationDir, file.Name()))
		}
	}

	// 执行每个迁移脚本
	for _, sqlFile := range sqlFiles {
		fmt.Printf("Executing migration: %s\n", sqlFile)

		content, err := ioutil.ReadFile(sqlFile)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", sqlFile, err)
		}

		// 执行 SQL
		_, err = db.Exec(string(content))
		if err != nil {
			log.Fatalf("Failed to execute migration %s: %v", sqlFile, err)
		}

		fmt.Printf("✓ Migration %s completed successfully\n", sqlFile)
	}

	fmt.Println("\n✅ All migrations completed successfully!")
}
