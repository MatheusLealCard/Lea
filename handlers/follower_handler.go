package handlers

import (
	"lea/database"
	"lea/entities"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm/clause"
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

type ProfileDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	AvatarURL   string `json:"avatar_url"`
	Followers   int64  `json:"followers"`
	Following   int64  `json:"following"`
	IsFollowing bool   `json:"is_following"`
	IsMe        bool   `json:"is_me"`
}

func GetProfile(c *fiber.Ctx) error {
	viewerID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}
	targetID, err := c.ParamsInt("id")
	if err != nil || targetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id invalido"})
	}

	var user entities.User
	if err := database.DB.First(&user, targetID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "usuario nao encontrado"})
	}

	// UserID = quem é seguido, FolowerID = quem segue
	var followers, following, rel int64
	database.DB.Model(&entities.Follower{}).Where("user_id = ?", targetID).Count(&followers)
	database.DB.Model(&entities.Follower{}).Where("folower_id = ?", targetID).Count(&following)
	database.DB.Model(&entities.Follower{}).
		Where("user_id = ? AND folower_id = ?", targetID, viewerID).Count(&rel)

	dto := ProfileDTO{
		ID:          user.ID,
		Name:        user.Name,
		AvatarURL:   uploadURL(user.Avatar),
		Followers:   followers,
		Following:   following,
		IsFollowing: rel > 0,
		IsMe:        viewerID == user.ID,
	}
	if dto.IsMe {
		dto.Email = user.Email
	}
	return c.JSON(dto)
}

func FollowUser(c *fiber.Ctx) error {
	viewerID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}
	targetID, err := c.ParamsInt("id")
	if err != nil || targetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id invalido"})
	}
	if uint(targetID) == viewerID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "voce nao pode seguir a si mesmo"})
	}

	var count int64
	database.DB.Model(&entities.User{}).Where("id = ?", targetID).Count(&count)
	if count == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "usuario nao existe"})
	}

	f := entities.Follower{UserID: uint(targetID), FolowerID: viewerID}
	// se já segue, não faz nada (sem erro)
	if err := database.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&f).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao seguir"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func UnfollowUser(c *fiber.Ctx) error {
	viewerID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}
	targetID, err := c.ParamsInt("id")
	if err != nil || targetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id invalido"})
	}

	// Unscoped = apaga de verdade. Com soft delete a linha continuaria
	// no banco e bloquearia seguir de novo (índice único user+follower).
	err = database.DB.Unscoped().
		Where("user_id = ? AND folower_id = ?", targetID, viewerID).
		Delete(&entities.Follower{}).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao deixar de seguir"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
