package SX1276

import (
	"errors"
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
			result := lu.setWriteMask(tt.input)
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
			result := lu.setReadMask(tt.input)
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
		{
			name:  "nil input",
			input: 100,
			want:  0x80 | 0x01,
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			result := lu.changeMode(tt.input)
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
			result := lu.setFreq(tt.freq)
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
			result := lu.setTxPower(tt.power)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestLoraUtils_CheckData(t *testing.T) {
	lu := newLoraUtils()
	tests := []struct {
		name string
		irq  byte
		want error
	}{
		{
			name: "Should Return nil with good irq",
			irq:  0x40,
			want: nil,
		},
		{
			name: "Should Return error if rx isn't done yet",
			irq:  0x01,
			want: errors.New("no Packet Received"),
		},
		{
			name: "Should Return error if crc invalid",
			irq:  0xA0,
			want: errors.New("packet damaged or lost in transmit"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lu.checkData(tt.irq)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestLoraUtils_SetBW(t *testing.T) {
	lu := newLoraUtils()
	type testA struct {
		name string
		bw   byte
		want byte
	}
	test := testA{
		name: "Should return a shifted 4 bit to msb",
		bw:   0x0f,
		want: 0xf0,
	}
	t.Run(test.name, func(t *testing.T) {
		result := lu.setBW(test.bw)
		assert.Equal(t, test.want, result)
	})
}

func TestLoraUtils_SetCodingRate(t *testing.T) {
	lu := newLoraUtils()
	test := struct {
		name        string
		cr          byte
		currentConf byte
		want        byte
	}{
		name:        "it Should Overwrite cr conf in current Modem config",
		cr:          0x00,
		currentConf: 0xff,
		want:        0xf1,
	}

	t.Run(test.name, func(t *testing.T) {
		result := lu.setCodingRate(test.cr, test.currentConf)
		assert.Equal(t, test.want, result)
	})
}

func TestLoraUtils_SetCrc(t *testing.T) {
	lu := newLoraUtils()
	tests := []struct {
		name        string
		crc         bool
		currentConf byte
		want        byte
	}{
		{
			name:        "it Should overwrite crcbit to 1 in current conf if crc true",
			currentConf: 0xdb,
			crc:         true,
			want:        0xdf,
		},
		{
			name:        "it Should overwrite crcbit to 0 in current conf if crc false",
			crc:         false,
			currentConf: 0xdf,
			want:        0xdb,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lu.setCrc(tt.crc, tt.currentConf)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestLoraUtils_SetHeader(t *testing.T) {
	lu := newLoraUtils()
	test := []struct {
		name        string
		header      bool
		currentConf byte
		want        byte
	}{
		{
			name:        "it Should Overwrite header bit to zero in explicit in current modem config",
			header:      true,
			currentConf: 0xfb,
			want:        0xfa,
		},
		{
			name:        "it Should Overwrite header bit to one in implicit in current modem config",
			header:      false,
			currentConf: 0xfa,
			want:        0xfb,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			result := lu.setHeader(bool(tt.header), tt.currentConf)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestLoraUtils_SetPreamble(t *testing.T) {
	lu := newLoraUtils()
	test := struct {
		name     string
		preamble uint16
		want     []byte
	}{
		name:     "It should slice into msb and lsb",
		preamble: 10245,
		want:     []byte{0x28, 0x05},
	}

	t.Run(test.name, func(t *testing.T) {
		results := lu.setPreamble(test.preamble)
		for idx, result := range results {
			assert.Equal(t, test.want[idx], result)
		}
	})
}

func TestLoraUtils_SetSF(t *testing.T) {
	lu := newLoraUtils()
	test := struct {
		name string
		sf   byte
		want byte
	}{
		name: "it should shift 4 step towards msb",
		sf:   0x0f,
		want: 0xf0,
	}

	t.Run(test.name, func(t *testing.T) {
		result := lu.setSF(test.sf)
		assert.Equal(t, test.want, result)
	})
}
