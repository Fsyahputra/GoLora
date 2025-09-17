package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newLoraUtils() *LoraUtils {
	return &LoraUtils{}
}

func TestLoraUtils_SetWriteMask(t *testing.T) {
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

func TestLoraUtils_SetReadMask(t *testing.T) {
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

func TestLoraUtils_ChangeMode(t *testing.T) {
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

func TestLoraUtils_SetFreq(t *testing.T) {
	lu := newLoraUtils()
	tests := []struct {
		name string
		freq uint64
		want []byte
	}{
		{
			name: "freq 0",
			freq: 0,
			want: []byte{0x00, 0x00, 0x00},
		},
		{
			name: "freq kecil",
			freq: 0x123456,
			want: []byte{0x12, 0x34, 0x56},
		},
		{
			name: "freq besar",
			freq: 0xFFFFFF,
			want: []byte{0xFF, 0xFF, 0xFF},
		},
		{
			name: "freq lebih dari 24 bit",
			freq: 0x12345678,
			want: []byte{0x34, 0x56, 0x78}, // cuma ambil 3 byte terakhir
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lu.SetFreq(tt.freq)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestLoraUtils_SetTxPower(t *testing.T) {
	lu := newLoraUtils()
	tests := []struct {
		power byte
		want  byte
	}{
		{power: 0x00, want: 0x80},
		{power: 0x01, want: 0x81},
		{power: 0x02, want: 0x82},
		{power: 0x03, want: 0x83},
		{power: 0x7f, want: 0xff},
	}

	for _, tt := range tests {
		t.Run("write Mask Test", func(t *testing.T) {
			result := lu.SetWriteMask(tt.power)
			assert.Equal(t, tt.want, result)
		})
	}
}
