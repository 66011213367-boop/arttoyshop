package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
)

func CheckoutPageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		userIDCookie, err := r.Cookie("user_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID, _ := strconv.Atoi(userIDCookie.Value)

		productID, _ := strconv.Atoi(r.URL.Query().Get("product_id"))
		quantity, _ := strconv.Atoi(r.URL.Query().Get("quantity"))

		if quantity <= 0 {
			quantity = 1
		}

		var p Product
		err = db.QueryRow("SELECT id, name, price, stock, image_url FROM products WHERE id = ?", productID).Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.ImageURL)
		if err != nil {
			http.Redirect(w, r, "/products", http.StatusSeeOther)
			return
		}

		var username string
		_ = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

		unitPrice := p.Price
		totalPrice := unitPrice * float64(quantity)

		_, _ = w.Write([]byte(`
	<!DOCTYPE html>
	<html lang="th">
	<head>
		<meta charset="UTF-8">
		<title>ยืนยันการสั่งซื้อ - Art Toy Shop</title>
		<style>
			* { box-sizing: border-box; }
			body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f0f4f8; margin: 0; padding: 0; color: #333; }
			.navbar { background-color: #0077b6; padding: 15px 30px; display: flex; justify-content: space-between; align-items: center; }
			.navbar h1 { color: white; margin: 0; font-size: 24px; font-weight: 700; }
			.navbar a { color: white; text-decoration: none; font-weight: 600; padding: 8px 16px; border-radius: 20px; background: rgba(255,255,255,0.18); font-size: 14px; }
			
			.container { max-width: 950px; margin: 40px auto; padding: 0 20px; }
			.checkout-grid { display: grid; grid-template-columns: 1fr 1.2fr; gap: 25px; align-items: start; }
			
			.card { background: white; border-radius: 16px; padding: 25px; box-shadow: 0 4px 15px rgba(0,0,0,0.05); border: 1px solid #e2e8f0; }
			.card-title { font-size: 20px; color: #03045e; margin-top: 0; margin-bottom: 20px; font-weight: 700; }
			
			.item-box { display: flex; gap: 15px; align-items: center; padding-bottom: 20px; border-bottom: 1px solid #e2e8f0; margin-bottom: 20px; }
			.item-img { width: 85px; height: 85px; object-fit: contain; background: #fafafa; border-radius: 10px; border: 1px solid #cbd5e1; padding: 5px; }
			.item-info h3 { margin: 0 0 5px 0; font-size: 18px; color: #023e8a; }
			.item-price-unit { color: #64748b; font-size: 14px; margin-bottom: 8px; }
			
			.qty-control { display: flex; align-items: center; gap: 10px; font-size: 13px; color: #475569; }
			.qty-input { width: 60px; padding: 6px; border: 1px solid #cbd5e1; border-radius: 6px; text-align: center; font-size: 14px; font-weight: bold; }
			.stock-text { color: #94a3b8; font-size: 12px; }

			.summary-row { display: flex; justify-content: space-between; font-size: 14px; color: #475569; margin-bottom: 10px; }
			.summary-total { display: flex; justify-content: space-between; font-size: 18px; font-weight: bold; color: #03045e; margin-top: 15px; padding-top: 15px; border-top: 2px dashed #cbd5e1; }
			.free-shipping { color: #2a9d8f; font-weight: bold; }

			.form-group { margin-bottom: 16px; }
			.form-group label { display: block; margin-bottom: 6px; font-weight: 600; font-size: 13px; color: #023e8a; }
			.form-group input, .form-group textarea, .form-group select { width: 100%; padding: 10px 12px; border: 1px solid #cbd5e1; border-radius: 8px; font-size: 14px; outline: none; background-color: #f8fafc; transition: 0.2s; }
			.form-group input:focus, .form-group textarea:focus, .form-group select:focus { background-color: #fff; border-color: #0096c7; }
			
			.btn-submit { width: 100%; background-color: #2b6cb0; color: white; border: none; padding: 14px; font-size: 16px; font-weight: bold; border-radius: 25px; cursor: pointer; transition: 0.2s; margin-top: 10px; }
			.btn-submit:hover { background-color: #2c5282; }
			.btn-cancel { display: block; text-align: center; color: #e53e3e; text-decoration: none; font-size: 13px; font-weight: 600; margin-top: 12px; }
		</style>
	</head>
	<body>
		<div class="navbar">
			<h1>Art Toy Shop</h1>
			<div>
				<a href="/products">กลับหน้าร้านค้า</a>
			</div>
		</div>

		<div class="container">
			<form method="POST" action="/order/buy">
				<input type="hidden" name="product_id" value="` + strconv.Itoa(p.ID) + `">

				<div class="checkout-grid">
					<!-- คอลัมน์ซ้าย: สรุปรายการสั่งซื้อ -->
					<div class="card">
						<h2 class="card-title">สรุปรายการสั่งซื้อ</h2>
						
						<div class="item-box">
							<img src="` + p.ImageURL + `" alt="` + p.Name + `" class="item-img">
							<div class="item-info">
								<h3>` + p.Name + `</h3>
								<div class="item-price-unit">ราคาชิ้นละ: ฿` + fmt.Sprintf("%.2f", unitPrice) + `</div>
								<div class="qty-control">
									<label>จำนวนที่ต้องการ:</label>
									<input type="number" id="qtyInput" name="quantity" value="` + strconv.Itoa(quantity) + `" min="1" max="` + strconv.Itoa(p.Stock) + `" class="qty-input" required onchange="updateTotal(this.value)">
									<span class="stock-text">(คงเหลือ ` + strconv.Itoa(p.Stock) + ` ชิ้น)</span>
								</div>
							</div>
						</div>

						<div class="summary-row">
							<span>ราคาสินค้ารวม</span>
							<span id="subtotalDisplay">฿` + fmt.Sprintf("%.2f", totalPrice) + `</span>
						</div>
						<div class="summary-row">
							<span>ค่าจัดส่ง</span>
							<span class="free-shipping">ฟรี</span>
						</div>

						<div class="summary-total">
							<span>ยอดชำระสุทธิ</span>
							<span id="totalDisplay" style="color:#0077b6;">฿` + fmt.Sprintf("%.2f", totalPrice) + `</span>
						</div>
					</div>

					<!-- คอลัมน์ขวา: ข้อมูลการจัดส่ง & ชำระเงิน -->
					<div class="card">
						<h2 class="card-title">ข้อมูลการจัดส่ง & ชำระเงิน</h2>

						<div class="form-group">
							<label>ชื่อ-นามสกุล ผู้รับ</label>
							<input type="text" name="fullname" placeholder="ระบุชื่อ-นามสกุล" required>
						</div>

						<div class="form-group">
							<label>เบอร์โทรศัพท์</label>
							<input type="tel" name="phone" placeholder="08X-XXX-XXXX" required>
						</div>

						<div class="form-group">
							<label>ที่อยู่จัดส่งสินค้า</label>
							<textarea name="address" rows="3" placeholder="ระบุบ้านเลขที่ ซอย ถนน ตำบล/แขวง อำเภอ/เขต จังหวัด รหัสไปรษณีย์" required></textarea>
						</div>

						<div class="form-group">
							<label>หมายเหตุถึงผู้ขาย (ไม่บังคับ)</label>
							<input type="text" name="note" placeholder="เช่น ฝากเขียนการ์ดอวยพร, ห่อของขวัญ, ระบุช่วงเวลาจัดส่ง">
						</div>

						<div class="form-group">
							<label>ช่องทางการชำระเงิน</label>
							<select name="payment_method">
								<option value="promptpay">โอนผ่านธนาคาร / PromptPay</option>
								<option value="cod">เก็บเงินปลายทาง (COD)</option>
							</select>
						</div>

						<button type="submit" class="btn-submit">ยืนยันและชำระเงิน</button>
						<a href="/products" class="btn-cancel">ยกเลิกรายการ</a>
					</div>
				</div>
			</form>
		</div>

		<script>
			const unitPrice = ` + fmt.Sprintf("%.2f", unitPrice) + `;
			function updateTotal(qty) {
				const q = parseInt(qty) || 1;
				const total = unitPrice * q;
				const formattedTotal = '฿' + total.toFixed(2);
				document.getElementById('subtotalDisplay').innerText = formattedTotal;
				document.getElementById('totalDisplay').innerText = formattedTotal;
			}
		</script>
	</body>
	</html>
	`))
	}
}
