package main

import (
	"lea/database"
	"lea/entities"
	"lea/handlers"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	database.Connect()

	//ROLES
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

	//CRIANDO POSTS
	post := entities.Post{
		Title:   "Primeiro post",
		Content: "Testando 1234",
		UserID:  1,
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

	app := fiber.New()

	app.Get("/getUsers", handlers.GetUsers)
	app.Post("/createUsers", handlers.CreateUser)

	app.Get("/getPosts", handlers.GetPosts)

	app.Get("/users/:id/followers", handlers.GetUserFolowers)

	app.Get("/getRoles", handlers.GetRoles)

	log.Println("Server running")
	log.Fatal(app.Listen(":3000"))
}
