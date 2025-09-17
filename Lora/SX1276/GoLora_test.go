package SX1276

import (
	"errors"
	"testing"

	"github.com/Fsyahputra/GoLora/driver"
	"github.com/stretchr/testify/assert"
)

type mockCbPin struct {
	readVal bool
}

func (cb *mockCbPin) ReadVal() (bool, error) {
	return cb.readVal, nil
}

func newMockCbPin() *mockCbPin {
	return &mockCbPin{
		readVal: true,
	}
}

type mockRstPin struct {
}

func (rst *mockRstPin) Low() error {
	return nil
}

func (rst *mockRstPin) High() error {
	return nil
}

type mockModConn struct {
	readval byte
}

func (m *mockModConn) SendToMod(reg, value byte) error {
	return nil
}

func (m *mockModConn) updateReadVal(newReadVal byte) {
	m.readval = newReadVal
}

func (m *mockModConn) ReadFromMod(reg byte) (byte, error) {
	return m.readval, nil
}

type mockHwDriver struct {
	mockCbPin
	mockRstPin
	mockModConn
}

func (mhw *mockHwDriver) Init() (*driver.Driver, error) {
	return &driver.Driver{
		RSTPin:  &mhw.mockRstPin,
		CbPin:   &mhw.mockCbPin,
		ModComm: &mhw.mockModConn,
	}, nil
}

func newDefLoraConf() LoraConf {
	return LoraConf{
		TxPower:        0,
		SF:             0,
		BW:             0,
		CodingRate:     0,
		PreambleLength: 0,
		SyncWord:       0,
		Frequency:      0,
		Header:         false,
		EnableCrc:      false,
	}
}

func TestGoLora_CheckConn(t *testing.T) {

	tests := []struct {
		name     string
		mockAddr byte
		want     error
	}{
		{
			name:     "Should return nil if addr is 0x12",
			mockAddr: 0x12,
			want:     nil,
		},
		{
			name:     "Should return err if addr isn't 0x12",
			mockAddr: 0xff,
			want:     errors.New("check Your Connection"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			hw := mockHwDriver{
				mockCbPin:   mockCbPin{readVal: false},
				mockRstPin:  mockRstPin{},
				mockModConn: mockModConn{readval: tt.mockAddr},
			}
			drv, _ := hw.Init()
			defConf := newDefLoraConf()
			gl := NewGoLoraSX1276(drv, defConf)
			err := gl.CheckConn()
			if tt.want == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.want.Error())
			}
		})
	}
}
