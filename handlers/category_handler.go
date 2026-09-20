package handlers

import (
	"lea/database"
	"lea/entities"

	"github.com/gofiber/fiber/v2"
)

func GetCategories(c *fiber.Ctx) error {
	var categories []entities.Category

	result := database.DB.Find(&categories)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao buscar categorias",
		})
	}
	return c.JSON(categories)
}
