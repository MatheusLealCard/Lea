package handlers

import (
	"context"
	"lea/database"

	"github.com/gofiber/contrib/websocket"
)

func ChatWS(c *websocket.Conn) {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := database.Rdb.Subscribe(ctx, database.UserChannel(userID))
	defer sub.Close()

	// Redis -> navegador (só esta goroutine escreve no socket)
	go func() {
		for msg := range sub.Channel() {
			if err := c.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				c.Close()
				return
			}
		}
	}()

	// fica lendo só pra perceber quando o navegador fecha a conexão
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			return
		}
	}
}
