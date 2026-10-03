package handlers

import (
	"lea/database"
	"lea/entities"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func DeleteGroup(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}
	groupID, err := c.ParamsInt("id")
	if err != nil || groupID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id invalido"})
	}

	var group entities.Group
	if err := database.DB.First(&group, groupID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "grupo nao existe"})
	}
	if group.OwnerID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "so o dono pode excluir o grupo"})
	}

	// guarda as fotos dos posts pra apagar do disco depois
	var photos []string
	database.DB.Model(&entities.Post{}).
		Where("group_id = ? AND photo <> ''", groupID).Pluck("photo", &photos)

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var postIDs []uint
		if err := tx.Model(&entities.Post{}).Where("group_id = ?", groupID).Pluck("id", &postIDs).Error; err != nil {
			return err
		}
		if len(postIDs) > 0 {
			if err := tx.Unscoped().Where("post_id IN ?", postIDs).Delete(&entities.Comment{}).Error; err != nil {
				return err
			}
			// mensagens que compartilharam esses posts ficam sem o card
			if err := tx.Model(&entities.Message{}).Where("post_id IN ?", postIDs).
				Update("post_id", nil).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Where("id IN ?", postIDs).Delete(&entities.Post{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Unscoped().Where("group_id = ?", groupID).Delete(&entities.GroupMember{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&group).Error
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao excluir grupo"})
	}

	// só apaga os arquivos depois que o banco confirmou
	for _, name := range photos {
		os.Remove(filepath.Join(uploadDir, filepath.Base(name)))
	}
	return c.SendStatus(fiber.StatusNoContent)
}
