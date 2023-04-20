package internal

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

func connectToLights(lights map[string]confLight) (map[string]*bluetooth.Device, error) {
	finalDevs := make(map[string]*bluetooth.Device)

	for ln, l := range lights {
		uuid, err := bluetooth.ParseUUID(l.UUID)
		if err != nil {
			for _, d := range finalDevs {
				d.Disconnect()
			}
			return nil, fmt.Errorf("failed to parse uuid address[%v]: %v", l.UUID, err)
		}

		address := bluetooth.Address{
			UUID: uuid,
		}

		// scan for the light we want to connect to
		scanChan := make(chan error, 1)
		go func() {
			log.Infof("scanning for light[%v] at %v...", ln, uuid.String())
			err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
				if device.Address.String() == uuid.String() {
					scanChan <- nil
				}
			})
			if err != nil {
				scanChan <- fmt.Errorf("failed to scan for ble devices: %v", err)
			}
		}()

		select {
		case scanRes := <-scanChan:
			if scanRes != nil {
				return nil, fmt.Errorf("error while scanning for light[%v] at %v: %v", ln, l.UUID, scanRes)
			} else {
				log.Infof("found light[%v] at %v", ln, l.UUID)
				adapter.StopScan()
			}
		case <-time.After(5 * time.Second):
			return nil, fmt.Errorf("failed to find light[%v] at %v. Is it in range?", ln, l.UUID)
		}

		// connect to the light that we found while scanning
		log.Infof("connecting to light[%v] at %v...", ln, l.UUID)
		dev, err := adapter.Connect(address, bluetooth.ConnectionParams{})
		if err != nil {
			for _, d := range finalDevs {
				d.Disconnect()
			}
			return nil, fmt.Errorf("failed to connect to device[%v]: %v", ln, err)
		}
		log.Infof("successfully connected to light[%v] at %v", ln, l.UUID)

		finalDevs[ln] = dev
	}

	return finalDevs, nil
}

func on(dChar *bluetooth.DeviceCharacteristic) error {
	_, err := dChar.WriteWithoutResponse([]byte{0xCC, 0x23, 0x33})
	return err
}

func off(dChar *bluetooth.DeviceCharacteristic) error {
	_, err := dChar.WriteWithoutResponse([]byte{0xCC, 0x24, 0x33})
	return err
}

func setColor(dChar *bluetooth.DeviceCharacteristic, red, green, blue byte) error {
	_, err := dChar.WriteWithoutResponse([]byte{0x56, red, green, blue, 0x00, 0xF0, 0xAA})
	return err
}

func (ml *merryLighting) listenForColor() {
	for {
		select {
		case color := <-ml.colorChan:
			setColor(color.char, color.Red, color.Green, color.Blue)
		}
	}
}
