package main

import (
	"fmt"
	"net/http"
	"os"
)

// Security Middleware ป้องกัน Clickjacking และ MIME-sniffing
func securityHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		next.ServeHTTP(w, r)
	})
}

func main() {
	db := ConnectDB()
	defer func() {
		db.Close()
		fmt.Println("DB Closed & Memory Released.")
	}()

	mux := http.NewServeMux()

	// Static Files (Images)
	fs := http.FileServer(http.Dir("./images"))
	mux.Handle("/images/", http.StripPrefix("/images/", fs))

	// Auth Routes
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			RegisterPageHandler(w, r)
		} else if r.Method == http.MethodPost {
			RegisterProcessHandler(db)(w, r)
		}
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			LoginPageHandler(w, r)
		} else if r.Method == http.MethodPost {
			LoginProcessHandler(db)(w, r)
		}
	})

	mux.HandleFunc("/logout", LogoutHandler)

	// User Routes
	mux.HandleFunc("/products", ProductsHandler(db))
	mux.HandleFunc("/checkout", CheckoutPageHandler(db))
	mux.HandleFunc("/order/buy", BuyProductHandler(db))
	mux.HandleFunc("/my-orders", UserOrdersHandler(db))

	// Review Routes (รองรับ Add, Update, Delete)
	mux.HandleFunc("/review/add", AddReviewHandler(db))
	mux.HandleFunc("/review/update", UpdateReviewHandler(db))
	mux.HandleFunc("/review/delete", DeleteReviewHandler(db))

	mux.HandleFunc("/", HomePageHandler)

	// Admin Routes
	mux.HandleFunc("/admin", AdminHandler(db))
	mux.HandleFunc("/admin/add-product", AddProductHandler(db))
	mux.HandleFunc("/admin/update-product", UpdateProductHandler(db))
	mux.HandleFunc("/admin/delete-product", DeleteProductHandler(db))
	mux.HandleFunc("/admin/update-order-status", UpdateOrderStatusHandler(db))
	mux.HandleFunc("/admin/delete-order", DeleteOrderHandler(db))
	mux.HandleFunc("/admin/delete-user", DeleteUserHandler(db))

	secureHandler := securityHeaderMiddleware(mux)

	// อ่านค่า PORT จาก Environment Variable บน Cloud (ถ้าไม่มีให้ใช้ 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port:", port)
	_ = http.ListenAndServe(":"+port, secureHandler)
}

func HomePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`
	<!DOCTYPE html>
	<html lang="th">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Art Toy Shop</title>
		<style>
			* { box-sizing: border-box; }
			body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f0f4f8; margin: 0; padding: 0; color: #333; }
			
			.navbar { background-color: #0077b6; padding: 15px 30px; display: flex; justify-content: space-between; align-items: center; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
			.navbar h1 { color: white; margin: 0; font-size: 24px; font-weight: 700; }
			.navbar a { color: white; text-decoration: none; margin-left: 10px; font-weight: 600; padding: 8px 18px; border-radius: 20px; background: rgba(255,255,255,0.18); transition: 0.2s; font-size: 14px; }
			.navbar a:hover { background: white; color: #0077b6; }
			
			.container { max-width: 1100px; margin: 0 auto; padding: 20px; }

			.hero { text-align: center; padding: 50px 20px; background: linear-gradient(135deg, #0096c7, #03045e); color: white; border-radius: 16px; margin-bottom: 40px; box-shadow: 0 6px 15px rgba(0,119,182,0.2); }
			.hero h2 { font-size: 32px; margin-bottom: 10px; }
			.hero p { font-size: 16px; opacity: 0.9; margin-bottom: 25px; }
			.btn-primary { display: inline-block; background-color: #48cae4; color: #03045e; padding: 12px 30px; text-decoration: none; border-radius: 25px; font-weight: bold; transition: 0.2s; }
			.btn-primary:hover { background-color: #90e0ef; transform: translateY(-2px); }

			.section-title { font-size: 22px; color: #03045e; margin-bottom: 20px; text-align: center; font-weight: 700; }
			.popular-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 20px; margin-bottom: 30px; }
			.popular-card { background: white; border-radius: 12px; padding: 15px; text-align: center; box-shadow: 0 4px 12px rgba(0,0,0,0.05); transition: 0.2s; border: 1px solid #e0e6ed; display: flex; flex-direction: column; align-items: center; justify-content: space-between; }
			.popular-card:hover { transform: translateY(-5px); box-shadow: 0 8px 20px rgba(0,0,0,0.1); }
			
			.popular-card img { width: 100%; height: 220px; object-fit: contain; border-radius: 8px; margin-bottom: 12px; background-color: #fafafa; }
			.popular-card h3 { margin: 8px 0 4px; font-size: 18px; color: #023e8a; }
			.popular-card p { font-size: 13px; color: #666; margin: 0; }
		</style>
	</head>
	<body>
		<div class="navbar">
			<h1>Art Toy Shop</h1>
			<div>
				<a href="/products">ดูสินค้าทั้งหมด</a>
				<a href="/login">เข้าสู่ระบบ</a>
				<a href="/register">สมัครสมาชิก</a>
			</div>
		</div>

		<div class="container">
			<div class="hero">
				<h2>ยินดีต้อนรับสู่โลกของ Art Toy</h2>
				<p>รวบรวมกล่องสุ่ม ฟิกเกอร์ยอดฮิต และผลงานจากศิลปินชื่อดัง</p>
				<a href="/products" class="btn-primary">เลือกซื้อสินค้าเลย</a>
			</div>

			<div class="section-title">คอลเลกชันยอดนิยม (Popular Characters)</div>
			<div class="popular-grid">
				<div class="popular-card">
					<img src="/images/labobu.png" alt="LABUBU">
					<div>
						<h3>LABUBU</h3>
						<p>The Monsters Series & Plush Doll</p>
					</div>
				</div>
				<div class="popular-card">
					<img src="/images/molly.jpg" alt="MOLLY">
					<div>
						<h3>MOLLY</h3>
						<p>MEGA SPACE MOLLY & Special Series</p>
					</div>
				</div>
				<div class="popular-card">
					<img src="/images/crybaby.jpg" alt="CRYBABY">
					<div>
						<h3>CRYBABY</h3>
						<p>Crying Parade & Emotional Figures</p>
					</div>
				</div>
				<div class="popular-card">
					<img src="/images/skullpanda.jpg" alt="SKULLPANDA">
					<div>
						<h3>SKULLPANDA</h3>
						<p>Ancient Stories & Art Figures</p>
					</div>
				</div>
			</div>
		</div>
	</body>
	</html>
	`))
}
