package socket

import (
	"context"
	"log"
)

// CheckContext kiểm tra xem context đã bị hủy chưa và log message nếu có.
// Trả về true nếu context đã bị hủy, ngược lại false.
func CheckContext(ctx context.Context, clientID string, logMessage string) bool {
	select {
	case <-ctx.Done():
		if clientID != "" { // Chỉ log nếu có clientID
			log.Printf("ContextUtils: %s for client %s. Reason: %v", logMessage, clientID, ctx.Err())
		} else {
			log.Printf("ContextUtils: %s. Reason: %v", logMessage, ctx.Err())
		}
		return true
	default:
		return false
	}
}
