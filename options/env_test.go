package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_env(t *testing.T) {
	conf := ReadEnv()
	assert.Equal(t, "system:node:docker-desktop", conf.SystemUsers[0])
	assert.Equal(t, "docker-for-desktop", conf.AdminUsers[0])
	assert.Equal(t, "user1", conf.AdminUsers[1])
	assert.Equal(t, "user2", conf.AdminUsers[2])
}
