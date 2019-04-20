package formatter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/formatter"
)

func TestJson_Format(t *testing.T) {
	tests := []struct {
		name     string
		entry    logger.Entry
		expected string
	}{
		{
			name:     "test simple message without context",
			entry:    logger.Entry{Message: "test message", Level: logger.DebugLevel, Context: nil},
			expected: "{\"Message\":\"test message\",\"Level\":7,\"Context\":null}",
		},
		{
			name:     "test simple message with context",
			entry:    logger.Entry{Message: "test message", Level: logger.WarningLevel, Context: logger.NewContext().Add("my_key", "my_value")},
			expected: "{\"Message\":\"test message\",\"Level\":4,\"Context\":{\"my_key\":\"my_value\"}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := formatter.NewJsonEncoder()

			assert.Equal(t, tt.expected, f.Format(tt.entry))
		})
	}
}
