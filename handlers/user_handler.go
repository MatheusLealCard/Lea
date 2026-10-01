package handlers

import (
	"lea/database"
	"lea/entities"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// lê na hora de usar, depois do godotenv.Load()
func jwtSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserInput struct {
	Name     string `json:"name"`
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
	var input CreateUserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dados invalidos"})
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Name == "" || input.Email == "" || len(input.Password) < 4 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Preencha nome, email e senha (min. 4 caracteres)",
		})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao processar senha"})
	}

	// sempre nasce com a role "User", ninguém se cadastra como admin
	var role entities.Role
	if err := database.DB.Where("role_name = ?", "User").First(&role).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Role padrao nao encontrada"})
	}

	user := entities.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hash),
		RoleID:   role.ID,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Erro ao criar usuario (email ja existe?)"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}

func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dados invalidos"})
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))

	var user entities.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Email ou senha incorretos"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Email ou senha incorretos"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	signedToken, err := token.SignedString(jwtSecret())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao gerar token"})
	}

	return c.JSON(fiber.Map{"token": signedToken})
}
