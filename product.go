package main

import (
	"database/sql"
	"fmt"
	"html"
	"net/http"
	"strconv"
)

type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
	Stock       int
	ImageURL    string
}

func ProductsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		currentUserID := 0
		currentUserRole := ""
		currentUsername := ""

		if userIDCookie, err := r.Cookie("user_id"); err == nil {
			currentUserID, _ = strconv.Atoi(userIDCookie.Value)
		}
		if roleCookie, err := r.Cookie("user_role"); err == nil {
			currentUserRole = roleCookie.Value
		}

		if currentUserID > 0 {
			_ = db.QueryRow("SELECT username FROM users WHERE id = ?", currentUserID).Scan(&currentUsername)
		}

		var navMenuHTML string
		if currentUserRole == "admin" {
			navMenuHTML = `
				<span class="user-badge admin-badge">ADMIN: ` + currentUsername + `</span>
				<a href="/my-orders" class="btn-orders-link">คำสั่งซื้อของฉัน</a>
				<a href="/admin" class="btn-admin-link">ไปหน้าระบบหลังบ้าน Admin</a>
				<a href="/logout" class="btn-logout-link">ออกจากระบบ</a>
			`
		} else if currentUserID > 0 {
			navMenuHTML = `
				<span class="user-badge">ผู้ใช้งาน: ` + currentUsername + `</span>
				<a href="/my-orders" class="btn-orders-link">คำสั่งซื้อของฉัน</a>
				<a href="/logout" class="btn-logout-link">ออกจากระบบ</a>
			`
		} else {
			navMenuHTML = `
				<a href="/login">เข้าสู่ระบบ</a>
				<a href="/register">สมัครสมาชิก</a>
			`
		}

		rows, err := db.Query("SELECT id, name, description, price, stock, image_url FROM products ORDER BY id ASC")
		if err != nil {
			http.Error(w, "Database Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var productsListHTML string
		for rows.Next() {
			var p Product
			_ = rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.ImageURL)

			reviewRows, rErr := db.Query(`
				SELECT r.id, r.user_id, u.username, r.comment, r.rating 
				FROM reviews r 
				JOIN users u ON r.user_id = u.id 
				WHERE r.product_id = ? 
				ORDER BY r.id DESC`, p.ID)

			var reviewsHTML string
			if rErr == nil {
				for reviewRows.Next() {
					var revID, revUserID, rating int
					var username, comment string
					_ = reviewRows.Scan(&revID, &revUserID, &username, &comment, &rating)

					actionBtnsHTML := ""
					if currentUserID > 0 && currentUserID == revUserID {
						cleanComment := html.EscapeString(comment)
						actionBtnsHTML = `
						<div style="display:flex; gap:4px;">
							<button type="button" class="btn-edit-review" onclick="openEditReview('` + strconv.Itoa(revID) + `', '` + cleanComment + `', '` + strconv.Itoa(rating) + `')">แก้ไข</button>
							<form method="POST" action="/review/delete" style="display:inline; margin:0;" onsubmit="return confirm('ยืนยันลบรีวิวนี้?');">
								<input type="hidden" name="review_id" value="` + strconv.Itoa(revID) + `">
								<button type="submit" class="btn-del-review">ลบรีวิว</button>
							</form>
						</div>`
					}

					reviewsHTML += `
					<div class="review-item">
						<div class="review-header">
							<div>
								<strong>` + username + `</strong> 
								<span class="rating-stars" style="margin-left:8px;">★ ` + strconv.Itoa(rating) + `/5</span>
							</div>
							` + actionBtnsHTML + `
						</div>
						<p class="review-comment">` + comment + `</p>
					</div>`
				}
				reviewRows.Close()
			}

			if reviewsHTML == "" {
				reviewsHTML = `<p style="color:#888; font-size:13px; margin:5px 0;">ยังไม่มีรีวิวสำหรับสินค้านี้</p>`
			}

			var reviewFormHTML string
			if currentUserID > 0 {
				reviewFormHTML = `
				<form method="POST" action="/review/add" class="review-form">
					<input type="hidden" name="product_id" value="` + strconv.Itoa(p.ID) + `">
					<div style="display:flex; gap:8px; margin-top:8px;">
						<input type="text" name="comment" placeholder="เขียนรีวิวสินค้าชิ้นนี้..." required style="flex:1; padding:8px; border:1px solid #cbd5e1; border-radius:6px; font-size:13px;">
						<select name="rating" style="padding:8px; border:1px solid #cbd5e1; border-radius:6px; font-size:13px;">
							<option value="5">5 ดาว</option>
							<option value="4">4 ดาว</option>
							<option value="3">3 ดาว</option>
							<option value="2">2 ดาว</option>
							<option value="1">1 ดาว</option>
						</select>
						<button type="submit" class="btn-submit-review">ส่งรีวิว</button>
					</div>
				</form>`
			} else {
				reviewFormHTML = `<p style="font-size:12px; color:#0077b6; margin-top:10px;"><a href="/login" style="color:#0077b6; font-weight:bold;">เข้าสู่ระบบ</a> เพื่อเขียนรีวิวสินค้า</p>`
			}

			var orderFormHTML string
			if currentUserID > 0 {
				if p.Stock > 0 {
					orderFormHTML = `
					<form method="GET" action="/checkout" style="margin-top:12px; padding-top:12px; border-top:1px dashed #e2e8f0; display:flex; gap:8px; align-items:center;">
						<input type="hidden" name="product_id" value="` + strconv.Itoa(p.ID) + `">
						<label style="font-size:13px; font-weight:600; color:#023e8a;">จำนวน:</label>
						<input type="number" name="quantity" value="1" min="1" max="` + strconv.Itoa(p.Stock) + `" style="width:60px; padding:6px; border:1px solid #cbd5e1; border-radius:6px; text-align:center; font-size:14px;" required>
						<button type="submit" class="btn-buy">สั่งซื้อสินค้า</button>
					</form>`
				} else {
					orderFormHTML = `<div style="margin-top:10px; color:#e63946; font-weight:bold; font-size:13px; text-align:center;">สินค้าหมดชั่วคราว</div>`
				}
			} else {
				orderFormHTML = `<a href="/login" class="btn-buy-login">เข้าสู่ระบบเพื่อสั่งซื้อ</a>`
			}

			productsListHTML += `
			<div class="product-card">
				<img src="` + p.ImageURL + `" alt="` + p.Name + `" class="product-img">
				<div class="product-info">
					<h3 class="product-title">` + p.Name + `</h3>
					<p class="product-desc">` + p.Description + `</p>
					<div class="product-meta">
						<span class="product-price">฿` + fmt.Sprintf("%.2f", p.Price) + `</span>
						<span class="product-stock">คงเหลือ ` + strconv.Itoa(p.Stock) + ` ชิ้น</span>
					</div>
					` + orderFormHTML + `
					<div class="review-section" style="margin-top:15px;">
						<h4 style="margin:5px 0 10px 0; font-size:14px; color:#023e8a;">รีวิวจากผู้ซื้อ</h4>
						` + reviewsHTML + `
						` + reviewFormHTML + `
					</div>
				</div>
			</div>`
		}

		if productsListHTML == "" {
			productsListHTML = `<div style="grid-column: 1/-1; text-align:center; padding:40px; color:#666;">ยังไม่มีรายการสินค้าในขณะนี้</div>`
		}

		_, _ = w.Write([]byte(`
	<!DOCTYPE html>
	<html lang="th">
	<head>
		<meta charset="UTF-8">
		<title>รายการสินค้าทั้งหมด - Art Toy Shop</title>
		<style>
			* { box-sizing: border-box; }
			body { font-family: 'Segoe UI', sans-serif; background-color: #f0f4f8; margin: 0; padding: 0; color: #333; }
			.navbar { background-color: #0077b6; padding: 15px 30px; display: flex; justify-content: space-between; align-items: center; }
			.navbar h1 { color: white; margin: 0; font-size: 24px; }
			.nav-right { display: flex; align-items: center; gap: 10px; }
			.navbar a { color: white; text-decoration: none; font-weight: 600; padding: 8px 16px; border-radius: 20px; background: rgba(255,255,255,0.18); font-size: 14px; }
			.user-badge { color: white; font-weight: bold; background: rgba(255,255,255,0.25); padding: 6px 14px; border-radius: 15px; font-size: 13px; }
			.admin-badge { background-color: #e63946 !important; border: 1px solid #ffb703; }
			.btn-orders-link { background-color: #0284c7 !important; }
			.btn-admin-link { background-color: #03045e !important; }
			.btn-logout-link { background-color: #d62828 !important; }
			.container { max-width: 1100px; margin: 30px auto; padding: 0 20px; }
			.page-title { font-size: 24px; color: #03045e; margin-bottom: 25px; font-weight: 700; }
			.products-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 25px; }
			.product-card { background: white; border-radius: 14px; overflow: hidden; box-shadow: 0 4px 12px rgba(0,0,0,0.05); border: 1px solid #e0e6ed; display: flex; flex-direction: column; }
			.product-img { width: 100%; height: 240px; object-fit: contain; background-color: #fafafa; padding: 15px; border-bottom: 1px solid #edf2f7; }
			.product-info { padding: 20px; flex: 1; display: flex; flex-direction: column; }
			.product-title { margin: 0 0 8px 0; font-size: 20px; color: #023e8a; }
			.product-desc { font-size: 14px; color: #666; margin: 0 0 15px 0; line-height: 1.4; flex: 1; }
			.product-meta { display: flex; justify-content: space-between; align-items: center; margin-bottom: 5px; }
			.product-price { font-size: 20px; font-weight: bold; color: #0077b6; }
			.product-stock { font-size: 13px; color: #2a9d8f; background: #e8f5e9; padding: 4px 10px; border-radius: 10px; font-weight: 600; }
			
			.btn-buy { flex: 1; background-color: #2a9d8f; color: white; border: none; padding: 8px 12px; font-size: 14px; font-weight: bold; border-radius: 8px; cursor: pointer; transition: 0.2s; }
			.btn-buy:hover { background-color: #218376; }
			.btn-buy-login { display: block; text-align: center; background-color: #0096c7; color: white; padding: 8px; border-radius: 8px; text-decoration: none; font-weight: bold; font-size: 13px; margin-top: 10px; }

			.review-section { background: #f8fafc; padding: 12px; border-radius: 8px; border: 1px solid #e2e8f0; }
			.review-item { background: white; padding: 10px 12px; border-radius: 8px; margin-bottom: 8px; border: 1px solid #e2e8f0; }
			.review-header { display: flex; justify-content: space-between; align-items: center; font-size: 13px; }
			.rating-stars { color: #f59e0b; font-weight: bold; }
			.review-comment { font-size: 13px; color: #475569; margin: 6px 0 0 0; }
			.btn-del-review { background-color: #ef4444; color: white; border: none; padding: 4px 10px; border-radius: 6px; font-size: 12px; font-weight: bold; cursor: pointer; }
			.btn-edit-review { background-color: #f59e0b; color: white; border: none; padding: 4px 10px; border-radius: 6px; font-size: 12px; font-weight: bold; cursor: pointer; }
			.btn-submit-review { background: #0096c7; color: white; border: none; padding: 8px 12px; border-radius: 6px; font-size: 13px; font-weight: bold; cursor: pointer; }

			.modal-overlay { display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); justify-content: center; align-items: center; z-index: 1000; }
			.modal-box { background: white; padding: 25px; border-radius: 12px; max-width: 400px; width: 90%; box-shadow: 0 4px 20px rgba(0,0,0,0.15); }
			.modal-box h3 { margin-top: 0; color: #03045e; border-bottom: 2px solid #0096c7; padding-bottom: 8px; }
		</style>
	</head>
	<body>
		<div class="navbar">
			<h1>Art Toy Shop</h1>
			<div class="nav-right">
				<a href="/">หน้าแรก</a>
				` + navMenuHTML + `
			</div>
		</div>

		<div class="container">
			<h2 class="page-title">รายการสินค้าทั้งหมด (All Products)</h2>
			<div class="products-grid">
				` + productsListHTML + `
			</div>
		</div>

		<!-- Modal สำหรับแก้ไขรีวิว -->
		<div id="editReviewModal" class="modal-overlay">
			<div class="modal-box">
				<h3>แก้ไขรีวิวของคุณ</h3>
				<form method="POST" action="/review/update">
					<input type="hidden" id="editReviewId" name="review_id">
					<div style="margin-bottom:12px;">
						<label style="display:block; font-size:13px; font-weight:bold; margin-bottom:5px; color:#023e8a;">ข้อความรีวิว:</label>
						<input type="text" id="editReviewComment" name="comment" required style="width:100%; padding:8px 10px; border:1px solid #cbd5e1; border-radius:6px; font-size:14px;">
					</div>
					<div style="margin-bottom:15px;">
						<label style="display:block; font-size:13px; font-weight:bold; margin-bottom:5px; color:#023e8a;">คะแนน:</label>
						<select id="editReviewRating" name="rating" style="width:100%; padding:8px 10px; border:1px solid #cbd5e1; border-radius:6px; font-size:14px;">
							<option value="5">5 ดาว</option>
							<option value="4">4 ดาว</option>
							<option value="3">3 ดาว</option>
							<option value="2">2 ดาว</option>
							<option value="1">1 ดาว</option>
						</select>
					</div>
					<button type="submit" style="width:100%; background:#2a9d8f; color:white; border:none; padding:10px; border-radius:6px; font-weight:bold; cursor:pointer; font-size:14px;">บันทึกการแก้ไข</button>
					<button type="button" onclick="closeEditModal()" style="width:100%; background:#64748b; color:white; border:none; padding:8px; border-radius:6px; font-weight:bold; cursor:pointer; margin-top:8px; font-size:13px;">ยกเลิก</button>
				</form>
			</div>
		</div>

		<script>
			function openEditReview(id, comment, rating) {
				document.getElementById('editReviewId').value = id;
				document.getElementById('editReviewComment').value = comment;
				document.getElementById('editReviewRating').value = rating;
				document.getElementById('editReviewModal').style.display = 'flex';
			}
			function closeEditModal() {
				document.getElementById('editReviewModal').style.display = 'none';
			}
		</script>
	</body>
	</html>
	`))
	}
}

func BuyProductHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		userIDCookie, err := r.Cookie("user_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID, _ := strconv.Atoi(userIDCookie.Value)

		productID, _ := strconv.Atoi(r.FormValue("product_id"))
		quantity, _ := strconv.Atoi(r.FormValue("quantity"))
		fullname := html.EscapeString(r.FormValue("fullname"))
		phone := html.EscapeString(r.FormValue("phone"))
		address := html.EscapeString(r.FormValue("address"))
		paymentMethod := html.EscapeString(r.FormValue("payment_method"))
		note := html.EscapeString(r.FormValue("note"))

		if quantity <= 0 {
			quantity = 1
		}

		var price float64
		var stock int
		err = db.QueryRow("SELECT price, stock FROM products WHERE id = ?", productID).Scan(&price, &stock)

		if err != nil || stock < quantity {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`
			<script>
				alert('ขออภัย สินค้าไม่พอสำหรับการสั่งซื้อ');
				window.location.href = '/products';
			</script>`))
			return
		}

		totalPrice := price * float64(quantity)

		_, err = db.Exec("UPDATE products SET stock = stock - ? WHERE id = ?", quantity, productID)
		if err != nil {
			http.Error(w, "Error updating stock: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(`
			INSERT INTO orders (user_id, product_id, quantity, total_price, status, fullname, phone, address, payment_method, note) 
			VALUES (?, ?, ?, ?, 'pending', ?, ?, ?, ?, ?)`,
			userID, productID, quantity, totalPrice, fullname, phone, address, paymentMethod, note)

		if err != nil {
			http.Error(w, "Error creating order: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`
		<script>
			alert('สั่งซื้อสินค้าสำเร็จแล้ว!');
			window.location.href = '/my-orders';
		</script>`))
	}
}
