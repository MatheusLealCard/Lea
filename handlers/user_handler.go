package handlers

import (
	"lea/database"
	"lea/entities"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(c *fiber.Ctx) error {
	var users []entities.User
	result := database.DB.Find(&users)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao buscar usuarios",
		})
	}
	return c.JSON(users)
}

func CreateUser(c *fiber.Ctx) error {
	var user entities.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Dados invalidos",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao processar senha",
		})
	}

	user.Password = string(hashedPassword)

	result := database.DB.Create(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao criar usuario",
		})
	}

	user.Password = ""
	return c.Status(fiber.StatusCreated).JSON(user)
}
