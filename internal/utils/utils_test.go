package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{
			name: "test 1: len of return function >0",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := GenerateRandomString()
			assert.Greater(t, len(gotResult), tt.want)
		})
	}
}
