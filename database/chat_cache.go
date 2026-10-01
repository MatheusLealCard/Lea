package database

import "fmt"

func ChatKey(a, b uint) string {
	if a > b {
		a, b = b, a
	}
	return fmt.Sprintf("chat:%d:%d", a, b)
}

func UserChannel(id uint) string {
	return fmt.Sprintf("user:%d:messages", id)
}
