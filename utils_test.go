package ttapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeMapPath(t *testing.T) {
	var m any
	m = map[string]any{"a": map[string]any{"b": map[string]any{"c": 123}}}
	assert.Equal(t, 123, safeMapPath(m, "a.b.c"))
	assert.Nil(t, safeMapPath(m, "a.b.d"))
	assert.Nil(t, safeMapPath(m, "a.c.d.e.f.g"))
}

func TestTrunkStr(t *testing.T) {
	assert.Equal(t, "some long string", truncStr("some long string", 16, "..."))
	assert.Equal(t, "some long string", truncStr("some long string", 100, "..."))
	assert.Equal(t, "some long ...", truncStr("some long string", 10, "..."))
}
