package lecho_test

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/ziflex/lecho/v2"
)

func TestCtx(t *testing.T) {
	b := &bytes.Buffer{}
	l := lecho.New(b)
	zerologger := l.Unwrap()
	ctx := l.WithContext(context.Background())

	assert.Equal(t, lecho.Ctx(ctx), &zerologger)
}
