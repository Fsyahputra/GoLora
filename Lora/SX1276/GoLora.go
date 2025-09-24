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
	BW             uint64
	Denum          uint8
	PreambleLength uint16
	SyncWord       uint8
	Frequency      physic.Frequency
	Header         Header
	EnableCrc      bool
}
type GoLora struct {
	*driver.Driver
	*LoraUtils
	Conf      LoraConf
	mu        sync.Mutex
	cb        func()
	cbStopper chan struct{}
	Mode      LoraMode
}

type RegVal struct {
	Reg byte
	Val byte
}

func (gl *GoLora) configure() error {
	var err error
	err = gl.SetTXPower(gl.Conf.TxPower)
	err = gl.SetSF(gl.Conf.SF)
	err = gl.SetBW(gl.Conf.BW)
	err = gl.SetCodingRate(gl.Conf.Denum)
	err = gl.SetPreamble(gl.Conf.PreambleLength)
	err = gl.SetSyncWord(gl.Conf.SyncWord)
	err = gl.SetFrequency(gl.Conf.Frequency)
	err = gl.SetHeader(gl.Conf.Header)
	err = gl.SetCrc(gl.Conf.EnableCrc)
	return err
}

func NewGoLoraSX1276(drv *driver.Driver, conf LoraConf) *GoLora {
	gl := &GoLora{
		Driver:    drv,
		LoraUtils: &LoraUtils{},
		Conf:      conf,
		mu:        sync.Mutex{},
		cb:        nil,
		cbStopper: nil,
	}

	return gl
}

func (gl *GoLora) Begin() error {
	if err := gl.Reset(); err != nil {
		return err
	}
	gl.mu.Lock()

	modVer, err := gl.readReg(internal.REG_VERSION)
	if err != nil {
		return err
	}
	if modVer != 0x12 {
		return fmt.Errorf("unsupported module version: got 0x%X", modVer)
	}
	if err := gl.ChangeMode(Sleep); err != nil {
		return fmt.Errorf("failed to set sleep mode: %w", err)
	}
	currentLna, err := gl.readReg(internal.REG_LNA)
	if err != nil {
		return err
	}
	registers := []byte{internal.REG_FIFO_RX_BASE_ADDR, internal.REG_FIFO_TX_BASE_ADDR, internal.REG_LNA, internal.REG_MODEM_CONFIG_3}
	values := []byte{0, 0, currentLna | 0x03, 0x04}
	if err := gl.writeRegMany(registers, values); err != nil {
		return err
	}
	gl.mu.Unlock()
	if err := gl.configure(); err != nil {
		return err
	}
	gl.mu.Lock()
	if err = gl.ChangeMode(Idle); err != nil {
		return fmt.Errorf("failed to set Idle mode: %w", err)
	}
	gl.mu.Unlock()
	return nil
}

func (gl *GoLora) readReg(reg byte) (byte, error) {
	readReg := gl.setReadMask(reg)

	rx, err := gl.ModComm.ReadFromMod(readReg)
	if err != nil {
		return 0, err
	}
	return rx, nil
}

func (gl *GoLora) writeReg(reg byte, value uint8) error {
	writeReg := gl.setWriteMask(reg)
	byteValue := value
	if err := gl.ModComm.SendToMod(writeReg, byteValue); err != nil {
		return err
	}
	return nil
}

func (gl *GoLora) Reset() error {
	err := gl.RSTPin.Low()
	if err != nil {
		return err
	}
	time.Sleep(300 * time.Millisecond)
	err = gl.RSTPin.High()
	if err != nil {
		return err
	}
	time.Sleep(1000 * time.Millisecond)
	return nil
}

func (gl *GoLora) ChangeMode(mode LoraMode) error {
	modeVal := gl.LoraUtils.changeMode(mode)
	if err := gl.writeReg(internal.REG_OP_MODE, modeVal); err != nil {
		return err
	}
	gl.Mode = mode
	return nil
}

func (gl *GoLora) SetTXPower(txPower uint8) error {
	var tx uint8
	if txPower < 2 {
		fmt.Println("txPower Too Low set default tx=2")
		tx = 2
	} else if txPower > 17 {
		fmt.Println("txPower Too High set to 17")
		tx = 17
	} else {
		tx = txPower
	}

	chipTx := tx - 2
	txReg := gl.LoraUtils.setTxPower(byte(chipTx))
	gl.mu.Lock()
	defer gl.mu.Unlock()
	if err := gl.writeReg(internal.REG_PA_CONFIG, txReg); err != nil {
		return err
	}
	fmt.Printf("tx %v", tx)
	gl.Conf.TxPower = tx
	return nil
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

func (gl *GoLora) SetFrequency(freq physic.Frequency) error {
	gl.Conf.Frequency = freq
	frf := (uint64(freq) << 19) / 32000000
	freqBytes := gl.LoraUtils.setFreq(frf)
	registers := []byte{internal.REG_FRF_MSB, internal.REG_FRF_MID, internal.REG_FRF_LSB}
	gl.mu.Lock()
	defer gl.mu.Unlock()
	if err := gl.writeRegMany(registers, freqBytes); err != nil {
		return err
	}
	return nil
}

func (gl *GoLora) SetSF(sf uint8) error {

	if sf <= 6 {
		sf = 6
		fmt.Println("SF Too low set to 6")
	} else if sf >= 12 {
		sf = 12
		fmt.Println("SF Too high set to 12")
	}
	gl.Conf.SF = sf
	sfReg := gl.LoraUtils.setSF(sf)

	gl.mu.Lock()
	defer gl.mu.Unlock()
	currentConf, err := gl.readReg(internal.REG_MODEM_CONFIG_2)
	if err != nil {
		return err
	}
	currentConfFourtMsb := currentConf & 0x0f
	overWritenConf := sfReg | currentConfFourtMsb
	values := []byte{overWritenConf, internal.SF_DEF_OPTIMIZE, internal.SF_DEF_THRESHOLD}
	registers := []byte{internal.REG_MODEM_CONFIG_2, internal.REG_DETECTION_OPTIMIZE, internal.REG_DETECTION_THRESHOLD}
	if sf == 6 {
		values[1] = internal.SF_6_OPTIMIZE
		values[2] = internal.SF_6_THRESHOLD
	}
	err = gl.writeRegMany(registers, values)
	if err != nil {
		return err
	}
	return nil
}

func (gl *GoLora) SetBW(bw uint64) error {
	var sbw uint8

	var threshold uint64
	type bwVal struct {
		bwThres BW
		value   byte
	}
	bwVals := []bwVal{
		{bwThres: BW_1, value: 1},
		{bwThres: BW_2, value: 2},
		{bwThres: BW_3, value: 3},
		{bwThres: BW_4, value: 4},
		{bwThres: BW_5, value: 5},
		{bwThres: BW_6, value: 6},
		{bwThres: BW_7, value: 7},
		{bwThres: BW_8, value: 8},
	}

	for _, bwval := range bwVals {
		if bw <= uint64(bwval.bwThres) {
			sbw = bwval.value
			threshold = uint64(bwval.bwThres)
			break
		} else {
			threshold = uint64(BW_8)
			sbw = 9
		}
	}
	gl.mu.Lock()
	defer gl.mu.Unlock()
	bwReg := gl.LoraUtils.setBW(sbw)
	currentConf, err := gl.readReg(internal.REG_MODEM_CONFIG_1)
	currentConfFourthMSB := currentConf & 0x0f
	overWrittenConf := currentConfFourthMSB | bwReg
	if err != nil {
		return err
	}
	if err := gl.writeReg(internal.REG_MODEM_CONFIG_1, overWrittenConf); err != nil {
		return err
	}
	gl.Conf.BW = threshold
	return nil
}

func (gl *GoLora) SetCrc(enable bool) error {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	currentModemConf, err := gl.readReg(internal.REG_MODEM_CONFIG_2)
	if err != nil {
		return err
	}
	updatedConf := gl.LoraUtils.setCrc(enable, currentModemConf)
	if err := gl.writeReg(internal.REG_MODEM_CONFIG_2, updatedConf); err != nil {
		return err
	}
	gl.Conf.EnableCrc = enable
	return nil
}

func (gl *GoLora) SetPreamble(length uint16) error {
	preambleReg := gl.LoraUtils.setPreamble(length)
	registers := []byte{internal.REG_PREAMBLE_MSB, internal.REG_PREAMBLE_LSB}
	gl.mu.Lock()
	defer gl.mu.Unlock()
	if err := gl.writeRegMany(registers, preambleReg); err != nil {
		return err
	}
	gl.Conf.PreambleLength = length
	return nil
}

func (gl *GoLora) SetSyncWord(syncWord uint8) error {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	if err := gl.writeReg(internal.REG_SYNC_WORD, syncWord); err != nil {
		return err
	}
	gl.Conf.SyncWord = syncWord
	return nil
}

func (gl *GoLora) CheckConn() error {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	version, err := gl.readReg(internal.REG_VERSION)
	if err != nil {
		return err
	}
	if version != 0x12 {
		return errors.New("check Your Connection")
	}
	return nil
}

func (gl *GoLora) setFifoPtr(ptr uint8) error {

	if err := gl.writeReg(internal.REG_FIFO_ADDR_PTR, ptr); err != nil {
		return err
	}
	return nil
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
	readVal, err := gl.readReg(internal.REG_IRQ_FLAGS)
	if err != nil {
		return err
	}

	for readVal&internal.IRQ_TX_DONE_MASK == 0 {
		time.Sleep(100 * time.Millisecond)
	}
	if err := gl.writeReg(internal.REG_IRQ_FLAGS, internal.IRQ_TX_DONE_MASK); err != nil {
		return err
	}
	return nil
}

func (gl *GoLora) SendPacket(buff []byte) error {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	if err := gl.ChangeMode(Idle); err != nil {
		return err
	}
	if err := gl.setFifoPtr(0); err != nil {
		return err
	}
	if err := gl.sendToFifo(buff); err != nil {
		return err
	}
	buffSize := byte(len(buff))
	if err := gl.writeReg(internal.REG_PAYLOAD_LENGTH, buffSize); err != nil {
		return err
	}
	if err := gl.ChangeMode(Tx); err != nil {
		return err
	}
	if err := gl.waitTxDone(); err != nil {
		return err
	}
	return nil
}

func (gl *GoLora) SetHeader(header Header) error {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.Conf.Header = header
	currentConf, err := gl.readReg(internal.REG_MODEM_CONFIG_1)
	if err != nil {
		return err
	}
	newConf := gl.LoraUtils.setHeader(bool(header), currentConf)
	if err := gl.writeReg(internal.REG_MODEM_CONFIG_1, newConf); err != nil {
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
	time.Sleep(1 * time.Second)
	if err := gl.LoraUtils.checkData(irq); err != nil {
		return nil, err
	}

	if err := gl.ChangeMode(Idle); err != nil {
		return nil, err
	}

	pktLen := byte(0)
	if gl.Conf.Header {
		pktLen, err = gl.readReg(internal.REG_RX_NB_BYTES)
	} else {
		pktLen, err = gl.readReg(internal.REG_PAYLOAD_LENGTH)
	}

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
	data := make([]byte, int(pktLen))
	for i := 0; uint8(i) < pktLen; i++ {
		data[i], err = gl.readReg(internal.REG_FIFO)
	}
	time.Sleep(1000 * time.Millisecond)
	if err := gl.writeReg(internal.REG_IRQ_FLAGS, 0xff); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	_ = gl.ChangeMode(RxContinuous)

	return data, nil
}

func (gl *GoLora) SetCodingRate(denum uint8) error {
	if denum < 5 {
		fmt.Println("Coding Rate To Low Setting to Low Default Value 5")
		denum = 5
	} else if denum > 8 {
		fmt.Println("Coding Rate To High Setting to High Default Value 8")
		denum = 8
	}
	var cr = denum - 4
	gl.mu.Lock()
	defer gl.mu.Unlock()
	currentModemConf, err := gl.readReg(internal.REG_MODEM_CONFIG_1)
	if err != nil {
		return err
	}
	crReg := gl.LoraUtils.setCodingRate(cr, currentModemConf)
	if err := gl.writeReg(internal.REG_MODEM_CONFIG_1, crReg); err != nil {
		return err
	}
	gl.Conf.Denum = denum
	return nil
}

func (gl *GoLora) IsReceived() (bool, error) {
	gl.mu.Lock()
	data, err := gl.readReg(internal.REG_IRQ_FLAGS)
	gl.mu.Unlock()
	data = data & internal.IRQ_RX_DONE_MASK
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

func (gl *GoLora) waitForInterrupt(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		ok, err := gl.CbPin.ReadVal()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("timeout reached")
		}
		time.Sleep(time.Millisecond)
	}
}

func (gl *GoLora) waitForPacket(millis time.Duration) error {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	//if err := gl.ChangeMode(Idle); err != nil {
	//	return err
	//}
	//
	//if err := gl.writeReg(internal.REG_DIO_MAPPING_1, 0x00); err != nil {
	//	return err
	//}

	if err := gl.ChangeMode(RxContinuous); err != nil {
		return err
	}
	if err := gl.waitForInterrupt(millis); err != nil {
		return err
	}
	return nil
}

func (gl *GoLora) rxDoneWrapper() func() bool {
	return func() bool {
		err := gl.waitForPacket(1000 * time.Millisecond)
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
	} else if event == OnTxDone {
		err = errors.New("OnTxDone Not Implemented Yet")
	} else {
		err = errors.New("event not recognized")
	}
	return checkerFunc, err
}

func (gl *GoLora) runCb() {
	if gl.cb != nil {
		gl.cb()
	}
}

func (gl *GoLora) cbDaemon(eventChecker func() bool, ch chan struct{}) {
	t := time.NewTicker(20 * time.Millisecond)
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

func (gl *GoLora) Destroy() error {
	defer func() {
		if gl.cbStopper != nil {
			close(gl.cbStopper)
			gl.cb = nil
			gl.cbStopper = nil
		}
	}()
	if err := gl.ChangeMode(Sleep); err != nil {
		return err
	}
	if err := gl.Reset(); err != nil {
		return err
	}
	return nil
}

func (gl *GoLora) DumpRegisters() ([]RegVal, error) {
	registers := make([]byte, 0x26)
	regVal := make([]RegVal, len(registers))
	values := make([]byte, len(registers))
	var err error
	for i := 0; i < 0x26; i++ {
		registers[i] = byte(i)
	}
	gl.mu.Lock()
	for idx, reg := range registers {
		values[idx], err = gl.readReg(reg)
		if err != nil {
			break
		}
		regVal[idx] = RegVal{
			Reg: reg,
			Val: values[idx],
		}
	}
	gl.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return regVal, nil
}

func (gl *GoLora) GetConf() LoraConf {
	return gl.Conf
}
