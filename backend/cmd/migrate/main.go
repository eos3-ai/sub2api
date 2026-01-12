// Package main 提供独立的数据库迁移工具
//
// 使用方式：
//   go run ./cmd/migrate [command]
//
// 命令：
//   apply   - 应用所有待执行的迁移（默认）
//   status  - 查看迁移状态
//   verify  - 验证约束是否生效
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/Wei-Shaw/sub2api/internal/repository"
)

func main() {
	// 从环境变量读取数据库配置
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/sub2api?sslmode=disable"
		log.Printf("DATABASE_URL not set, using default: %s", dbURL)
	}

	// 连接数据库
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Database connection successful")

	// 解析命令
	command := "apply"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "apply":
		applyMigrations(db)
	case "status":
		showStatus(db)
	case "verify":
		verifyConstraint(db)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("\nAvailable commands:")
		fmt.Println("  apply   - Apply all pending migrations")
		fmt.Println("  status  - Show migration status")
		fmt.Println("  verify  - Verify balance constraint")
		os.Exit(1)
	}
}

// applyMigrations 应用所有待执行的迁移
func applyMigrations(db *sql.DB) {
	log.Println("Applying migrations...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := repository.ApplyMigrations(ctx, db); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	log.Println("✓ All migrations applied successfully")

	// 显示迁移状态
	showStatus(db)
}

// showStatus 显示迁移状态
func showStatus(db *sql.DB) {
	log.Println("\nMigration Status:")
	log.Println("================")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 查询所有已应用的迁移
	rows, err := db.QueryContext(ctx, `
		SELECT filename, applied_at
		FROM schema_migrations
		ORDER BY filename
	`)
	if err != nil {
		log.Printf("Warning: failed to query migrations: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var filename string
		var appliedAt time.Time

		if err := rows.Scan(&filename, &appliedAt); err != nil {
			log.Printf("Warning: failed to scan row: %v", err)
			continue
		}

		log.Printf("  ✓ %s (applied: %s)", filename, appliedAt.Format("2006-01-02 15:04:05"))
		count++
	}

	log.Printf("\nTotal: %d migrations applied\n", count)

	// 检查余额约束
	var constraintExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_constraint
			WHERE conname = 'check_balance_non_negative'
			AND conrelid = 'users'::regclass
		)
	`).Scan(&constraintExists)

	if err != nil {
		log.Printf("Warning: failed to check constraint: %v", err)
	} else if constraintExists {
		log.Println("✓ Balance constraint (check_balance_non_negative) is active")
	} else {
		log.Println("⚠ Balance constraint NOT found")
	}
}

// verifyConstraint 验证余额约束是否生效
func verifyConstraint(db *sql.DB) {
	log.Println("Verifying balance constraint...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试1：检查约束是否存在
	var constraintExists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_constraint
			WHERE conname = 'check_balance_non_negative'
			AND conrelid = 'users'::regclass
		)
	`).Scan(&constraintExists)

	if err != nil {
		log.Fatalf("Failed to check constraint existence: %v", err)
	}

	if !constraintExists {
		log.Println("✗ Constraint does NOT exist")
		log.Println("  Run: go run ./cmd/migrate apply")
		os.Exit(1)
	}

	log.Println("✓ Constraint exists in database")

	// 测试2：尝试插入负余额（应失败）
	log.Println("\nTesting constraint enforcement...")
	log.Println("  Attempting to insert user with negative balance...")

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback() // 无论如何都回滚测试数据

	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (email, password_hash, balance)
		VALUES ('test_negative_balance@test.com', 'test_hash', -10.0)
	`)

	if err != nil {
		// 预期应该失败
		if contains(err.Error(), "check_balance_non_negative") {
			log.Println("  ✓ Constraint correctly blocked negative balance")
		} else {
			log.Printf("  ✗ Unexpected error: %v", err)
			os.Exit(1)
		}
	} else {
		log.Println("  ✗ Constraint DID NOT block negative balance!")
		log.Println("  This is a critical security issue!")
		os.Exit(1)
	}

	// 测试3：检查是否有现存负余额用户
	log.Println("\nChecking for existing negative balances...")

	var negativeCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users WHERE balance < 0
	`).Scan(&negativeCount)

	if err != nil {
		log.Printf("Warning: failed to count negative balances: %v", err)
	} else if negativeCount > 0 {
		log.Printf("  ⚠ Found %d users with negative balance", negativeCount)
		log.Println("  This should not happen with the constraint active!")
		log.Println("  Please investigate immediately.")
	} else {
		log.Println("  ✓ No users with negative balance")
	}

	log.Println("\n✓ All verification tests passed")
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
