package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
)

type MyOrderItem struct {
	ID            int
	ProductName   string
	ImageURL      string
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

func UserOrdersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		userIDCookie, err := r.Cookie("user_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID, _ := strconv.Atoi(userIDCookie.Value)

		var username string
		_ = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

		rows, err := db.Query(`
			SELECT o.id, IFNULL(p.name, 'สินค้าถูกลบไปแล้ว'), IFNULL(p.image_url, '/images/labobu.png'),
			       o.quantity, o.total_price, o.status, 
			       IFNULL(o.fullname, ''), IFNULL(o.phone, ''), IFNULL(o.address, ''), IFNULL(o.payment_method, 'promptpay'), IFNULL(o.note, ''), o.created_at
			FROM orders o
			LEFT JOIN products p ON o.product_id = p.id
			WHERE o.user_id = ?
			ORDER BY o.id DESC`, userID)

		var ordersHTML string
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var item MyOrderItem
				_ = rows.Scan(&item.ID, &item.ProductName, &item.ImageURL, &item.Quantity, &item.TotalPrice, &item.Status,
					&item.Fullname, &item.Phone, &item.Address, &item.PaymentMethod, &item.Note, &item.CreatedAt)

				statusBadge := `<span class="badge badge-pending">รอดำเนินการ</span>`
				if item.Status == "paid" || item.Status == "ชำระเงินแล้ว" {
					statusBadge = `<span class="badge badge-paid">ชำระเงินแล้ว</span>`
				} else if item.Status == "shipped" || item.Status == "จัดส่งแล้ว" {
					statusBadge = `<span class="badge badge-shipped">จัดส่งเรียบร้อย</span>`
				} else if item.Status == "cancelled" || item.Status == "ยกเลิก" {
					statusBadge = `<span class="badge badge-cancelled">ยกเลิกคำสั่งซื้อ</span>`
				}

				payText := "โอนผ่านธนาคาร / PromptPay"
				if item.PaymentMethod == "cod" {
					payText = "เก็บเงินปลายทาง (COD)"
				}

				noteHTML := ""
				if item.Note != "" {
					noteHTML = `<p>หมายเหตุ: <span style="color:#0284c7;">` + item.Note + `</span></p>`
				}

				ordersHTML += `
				<div class="order-card">
					<div class="order-header">
						<span>หมายเลขคำสั่งซื้อ: #` + strconv.Itoa(item.ID) + `</span>
						<span>วันที่: ` + item.CreatedAt + `</span>
					</div>
					<div class="order-body">
						<img src="` + item.ImageURL + `" alt="` + item.ProductName + `" class="order-img">
						<div class="order-details">
							<h3 class="order-title">` + item.ProductName + `</h3>
							<p>จำนวน: <strong>` + strconv.Itoa(item.Quantity) + `</strong> ชิ้น</p>
							<p>ช่องทางชำระเงิน: ` + payText + `</p>
							<p>ที่อยู่จัดส่ง: ` + item.Fullname + ` (` + item.Phone + `) ` + item.Address + `</p>
							` + noteHTML + `
						</div>
						<div class="order-summary">
							<div class="order-price">฿` + fmt.Sprintf("%.2f", item.TotalPrice) + `</div>
							<div style="margin-top:10px;">` + statusBadge + `</div>
						</div>
					</div>
				</div>`
			}
		}

		if ordersHTML == "" {
			ordersHTML = `
			<div style="text-align:center; padding:50px 20px; background:white; border-radius:12px; border:1px solid #e0e6ed;">
				<h3 style="color:#64748b; margin-bottom:10px;">คุณยังไม่มีรายการสั่งซื้อ</h3>
				<a href="/products" style="display:inline-block; background:#0096c7; color:white; padding:10px 20px; border-radius:20px; text-decoration:none; font-weight:bold;">ไปเลือกซื้อสินค้าเลย</a>
			</div>`
		}

		_, _ = w.Write([]byte(`
	<!DOCTYPE html>
	<html lang="th">
	<head>
		<meta charset="UTF-8">
		<title>คำสั่งซื้อของฉัน - Art Toy Shop</title>
		<style>
			* { box-sizing: border-box; }
			body { font-family: 'Segoe UI', sans-serif; background-color: #f0f4f8; margin: 0; padding: 0; color: #333; }
			.navbar { background-color: #0077b6; padding: 15px 30px; display: flex; justify-content: space-between; align-items: center; }
			.navbar h1 { color: white; margin: 0; font-size: 24px; }
			.nav-right { display: flex; align-items: center; gap: 10px; }
			.navbar a { color: white; text-decoration: none; font-weight: 600; padding: 8px 16px; border-radius: 20px; background: rgba(255,255,255,0.18); font-size: 14px; }
			.user-badge { color: white; font-weight: bold; background: rgba(255,255,255,0.25); padding: 6px 14px; border-radius: 15px; font-size: 13px; }
			
			.container { max-width: 900px; margin: 30px auto; padding: 0 20px; }
			.page-title { font-size: 24px; color: #03045e; margin-bottom: 20px; font-weight: 700; }
			
			.order-card { background: white; border-radius: 12px; border: 1px solid #e0e6ed; box-shadow: 0 4px 12px rgba(0,0,0,0.04); margin-bottom: 20px; overflow: hidden; }
			.order-header { background: #f8fafc; padding: 12px 20px; border-bottom: 1px solid #edf2f7; display: flex; justify-content: space-between; font-size: 13px; color: #64748b; font-weight: 600; }
			.order-body { padding: 20px; display: flex; gap: 20px; align-items: center; }
			.order-img { width: 90px; height: 90px; object-fit: contain; background: #fafafa; border-radius: 8px; border: 1px solid #e2e8f0; padding: 5px; }
			.order-details { flex: 1; }
			.order-title { margin: 0 0 8px 0; font-size: 18px; color: #023e8a; }
			.order-details p { margin: 4px 0; font-size: 13px; color: #475569; }
			.order-summary { text-align: right; min-width: 140px; }
			.order-price { font-size: 20px; font-weight: bold; color: #0077b6; }
			
			.badge { padding: 6px 12px; border-radius: 15px; font-size: 12px; font-weight: bold; display: inline-block; }
			.badge-pending { background-color: #fef08a; color: #854d0e; }
			.badge-paid { background-color: #bbf7d0; color: #166534; }
			.badge-shipped { background-color: #bfdbfe; color: #1e40af; }
			.badge-cancelled { background-color: #fecaca; color: #991b1b; }
		</style>
	</head>
	<body>
		<div class="navbar">
			<h1>Art Toy Shop</h1>
			<div class="nav-right">
				<a href="/">หน้าแรก</a>
				<a href="/products">หน้าร้านค้า</a>
				<span class="user-badge">ผู้ใช้งาน: ` + username + `</span>
				<a href="/logout" style="background:#d62828;">ออกจากระบบ</a>
			</div>
		</div>

		<div class="container">
			<h2 class="page-title">รายการคำสั่งซื้อของฉัน (My Orders)</h2>
			` + ordersHTML + `
		</div>
	</body>
	</html>
	`))
	}
}
