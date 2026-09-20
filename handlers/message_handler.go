package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"lea/database"
	"lea/entities"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetUserMessage(c *fiber.Ctx) error {
	userID := c.Params("id")
	var messages []entities.Message

	result := database.DB.Preload("Sender").Preload("Reciever").Where("sender_id = ?", userID, userID).Find(&messages)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao buscar mensagens",
		})
	}
	return c.JSON(messages)
}

func GetChatBetweenUsers(c *fiber.Ctx) error {
	user1 := c.Params("user1")
	user2 := c.Params("user2")
	var messages []entities.Message

	// Busca mensagens onde (Sender = 1 E Reciever = 2) OU (Sender = 2 E Reciever = 1)
	result := database.DB.Preload("Sender").Preload("Reciever").
		Where("(sender_id = ? AND reciever_id = ?) OR (sender_id = ? AND reciever_id = ?)", user1, user2, user2, user1).
		Order("created_at asc"). // Ordena para a conversa ficar na ordem cronológica certa
		Find(&messages)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao buscar histórico do chat",
		})
	}

	return c.JSON(messages)
}

type CreateMessageInput struct {
	RecieverID uint   `json:"reciever_id"`
	Content    string `json:"content"`
}

func CreateMessage(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Não autorizado",
		})
	}

	var input CreateMessageInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "mensagem invalida",
		})
	}

	message := entities.Message{
		SenderID:   userID,
		RecieverID: input.RecieverID,
		Content:    input.Content,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "erro ao salvar mensagem no banco",
		})
	}

	messageJSON, err := json.Marshal(message)
	if err == nil {
		ctx := context.Background()

		cacheKey := fmt.Sprintf("chat:last:%d:%d", userID, input.RecieverID)
		database.Rdb.Set(ctx, cacheKey, messageJSON, 36*time.Hour)
		database.Rdb.Publish(ctx, "chat_channel", messageJSON)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "mensagem enviada com sucesso",
		"data":    message,
	})
}
