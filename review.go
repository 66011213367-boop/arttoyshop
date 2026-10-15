package main

import (
	"database/sql"
	"html"
	"net/http"
	"strconv"
)

// เพิ่มรีวิวใหม่
func AddReviewHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		userIDCookie, err := r.Cookie("user_id")
		if err != nil || userIDCookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userID, _ := strconv.Atoi(userIDCookie.Value)
		productID, _ := strconv.Atoi(r.FormValue("product_id"))
		comment := html.EscapeString(r.FormValue("comment"))
		rating, _ := strconv.Atoi(r.FormValue("rating"))

		_, _ = db.Exec("INSERT INTO reviews (product_id, user_id, comment, rating) VALUES (?, ?, ?, ?)",
			productID, userID, comment, rating)

		http.Redirect(w, r, "/products", http.StatusSeeOther)
	}
}

// ลบรีวิว + ตรวจสอบ Ownership Control
func DeleteReviewHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		reviewID := r.FormValue("review_id")
		userIDCookie, err := r.Cookie("user_id")
		userRoleCookie, _ := r.Cookie("user_role")

		if err != nil || userIDCookie.Value == "" {
			http.Error(w, "กรุณาเข้าสู่ระบบก่อนทำรายการ", http.StatusUnauthorized)
			return
		}

		currentUserID, _ := strconv.Atoi(userIDCookie.Value)

		var ownerID int
		err = db.QueryRow("SELECT user_id FROM reviews WHERE id = ?", reviewID).Scan(&ownerID)
		if err != nil {
			http.Error(w, "ไม่พบข้อมูลรีวิว", http.StatusNotFound)
			return
		}

		isAdmin := userRoleCookie != nil && userRoleCookie.Value == "admin"
		if ownerID != currentUserID && !isAdmin {
			http.Error(w, "403 Forbidden: คุณไม่มีสิทธิ์ลบข้อมูลของผู้ใช้อื่น", http.StatusForbidden)
			return
		}

		_, _ = db.Exec("DELETE FROM reviews WHERE id = ?", reviewID)
		http.Redirect(w, r, "/products", http.StatusSeeOther)
	}
}

// แก้ไขรีวิว + ตรวจสอบ Ownership Control
func UpdateReviewHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		reviewID := r.FormValue("review_id")
		newComment := html.EscapeString(r.FormValue("comment"))
		newRatingStr := r.FormValue("rating")
		userIDCookie, err := r.Cookie("user_id")

		if err != nil || userIDCookie.Value == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		currentUserID, _ := strconv.Atoi(userIDCookie.Value)
		newRating, _ := strconv.Atoi(newRatingStr)

		var ownerID int
		err = db.QueryRow("SELECT user_id FROM reviews WHERE id = ?", reviewID).Scan(&ownerID)
		if err != nil || ownerID != currentUserID {
			http.Error(w, "403 Forbidden: คุณไม่มีสิทธิ์แก้ไขข้อมูลของผู้ใช้อื่น", http.StatusForbidden)
			return
		}

		_, _ = db.Exec("UPDATE reviews SET comment = ?, rating = ? WHERE id = ?", newComment, newRating, reviewID)
		http.Redirect(w, r, "/products", http.StatusSeeOther)
	}
}
