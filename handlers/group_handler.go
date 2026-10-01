package handlers

import (
	"lea/database"
	"lea/entities"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	groupRoleOwner  = "owner"
	groupRoleMember = "member"
)

type CreateGroupInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GroupDTO struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uint      `json:"owner_id"`
	OwnerName   string    `json:"owner_name"`
	Members     int64     `json:"members"`
	IsMember    bool      `json:"is_member"`
	IsOwner     bool      `json:"is_owner"`
	CreatedAt   time.Time `json:"created_at"`
}

func GetGroups(c *fiber.Ctx) error {
	viewerID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}

	var groups []entities.Group
	err := database.DB.Preload("Owner").Order("created_at desc, id desc").Find(&groups).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao buscar grupos"})
	}

	// contagem de membros de todos os grupos numa query só
	type row struct {
		GroupID uint
		N       int64
	}
	var counts []row
	database.DB.Model(&entities.GroupMember{}).
		Select("group_id, count(*) as n").Group("group_id").Scan(&counts)
	members := make(map[uint]int64, len(counts))
	for _, r := range counts {
		members[r.GroupID] = r.N
	}

	// em quais grupos o usuário logado já está
	var mine []uint
	database.DB.Model(&entities.GroupMember{}).Where("user_id = ?", viewerID).Pluck("group_id", &mine)
	isMember := make(map[uint]bool, len(mine))
	for _, id := range mine {
		isMember[id] = true
	}

	out := make([]GroupDTO, 0, len(groups))
	for _, g := range groups {
		out = append(out, GroupDTO{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			OwnerID:     g.OwnerID,
			OwnerName:   g.Owner.Name,
			Members:     members[g.ID],
			IsMember:    isMember[g.ID],
			IsOwner:     g.OwnerID == viewerID,
			CreatedAt:   g.CreatedAt,
		})
	}
	return c.JSON(out)
}

func CreateGroup(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}

	var input CreateGroupInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "dados invalidos"})
	}
	name := strings.TrimSpace(input.Name)
	desc := strings.TrimSpace(input.Description)
	if n := utf8.RuneCountInString(name); n < 3 || n > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "o nome deve ter entre 3 e 100 caracteres"})
	}
	if utf8.RuneCountInString(desc) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "a descricao pode ter no maximo 500 caracteres"})
	}

	var exists int64
	database.DB.Model(&entities.Group{}).Where("lower(name) = lower(?)", name).Count(&exists)
	if exists > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "ja existe um grupo com esse nome"})
	}

	group := entities.Group{Name: name, Description: desc, OwnerID: userID}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		// o dono entra como membro com role "owner"
		return tx.Create(&entities.GroupMember{
			GroupID: group.ID, UserID: userID, Role: groupRoleOwner,
		}).Error
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao criar grupo"})
	}

	database.DB.Preload("Owner").First(&group, group.ID)
	return c.Status(fiber.StatusCreated).JSON(GroupDTO{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		OwnerID:     group.OwnerID,
		OwnerName:   group.Owner.Name,
		Members:     1,
		IsMember:    true,
		IsOwner:     true,
		CreatedAt:   group.CreatedAt,
	})
}

func JoinGroup(c *fiber.Ctx) error {
	viewerID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}
	groupID, err := c.ParamsInt("id")
	if err != nil || groupID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id invalido"})
	}

	var count int64
	database.DB.Model(&entities.Group{}).Where("id = ?", groupID).Count(&count)
	if count == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "grupo nao existe"})
	}

	m := entities.GroupMember{GroupID: uint(groupID), UserID: viewerID, Role: groupRoleMember}
	// se já é membro, não faz nada (sem erro)
	if err := database.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&m).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao entrar no grupo"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func LeaveGroup(c *fiber.Ctx) error {
	viewerID, ok := c.Locals("user_id").(uint)
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
	if group.OwnerID == viewerID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "o dono nao pode sair do proprio grupo"})
	}

	// Unscoped = apaga de verdade, senão o índice único (grupo+usuário)
	// bloquearia entrar de novo depois
	err = database.DB.Unscoped().
		Where("group_id = ? AND user_id = ?", groupID, viewerID).
		Delete(&entities.GroupMember{}).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao sair do grupo"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func GetGroupPosts(c *fiber.Ctx) error {
	if _, ok := c.Locals("user_id").(uint); !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Não autorizado"})
	}
	groupID, err := c.ParamsInt("id")
	if err != nil || groupID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id invalido"})
	}

	var count int64
	database.DB.Model(&entities.Group{}).Where("id = ?", groupID).Count(&count)
	if count == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "grupo nao existe"})
	}

	var posts []entities.Post
	err = database.DB.Preload("User").Preload("Category").
		Where("group_id = ?", groupID).
		Order("created_at desc, id desc").Find(&posts).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar posts do grupo"})
	}

	out := make([]PostDTO, 0, len(posts))
	for _, p := range posts {
		out = append(out, toPostDTO(p))
	}
	return c.JSON(out)
}
