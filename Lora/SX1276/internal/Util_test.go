package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newLoraUtils() *LoraUtils {
	return &LoraUtils{}
}

func TestSetWriteMask(t *testing.T) {
	lu := newLoraUtils()
	tests := []struct {
		input byte
		want  byte
	}{
		{input: 0x00, want: 0x80},
		{input: 0x01, want: 0x81},
		{input: 0x02, want: 0x82},
		{input: 0x03, want: 0x83},
	}

	for _, tt := range tests {
		t.Run("write Mask Test", func(t *testing.T) {
			result := lu.SetWriteMask(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSetReadMask(t *testing.T) {
	lu := newLoraUtils()
	tests := []struct {
		input byte
		want  byte
	}{
		{input: 0x00, want: 0x00},
		{0xff, 0x7f},
	}

	for _, tt := range tests {
		t.Run("read Mask Test", func(t *testing.T) {
			result := lu.SetReadMask(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestChangeMode(t *testing.T) {
	lu := newLoraUtils()
	test := []struct {
		name  string
		input LoraMode
		want  byte
	}{
		{
			name:  "Sleep Mode",
			input: Sleep,
			want:  0x80 | 0x00,
		},
		{
			name:  "Idle Mode",
			input: Idle,
			want:  0x80 | 0x01,
		},
		{
			name:  "Tx Mode",
			input: Tx,
			want:  0x80 | 0x03,
		},
		{
			name:  "RxContinuous Mode",
			input: RxContinuous,
			want:  0x80 | 0x05,
		},
		{
			name:  "RxSingle Mode",
			input: RxSingle,
			want:  0x80 | 0x06,
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			result := lu.ChangeMode(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}

}
