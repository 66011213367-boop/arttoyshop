package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// ConnectDB ทำหน้าที่เชื่อมต่อฐานข้อมูล MySQL
func ConnectDB() *sql.DB {
	// อ่านค่า DSN จาก Environment Variable ของ Railway
	dsn := os.Getenv("MYSQL_URL")
	if dsn == "" {
		// ค่าสำรองสำหรับรันบน เครื่อง Localhost
		dsn = "root:@tcp(127.0.0.1:3306)/arttoy_db?charset=utf8mb4&parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error creating DB pool: %v", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(3 * time.Minute)

	err = db.Ping()
	if err != nil {
		log.Fatalf("Cannot connect to MySQL: %v", err)
	}

	autoSetupDatabase(db)

	fmt.Println("Successfully connected to MySQL Database!")
	return db
}

// autoSetupDatabase ตรวจสอบและจัดการโครงสร้างตาราง พร้อมตั้งค่า Admin
func autoSetupDatabase(db *sql.DB) {
	// 1. สร้างตาราง users หากยังไม่มี
	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(100) NOT NULL UNIQUE,
		email VARCHAR(255) DEFAULT '',
		password VARCHAR(255) NOT NULL DEFAULT '',
		role VARCHAR(50) DEFAULT 'user',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, _ = db.Exec(createUsersTableSQL)

	// จัดการคอลัมน์ในตาราง users ให้ถูกต้องเรียบร้อย
	userCols := []string{
		"ALTER TABLE users ADD COLUMN email VARCHAR(255) DEFAULT ''",
		"ALTER TABLE users ADD COLUMN password VARCHAR(255) NOT NULL DEFAULT ''",
		"ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user'",
	}
	for _, colSQL := range userCols {
		_, _ = db.Exec(colSQL)
	}

	// หากใน DB มีคอลัมน์ password_hash อยู่ ให้ย้ายข้อมูลมาที่ password แล้วลบ password_hash ทิ้ง
	var hasOldHashCol int
	_ = db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = 'arttoy_db' AND TABLE_NAME = 'users' AND COLUMN_NAME = 'password_hash'").Scan(&hasOldHashCol)
	if hasOldHashCol > 0 {
		_, _ = db.Exec("UPDATE users SET password = password_hash WHERE (password IS NULL OR password = '') AND password_hash IS NOT NULL AND password_hash != ''")
		_, _ = db.Exec("ALTER TABLE users DROP COLUMN password_hash")
	}

	// 2. บังคับอัปเดต/รีเซ็ตรหัสผ่านบัญชี admin ให้เป็น bcrypt hash (รหัสผ่านคือ 123456)
	adminHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err == nil {
		var adminCount int
		_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&adminCount)
		if adminCount == 0 {
			_, _ = db.Exec("INSERT INTO users (username, email, password, role) VALUES ('admin', 'admin@arttoy.com', ?, 'admin')", string(adminHash))
		} else {
			_, _ = db.Exec("UPDATE users SET password = ?, role = 'admin' WHERE username = 'admin'", string(adminHash))
		}
	}

	// 3. สร้างตาราง products
	createProductsTableSQL := `
	CREATE TABLE IF NOT EXISTS products (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price DECIMAL(10, 2) NOT NULL,
		stock INT NOT NULL DEFAULT 0,
		user_id INT,
		image_url VARCHAR(255) DEFAULT '/images/labobu.png',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, _ = db.Exec(createProductsTableSQL)

	// 4. สร้างตาราง orders
	createOrdersTableSQL := `
	CREATE TABLE IF NOT EXISTS orders (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		product_id INT NOT NULL,
		quantity INT NOT NULL DEFAULT 1,
		total_price DECIMAL(10, 2) NOT NULL,
		status VARCHAR(50) DEFAULT 'pending',
		fullname VARCHAR(255) DEFAULT '',
		phone VARCHAR(50) DEFAULT '',
		address TEXT,
		payment_method VARCHAR(50) DEFAULT 'promptpay',
		note TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
	);`
	_, _ = db.Exec(createOrdersTableSQL)

	orderCols := []string{
		"ALTER TABLE orders ADD COLUMN status VARCHAR(50) DEFAULT 'pending'",
		"ALTER TABLE orders ADD COLUMN fullname VARCHAR(255) DEFAULT ''",
		"ALTER TABLE orders ADD COLUMN phone VARCHAR(50) DEFAULT ''",
		"ALTER TABLE orders ADD COLUMN address TEXT",
		"ALTER TABLE orders ADD COLUMN payment_method VARCHAR(50) DEFAULT 'promptpay'",
		"ALTER TABLE orders ADD COLUMN note TEXT",
	}
	for _, colSQL := range orderCols {
		_, _ = db.Exec(colSQL)
	}

	// 5. สร้างตาราง reviews
	createReviewsTableSQL := `
	CREATE TABLE IF NOT EXISTS reviews (
		id INT AUTO_INCREMENT PRIMARY KEY,
		product_id INT NOT NULL,
		user_id INT NOT NULL,
		rating INT NOT NULL DEFAULT 5,
		comment TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
	);`
	_, _ = db.Exec(createReviewsTableSQL)
}
