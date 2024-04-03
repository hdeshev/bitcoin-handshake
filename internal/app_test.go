package internal

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewApplication(t *testing.T) {
	log := slog.Default()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := NewApplication(ctx, log)

	assert.NotNil(t, a.config)
	assert.NotNil(t, a.log)
}
