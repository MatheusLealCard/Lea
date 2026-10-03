package handlers

import (
	"lea/database"
	"lea/entities"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func DeletePost(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id invalido"})
	}

	var post entities.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post nao encontrado"})
	}

	allowed := post.UserID == userID
	// o dono do grupo também pode apagar posts do grupo (tira este bloco se não quiser)
	if !allowed && post.GroupID != nil {
		var owner int64
		database.DB.Model(&entities.Group{}).
			Where("id = ? AND owner_id = ?", *post.GroupID, userID).Count(&owner)
		allowed = owner > 0
	}
	if !allowed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "voce nao pode excluir esse post"})
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("post_id = ?", post.ID).Delete(&entities.Comment{}).Error; err != nil {
			return err
		}
		// mensagens que compartilharam o post ficam sem o card
		if err := tx.Model(&entities.Message{}).Where("post_id = ?", post.ID).
			Update("post_id", nil).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&post).Error
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao excluir post"})
	}

	// só apaga a foto depois que o banco confirmou
	if post.Photo != "" {
		os.Remove(filepath.Join(uploadDir, filepath.Base(post.Photo)))
	}
	return c.SendStatus(fiber.StatusNoContent)
}
