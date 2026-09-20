package main

import (
	"lea/database"
	"lea/entities"
	"lea/handlers"
	"lea/middlewares"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	database.Connect()
	database.ConnectRedis()

	//CRIANDO ROLES
	user_role := entities.Role{
		RoleName: "User",
	}
	database.DB.FirstOrCreate(&user_role, entities.Role{
		RoleName: user_role.RoleName,
	})

	creator_role := entities.Role{
		RoleName: "Creator",
	}
	database.DB.FirstOrCreate(&creator_role, entities.Role{
		RoleName: creator_role.RoleName,
	})

	//CRIANDO USERS
	user := entities.User{
		Name:     "Carlos",
		Email:    "carlos@gmail.com",
		Password: "12345",
		RoleID:   1,
	}
	database.DB.FirstOrCreate(&user, entities.User{
		Email: user.Email,
	})

	user2 := entities.User{
		Name:     "Matheus",
		Email:    "matheus@gmail.com",
		Password: "mm1234",
		RoleID:   2,
	}
	database.DB.FirstOrCreate(&user2, entities.User{
		Email: user2.Email,
	})

	//CRIANDO CATEGORIAS
	matematica_category := entities.Category{
		CategoryName: "matematica",
	}
	database.DB.FirstOrCreate(&matematica_category, entities.Category{
		CategoryName: matematica_category.CategoryName,
	})

	//CRIANDO POSTS
	post := entities.Post{
		Title:      "Primeiro post",
		Content:    "Testando 1234",
		CategoryID: 1,
		UserID:     1,
	}
	database.DB.FirstOrCreate(&post, entities.Post{
		Title: post.Title,
	})

	//CRIANDO FOLOWERS
	follower1 := entities.Follower{
		UserID:    1,
		FolowerID: 2,
	}
	database.DB.FirstOrCreate(&follower1, entities.Follower{
		UserID:    user.ID,
		FolowerID: user2.ID,
	})

	//CRIANDO MENSAGEM
	mensagem1 := entities.Message{
		SenderID:   1,
		RecieverID: 2,
		Content:    "Eae fdp",
	}
	database.DB.Create(&mensagem1)

	mensagem2 := entities.Message{
		SenderID:   2,
		RecieverID: 1,
		Content:    "Fala arrombado",
	}
	database.DB.Create(&mensagem2)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	//USERS ROTAS
	app.Get("/getUsers", handlers.GetUsers)
	app.Post("/createUsers", handlers.CreateUser)
	app.Post("/login", handlers.Login)

	//POSTS ROTAS
	app.Get("/getPosts", handlers.GetPosts)

	//FOLOWERS ROTAS
	app.Get("/users/:id/followers", handlers.GetUserFolowers)

	//ROLES ROTAS
	app.Get("/getRoles", handlers.GetRoles)

	//CATEGORIAS ROTAS
	app.Get("getCategories", handlers.GetCategories)

	//app.Post("/sendMessage", handlers.CreateMessage)

	//ROTAS PROTEGIDAS
	api := app.Group("/api", middlewares.Protected())

	//ROTA PROTEGIDA MENSAGEM
	api.Post("/sendMessage", handlers.CreateMessage)

	//MENSAGEMS ROTAS
	api.Get("/getChatMessages/:user1/:user2", handlers.GetChatBetweenUsers)

	log.Println("Server running")
	log.Fatal(app.Listen(":3000"))
}
