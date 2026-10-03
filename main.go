package main

import (
	"lea/database"
	"lea/entities"
	"lea/handlers"
	"lea/middlewares"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func seedUser(name, email, password string, roleID uint) entities.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	var u entities.User
	if err := database.DB.Where("email = ?", email).First(&u).Error; err != nil {
		u = entities.User{Name: name, Email: email, Password: string(hash), RoleID: roleID}
		if err := database.DB.Create(&u).Error; err != nil {
			log.Fatal(err)
		}
		log.Println("seed: criado", email)
		return u
	}

	err = database.DB.Model(&u).Updates(map[string]interface{}{
		"name":     name,
		"password": string(hash),
		"role_id":  roleID,
	}).Error
	if err != nil {
		log.Fatal(err)
	}
	log.Println("seed: atualizado", email)
	return u
}

func main() {
	godotenv.Load()
	database.Connect()
	database.ConnectRedis()

	// ROLES
	var userRole, creatorRole, adminRole entities.Role
	database.DB.Where(entities.Role{RoleName: "User"}).FirstOrCreate(&userRole)
	database.DB.Where(entities.Role{RoleName: "Creator"}).FirstOrCreate(&creatorRole)
	database.DB.Where(entities.Role{RoleName: "Admin"}).FirstOrCreate(&adminRole)

	// ADMINS
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		log.Fatal("defina ADMIN_PASSWORD no .env")
	}
	matheus := seedUser("Matheus", "matheus@gmail.com", adminPass, adminRole.ID)
	marcos := seedUser("Marcos", "marcos@gmail.com", adminPass, adminRole.ID)

	// CATEGORIA, POST E FOLLOWER
	var category entities.Category
	database.DB.Where(entities.Category{CategoryName: "matematica"}).FirstOrCreate(&category)

	post := entities.Post{Title: "Primeiro post", Content: "Testando 1234", CategoryID: category.ID, UserID: matheus.ID}
	database.DB.Where(entities.Post{Title: post.Title}).FirstOrCreate(&post)

	follower := entities.Follower{UserID: matheus.ID, FolowerID: marcos.ID}
	database.DB.Where(follower).FirstOrCreate(&follower)

	app := fiber.New(fiber.Config{BodyLimit: 6 * 1024 * 1024})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// públicas
	app.Static("/uploads", "./uploads")
	app.Get("/getUsers", handlers.GetUsers)
	app.Post("/createUsers", handlers.CreateUser)
	app.Post("/login", handlers.Login)
	app.Get("/getPosts", handlers.GetPosts)
	app.Get("/users/:id/followers", handlers.GetUserFolowers)
	app.Get("/getRoles", handlers.GetRoles)
	app.Get("/getCategories", handlers.GetCategories)

	// protegidas
	api := app.Group("/api", middlewares.Protected())
	api.Post("/sendMessage", handlers.CreateMessage)
	api.Get("/chat/:otherId", handlers.GetChat)
	api.Post("/posts", handlers.CreatePost)
	api.Get("/users/:id/profile", handlers.GetProfile)
	api.Post("/users/:id/follow", handlers.FollowUser)
	api.Delete("/users/:id/follow", handlers.UnfollowUser)
	api.Post("/me/avatar", handlers.UploadAvatar)
	api.Get("/groups", handlers.GetGroups)
	api.Post("/groups", handlers.CreateGroup)
	api.Post("/groups/:id/join", handlers.JoinGroup)
	api.Delete("/groups/:id/join", handlers.LeaveGroup)
	api.Get("/groups/:id/posts", handlers.GetGroupPosts)
	api.Get("/posts/:id", handlers.GetPost)
	api.Get("/posts/:id/comments", handlers.GetComments)
	api.Post("/posts/:id/comments", handlers.CreateComment)
	api.Delete("/comments/:id", handlers.DeleteComment)
	api.Delete("/groups/:id", handlers.DeleteGroup)
	api.Delete("/posts/:id", handlers.DeletePost)

	log.Println("Server running")
	log.Fatal(app.Listen(":3000"))
}
