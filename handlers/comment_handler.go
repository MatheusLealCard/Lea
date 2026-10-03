package handlers

import (
	"lea/database"
	"lea/entities"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
)

const maxCommentLen = 1000

type CreateCommentInput struct {
	Content string `json:"content"`
}

type CommentDTO struct {
	ID              uint      `json:"id"`
	PostID          uint      `json:"post_id"`
	AuthorID        uint      `json:"author_id"`
	AuthorName      string    `json:"author_name"`
	AuthorAvatarURL string    `json:"author_avatar_url"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
	CanDelete       bool      `json:"can_delete"`
}

func toCommentDTO(cm entities.Comment, viewerID uint) CommentDTO {
	return CommentDTO{
		ID:              cm.ID,
		PostID:          cm.PostID,
		AuthorID:        cm.UserID,
		AuthorName:      cm.User.Name,
		AuthorAvatarURL: uploadURL(cm.User.Avatar),
		Content:         cm.Content,
		CreatedAt:       cm.CreatedAt,
		CanDelete:       cm.UserID == viewerID,
	}
}

func GetComments(c *fiber.Ctx) error {
	viewerID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}
	postID, err := c.ParamsInt("id")
	if err != nil || postID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id invalido"})
	}

	var count int64
	database.DB.Model(&entities.Post{}).Where("id = ?", postID).Count(&count)
	if count == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post nao encontrado"})
	}

	var comments []entities.Comment
	err = database.DB.Preload("User").Where("post_id = ?", postID).
		Order("created_at asc, id asc").Find(&comments).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar comentarios"})
	}

	out := make([]CommentDTO, 0, len(comments))
	for _, cm := range comments {
		out = append(out, toCommentDTO(cm, viewerID))
	}
	return c.JSON(out)
}

func CreateComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}
	postID, err := c.ParamsInt("id")
	if err != nil || postID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id invalido"})
	}

	var input CreateCommentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "dados invalidos"})
	}
	content := strings.TrimSpace(input.Content)
	if n := utf8.RuneCountInString(content); n == 0 || n > maxCommentLen {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "o comentario deve ter entre 1 e 1000 caracteres",
		})
	}

	var count int64
	database.DB.Model(&entities.Post{}).Where("id = ?", postID).Count(&count)
	if count == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post nao encontrado"})
	}

	cm := entities.Comment{PostID: uint(postID), UserID: userID, Content: content}
	if err := database.DB.Create(&cm).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao comentar"})
	}

	database.DB.Preload("User").First(&cm, cm.ID)
	return c.Status(fiber.StatusCreated).JSON(toCommentDTO(cm, userID))
}

func DeleteComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id invalido"})
	}

	var cm entities.Comment
	if err := database.DB.First(&cm, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "comentario nao encontrado"})
	}
	if cm.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "so o autor pode excluir o comentario"})
	}

	if err := database.DB.Delete(&cm).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao excluir comentario"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
