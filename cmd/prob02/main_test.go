package main

import "testing"

func TestTwosComplement(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected int32
	}{
		{
			name:     "negative number",
			input:    []byte{0xff, 0xff, 0xff, 0xfb},
			expected: int32(-5),
		},
		{
			name:     "positive number",
			input:    []byte{0x00, 0x00, 0x30, 0x39},
			expected: int32(12345),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := twosComplement(tt.input)
			if got != tt.expected {
				t.Errorf("got %v, expected %v", got, tt.expected)
			}
		})
	}
}
