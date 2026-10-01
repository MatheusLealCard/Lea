package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"lea/database"
	"lea/entities"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	uploadDir    = "./uploads"
	maxPhotoSize = 5 << 20 // 5MB
)

var allowedImages = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type PostDTO struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	PhotoURL     string    `json:"photo_url"`
	AuthorID     uint      `json:"author_id"`
	AuthorName   string    `json:"author_name"`
	CategoryName string    `json:"category_name"`
	CreatedAt    time.Time `json:"created_at"`
}

func toPostDTO(p entities.Post) PostDTO {
	url := ""
	if p.Photo != "" {
		url = "/uploads/" + p.Photo
	}
	return PostDTO{p.ID, p.Title, p.Content, url, p.UserID, p.User.Name, p.Category.CategoryName, p.CreatedAt}
}

// salva a foto (se veio) e devolve só o nome do arquivo
func savePhoto(c *fiber.Ctx) (string, error) {
	fh, err := c.FormFile("photo")
	if err != nil {
		return "", nil // post sem foto
	}
	if fh.Size > maxPhotoSize {
		return "", fiber.NewError(fiber.StatusRequestEntityTooLarge, "foto maior que 5MB")
	}

	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	head := make([]byte, 512)
	n, _ := f.Read(head)
	f.Close()

	ext, ok := allowedImages[http.DetectContentType(head[:n])]
	if !ok {
		return "", fiber.NewError(fiber.StatusUnsupportedMediaType, "formato invalido (use jpg, png, webp ou gif)")
	}

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", err
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	name := hex.EncodeToString(b) + ext

	if err := c.SaveFile(fh, filepath.Join(uploadDir, name)); err != nil {
		return "", err
	}
	return name, nil
}

func GetPosts(c *fiber.Ctx) error {
	var posts []entities.Post
	err := database.DB.Preload("User").Preload("Category").
		Order("created_at desc, id desc").Find(&posts).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Nao foi possivel buscar posts",
		})
	}

	out := make([]PostDTO, 0, len(posts))
	for _, p := range posts {
		out = append(out, toPostDTO(p))
	}
	return c.JSON(out)
}

func CreatePost(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}

	title := strings.TrimSpace(c.FormValue("title"))
	content := strings.TrimSpace(c.FormValue("content"))
	catID, err := strconv.Atoi(c.FormValue("category_id"))
	if title == "" || content == "" || len(title) > 150 || err != nil || catID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Preencha titulo (max 150), conteudo e categoria",
		})
	}

	var count int64
	database.DB.Model(&entities.Category{}).Where("id = ?", catID).Count(&count)
	if count == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "categoria nao existe"})
	}

	photo, err := savePhoto(c)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao salvar foto"})
	}

	post := entities.Post{
		Title:      title,
		Content:    content,
		Photo:      photo,
		UserID:     userID,
		CategoryID: uint(catID),
	}
	if err := database.DB.Create(&post).Error; err != nil {
		if photo != "" {
			os.Remove(filepath.Join(uploadDir, photo)) // não deixa foto órfã
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao criar post"})
	}

	database.DB.Preload("User").Preload("Category").First(&post, post.ID)
	return c.Status(fiber.StatusCreated).JSON(toPostDTO(post))
}
