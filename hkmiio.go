package main

import (
	"flag"

	"github.com/nickw444/miio-go"
	"github.com/nickw444/miio-go/common"
	device "github.com/nickw444/miio-go/device"
	"github.com/sirupsen/logrus"
)

var devices = map[uint32]HKDevice{}
var log = logrus.New()

func NewDevice(dev common.Device, pin string) {
	if _, ok := devices[dev.ID()]; ok {
		log.Infof("New Device event for already known device: %d", dev.ID())
		return
	}

	var hkDev HKDevice
	switch dev.(type) {
	case *device.PowerPlug:
		hkDev = NewHKPowerPlug(dev.(*device.PowerPlug), pin)
	case *device.Yeelight:
		hkDev = NewHKYeelight(dev.(*device.Yeelight), pin)
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
		delete(devices, device.ID())
	}
}

func main() {
	pin := flag.String("pin", "", "Homekit device pin")
	debug := flag.Bool("debug", false, "Enable debug")
	miioDebug := flag.Bool("miio-debug", false, "Enable miio debug")
	flag.Parse()

	if *pin == "" {
		log.Panicf("Must provide a pin")
	}

	if *debug {
		log.SetLevel(logrus.DebugLevel)
	}

	if *miioDebug {
		miioLogger := logrus.New()
		miioLogger.SetLevel(logrus.DebugLevel)
		common.SetLogger(miioLogger)
	}

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
				go NewDevice(event.(common.EventNewDevice).Device, *pin)
			case common.EventExpiredDevice:
				go ExpiredDevice(event.(common.EventExpiredDevice).Device)
			}
		}
	}
}
