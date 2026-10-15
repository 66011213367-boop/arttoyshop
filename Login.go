package main

import (
	"database/sql"
	"html"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

// LoginPageHandler แสดงหน้าต่างเข้าสู่ระบบ
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`
	<!DOCTYPE html>
	<html lang="th">
	<head>
		<meta charset="UTF-8">
		<title>เข้าสู่ระบบ - Art Toy Shop</title>
		<style>
			* { box-sizing: border-box; }
			body { font-family: 'Segoe UI', sans-serif; background-color: #f0f4f8; margin: 0; padding: 0; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
			.card { background: white; padding: 30px; border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.08); width: 100%; max-width: 380px; border: 1px solid #e0e6ed; }
			h2 { margin-top: 0; color: #03045e; text-align: center; font-size: 22px; margin-bottom: 20px; }
			.form-group { margin-bottom: 15px; }
			label { display: block; margin-bottom: 5px; font-weight: 600; font-size: 13px; color: #023e8a; }
			input { width: 100%; padding: 10px 12px; border: 1px solid #cbd5e1; border-radius: 6px; font-size: 14px; outline: none; }
			button { width: 100%; background-color: #0077b6; color: white; border: none; padding: 12px; font-size: 15px; font-weight: bold; border-radius: 20px; cursor: pointer; transition: 0.2s; margin-top: 10px; }
			button:hover { background-color: #0096c7; }
			.footer-links { text-align: center; margin-top: 15px; font-size: 13px; }
			.footer-links a { color: #0077b6; text-decoration: none; font-weight: 600; }
		</style>
	</head>
	<body>
		<div class="card">
			<h2>เข้าสู่ระบบ</h2>
			<form method="POST" action="/login">
				<div class="form-group">
					<label>ชื่อผู้ใช้งาน (Username)</label>
					<input type="text" name="username" placeholder="ระบุ Username" required>
				</div>
				<div class="form-group">
					<label>รหัสผ่าน (Password)</label>
					<input type="password" name="password" placeholder="ระบุ Password" required>
				</div>
				<button type="submit">เข้าสู่ระบบ</button>
			</form>
			<div class="footer-links">
				ยังไม่มีบัญชี? <a href="/register">สมัครสมาชิก</a> | <a href="/">กลับหน้าแรก</a>
			</div>
		</div>
	</body>
	</html>
	`))
}

// LoginProcessHandler ตรวจสอบการเข้าสู่ระบบด้วย bcrypt
func LoginProcessHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		username := html.EscapeString(r.FormValue("username"))
		plainPassword := r.FormValue("password")

		var id int
		var role string
		var dbHashedPassword string

		// 1. ดึงข้อมูลผู้ใช้จาก DB
		err := db.QueryRow("SELECT id, password, role FROM users WHERE username = ?", username).Scan(&id, &dbHashedPassword, &role)

		// 2. ตรวจสอบคำสั่งเปรียบเทียบรหัสผ่าน bcrypt
		if err != nil || bcrypt.CompareHashAndPassword([]byte(dbHashedPassword), []byte(plainPassword)) != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`
			<script>
				alert('ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง');
				window.location.href = '/login';
			</script>`))
			return
		}

		// 3. ตั้งค่า Secure Cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "user_id",
			Value:    strconv.Itoa(id),
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "user_role",
			Value:    role,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		if role == "admin" {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/products", http.StatusSeeOther)
	}
}

// LogoutHandler ออกจากระบบและล้าง Cookie
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "user_role",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
