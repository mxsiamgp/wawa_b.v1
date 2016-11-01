package js_regex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeTextPattern(t *testing.T) {
	assert.Equal(t, "\\^test\\$", EscapeTextPattern("^test$"))
	assert.Equal(t, "\\(test\\)", EscapeTextPattern("(test)"))
	assert.Equal(t, "\\[test\\]", EscapeTextPattern("[test]"))
	assert.Equal(t, "\\{test\\}", EscapeTextPattern("{test}"))
	assert.Equal(t, "\\\\\\.\\*\\+\\?\\|test", EscapeTextPattern("\\.*+?|test"))
}
