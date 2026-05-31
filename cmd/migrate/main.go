package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/husky/husky/internal/config"
	_ "github.com/lib/pq"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config YAML file")
	flag.Parse()

	// 从 YAML 加载配置获取数据库连接信息
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)

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
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		log.Fatalf("Failed to read migration directory: %v", err)
	}

	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, filepath.Join(migrationDir, file.Name()))
		}
	}

	// 执行每个迁移脚本
	for _, sqlFile := range sqlFiles {
		fmt.Printf("Executing migration: %s\n", sqlFile)

		content, err := os.ReadFile(sqlFile)
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
