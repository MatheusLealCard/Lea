package handlers

import (
	"lea/database"
	"lea/entities"

	"github.com/gofiber/fiber/v2"
)

func GetPosts(c *fiber.Ctx) error {
	var posts []entities.Post
	result := database.DB.Preload("User").Find(&posts)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Nao foi possivel buscar posts",
		})
	}

	return c.JSON(posts)
}

func CreatePost(c *fiber.Ctx) error {
	var post entities.Post

	if err := c.BodyParser(&post); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "erro ao criar post",
		})
	}

	database.DB.Preload("User").First(&post, post.ID)
	return c.Status(fiber.StatusCreated).JSON(post)
}
