package SX1276

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Fsyahputra/GoLora/Lora/SX1276/internal"
	"github.com/Fsyahputra/GoLora/driver"
	"periph.io/x/conn/v3/physic"
)

type LoraConf struct {
	TxPower        uint8
	SF             uint8
	BW             uint8
	CodingRate     uint8
	PreambleLength uint8
	SyncWord       uint8
	Frequency      physic.Frequency
	Header         Header
	EnableCrc      bool
}
type GoLora struct {
	driver.ModComm
	driver.RSTPin
	driver.CbPin
	internal.LoraUtils
	Conf      LoraConf
	mu        sync.Mutex
	cb        func()
	cbStopper chan struct{}
}

type RegVal struct {
	Reg byte
	Val byte
}

func NewGoLoraSX1276(modComm driver.ModComm, rstPin driver.RSTPin, conf LoraConf) *GoLora {
	gl := &GoLora{
		ModComm: modComm,
		RSTPin:  rstPin,
		Conf:    conf,
	}
	return gl
}

func (gl *GoLora) readReg(reg byte) (byte, error) {
	readReg := gl.SetReadMask(reg)
	rx, err := gl.ModComm.ReadFromMod(readReg)
	if err != nil {
		return 0, err
	}
	return rx, nil
}

func (gl *GoLora) writeReg(reg byte, value uint8) error {
	writeReg := gl.SetWriteMask(reg)
	byteValue := value
	if err := gl.ModComm.SendToMod(writeReg, byteValue); err != nil {
		return err
	}
	return nil
}

func (gl *GoLora) Reset() {
	gl.RSTPin.Low()
	time.Sleep(300 * time.Millisecond)
	gl.RSTPin.High()
	time.Sleep(1000 * time.Millisecond)
}

func (gl *GoLora) ChangeMode(mode internal.LoraMode) {
	modeVal := gl.LoraUtils.ChangeMode(mode)
	if err := gl.writeReg(internal.REG_OP_MODE, modeVal); err != nil {
		return
	}
}

func (gl *GoLora) SetTXPower(txPower uint8) {
	gl.Conf.TxPower = txPower
	tx := 0
	if txPower < 2 {
		fmt.Println("txPower Too Low set default tx=2")
		tx = 2
	} else if txPower > 17 {
		fmt.Println("txPower Too High set to 17")
		tx = 17
	}
	tx = tx - 2
	txReg := gl.LoraUtils.SetTxPower(byte(tx))
	if err := gl.writeReg(internal.REG_PA_CONFIG, txReg); err != nil {
		return
	}
}

func (gl *GoLora) writeRegMany(Regs []byte, Values []byte) error {
	if len(Regs) != len(Values) {
		return errors.New("len(Regs) != len(Values)")
	}
	type regValue struct {
		Reg   byte
		Value byte
	}
	regValues := make([]regValue, len(Regs))
	for i, Reg := range Regs {
		regValues[i] = regValue{Reg: Reg, Value: Values[i]}
	}
	for _, reg := range regValues {
		if err := gl.writeReg(reg.Reg, reg.Value); err != nil {
			return err
		}
	}
	return nil
}

func (gl *GoLora) SetFrequency(freq physic.Frequency) {
	gl.Conf.Frequency = freq
	frf := (uint64(freq) << 19) / 32000000
	freqBytes := gl.LoraUtils.SetFreq(byte(frf))
	registers := []byte{internal.REG_FRF_MSB, internal.REG_FRF_MID, internal.REG_FRF_LSB}
	gl.mu.Lock()
	defer gl.mu.Unlock()
	if err := gl.writeRegMany(registers, freqBytes); err != nil {
		return
	}
}

func (gl *GoLora) SetSF(sf uint8) {

	if sf < 6 {
		sf = 6
		fmt.Println("SF Too low set to 6")
	} else if sf > 12 {
		sf = 12
		fmt.Println("SF Too high set to 12")
	}
	gl.Conf.SF = sf
	sfReg := gl.LoraUtils.SetSF(sf)
	values := []byte{sfReg, internal.SF_DEF_OPTIMIZE, internal.SF_DEF_THRESHOLD}
	registers := []byte{internal.REG_MODEM_CONFIG_2, internal.REG_DETECTION_OPTIMIZE, internal.REG_DETECTION_THRESHOLD}

	gl.mu.Lock()
	defer gl.mu.Unlock()
	if sf == 6 {
		values[1] = internal.SF_6_OPTIMIZE
		values[2] = internal.SF_6_THRESHOLD
	}
	_ = gl.writeRegMany(registers, values)
}

func (gl *GoLora) SetBW(bw uint64) {
	var sbw uint8
	bwMaps := map[BW]uint8{
		BW_1: 1,
		BW_2: 2,
		BW_3: 3,
		BW_4: 4,
		BW_5: 5,
		BW_6: 6,
		BW_7: 7,
		BW_8: 8,
	}

	for thres, value := range bwMaps {
		if bw < uint64(thres) {
			sbw = value
		} else {
			sbw = 9
		}
	}
	gl.Conf.BW = sbw
	bwReg := gl.LoraUtils.SetBW(sbw)
	gl.mu.Lock()
	defer gl.mu.Unlock()
	_ = gl.writeReg(internal.REG_MODEM_CONFIG_1, bwReg)

}

func (gl *GoLora) SetCrc(enable bool) {
	crcReg := gl.LoraUtils.SetCrc(enable)
	gl.mu.Lock()
	defer gl.mu.Unlock()
	_ = gl.writeReg(internal.REG_MODEM_CONFIG_2, crcReg)
}

func (gl *GoLora) SetPreamble(length uint16) {
	preambleReg := gl.LoraUtils.SetPreamble(length)
	registers := []byte{internal.REG_PREAMBLE_MSB, internal.REG_PREAMBLE_LSB}
	gl.mu.Lock()
	defer gl.mu.Unlock()
	_ = gl.writeRegMany(registers, preambleReg)
}

func (gl *GoLora) SetSyncWord(syncWord uint8) {
	syncWordValue := gl.LoraUtils.SetSyncWord(syncWord)
	gl.mu.Lock()
	defer gl.mu.Unlock()
	_ = gl.writeReg(internal.REG_SYNC_WORD, syncWordValue)
}

func (gl *GoLora) CheckConn() error {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	version, err := gl.readReg(internal.REG_VERSION)
	if err != nil {
		return err
	}
	if version != 0x12 {
		return errors.New("Check Your Connection")
	}
	return nil
}

func (gl *GoLora) setFifoPtr(ptr uint8) {
	_ = gl.writeReg(internal.REG_FIFO_ADDR_PTR, ptr)
}

func (gl *GoLora) sendToFifo(buff []byte) error {
	for _, data := range buff {
		if err := gl.writeReg(internal.REG_FIFO, data); err != nil {
			return err
		}
	}
	return nil
}

func (gl *GoLora) waitTxDone() error {
	gl.mu.Lock()
	readVal, err := gl.readReg(internal.REG_IRQ_FLAGS)
	if err != nil {
		return err
	}
	gl.mu.Unlock()

	for readVal&internal.IRQ_TX_DONE_MASK == 0 {
		time.Sleep(100 * time.Millisecond)
	}
	gl.mu.Lock()
	if err := gl.writeReg(internal.REG_IRQ_FLAGS, internal.IRQ_TX_DONE_MASK); err != nil {
		return err
	}
	gl.mu.Unlock()
	return nil
}

func (gl *GoLora) SendPacket(buff []byte) error {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.ChangeMode(internal.Idle)
	gl.setFifoPtr(0)
	if err := gl.sendToFifo(buff); err != nil {
		return err
	}
	buffSize := byte(len(buff))
	if err := gl.writeReg(internal.REG_PAYLOAD_LENGTH, buffSize); err != nil {
		return err
	}

	gl.ChangeMode(internal.Tx)
	if err := gl.waitTxDone(); err != nil {
		return err
	}
	return nil
}

func (gl *GoLora) ReceivePacket() ([]byte, error) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	irq, err := gl.readReg(internal.REG_IRQ_FLAGS)
	if err != nil {
		return nil, err
	}

	if err := gl.LoraUtils.CheckData(irq); err != nil {
		return nil, err
	}

	pktLen, err := gl.LoraUtils.GetPktLenByHeader(bool(gl.Conf.Header))
	if err != nil {
		return nil, err
	}

	currentPtr, err := gl.readReg(internal.REG_FIFO_RX_CURRENT_ADDR)
	if err != nil {
		return nil, err
	}

	if err := gl.writeReg(internal.REG_FIFO_ADDR_PTR, currentPtr); err != nil {
		return nil, err
	}

	data := make([]byte, pktLen)
	for i := 0; uint8(i) < pktLen; i++ {
		data[i], err = gl.readReg(internal.REG_FIFO)
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (gl *GoLora) SetCodingRate(denum uint8) {
	if denum < 5 {
		denum = 5
	} else if denum > 8 {
		denum = 8
	}

	var cr uint8 = denum - 4
	modemConf, _ := gl.readReg(internal.REG_MODEM_CONFIG_2)
	crReg := gl.LoraUtils.SetCodingRate(cr, modemConf)
	gl.mu.Lock()
	defer gl.mu.Unlock()
	_ = gl.writeReg(internal.REG_MODEM_CONFIG_1, crReg)
}

func (gl *GoLora) IsReceived() (bool, error) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	data, err := gl.readReg(internal.REG_IRQ_FLAGS & internal.IRQ_RX_DONE_MASK)
	isExists := false
	if err != nil {
		return false, err
	}

	if data != 0 {
		isExists = true
	} else {
		isExists = false
	}

	return isExists, nil
}

func timer(millis time.Duration, ch chan bool) {
	ch <- false
	for i := 0; i < int(millis); i++ {
		time.Sleep(time.Millisecond)
	}
	ch <- true
	close(ch)
}

func (gl *GoLora) waitForInterrupt(millis time.Duration) error {
	timeOutChan := make(chan bool)
	go timer(millis, timeOutChan)
	for {
		select {
		case isTimeOut := <-timeOutChan:
			if isTimeOut {
				return errors.New("timeout reached")
			}
		default:

		}
		ok, err := gl.CbPin.ReadVal()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
}

func (gl *GoLora) waitForPacket(millis time.Duration) error {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.ChangeMode(internal.Idle)
	err := gl.writeReg(internal.REG_IRQ_FLAGS, 0x9f)
	if err != nil {
		return err
	}
	err = gl.writeReg(internal.REG_DIO_MAPPING_1, 0x00)
	if err != nil {
		return err
	}
	gl.ChangeMode(internal.RxContinuous)
	if err := gl.waitForInterrupt(millis); err != nil {
		return err
	}
	return nil
}

func (gl *GoLora) rxDoneWrapper() func() bool {
	return func() bool {
		err := gl.waitForPacket(1)
		if err != nil {
			return false
		}
		return true
	}
}
func (gl *GoLora) eventChecker(event Event) (func() bool, error) {
	var checkerFunc func() bool = nil
	var err error
	if event == OnRxDone {
		checkerFunc = gl.rxDoneWrapper()
	} else {
		err = errors.New("Not Implemented")
	}
	return checkerFunc, err
}

func (gl *GoLora) runCb() {
	if gl.cb != nil {
		gl.cb()
	}
}

func (gl *GoLora) cbDaemon(eventChecker func() bool, ch chan struct{}) {
	t := time.NewTicker(1 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ch:
			return
		case <-t.C:
			isHappen := eventChecker()
			if isHappen {
				gl.runCb()
			} else {
				continue
			}
		}
	}
}

func (gl *GoLora) RegisterCb(event Event, cb func()) (chan struct{}, error) {
	thStopper := make(chan struct{})
	checkerFunc, err := gl.eventChecker(event)
	if err != nil {
		return nil, err
	}
	gl.cb = cb
	gl.cbStopper = thStopper
	go gl.cbDaemon(checkerFunc, thStopper)
	return thStopper, nil
}

func (gl *GoLora) GetLastPktRSSI() (uint8, error) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	rssi, err := gl.readReg(internal.REG_PKT_RSSI_VALUE)
	if err != nil {
		return 0, err
	}
	return rssi, nil
}

func (gl *GoLora) GetLastPktSNR() (uint8, error) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	snr, err := gl.readReg(internal.REG_PKT_SNR_VALUE)
	if err != nil {
		return 0, err
	}
	return snr, nil
}

func (gl *GoLora) Destroy() {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.ChangeMode(internal.Sleep)

	if gl.cb != nil {
		close(gl.cbStopper)
		gl.cb = nil
	}
}

func (gl *GoLora) DumpRegisters() ([]RegVal, error) {
	regVal := make([]RegVal, 0)
	registers := make([]byte, 0)
	values := make([]byte, 0)
	var err error
	for i := 0; i < 0x26; i++ {
		registers = append(registers, byte(i))
	}

	for idx, reg := range registers {
		values[idx], err = gl.readReg(reg)
		regVal[idx] = RegVal{
			Reg: reg,
			Val: values[idx],
		}
	}

	if err != nil {
		return nil, err
	}

	return regVal, nil
}
