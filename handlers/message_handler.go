package handlers

import (
	"encoding/json"
	"lea/database"
	"lea/entities"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	chatCacheSize = 50
	chatCacheTTL  = 36 * time.Hour
)

type CreateMessageInput struct {
	RecieverID uint   `json:"reciever_id"`
	Content    string `json:"content"`
	PostID     *uint  `json:"post_id"`
}

// resumo do post que aparece dentro do chat
type SharedPostDTO struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	Snippet    string `json:"snippet"`
	PhotoURL   string `json:"photo_url"`
	AuthorName string `json:"author_name"`
}

type MessageDTO struct {
	ID         uint           `json:"id"`
	SenderID   uint           `json:"sender_id"`
	RecieverID uint           `json:"reciever_id"`
	Content    string         `json:"content"`
	Post       *SharedPostDTO `json:"post,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

func snippet(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func toDTO(m entities.Message) MessageDTO {
	dto := MessageDTO{
		ID:         m.ID,
		SenderID:   m.SenderID,
		RecieverID: m.RecieverID,
		Content:    m.Content,
		CreatedAt:  m.CreatedAt,
	}
	if m.Post != nil {
		dto.Post = &SharedPostDTO{
			ID:         m.Post.ID,
			Title:      m.Post.Title,
			Snippet:    snippet(m.Post.Content, 120),
			PhotoURL:   uploadURL(m.Post.Photo),
			AuthorName: m.Post.User.Name,
		}
	}
	return dto
}

func CreateMessage(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}

	var input CreateMessageInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "mensagem invalida"})
	}
	input.Content = strings.TrimSpace(input.Content)
	// texto só pode ser vazio se estiver compartilhando um post
	if input.Content == "" && input.PostID == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "mensagem invalida"})
	}
	if input.RecieverID == 0 || input.RecieverID == userID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "destinatario invalido"})
	}

	var count int64
	database.DB.Model(&entities.User{}).Where("id = ?", input.RecieverID).Count(&count)
	if count == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "destinatario nao existe"})
	}

	var shared *entities.Post
	if input.PostID != nil {
		var p entities.Post
		if err := database.DB.Preload("User").First(&p, *input.PostID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "post nao existe"})
		}
		shared = &p
	}

	message := entities.Message{
		SenderID:   userID,
		RecieverID: input.RecieverID,
		Content:    input.Content,
		PostID:     input.PostID,
	}
	if err := database.DB.Create(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao salvar mensagem"})
	}

	message.Post = shared // só pra montar o DTO, depois do Create
	dto := toDTO(message)
	if payload, err := json.Marshal(dto); err == nil {
		ctx := c.Context()
		key := database.ChatKey(userID, input.RecieverID)

		pipe := database.Rdb.Pipeline()
		pipe.LPushX(ctx, key, payload) // só se o cache já existe
		pipe.LTrim(ctx, key, 0, chatCacheSize-1)
		pipe.Expire(ctx, key, chatCacheTTL)
		pipe.Publish(ctx, database.UserChannel(input.RecieverID), payload)
		pipe.Exec(ctx) // falha no Redis não derruba o envio
	}

	return c.Status(fiber.StatusCreated).JSON(dto)
}

func GetChat(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}
	otherID, err := c.ParamsInt("otherId")
	if err != nil || otherID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id invalido"})
	}

	ctx := c.Context()
	key := database.ChatKey(userID, uint(otherID))

	// 1. tenta o cache (lista guardada do mais novo pro mais antigo)
	if raw, err := database.Rdb.LRange(ctx, key, 0, -1).Result(); err == nil && len(raw) > 0 {
		msgs := make([]MessageDTO, 0, len(raw))
		for i := len(raw) - 1; i >= 0; i-- { // inverte pra ordem cronológica
			var m MessageDTO
			if json.Unmarshal([]byte(raw[i]), &m) == nil {
				msgs = append(msgs, m)
			}
		}
		return c.JSON(msgs)
	}

	// 2. cache miss: busca no banco (já com o post compartilhado e o autor dele)
	var rows []entities.Message
	err = database.DB.Preload("Post.User").
		Where("(sender_id = ? AND reciever_id = ?) OR (sender_id = ? AND reciever_id = ?)",
			userID, otherID, otherID, userID).
		Order("created_at desc, id desc").Limit(chatCacheSize).Find(&rows).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao buscar chat"})
	}

	// 3. repopula o cache (rows já tá do mais novo pro mais antigo)
	if len(rows) > 0 {
		pipe := database.Rdb.Pipeline()
		for i := len(rows) - 1; i >= 0; i-- { // empurra do mais antigo pro mais novo
			if p, err := json.Marshal(toDTO(rows[i])); err == nil {
				pipe.LPush(ctx, key, p)
			}
		}
		pipe.Expire(ctx, key, chatCacheTTL)
		pipe.Exec(ctx)
	}

	msgs := make([]MessageDTO, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		msgs = append(msgs, toDTO(rows[i]))
	}
	return c.JSON(msgs)
}
