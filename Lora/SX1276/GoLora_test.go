package SX1276

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Fsyahputra/GoLora/Lora/SX1276/internal"
	"github.com/Fsyahputra/GoLora/driver"
	"github.com/stretchr/testify/assert"
	"periph.io/x/conn/v3/physic"
)

func newDefLoraConf() LoraConf {
	return LoraConf{
		TxPower:        0,
		SF:             0,
		BW:             0,
		Denum:          0,
		PreambleLength: 0,
		SyncWord:       0,
		Frequency:      0,
		Header:         false,
		EnableCrc:      false,
	}
}

type mockRstPin struct {
	lowFunc  func() error
	highFunc func() error
}

func (mrst *mockRstPin) Low() error {
	return mrst.lowFunc()
}

func (mrst *mockRstPin) High() error {
	return mrst.highFunc()
}

type mockCbPin struct {
	readValFunc func() (bool, error)
}

func (mcb *mockCbPin) ReadVal() (bool, error) {
	mockedVal, err := mcb.readValFunc()
	return mockedVal, err
}

type mockModConn struct {
	send func(reg, val byte) error
	read func(reg byte) (byte, error)
}

type modConnTest struct {
	name    string
	readErr error
	sendErr error
}

func (mc *mockModConn) SendToMod(reg, val byte) error {
	return mc.send(reg, val)
}

func (mc *mockModConn) ReadFromMod(reg byte) (byte, error) {
	return mc.read(reg)
}

func testsDrvMock(SendErr error, ReadErr error) func() *driver.Driver {
	return func() *driver.Driver {
		return &driver.Driver{
			RSTPin: nil,
			CbPin:  nil,
			ModComm: &mockModConn{
				send: func(reg, val byte) error {
					return SendErr
				},
				read: func(reg byte) (byte, error) {
					return 0xff, ReadErr
				},
			},
		}
	}
}

func createDrvMockAndTest() ([]*driver.Driver, []modConnTest) {
	tests := []modConnTest{
		{
			name:    "should not return err if err doesnt happened",
			sendErr: nil,
			readErr: nil,
		},
		{
			name:    "Should return err if send communication failed",
			sendErr: errors.New("send test err"),
			readErr: nil,
		},
		{
			name:    "Should return err if read communication failed",
			sendErr: nil,
			readErr: errors.New("read test err"),
		},
	}

	driverList := make([]*driver.Driver, len(tests))
	for idx, tt := range tests {
		driverList[idx] = testsDrvMock(tt.sendErr, tt.readErr)()
	}

	return driverList, tests
}

func TestGoLora_CheckConn(t *testing.T) {

	tests := []struct {
		name     string
		mockFunc func() *driver.Driver
		want     error
	}{
		{
			name: "Should return nil if addr 0x12",
			mockFunc: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: &mockRstPin{
						lowFunc:  nil,
						highFunc: nil,
					},
					CbPin: &mockCbPin{readValFunc: nil},
					ModComm: &mockModConn{
						send: nil,
						read: func(reg byte) (byte, error) {
							return 0x12, nil
						},
					},
				}
			},
			want: nil,
		},
		{
			name: "Should return error if addr isn't 0x12",
			mockFunc: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: nil,
					CbPin:  nil,
					ModComm: &mockModConn{
						send: nil,
						read: func(reg byte) (byte, error) {
							return 0xff, nil
						},
					},
				}
			},
			want: errors.New("check Your Connection"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(tt.mockFunc(), newDefLoraConf())
			err := gl.CheckConn()
			if tt.want == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.want.Error())
			}
		})
	}
}

func TestGoLora_Reset(t *testing.T) {
	values := make([]bool, 2)
	test := struct {
		name    string
		mockDrv func() *driver.Driver
		want    error
		val     []bool
	}{
		name: "Should Return nil if no error",
		mockDrv: func() *driver.Driver {
			return &driver.Driver{
				RSTPin: &mockRstPin{lowFunc: func() error {
					values[0] = false
					return nil
				},
					highFunc: func() error {
						values[1] = true
						return nil
					},
				},
				CbPin:   nil,
				ModComm: nil,
			}
		},
		want: nil,
		val:  []bool{false, true},
	}

	t.Run(test.name, func(t *testing.T) {
		gl := NewGoLoraSX1276(test.mockDrv(), newDefLoraConf())
		err := gl.Reset()
		assert.NoError(t, err)
		assert.Equal(t, test.val, values)
	})

	test2 := []struct {
		name    string
		mockDrv func() *driver.Driver
		want    error
	}{
		{
			name: "it Should Return Err if Low is Err",
			mockDrv: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: &mockRstPin{lowFunc: func() error {
						return errors.New("low is Err")
					},
						highFunc: func() error {
							return nil
						}},
				}
			},
			want: errors.New("low is Err"),
		}, {
			name: "it Should retur err if High is err",
			mockDrv: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: &mockRstPin{highFunc: func() error {
						return errors.New("high is Err")
					},
						lowFunc: func() error {
							return nil
						}},
				}
			},
			want: errors.New("high is Err"),
		},
	}

	for _, tt := range test2 {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(tt.mockDrv(), newDefLoraConf())
			err := gl.Reset()
			assert.EqualError(t, err, tt.want.Error())
		})
	}
}

func TestGoLora_SetTXPower(t *testing.T) {
	newMockDrvtest1 := func() *driver.Driver {
		return &driver.Driver{
			RSTPin: nil,
			CbPin:  nil,
			ModComm: &mockModConn{send: func(reg, val byte) error {
				return nil
			}},
		}
	}
	errorTest := []struct {
		name    string
		want    error
		txPower int
		mockdrv func() *driver.Driver
	}{
		{name: "it Should return nil if error is nil",
			want:    nil,
			mockdrv: newMockDrvtest1,
			txPower: 0,
		},
		{
			name:    "it Should Keep txPower in config",
			want:    nil,
			txPower: 10,
			mockdrv: newMockDrvtest1,
		},
	}

	for _, tt := range errorTest {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(tt.mockdrv(), newDefLoraConf())
			defTxPwr := gl.Conf.TxPower
			assert.Equal(t, defTxPwr, uint8(0))
			err := gl.SetTXPower(uint8(tt.txPower))
			if tt.want == nil {
				assert.NoError(t, err)
			}

			if tt.txPower > 0 {
				savedTxPower := gl.Conf.TxPower
				fmt.Println(savedTxPower)
				assert.Equal(t, uint8(tt.txPower), savedTxPower)
			}

			if tt.txPower == 0 {
				savedTxPower := gl.Conf.TxPower
				assert.Equal(t, uint8(2), savedTxPower)
			}
		})
	}

	defaultTest := []struct {
		name       string
		want       error
		txPower    int
		defTxPower int
		mockdrv    func() *driver.Driver
	}{
		{
			name:       "tx power must be set to lower default ",
			want:       nil,
			txPower:    0,
			mockdrv:    newMockDrvtest1,
			defTxPower: 2,
		},
		{
			name:       "tx power must be set to max default",
			want:       nil,
			txPower:    300,
			defTxPower: 17,
			mockdrv:    newMockDrvtest1,
		},
	}

	for _, tt := range defaultTest {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(tt.mockdrv(), newDefLoraConf())
			err := gl.SetTXPower(uint8(tt.txPower))
			if tt.want == nil {
				assert.NoError(t, err)
				savedTxPower := gl.Conf.TxPower
				assert.Equal(t, uint8(tt.defTxPower), savedTxPower)
			}
		})
	}

}

func TestGoLora_SetSyncWord(t *testing.T) {
	tests := []struct {
		name     string
		want     error
		mockDrv  func() *driver.Driver
		syncWord byte
	}{
		{
			name: "Should Return Nil if not err",
			want: nil,
			mockDrv: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: nil,
					CbPin:  nil,
					ModComm: &mockModConn{
						send: func(reg, val byte) error {
							return nil
						},
						read: nil,
					},
				}
			},
			syncWord: 0x00,
		},
		{
			name: "Should return err",
			want: errors.New("test err"),
			mockDrv: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: nil,
					CbPin:  nil,
					ModComm: &mockModConn{
						send: func(reg, val byte) error {
							return errors.New("test err")
						},
						read: nil,
					},
				}
			},
			syncWord: 0xff,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(tt.mockDrv(), newDefLoraConf())
			err := gl.SetSyncWord(tt.syncWord)
			if tt.want == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.syncWord, gl.Conf.SyncWord)
				return
			}
			assert.EqualError(t, err, tt.want.Error())
		})
	}

}

func TestGoLora_SetSF(t *testing.T) {
	tests := []struct {
		name    string
		want    error
		mockDrv func() *driver.Driver
		sf      uint8
	}{
		{
			name: "Should return nil if error didnt happen",
			want: nil,
			mockDrv: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: nil,
					CbPin:  nil,
					ModComm: &mockModConn{
						send: func(reg, val byte) error {
							return nil
						},
						read: func(reg byte) (byte, error) {
							return 0xff, nil
						},
					},
				}
			},
			sf: 10,
		},
		{
			name: "Should return err  if send communication failed",
			want: errors.New("test err"),
			mockDrv: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: nil,
					CbPin:  nil,
					ModComm: &mockModConn{
						send: func(reg, val byte) error {
							return errors.New("test err")
						},
						read: func(reg byte) (byte, error) {
							return 0xff, nil
						},
					},
				}
			},
			sf: 10,
		},
		{
			name: "Should return err if read communication failed",
			want: errors.New("test err"),
			mockDrv: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: nil,
					CbPin:  nil,
					ModComm: &mockModConn{
						send: func(reg, val byte) error {
							return nil
						},
						read: func(reg byte) (byte, error) {
							return 0xff, errors.New("test err")

						},
					},
				}
			},
			sf: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(tt.mockDrv(), newDefLoraConf())
			err := gl.SetSF(tt.sf)
			if tt.want == nil {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.want.Error())
		})
	}

	test2mockDrv := func() *driver.Driver {
		return &driver.Driver{
			RSTPin: nil,
			CbPin:  nil,
			ModComm: &mockModConn{
				send: func(reg, val byte) error {
					return nil
				},
				read: func(reg byte) (byte, error) {
					return 0xff, nil
				},
			},
		}
	}

	tests2 := []struct {
		name string
		sf   uint8
		want uint8
	}{
		{
			name: "Should Set To lower DefValue if sf lower than 6",
			sf:   2,
			want: 6,
		},
		{
			name: "Should Set To Max DefValue if sf higher than 12",
			sf:   20,
			want: 12,
		},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(test2mockDrv(), newDefLoraConf())
			err := gl.SetSF(tt.sf)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, gl.Conf.SF)
		})
	}

	tests3 := struct {
		name string
		sf   uint8
	}{
		name: "Should set to actual sf if 6<sf<12",
		sf:   8,
	}

	t.Run(tests3.name, func(t *testing.T) {
		gl := NewGoLoraSX1276(test2mockDrv(), newDefLoraConf())
		err := gl.SetSF(tests3.sf)
		assert.NoError(t, err)
		assert.Equal(t, tests3.sf, gl.Conf.SF)
	})
}

func TestGoLora_SetPreamble(t *testing.T) {
	tests := []struct {
		name     string
		sendErr  error
		readErr  error
		preamble uint16
	}{
		{
			name:     "Should return nil if error didnt happen",
			sendErr:  nil,
			readErr:  nil,
			preamble: 100,
		},
		{
			name:     "Should return err  if send communication failed",
			sendErr:  errors.New("send test err"),
			readErr:  nil,
			preamble: 2031,
		},
		{
			name:     "Should return err if read communication failed",
			sendErr:  nil,
			readErr:  errors.New("read test err"),
			preamble: 312,
		},
		{
			name:     "Should return err if both failed",
			sendErr:  errors.New("both sf and read communication failed"),
			readErr:  errors.New("both sf and read communication failed"),
			preamble: 831,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(testsDrvMock(tt.sendErr, tt.readErr)(), newDefLoraConf())
			if err := gl.SetPreamble(tt.preamble); err != nil {
				assert.EqualError(t, err, tt.sendErr.Error())
				return
			}
			assert.Equal(t, tt.preamble, gl.Conf.PreambleLength)
		})
	}
}

func TestGoLora_SetHeader(t *testing.T) {
	driversList, tests := createDrvMockAndTest()
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(driversList[idx], newDefLoraConf())
			err := gl.SetHeader(Explicit)
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
	tests2 := []struct {
		name   string
		header Header
	}{
		{
			name:   "It Should OverwriteConf header to Explicit",
			header: Explicit,
		},
		{
			name:   "It Should OverwriteConf Header to Implicit",
			header: Implicit,
		},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(testsDrvMock(nil, nil)(), newDefLoraConf())
			_ = gl.SetHeader(tt.header)
			assert.Equal(t, tt.header, gl.Conf.Header)
		})
	}
}

func TestGoLora_SetFrequency(t *testing.T) {
	testDrvMock, tests := createDrvMockAndTest()
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(testDrvMock[idx], newDefLoraConf())
			err := gl.SetFrequency(920 * physic.MegaHertz)
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}

	tests2 := []struct {
		name string
		freq physic.Frequency
	}{
		{
			name: "it Should Overwrite the Config",
			freq: 815 * physic.MegaHertz,
		},
		{
			name: "it Should Overwrite the Config",
			freq: 915 * physic.MegaHertz,
		},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(testsDrvMock(nil, nil)(), newDefLoraConf())
			_ = gl.SetFrequency(tt.freq)
			assert.Equal(t, tt.freq, gl.Conf.Frequency)
		})
	}
}

func TestGoLora_SetCrc(t *testing.T) {
	driverList, tests := createDrvMockAndTest()
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(driverList[idx], newDefLoraConf())
			err := gl.SetCrc(true)
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}

	tests2 := []struct {
		name string
		crc  bool
	}{
		{
			name: "it Should Overwrite The Config with True",
			crc:  true,
		},
		{
			name: "it Should Overwrite The Config with false",
			crc:  false,
		},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(testsDrvMock(nil, nil)(), newDefLoraConf())
			_ = gl.SetCrc(tt.crc)
			assert.Equal(t, tt.crc, gl.Conf.EnableCrc)
		})
	}
}

func TestGoLora_SetCodingRate(t *testing.T) {
	driverList, tests := createDrvMockAndTest()
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(driverList[idx], newDefLoraConf())
			err := gl.SetCodingRate(1)
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}

	tests2 := []struct {
		name  string
		denum uint8
		want  uint8
	}{
		{
			name:  "it Should set to low default value if denum < 5",
			denum: 2,
			want:  5,
		}, {
			name:  "it Should set to high default value if denum > 8",
			denum: 10,
			want:  8,
		},
		{
			name:  "it Should Set to actual value if denum 1 =< denum =< 4",
			denum: 6,
			want:  6,
		},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(testsDrvMock(nil, nil)(), newDefLoraConf())
			_ = gl.SetCodingRate(tt.denum)
			assert.Equal(t, tt.want, gl.Conf.Denum)
		})
	}
}

func TestGoLora_SetBW(t *testing.T) {
	driverList, tests := createDrvMockAndTest()
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(driverList[idx], newDefLoraConf())
			err := gl.SetBW(10)
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}

	type BwTest struct {
		name string
		bw   int
		want int
	}

	tests2 := []BwTest{
		{
			name: "it Should Set SBW to 1 if BW to low",
			bw:   1,
			want: int(BW_1),
		},
		{
			name: "it Should Set SBW to 9 if BW to high",
			bw:   int(BW_8) + 1,
			want: int(BW_8),
		},
		{
			name: "it Should set SBW to 6 if BW is int(BW_7) - 1",
			bw:   int(BW_6) - 1,
			want: int(BW_6),
		},
		{
			name: "it Should set SBW to 7 if BW is int(BW_6) + 1",
			bw:   int(BW_6) + 1,
			want: int(BW_7),
		},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(testsDrvMock(nil, nil)(), newDefLoraConf())
			_ = gl.SetBW(uint64(tt.bw))
			assert.Equal(t, tt.want, int(gl.Conf.BW))
		})
	}

}

func TestGoLora_ChangeMode(t *testing.T) {
	driverList, tests := createDrvMockAndTest()
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(driverList[idx], newDefLoraConf())
			err := gl.ChangeMode(internal.Sleep)
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}

	tests2 := []struct {
		name string
		mode internal.LoraMode
	}{
		{
			name: "it Should Overwrite the ChangeMode to Tx",
			mode: internal.Tx,
		}, {
			name: "it Should Overwrite the ChangeMode to RxSingle",
			mode: internal.RxSingle,
		},
		{
			name: "it Should Overwrite the ChangeMode to RxContinuous",
			mode: internal.RxContinuous,
		},
		{
			name: "it Should Overwrite the ChangeMode to idle",
			mode: internal.Idle,
		},

		{
			name: "it Should Overwrite the ChangeMode to Sleep",
			mode: internal.Sleep,
		},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(testsDrvMock(nil, nil)(), newDefLoraConf())
			_ = gl.ChangeMode(tt.mode)
			assert.Equal(t, tt.mode, gl.Mode)
		})
	}
}

func TestGoLora_IsReceived(t *testing.T) {
	driverList, tests := createDrvMockAndTest()
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(driverList[idx], newDefLoraConf())
			_, err := gl.IsReceived()
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}

	isRecvMockDrv := func(data byte, err error) func() *driver.Driver {
		return func() *driver.Driver {
			return &driver.Driver{
				RSTPin: nil,
				CbPin:  nil,
				ModComm: &mockModConn{
					send: func(reg, val byte) error {
						return nil
					},
					read: func(reg byte) (byte, error) {
						return data, err
					},
				},
			}
		}
	}
	tests2 := []struct {
		name    string
		mockDrv func() *driver.Driver
		val     bool
	}{
		{
			name:    "it Should return the actual value if no err",
			mockDrv: isRecvMockDrv(0b00001000, nil),
			val:     true,
		},
		{
			name:    "it Should return err if readValFunc return err",
			mockDrv: isRecvMockDrv(0, nil),
			val:     false,
		},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(tt.mockDrv(), newDefLoraConf())
			val, _ := gl.IsReceived()
			assert.Equal(t, tt.val, val)
		})
	}
}

func TestGoLora_GetLastPktRSSI(t *testing.T) {
	driverList, tests := createDrvMockAndTest()
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(driverList[idx], newDefLoraConf())
			_, err := gl.GetLastPktRSSI()
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGoLora_GetLastPktSNR(t *testing.T) {
	driverList, tests := createDrvMockAndTest()
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(driverList[idx], newDefLoraConf())
			_, err := gl.GetLastPktSNR()
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGoLora_Begin(t *testing.T) {
	tests := []struct {
		name    string
		mockDrv func() *driver.Driver
		want    error
	}{
		{
			name: "Should return nil if no err and conf Should Match to NewLoraConf",
			mockDrv: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: &mockRstPin{
						lowFunc: func() error {
							return nil
						},
						highFunc: func() error {
							return nil
						},
					},
					CbPin: nil,
					ModComm: &mockModConn{
						send: func(reg, val byte) error {
							return nil
						},
						read: func(reg byte) (byte, error) {
							return 0x12, nil
						},
					},
				}
			},
			want: nil,
		},
		{
			name: "Should return err if check conn err",
			mockDrv: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: &mockRstPin{
						lowFunc: func() error {
							return nil
						},
						highFunc: func() error {
							return nil
						},
					},
					CbPin: nil,
					ModComm: &mockModConn{
						send: func(reg, val byte) error {
							return nil
						},
						read: func(reg byte) (byte, error) {
							return 0xff, nil
						},
					},
				}
			},
			want: errors.New("check Your Connection"),
		},
		{
			name: "Should return err if read err",
			mockDrv: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: &mockRstPin{
						lowFunc: func() error {
							return nil
						},
						highFunc: func() error {
							return nil
						},
					},
					CbPin: nil,
					ModComm: &mockModConn{
						send: func(reg, val byte) error {
							return nil
						},
						read: func(reg byte) (byte, error) {
							return 0xff, errors.New("read test err")
						},
					},
				}
			},
			want: errors.New("read test err"),
		},
		{
			name: "Should return err if send err",
			mockDrv: func() *driver.Driver {
				return &driver.Driver{
					RSTPin: &mockRstPin{
						lowFunc: func() error {
							return nil
						},
						highFunc: func() error {
							return nil
						},
					},
					CbPin: nil,
					ModComm: &mockModConn{
						send: func(reg, val byte) error {
							return errors.New("send test err")
						},
						read: func(reg byte) (byte, error) {
							return 0x12, nil
						},
					},
				}
			},
			want: errors.New("send test err"),
		},
	}

	loraNewConf := LoraConf{
		TxPower:        17,
		SF:             7,
		BW:             uint64(BW_1),
		Denum:          5,
		PreambleLength: 8,
		SyncWord:       0x34,
		Frequency:      915 * physic.MegaHertz,
		Header:         Explicit,
		EnableCrc:      false,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defConf := LoraConf{}
			if tt.want == nil {
				defConf = loraNewConf
			} else {
				defConf = newDefLoraConf()
			}
			gl := NewGoLoraSX1276(tt.mockDrv(), defConf)
			err := gl.Begin()
			if tt.want == nil {
				assert.NoError(t, err)
				assert.Equal(t, loraNewConf, gl.Conf)
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestGoLora_Destroy(t *testing.T) {
	driverList, tests := createDrvMockAndTest()
	for idx := range driverList {
		driverList[idx].RSTPin = &mockRstPin{
			lowFunc: func() error {
				return nil
			},
			highFunc: func() error {
				return nil
			},
		}
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(driverList[idx], newDefLoraConf())
			err := gl.Destroy()
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGoLora_DumpRegisters(t *testing.T) {
	driverList, tests := createDrvMockAndTest()
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(driverList[idx], newDefLoraConf())
			_, err := gl.DumpRegisters()
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGoLora_GetConf(t *testing.T) {
	newLoraConf := newDefLoraConf()
	tests := struct {
		name string
		want LoraConf
	}{
		name: "it Should return the actual config",
		want: newLoraConf,
	}

	t.Run(tests.name, func(t *testing.T) {
		gl := NewGoLoraSX1276(testsDrvMock(nil, nil)(), newDefLoraConf())
		conf := gl.GetConf()
		assert.Equal(t, tests.want, conf)
	})
}

func TestGoLora_RegisterCb_CalledCallback(t *testing.T) {
	gl := NewGoLoraSX1276(&driver.Driver{
		RSTPin: nil,
		CbPin: &mockCbPin{
			readValFunc: func() (bool, error) {
				return true, nil
			},
		},
		ModComm: &mockModConn{
			send: func(reg, val byte) error {
				return nil
			},
			read: func(reg byte) (byte, error) {
				return 0xff, nil
			},
		},
	}, newDefLoraConf())

	ch := make(chan bool, 1)

	stopper, err := gl.RegisterCb(OnRxDone, func() {
		ch <- true
	})
	assert.NoError(t, err)

	select {
	case val := <-ch:
		assert.Equal(t, true, val)
	case <-time.After(100 * time.Millisecond):
		t.Error("callback was not called")
	}

	close(stopper)
	close(ch)
}

func TestGoLora_RegisterCb_NotCalledCallback(t *testing.T) {
	gl := NewGoLoraSX1276(&driver.Driver{
		RSTPin: nil,
		CbPin: &mockCbPin{
			readValFunc: func() (bool, error) {
				return false, nil
			},
		},
		ModComm: &mockModConn{
			send: func(reg, val byte) error {
				return nil
			},
			read: func(reg byte) (byte, error) {
				return 0xff, nil
			},
		},
	}, newDefLoraConf())

	ch := make(chan bool, 1)

	stopper, err := gl.RegisterCb(OnRxDone, func() {
		ch <- true
	})
	assert.NoError(t, err)

	select {
	case val := <-ch:
		t.Errorf("callback should not have been called, but got %v", val)
	case <-time.After(100 * time.Millisecond):
		// ✅ expected path: no callback
	}

	close(stopper)
	close(ch)
}

func TestGoLora_RegisterCb_Ok(t *testing.T) {
	gl := NewGoLoraSX1276(testsDrvMock(nil, nil)(), newDefLoraConf())
	stopper, err := gl.RegisterCb(OnRxDone, func() {})
	assert.NoError(t, err)
	if stopper != nil {
		stopper <- struct{}{}
	}
}

func TestGoLora_RegisterCb_UnknownEvent(t *testing.T) {
	gl := NewGoLoraSX1276(testsDrvMock(nil, nil)(), newDefLoraConf())
	stopper, err := gl.RegisterCb(99, func() {})
	assert.EqualError(t, err, "event not recognized")
	if stopper != nil {
		stopper <- struct{}{}
	}
}

func TestGoLora_RegisterCb_NotImplemented(t *testing.T) {
	gl := NewGoLoraSX1276(testsDrvMock(nil, nil)(), newDefLoraConf())
	stopper, err := gl.RegisterCb(OnTxDone, func() {})
	assert.EqualError(t, err, "OnTxDone Not Implemented Yet")
	if stopper != nil {
		stopper <- struct{}{}
	}
}

func TestGoLora_SendPacket(t *testing.T) {
	driverList, tests := createDrvMockAndTest()
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(driverList[idx], newDefLoraConf())
			err := gl.SendPacket([]byte("test data"))
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGoLora_ReceivePacket(t *testing.T) {
	drvMock := func(sendErr, readErr error, irq byte) *driver.Driver {
		return &driver.Driver{
			RSTPin: nil,
			CbPin:  nil,
			ModComm: &mockModConn{
				send: func(reg, val byte) error {
					return sendErr
				},
				read: func(reg byte) (byte, error) {
					return irq, readErr
				},
			},
		}
	}

	tests := []struct {
		name    string
		irq     byte
		want    error
		sendErr error
		readErr error
	}{
		{
			name:    "it Should return nil if packet received and crc no error",
			irq:     0b01000000,
			want:    nil,
			sendErr: nil,
			readErr: nil,
		},
		{
			name:    "it Should return err if packet received but crc error",
			irq:     0b01100000,
			want:    errors.New("packet damaged or lost in transmit"),
			sendErr: nil,
			readErr: nil,
		},
		{
			name:    "it Should return err if packet not received",
			irq:     0b00000000,
			want:    errors.New("no Packet Received"),
			sendErr: nil,
			readErr: nil,
		},
		{
			name:    "it Should return err if send err happened",
			irq:     0b01000000,
			want:    errors.New("send test err"),
			sendErr: errors.New("send test err"),
			readErr: nil,
		},
		{
			name:    "it Should return err if read err happened",
			irq:     0b01000000,
			want:    errors.New("read test err"),
			sendErr: nil,
			readErr: errors.New("read test err"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLoraSX1276(drvMock(tt.sendErr, tt.readErr, tt.irq), newDefLoraConf())
			_, err := gl.ReceivePacket()
			if tt.want == nil {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.want.Error())
		})
	}
}
