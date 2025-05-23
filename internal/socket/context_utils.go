package socket

import (
	"context"
	"log"
)

func CheckContext(ctx context.Context, clientID string, logMessage string) bool {
	select {
	case <-ctx.Done():
		log.Printf("%s for client %s", logMessage, clientID)
		return true
	default:
		return false
	}
}
