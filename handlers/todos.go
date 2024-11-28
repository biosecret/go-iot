package handlers

import (
	"database/sql"
	"time"

	"github.com/biosecret/go-iot/database"
	"github.com/biosecret/go-iot/models"
	"github.com/biosecret/go-iot/utils"
	"github.com/gofiber/fiber/v2"
)

// Lấy tất cả Todos
func HandleAllTodos(c *fiber.Ctx) error {
	rows, err := database.GetDB().Query(
		"SELECT id, title, completed, description, date, updated_at FROM todos ORDER BY updated_at DESC",
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	todos := []models.Todo{}
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.Description, &todo.Date, &todo.UpdatedAt); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		todos = append(todos, todo)
	}

	return c.Status(200).JSON(todos)
}

// Tạo mới một Todo
func HandleCreateTodo(c *fiber.Ctx) error {
	nTodo := new(models.Todo)
	if err := c.BodyParser(nTodo); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Thử tạo ID tối đa 3 lần nếu ID bị trùng
	var id string
	var err error
	for i := 0; i < 3; i++ {
		id, err = utils.GenerateRandomID()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to generate ID"})
		}

		// Kiểm tra ID có tồn tại trong database không
		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM todos WHERE id=$1)"
		if err := database.GetDB().QueryRow(query, id).Scan(&exists); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Nếu ID chưa tồn tại, thoát khỏi vòng lặp
		if !exists {
			break
		}

		// Nếu đã thử 3 lần và vẫn trùng, trả về lỗi
		if i == 2 {
			return c.Status(500).JSON(fiber.Map{"error": "failed to generate a unique ID"})
		}
	}

	// Gán ID ngẫu nhiên cho Todo
	nTodo.ID = id

	// Chèn Todo vào database
	query := "INSERT INTO todos (id, title, completed, description, date) VALUES ($1, $2, $3, $4, $5)"
	_, err = database.GetDB().Exec(query, nTodo.ID, nTodo.Title, nTodo.Completed, nTodo.Description, nTodo.Date)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"inserted_id": nTodo.ID})
}

// Lấy một Todo theo ID
func HandleGetOneTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	var todo models.Todo
	err := database.GetDB().QueryRow(
		"SELECT id, title, completed, description, date FROM todos WHERE id = $1", id,
	).Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.Description, &todo.Date)

	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	} else if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(200).JSON(todo)
}

// Cập nhật một Todo
func HandleUpdateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	uTodo := new(models.Todo)

	// Parse request body
	if err := c.BodyParser(uTodo); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Cập nhật Todo và lưu thời gian cập nhật hiện tại
	res, err := database.GetDB().Exec(
		"UPDATE todos SET title=$1, completed=$2, description=$3, date=$4, updated_at=$5 WHERE id=$6",
		uTodo.Title, uTodo.Completed, uTodo.Description, uTodo.Date, time.Now(), id,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Lấy số hàng bị ảnh hưởng
	count, _ := res.RowsAffected()

	return c.Status(200).JSON(fiber.Map{"updated_count": count})
}

// Xóa một Todo
func HandleDeleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	res, err := database.GetDB().Exec("DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	count, _ := res.RowsAffected()
	return c.Status(200).JSON(fiber.Map{"deleted_count": count})
}
