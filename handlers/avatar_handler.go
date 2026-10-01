package handlers

import (
	"lea/database"
	"lea/entities"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func uploadURL(name string) string {
	if name == "" {
		return ""
	}
	return "/uploads/" + name
}

func UploadAvatar(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}

	name, err := savePhoto(c) // campo "photo", mesma validação dos posts
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao salvar foto"})
	}
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "envie uma imagem no campo photo"})
	}

	var user entities.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		os.Remove(filepath.Join(uploadDir, name))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "usuario nao encontrado"})
	}

	old := user.Avatar
	if err := database.DB.Model(&user).Update("avatar", name).Error; err != nil {
		os.Remove(filepath.Join(uploadDir, name))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao salvar avatar"})
	}
	if old != "" {
		os.Remove(filepath.Join(uploadDir, filepath.Base(old))) // apaga a foto antiga
	}

	return c.JSON(fiber.Map{"avatar_url": uploadURL(name)})
}
