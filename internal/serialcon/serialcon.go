package serialcon

import (
	"errors"
	"log"
	"yesbotics/ysm/internal/config"

	"go.bug.st/serial"
)

type ReceiveCallback func(data string)

type Serialcon struct {
	config          config.SerialConfig
	serialPort      *serial.Port
	messageCallback ReceiveCallback
	connected       bool
}

func GetSerialPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}

	var serialPorts []string
	for _, port := range ports {
		serialPorts = append(serialPorts, port)
	}
	return serialPorts, nil
}

func GetBaudrates() []int {
	return []int{
		300,
		600,
		1200,
		2400,
		4800,
		9600,
		19200,
		28800,
		38400,
		57600,
		76800,
		115200,
	}
}

func New() *Serialcon {
	m := Serialcon{
		serialPort: nil,
		connected:  false,
	}

	return &m
}

func (m *Serialcon) IsConnected() bool {
	return m.connected
}

func (m *Serialcon) Open(serialConfig config.SerialConfig, callback ReceiveCallback) error {
	port, err := serial.Open(serialConfig.SerialPort, &serialConfig.SerialMode)
	if err != nil {
		return err
	}

	m.connected = true
	m.config = serialConfig
	m.serialPort = &port
	m.messageCallback = callback
	go m.readSerialData()
	return nil
}

func (m *Serialcon) Close() error {
	m.messageCallback = nil
	m.connected = false
	if m.serialPort != nil {
		err := (*m.serialPort).Close()
		if err != nil {
			log.Println("Could not close serial connection:", err)
			return err
		}
	}
	return nil
}

func (m *Serialcon) GetCurrentPortName() (string, error) {
	if m.serialPort != nil {
		return m.config.SerialPort, nil
	} else {
		return "", errors.New("no port available")
	}
}

func (m *Serialcon) GetCurrentBaudrate() int {
	return m.config.SerialMode.BaudRate
}

func (m *Serialcon) Write(buffer []byte) (int, error) {
	return (*m.serialPort).Write(buffer)
}

func (m *Serialcon) readSerialData() {
	buffer := make([]byte, 128)
	for {
		n, err := (*m.serialPort).Read(buffer)
		if err != nil {
			log.Println("Could not read serial data:", err)
			return
		}

		data := string(buffer[:n])

		if m.messageCallback != nil {
			m.messageCallback(data)
		}
	}
}
