package app

import (
	"os"

	"github.com/biosecret/go-iot/config"
	"github.com/biosecret/go-iot/database"
	"github.com/biosecret/go-iot/handlers"
	"github.com/biosecret/go-iot/router"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SetupAndRunApp khởi động ứng dụng Fiber
func SetupAndRunApp() error {
	// Load biến môi trường từ file .env
	err := config.LoadENV()
	if err != nil {
		return err
	}

	// Khởi động PostgreSQL
	err = database.StartPostgreSQL()
	if err != nil {
		return err
	}

	// Đảm bảo kết nối với cơ sở dữ liệu được đóng sau khi ứng dụng kết thúc
	defer database.ClosePostgreSQL()

	// Tạo ứng dụng Fiber
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",                           // Cho phép tất cả các nguồn (có thể điều chỉnh)
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS", // Các phương thức được phép
	}))

	// Đính kèm middleware để xử lý lỗi và ghi log
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path} ${latency}\n",
	}))

	// Thiết lập route cho ứng dụng
	router.SetupRoutes(app)

	// Đính kèm Swagger (nếu cần)
	config.AddSwaggerRoutes(app)

	go handlers.InitMQTTSubscriber()

	// Lấy port từ biến môi trường và bắt đầu lắng nghe kết nối
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Giá trị mặc định nếu PORT không được thiết lập
	}

	// Lắng nghe trên cổng chỉ định
	return app.Listen(":" + port)
}
