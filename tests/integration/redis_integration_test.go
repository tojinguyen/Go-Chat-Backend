//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedisIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("set_and_get_string_integration", func(t *testing.T) {
		key := "test:string:key"
		value := "test string value"

		// Set value
		err := RedisService.Set(ctx, key, value, 5*time.Minute)
		assert.NoError(t, err)

		// Get value
		retrievedValue, err := RedisService.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrievedValue)

		// Clean up
		RedisService.Delete(ctx, key)
	})

	t.Run("set_and_get_with_expiration_integration", func(t *testing.T) {
		key := "test:expiration:key"
		value := "test expiration value"
		expiration := 2 * time.Second

		// Set value with short expiration
		err := RedisService.Set(ctx, key, value, expiration)
		assert.NoError(t, err)

		// Get value immediately
		retrievedValue, err := RedisService.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrievedValue)

		// Wait for expiration
		time.Sleep(3 * time.Second)

		// Try to get expired value
		retrievedValue, err = RedisService.Get(ctx, key)
		assert.Error(t, err)
		assert.Equal(t, "", retrievedValue)
		assert.Contains(t, err.Error(), "redis: nil")
	})

	t.Run("delete_key_integration", func(t *testing.T) {
		key := "test:delete:key"
		value := "test delete value"

		// Set value
		err := RedisService.Set(ctx, key, value, 5*time.Minute)
		assert.NoError(t, err)

		// Verify it exists
		retrievedValue, err := RedisService.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrievedValue)

		// Delete key
		err = RedisService.Delete(ctx, key)
		assert.NoError(t, err)

		// Verify it's deleted
		retrievedValue, err = RedisService.Get(ctx, key)
		assert.Error(t, err)
		assert.Equal(t, "", retrievedValue)
	})

	t.Run("get_nonexistent_key_integration", func(t *testing.T) {
		key := "test:nonexistent:key"

		// Try to get non-existent key
		value, err := RedisService.Get(ctx, key)
		assert.Error(t, err)
		assert.Equal(t, "", value)
		assert.Contains(t, err.Error(), "redis: nil")
	})

	t.Run("refresh_token_scenario_integration", func(t *testing.T) {
		userId := "user-123"
		refreshTokenKey := "refresh_token:" + userId
		refreshToken := "refresh-token-abc123"
		expiration := 24 * time.Hour

		// Store refresh token
		err := RedisService.Set(ctx, refreshTokenKey, refreshToken, expiration)
		assert.NoError(t, err)

		// Retrieve refresh token
		storedToken, err := RedisService.Get(ctx, refreshTokenKey)
		assert.NoError(t, err)
		assert.Equal(t, refreshToken, storedToken)

		// Update refresh token (simulate token refresh)
		newRefreshToken := "new-refresh-token-xyz789"
		err = RedisService.Set(ctx, refreshTokenKey, newRefreshToken, expiration)
		assert.NoError(t, err)

		// Verify new token
		storedToken, err = RedisService.Get(ctx, refreshTokenKey)
		assert.NoError(t, err)
		assert.Equal(t, newRefreshToken, storedToken)

		// Clean up
		RedisService.Delete(ctx, refreshTokenKey)
	})

	t.Run("concurrent_operations_integration", func(t *testing.T) {
		key := "test:concurrent:key"
		value1 := "concurrent value 1"
		value2 := "concurrent value 2"

		// Set first value
		err := RedisService.Set(ctx, key, value1, 5*time.Minute)
		assert.NoError(t, err)

		// Immediately set second value (should overwrite)
		err = RedisService.Set(ctx, key, value2, 5*time.Minute)
		assert.NoError(t, err)

		// Get final value
		retrievedValue, err := RedisService.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value2, retrievedValue)

		// Clean up
		RedisService.Delete(ctx, key)
	})

	t.Run("multiple_keys_operations_integration", func(t *testing.T) {
		keys := []string{
			"test:multi:key1",
			"test:multi:key2",
			"test:multi:key3",
		}
		values := []string{
			"value1",
			"value2",
			"value3",
		}

		// Set multiple keys
		for i, key := range keys {
			err := RedisService.Set(ctx, key, values[i], 5*time.Minute)
			assert.NoError(t, err)
		}

		// Get and verify all keys
		for i, key := range keys {
			retrievedValue, err := RedisService.Get(ctx, key)
			assert.NoError(t, err)
			assert.Equal(t, values[i], retrievedValue)
		}

		// Clean up all keys
		for _, key := range keys {
			RedisService.Delete(ctx, key)
		}

		// Verify all keys are deleted
		for _, key := range keys {
			_, err := RedisService.Get(ctx, key)
			assert.Error(t, err)
		}
	})

	t.Run("redis_connection_health_check_integration", func(t *testing.T) {
		// Test basic ping
		err := TestRedis.Ping(ctx).Err()
		assert.NoError(t, err)

		// Test with a simple operation
		testKey := "test:health:check"
		testValue := "health check value"

		err = RedisService.Set(ctx, testKey, testValue, time.Minute)
		assert.NoError(t, err)

		retrievedValue, err := RedisService.Get(ctx, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testValue, retrievedValue)

		// Clean up
		RedisService.Delete(ctx, testKey)
	})
}
