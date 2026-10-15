package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
)

func AdminHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAdminAuth(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// 1. ดึงรายการสินค้า
		productRows, err := db.Query("SELECT id, name, description, price, stock, image_url FROM products ORDER BY id ASC")
		var productsHTML string
		if err == nil {
			defer productRows.Close()
			for productRows.Next() {
				var p Product
				_ = productRows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.ImageURL)

				cleanDesc := strings.ReplaceAll(p.Description, "'", "\\'")
				cleanName := strings.ReplaceAll(p.Name, "'", "\\'")

				productsHTML += `
				<tr>
					<td>` + strconv.Itoa(p.ID) + `</td>
					<td><img src="` + p.ImageURL + `" class="tbl-img" alt="` + p.Name + `"></td>
					<td style="font-weight:600;">` + p.Name + `</td>
					<td style="color:#0077b6; font-weight:bold;">฿` + sprintf("%.2f", p.Price) + `</td>
					<td><span class="stock-badge">` + strconv.Itoa(p.Stock) + ` ชิ้น</span></td>
					<td>
						<div style="display:flex; gap:6px;">
							<button type="button" class="btn-action btn-edit" onclick="openEditModal('` + strconv.Itoa(p.ID) + `', '` + cleanName + `', '` + cleanDesc + `', '` + sprintf("%.2f", p.Price) + `', '` + strconv.Itoa(p.Stock) + `', '` + p.ImageURL + `')">แก้ไข</button>
							<form method="POST" action="/admin/delete-product" style="margin:0;" onsubmit="return confirm('ยืนยันการลบสินค้านี้?');">
								<input type="hidden" name="id" value="` + strconv.Itoa(p.ID) + `">
								<button type="submit" class="btn-action btn-delete">ลบ</button>
							</form>
						</div>
					</td>
				</tr>`
			}
		}

		// 2. ดึงรายการคำสั่งซื้อ
		type OrderItem struct {
			ID            int
			Username      string
			ProductName   string
			Quantity      int
			TotalPrice    float64
			Status        string
			Fullname      string
			Phone         string
			Address       string
			PaymentMethod string
			Note          string
			CreatedAt     string
		}
		orderRows, oErr := db.Query(`
			SELECT o.id, IFNULL(u.username, 'ไม่ระบุ'), IFNULL(p.name, 'สินค้าที่ถูกลบ'), 
			       o.quantity, o.total_price, o.status, 
			       IFNULL(o.fullname, ''), IFNULL(o.phone, ''), IFNULL(o.address, ''), IFNULL(o.payment_method, 'promptpay'), IFNULL(o.note, ''), o.created_at 
			FROM orders o 
			LEFT JOIN users u ON o.user_id = u.id 
			LEFT JOIN products p ON o.product_id = p.id 
			ORDER BY o.id DESC`)

		var ordersHTML string
		if oErr == nil {
			defer orderRows.Close()
			for orderRows.Next() {
				var ord OrderItem
				_ = orderRows.Scan(&ord.ID, &ord.Username, &ord.ProductName, &ord.Quantity, &ord.TotalPrice, &ord.Status,
					&ord.Fullname, &ord.Phone, &ord.Address, &ord.PaymentMethod, &ord.Note, &ord.CreatedAt)

				statusBadge := `<span class="badge badge-pending">รอดำเนินการ</span>`
				if ord.Status == "paid" || ord.Status == "ชำระเงินแล้ว" {
					statusBadge = `<span class="badge badge-paid">ชำระเงินแล้ว</span>`
				} else if ord.Status == "shipped" || ord.Status == "จัดส่งแล้ว" {
					statusBadge = `<span class="badge badge-shipped">จัดส่งแล้ว</span>`
				} else if ord.Status == "cancelled" || ord.Status == "ยกเลิก" {
					statusBadge = `<span class="badge badge-cancelled">ยกเลิก</span>`
				}

				payText := "โอนผ่านธนาคาร / PromptPay"
				if ord.PaymentMethod == "cod" {
					payText = "เก็บเงินปลายทาง (COD)"
				}

				cleanNote := strings.ReplaceAll(ord.Note, "'", "\\'")
				cleanAddress := strings.ReplaceAll(ord.Address, "'", "\\'")

				ordersHTML += `
				<tr>
					<td>` + strconv.Itoa(ord.ID) + `</td>
					<td style="font-weight:600;">` + ord.Username + `</td>
					<td>` + ord.ProductName + `</td>
					<td>` + strconv.Itoa(ord.Quantity) + `</td>
					<td style="color:#0077b6; font-weight:bold;">฿` + sprintf("%.2f", ord.TotalPrice) + `</td>
					<td>` + statusBadge + `</td>
					<td style="font-size:12px; color:#666;">` + ord.CreatedAt + `</td>
					<td>
						<div style="display:flex; gap:6px; align-items:center;">
							<button type="button" class="btn-action btn-detail" onclick="showOrderDetail('` + strconv.Itoa(ord.ID) + `', '` + ord.Username + `', '` + ord.Fullname + `', '` + ord.Phone + `', '` + cleanAddress + `', '` + payText + `', '` + cleanNote + `')">รายละเอียด</button>

							<form method="POST" action="/admin/update-order-status" style="margin:0; display:flex; gap:4px;">
								<input type="hidden" name="order_id" value="` + strconv.Itoa(ord.ID) + `">
								<select name="status" style="padding:4px; font-size:12px; border-radius:4px; border:1px solid #ccc;">
									<option value="pending" ` + checkSelected(ord.Status, "pending") + `>รอดำเนินการ</option>
									<option value="paid" ` + checkSelected(ord.Status, "paid") + `>ชำระเงินแล้ว</option>
									<option value="shipped" ` + checkSelected(ord.Status, "shipped") + `>จัดส่งแล้ว</option>
									<option value="cancelled" ` + checkSelected(ord.Status, "cancelled") + `>ยกเลิก</option>
								</select>
								<button type="submit" class="btn-action btn-update">ยืนยัน</button>
							</form>

							<form method="POST" action="/admin/delete-order" style="margin:0;" onsubmit="return confirm('ยืนยันการลบคำสั่งซื้อนี้?');">
								<input type="hidden" name="order_id" value="` + strconv.Itoa(ord.ID) + `">
								<button type="submit" class="btn-action btn-delete">ลบ</button>
							</form>
						</div>
					</td>
				</tr>`
			}
		}

		if ordersHTML == "" {
			ordersHTML = `<tr><td colspan="8" style="text-align:center; color:#888; padding:20px;">ยังไม่มีรายการสั่งซื้อ</td></tr>`
		}

		// 3. ดึงข้อมูลรีวิวของลูกค้า
		type AdminReviewItem struct {
			ID          int
			Username    string
			ProductName string
			Rating      int
			Comment     string
			CreatedAt   string
		}
		reviewRows, rErr := db.Query(`
			SELECT r.id, IFNULL(u.username, 'ผู้ใช้ทั่วไป'), IFNULL(p.name, 'สินค้าที่ถูกลบ'), r.rating, r.comment, r.created_at 
			FROM reviews r 
			LEFT JOIN users u ON r.user_id = u.id 
			LEFT JOIN products p ON r.product_id = p.id 
			ORDER BY r.id DESC`)

		var reviewsHTML string
		if rErr == nil {
			defer reviewRows.Close()
			for reviewRows.Next() {
				var rev AdminReviewItem
				_ = reviewRows.Scan(&rev.ID, &rev.Username, &rev.ProductName, &rev.Rating, &rev.Comment, &rev.CreatedAt)

				reviewsHTML += `
				<tr>
					<td>` + strconv.Itoa(rev.ID) + `</td>
					<td style="font-weight:600;">` + rev.Username + `</td>
					<td style="color:#023e8a; font-weight:600;">` + rev.ProductName + `</td>
					<td style="color:#f59e0b; font-weight:bold;">★ ` + strconv.Itoa(rev.Rating) + `/5</td>
					<td>` + rev.Comment + `</td>
					<td style="font-size:12px; color:#666;">` + rev.CreatedAt + `</td>
				</tr>`
			}
		}

		if reviewsHTML == "" {
			reviewsHTML = `<tr><td colspan="6" style="text-align:center; color:#888; padding:20px;">ยังไม่มีรีวิวจากลูกค้า</td></tr>`
		}

		// 4. ดึงข้อมูลผู้ใช้งาน
		type UserItem struct {
			ID        int
			Username  string
			Email     string
			Role      string
			CreatedAt string
		}
		userRows, err := db.Query("SELECT id, username, email, role, created_at FROM users ORDER BY id DESC")
		var usersHTML string
		if err == nil {
			defer userRows.Close()
			for userRows.Next() {
				var u UserItem
				_ = userRows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt)

				roleBadge := `<span class="badge badge-user">User</span>`
				if u.Role == "admin" {
					roleBadge = `<span class="badge badge-admin">Admin</span>`
				}

				usersHTML += `
				<tr>
					<td>` + strconv.Itoa(u.ID) + `</td>
					<td style="font-weight:600;">` + u.Username + `</td>
					<td>` + u.Email + `</td>
					<td>` + roleBadge + `</td>
					<td style="font-size:12px; color:#666;">` + u.CreatedAt + `</td>
					<td>
						<form method="POST" action="/admin/delete-user" style="margin:0;" onsubmit="return confirm('ยืนยันการลบผู้ใช้คนนี้?');">
							<input type="hidden" name="id" value="` + strconv.Itoa(u.ID) + `">
							<button type="submit" class="btn-action btn-delete">ลบผู้ใช้</button>
						</form>
					</td>
				</tr>`
			}
		}

		_, _ = w.Write([]byte(`
	<!DOCTYPE html>
	<html lang="th">
	<head>
		<meta charset="UTF-8">
		<title>Admin Dashboard - Art Toy Shop</title>
		<style>
			* { box-sizing: border-box; }
			body { font-family: 'Segoe UI', sans-serif; background-color: #f0f4f8; margin: 0; padding: 0; color: #333; }
			.navbar { background-color: #03045e; padding: 15px 30px; display: flex; justify-content: space-between; align-items: center; }
			.navbar h1 { color: white; margin: 0; font-size: 22px; }
			.navbar a { color: white; text-decoration: none; font-weight: 600; padding: 8px 18px; border-radius: 20px; background: rgba(255,255,255,0.15); font-size: 14px; }
			.btn-logout { background-color: #e63946 !important; }
			.container { max-width: 1150px; margin: 30px auto; padding: 0 20px; display: flex; flex-direction: column; gap: 30px; }
			.grid-top { display: grid; grid-template-columns: 1fr 2fr; gap: 25px; }
			.card { background: white; border-radius: 14px; padding: 25px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); border: 1px solid #e0e6ed; }
			.card h2 { margin-top: 0; color: #03045e; font-size: 20px; border-bottom: 2px solid #0096c7; padding-bottom: 10px; margin-bottom: 20px; }
			.form-group { margin-bottom: 15px; }
			.form-group label { display: block; margin-bottom: 5px; font-weight: 600; font-size: 13px; color: #023e8a; }
			.form-group input, .form-group textarea, .form-group select { width: 100%; padding: 10px 12px; border: 1px solid #cbd5e1; border-radius: 8px; font-size: 14px; outline: none; }
			.btn-add { width: 100%; background-color: #0096c7; color: white; border: none; padding: 12px; font-size: 15px; font-weight: bold; border-radius: 20px; cursor: pointer; }
			table { width: 100%; border-collapse: collapse; margin-top: 10px; }
			th { background-color: #f8fafc; color: #023e8a; text-align: left; padding: 12px; font-size: 14px; border-bottom: 2px solid #e2e8f0; }
			td { padding: 10px 12px; border-bottom: 1px solid #edf2f7; font-size: 14px; vertical-align: middle; }
			.tbl-img { width: 45px; height: 45px; object-fit: contain; border-radius: 6px; background: #f8fafc; border: 1px solid #e2e8f0; }
			.stock-badge { background: #e0f2fe; color: #0369a1; padding: 4px 8px; border-radius: 6px; font-weight: bold; font-size: 12px; }
			.badge { padding: 4px 10px; border-radius: 12px; font-size: 12px; font-weight: bold; display: inline-block; }
			.badge-user { background-color: #e2e8f0; color: #334155; }
			.badge-admin { background-color: #0096c7; color: white; }
			.badge-pending { background-color: #fef08a; color: #854d0e; }
			.badge-paid { background-color: #bbf7d0; color: #166534; }
			.badge-shipped { background-color: #bfdbfe; color: #1e40af; }
			.badge-cancelled { background-color: #fecaca; color: #991b1b; }
			.btn-action { border: none; padding: 5px 10px; border-radius: 6px; cursor: pointer; font-size: 12px; font-weight: bold; }
			.btn-delete { background-color: #ff4d4d; color: white; }
			.btn-edit { background-color: #f59e0b; color: white; }
			.btn-update { background-color: #0284c7; color: white; }
			.btn-detail { background-color: #64748b; color: white; }

			.modal-overlay { display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); justify-content: center; align-items: center; z-index: 1000; }
			.modal-box { background: white; padding: 25px; border-radius: 12px; max-width: 500px; width: 90%; box-shadow: 0 4px 20px rgba(0,0,0,0.15); max-height: 90vh; overflow-y: auto; }
			.modal-box h3 { margin-top: 0; color: #03045e; border-bottom: 2px solid #0096c7; padding-bottom: 8px; }
			.modal-box p { font-size: 14px; margin: 8px 0; color: #475569; }
			.btn-close-modal { width: 100%; background: #64748b; color: white; border: none; padding: 10px; border-radius: 6px; font-weight: bold; cursor: pointer; margin-top: 10px; }
			.btn-save-edit { width: 100%; background: #2a9d8f; color: white; border: none; padding: 12px; border-radius: 6px; font-weight: bold; cursor: pointer; font-size: 15px; margin-top: 15px; }
		</style>
	</head>
	<body>
		<div class="navbar">
			<h1>ระบบหลังบ้าน Admin</h1>
			<div>
				<a href="/">หน้าแรก</a>
				<a href="/products">หน้าร้านค้า</a>
				<a href="/logout" class="btn-logout">ออกจากระบบ</a>
			</div>
		</div>

		<div class="container">
			<div class="grid-top">
				<div class="card">
					<h2>เพิ่มสินค้าใหม่</h2>
					<form method="POST" action="/admin/add-product">
						<div class="form-group"><label>ชื่อสินค้า</label><input type="text" name="name" placeholder="เช่น Hirono Little Mischief" required></div>
						<div class="form-group"><label>รายละเอียด</label><textarea name="description" rows="2" placeholder="รายละเอียดสินค้า..."></textarea></div>
						<div class="form-group"><label>ราคา (บาท)</label><input type="number" step="0.01" name="price" placeholder="550.00" required></div>
						<div class="form-group"><label>จำนวนสต็อก</label><input type="number" name="stock" placeholder="10" required></div>
						
						<div class="form-group">
							<label>เลือกพรีเซ็ต หรือระบุ URL / พาธรูปภาพ</label>
							<select id="imgSelect" onchange="document.getElementById('imgUrlInput').value = this.value" style="margin-bottom:8px;">
								<option value="/images/hirono.png">Hirono (.png)</option>
								<option value="/images/hirono v.2.png">Hirono V.2 (.png)</option>
								<option value="/images/labobu.png">LABUBU (.png)</option>
								<option value="/images/molly.jpg">MOLLY (.jpg)</option>
								<option value="/images/crybaby.jpg">CRYBABY (.jpg)</option>
								<option value="/images/skullpanda.jpg">SKULLPANDA (.jpg)</option>
								<option value="/images/twinkletwinkle.png">Twinkle Twinkle (.png)</option>
								<option value="/images/twinkletwinkle v.2.png">Twinkle Twinkle V.2 (.png)</option>
							</select>
							<input type="text" id="imgUrlInput" name="image_url" value="/images/hirono.png" placeholder="ระบุ /images/filename.png หรือ URL ออนไลน์" required>
						</div>

						<button type="submit" class="btn-add">บันทึกสินค้า</button>
					</form>
				</div>

				<div class="card">
					<h2>รายการสินค้าทั้งหมด</h2>
					<table>
						<thead>
							<tr><th>ID</th><th>รูป</th><th>ชื่อสินค้า</th><th>ราคา</th><th>สต็อก</th><th>จัดการ</th></tr>
						</thead>
						<tbody>` + productsHTML + `</tbody>
					</table>
				</div>
			</div>

			<!-- ส่วนจัดการคำสั่งซื้อ -->
			<div class="card">
				<h2>จัดการคำสั่งซื้อ (Orders Management)</h2>
				<table>
					<thead>
						<tr><th>Order ID</th><th>ผู้ซื้อ</th><th>สินค้า</th><th>จำนวน</th><th>ราคารวม</th><th>สถานะ</th><th>วันที่สั่งซื้อ</th><th>จัดการ</th></tr>
					</thead>
					<tbody>` + ordersHTML + `</tbody>
				</table>
			</div>

			<!-- ส่วนรายการรีวิว -->
			<div class="card">
				<h2>รายการรีวิวจากลูกค้า (Customer Reviews)</h2>
				<table>
					<thead>
						<tr><th>Review ID</th><th>ผู้รีวิว</th><th>สินค้า</th><th>คะแนน</th><th>ข้อความรีวิว</th><th>วันที่เขียน</th></tr>
					</thead>
					<tbody>` + reviewsHTML + `</tbody>
				</table>
			</div>

			<!-- ส่วนจัดการผู้ใช้งาน -->
			<div class="card">
				<h2>ข้อมูลผู้ใช้งานในระบบ (Users Management)</h2>
				<table>
					<thead>
						<tr><th>ID</th><th>Username</th><th>Email</th><th>สิทธิ์ (Role)</th><th>วันที่สมัคร</th><th>จัดการ</th></tr>
					</thead>
					<tbody>` + usersHTML + `</tbody>
				</table>
			</div>
		</div>

		<!-- Modal ดูรายละเอียดออเดอร์ -->
		<div id="orderModal" class="modal-overlay">
			<div class="modal-box">
				<h3>รายละเอียดการจัดส่ง (Order #<span id="mOrderId"></span>)</h3>
				<p><strong>Username ผู้ซื้อ:</strong> <span id="mUsername"></span></p>
				<p><strong>ชื่อผู้รับ:</strong> <span id="mFullname"></span></p>
				<p><strong>เบอร์โทรศัพท์:</strong> <span id="mPhone"></span></p>
				<p><strong>ที่อยู่จัดส่ง:</strong> <span id="mAddress"></span></p>
				<p><strong>ชำระเงินโดย:</strong> <span id="mPayment"></span></p>
				<p><strong>หมายเหตุถึงผู้ขาย:</strong> <span id="mNote" style="color:#0284c7; font-weight:bold;"></span></p>
				<button type="button" class="btn-close-modal" onclick="closeModal('orderModal')">ปิดหน้าต่าง</button>
			</div>
		</div>

		<!-- Modal แก้ไขข้อมูลสินค้า -->
		<div id="editProductModal" class="modal-overlay">
			<div class="modal-box">
				<h3>แก้ไขข้อมูลสินค้า (ID: <span id="editPIdText"></span>)</h3>
				<form method="POST" action="/admin/update-product">
					<input type="hidden" id="editPId" name="id">
					
					<div class="form-group">
						<label>ชื่อสินค้า</label>
						<input type="text" id="editPName" name="name" required>
					</div>
					<div class="form-group">
						<label>รายละเอียดสินค้า</label>
						<textarea id="editPDesc" name="description" rows="2"></textarea>
					</div>
					<div class="form-group">
						<label>ราคา (บาท)</label>
						<input type="number" step="0.01" id="editPPrice" name="price" required>
					</div>
					<div class="form-group">
						<label>จำนวนสต็อกคงเหลือ</label>
						<input type="number" id="editPStock" name="stock" required>
					</div>
					<div class="form-group">
						<label>URL / พาธรูปภาพสินค้า</label>
						<input type="text" id="editPImg" name="image_url" required>
					</div>

					<button type="submit" class="btn-save-edit">บันทึกการแก้ไข</button>
					<button type="button" class="btn-close-modal" onclick="closeModal('editProductModal')">ยกเลิก</button>
				</form>
			</div>
		</div>

		<script>
			function showOrderDetail(id, username, fullname, phone, address, payment, note) {
				document.getElementById('mOrderId').innerText = id;
				document.getElementById('mUsername').innerText = username;
				document.getElementById('mFullname').innerText = fullname || 'ไม่ได้ระบุ';
				document.getElementById('mPhone').innerText = phone || 'ไม่ได้ระบุ';
				document.getElementById('mAddress').innerText = address || 'ไม่ได้ระบุ';
				document.getElementById('mPayment').innerText = payment || 'โอนผ่านธนาคาร / PromptPay';
				document.getElementById('mNote').innerText = note || 'ไม่มีหมายเหตุ';
				document.getElementById('orderModal').style.display = 'flex';
			}

			function openEditModal(id, name, desc, price, stock, img) {
				document.getElementById('editPIdText').innerText = id;
				document.getElementById('editPId').value = id;
				document.getElementById('editPName').value = name;
				document.getElementById('editPDesc').value = desc;
				document.getElementById('editPPrice').value = price;
				document.getElementById('editPStock').value = stock;
				document.getElementById('editPImg').value = img;
				document.getElementById('editProductModal').style.display = 'flex';
			}

			function closeModal(modalId) {
				document.getElementById(modalId).style.display = 'none';
			}
		</script>
	</body>
	</html>
	`))
	}
}

func sprintf(format string, value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func checkSelected(currentStatus, targetStatus string) string {
	if currentStatus == targetStatus {
		return "selected"
	}
	return ""
}

func AddProductHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAdminAuth(r) {
			http.Error(w, "Unauthorized access", http.StatusUnauthorized)
			return
		}
		if r.Method == http.MethodPost {
			name := r.FormValue("name")
			description := r.FormValue("description")
			price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
			stock, _ := strconv.Atoi(r.FormValue("stock"))
			imageURL := r.FormValue("image_url")

			if imageURL == "" {
				imageURL = "/images/labobu.png"
			}

			adminID := 1
			if cookie, err := r.Cookie("user_id"); err == nil {
				adminID, _ = strconv.Atoi(cookie.Value)
			}

			_, _ = db.Exec("INSERT INTO products (name, description, price, stock, user_id, image_url) VALUES (?, ?, ?, ?, ?, ?)",
				name, description, price, stock, adminID, imageURL)

			reorderProductIDs(db)
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func UpdateProductHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAdminAuth(r) {
			return
		}
		if r.Method == http.MethodPost {
			id := r.FormValue("id")
			name := r.FormValue("name")
			description := r.FormValue("description")
			price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
			stock, _ := strconv.Atoi(r.FormValue("stock"))
			imageURL := r.FormValue("image_url")

			_, _ = db.Exec(`
				UPDATE products 
				SET name = ?, description = ?, price = ?, stock = ?, image_url = ? 
				WHERE id = ?`,
				name, description, price, stock, imageURL, id)
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func DeleteProductHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAdminAuth(r) {
			return
		}
		if r.Method == http.MethodPost {
			id := r.FormValue("id")
			_, _ = db.Exec("DELETE FROM products WHERE id = ?", id)

			reorderProductIDs(db)
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func reorderProductIDs(db *sql.DB) {
	_, _ = db.Exec("SET @count = 0;")
	_, _ = db.Exec("UPDATE products SET id = (@count := @count + 1);")
	_, _ = db.Exec("ALTER TABLE products AUTO_INCREMENT = 1;")
}

func UpdateOrderStatusHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAdminAuth(r) {
			return
		}
		if r.Method == http.MethodPost {
			orderID := r.FormValue("order_id")
			status := r.FormValue("status")
			_, _ = db.Exec("UPDATE orders SET status = ? WHERE id = ?", status, orderID)
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func DeleteOrderHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAdminAuth(r) {
			return
		}
		if r.Method == http.MethodPost {
			orderID := r.FormValue("order_id")
			_, _ = db.Exec("DELETE FROM orders WHERE id = ?", orderID)
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func DeleteUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAdminAuth(r) {
			return
		}
		if r.Method == http.MethodPost {
			id := r.FormValue("id")
			_, _ = db.Exec("DELETE FROM users WHERE id = ?", id)
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func checkAdminAuth(r *http.Request) bool {
	cookie, err := r.Cookie("user_role")
	return err == nil && cookie.Value == "admin"
}
