package main

import (
	"context"
	"mime/multipart"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// UploadToCloudinary รับไฟล์จาก Form แล้วส่งไป Cloudinary
func UploadToCloudinary(fileHeader *multipart.FileHeader) (string, error) {
	// ดึงค่า Credentials จาก Environment Variables ที่เราตั้งไว้บน Render
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return "", err
	}

	// เปิดอ่านไฟล์
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// อัปโหลดไฟล์ขึ้น Cloudinary
	ctx := context.Background()
	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "products", // ชื่อโฟลเดอร์ใน Cloudinary
	})
	if err != nil {
		return "", err
	}

	// ส่งคืน URL รูปภาพถาวร (SecureURL)
	return uploadResult.SecureURL, nil
}
