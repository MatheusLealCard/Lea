package handlers

import (
	"lea/database"
	"lea/entities"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var JwtSecret = []byte(os.Getenv("JWT_SECRET"))

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

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

func Login(c *fiber.Ctx) error {
	var input LoginInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dados invalidos"})
	}

	var user entities.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		println("ERRO: Usuário não encontrado no banco:", err.Error()) // DEBUG
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Email ou senha incorretos"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		println("ERRO: Senha não bate com o hash:", err.Error()) // DEBUG
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Email ou senha incorretos"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	signedToken, err := token.SignedString(JwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao gerar token",
		})
	}

	return c.JSON(fiber.Map{
		"token": signedToken,
	})
}
