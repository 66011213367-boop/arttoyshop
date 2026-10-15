package main

import (
	"database/sql"
	"fmt"
	"html"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// RegisterPageHandler แสดงหน้าต่างสมัครสมาชิก
func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`
	<!DOCTYPE html>
	<html lang="th">
	<head>
		<meta charset="UTF-8">
		<title>สมัครสมาชิก - Art Toy Shop</title>
		<style>
			* { box-sizing: border-box; }
			body { font-family: 'Segoe UI', sans-serif; background-color: #f0f4f8; margin: 0; padding: 0; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
			.card { background: white; padding: 30px; border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.08); width: 100%; max-width: 380px; border: 1px solid #e0e6ed; }
			h2 { margin-top: 0; color: #03045e; text-align: center; font-size: 22px; margin-bottom: 20px; }
			.form-group { margin-bottom: 15px; }
			label { display: block; margin-bottom: 5px; font-weight: 600; font-size: 13px; color: #023e8a; }
			input { width: 100%; padding: 10px 12px; border: 1px solid #cbd5e1; border-radius: 6px; font-size: 14px; outline: none; }
			button { width: 100%; background-color: #2a9d8f; color: white; border: none; padding: 12px; font-size: 15px; font-weight: bold; border-radius: 20px; cursor: pointer; transition: 0.2s; margin-top: 10px; }
			button:hover { background-color: #218376; }
			.footer-links { text-align: center; margin-top: 15px; font-size: 13px; }
			.footer-links a { color: #0077b6; text-decoration: none; font-weight: 600; }
		</style>
	</head>
	<body>
		<div class="card">
			<h2>สมัครสมาชิก</h2>
			<form method="POST" action="/register">
				<div class="form-group">
					<label>ชื่อผู้ใช้งาน (Username)</label>
					<input type="text" name="username" placeholder="ตั้งชื่อ Username" required>
				</div>
				<div class="form-group">
					<label>อีเมล (Email)</label>
					<input type="email" name="email" placeholder="example@email.com" required>
				</div>
				<div class="form-group">
					<label>รหัสผ่าน (Password)</label>
					<input type="password" name="password" placeholder="ตั้งรหัสผ่าน" required>
				</div>
				<button type="submit">ยืนยันการสมัครสมาชิก</button>
			</form>
			<div class="footer-links">
				มีบัญชีอยู่แล้ว? <a href="/login">เข้าสู่ระบบ</a> | <a href="/">กลับหน้าแรก</a>
			</div>
		</div>
	</body>
	</html>
	`))
}

// RegisterProcessHandler ประมวลผลการสมัครสมาชิกพร้อม Bcrypt Hashing
func RegisterProcessHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		// 1. อ่านค่าจาก Form ให้ครบถ้วน
		username := html.EscapeString(r.FormValue("username"))
		email := html.EscapeString(r.FormValue("email"))
		plainPassword := r.FormValue("password")

		// 2. ตรวจสอบว่า Username ซ้ำหรือไม่
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
		if err == nil && count > 0 {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`
			<script>
				alert('ชื่อผู้ใช้งานนี้ถูกใช้ไปแล้ว กรุณาใช้ชื่ออื่น');
				window.location.href = '/register';
			</script>`))
			return
		}

		// 3. แปลงรหัสผ่านเป็น Hash ด้วย bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 4. บันทึกข้อมูลลงฐานข้อมูล MySQL
		_, err = db.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, 'user')", username, email, string(hashedPassword))
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			alertMsg := fmt.Sprintf("เกิดข้อผิดพลาดในการบันทึกข้อมูล: %s", err.Error())
			_, _ = w.Write([]byte(fmt.Sprintf(`
			<script>
				alert('%s');
				window.location.href = '/register';
			</script>`, alertMsg)))
			return
		}

		// 5. สมัครสำเร็จ แจ้งเตือนและนำไปหน้า Login
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`
		<script>
			alert('สมัครสมาชิกสำเร็จแล้ว! กรุณาเข้าสู่ระบบ');
			window.location.href = '/login';
		</script>`))
	}
}
