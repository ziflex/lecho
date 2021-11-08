package lecho_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ziflex/lecho/v3"
)

func TestCtx(t *testing.T) {
	b := &bytes.Buffer{}
	l := lecho.New(b)
	zerologger := l.Unwrap()
	ctx := l.WithContext(context.Background())

	assert.Equal(t, lecho.Ctx(ctx), &zerologger)
}
