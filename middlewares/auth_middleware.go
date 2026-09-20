package middlewares

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var JwtSecret = []byte(os.Getenv("JWT_SECRET"))

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Pega o cabeçalho Authorization da requisição
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token de autenticacao nao fornecido",
			})
		}

		// 2. O formato padrão é "Bearer <token>", então separamos por espaço
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Formato de token invalido. Use: Bearer <token>",
			})
		}

		tokenString := parts[1]

		// 3. Valida e desserializa o token usando a chave secreta
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Garante que o método de assinatura é o mesmo (HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return JwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token invalido ou expirado",
			})
		}

		// 4. Extrai as claims e guarda o user_id no contexto do Fiber (c.Locals)
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// No login nós salvamos o ID no campo "sub"
			if userID, ok := claims["sub"].(float64); ok {
				c.Locals("user_id", uint(userID))
			}
		}

		// 5. Se tudo estiver ok, deixa a requisição seguir para o handler final
		return c.Next()
	}
}
