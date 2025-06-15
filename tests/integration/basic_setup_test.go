//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicIntegrationSetup(t *testing.T) {
	t.Run("test_environment_loaded", func(t *testing.T) {
		assert.NotNil(t, TestEnv)
		assert.NotEmpty(t, TestEnv.MysqlDatabase)
		assert.Equal(t, "test", TestEnv.RunMode)
	})

	t.Run("database_connection_working", func(t *testing.T) {
		assert.NotNil(t, TestDB)
		err := TestDB.Ping()
		assert.NoError(t, err)

		assert.NotNil(t, MySQLService)
		assert.NotNil(t, MySQLService.DB)
	})

	t.Run("redis_connection_working", func(t *testing.T) {
		assert.NotNil(t, TestRedis)
		ctx := timeoutContext()
		err := TestRedis.Ping(ctx).Err()
		assert.NoError(t, err)

		assert.NotNil(t, RedisService)
		assert.NotNil(t, RedisService.Client)
	})
}
