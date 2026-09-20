package handlers

import (
	"lea/database"
	"lea/entities"

	"github.com/gofiber/fiber/v2"
)

func GetUserFolowers(c *fiber.Ctx) error {
	userID := c.Params("id")
	var folowers []entities.Follower

	result := database.DB.Preload("Folower").Where("user_id = ?", userID).Find(&folowers)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao buscar seguidores",
		})
	}
	return c.JSON(folowers)
}
