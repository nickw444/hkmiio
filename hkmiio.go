package main

import (
	"github.com/nickw444/miio-go"
	"github.com/nickw444/miio-go/common"
	device "github.com/nickw444/miio-go/device"
	"github.com/sirupsen/logrus"
)

var devices = map[uint32]HKDevice{}
var log = logrus.New()

func NewDevice(dev common.Device) {
	if _, ok := devices[dev.ID()]; ok {
		log.Infof("New Device event for already known device: %d", dev.ID())
		return
	}

	var hkDev HKDevice
	switch dev.(type) {
	case *device.PowerPlug:
		hkDev = NewHKPowerPlug(dev.(*device.PowerPlug), "12341234")
	case *device.Yeelight:
		hkDev = NewHKYeelight(dev.(*device.Yeelight))
	}

	err := hkDev.Start()
	if err != nil {
		log.Panic(err)
	}
	devices[dev.ID()] = hkDev
	log.Infof("New device %T", hkDev)
}

func ExpiredDevice(device common.Device) {
	if hkDev, ok := devices[device.ID()]; ok {
		hkDev.Stop()
	}
}

func main() {
	client, err := miio.NewClient()
	if err != nil {
		panic(err)
	}

	sub, err := client.NewSubscription()
	if err != nil {
		panic(err)
	}

	for {
		select {
		case event := <-sub.Events():
			log.Infof("Handling event: %T", event)

			switch event.(type) {
			case common.EventNewDevice:
				go NewDevice(event.(common.EventNewDevice).Device)
			case common.EventExpiredDevice:
				go ExpiredDevice(event.(common.EventExpiredDevice).Device)
			}
		}
	}
}
