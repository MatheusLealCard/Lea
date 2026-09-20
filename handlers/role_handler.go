package handlers

import (
	"lea/database"
	"lea/entities"

	"github.com/gofiber/fiber/v2"
)

func GetRoles(c *fiber.Ctx) error {
	var roles []entities.Role
	result := database.DB.Find(&roles)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Nao foi possivel buscar roles",
		})
	}

	return c.JSON(roles)
}
