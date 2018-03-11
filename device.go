package main

import (
	"fmt"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	common "github.com/nickw444/miio-go/common"
	"github.com/nickw444/miio-go/device"
	"github.com/nickw444/miio-go/subscription"
	"github.com/sirupsen/logrus"
)

type HKDevice interface {
	Start() error
	Stop() error
}

type hkDevice struct {
	log       *logrus.Entry
	transport hc.Transport
	pin       string
}

func (d *hkDevice) StartTransport(id uint32, acc *accessory.Accessory) {
	transport, err := hc.NewIPTransport(hc.Config{
		Pin:         d.pin,
		StoragePath: fmt.Sprintf("dev-%d", id),
	}, acc)
	if err != nil {
		d.log.Panic(err)
	}

	d.transport = transport
	d.log.Infof("Starting transport with pin %s", d.pin)
	transport.Start()
}

func (d *hkDevice) StopTransport() {
	<-d.transport.Stop()
}

type HKPowerPlug struct {
	hkDevice
	dev *device.PowerPlug
	sub subscription.Subscription
}

func NewHKPowerPlug(dev *device.PowerPlug, pin string) *HKPowerPlug {
	log := log.WithField("device", "HKPowerPlug").WithField("id", dev.ID())
	plug := &HKPowerPlug{
		dev:      dev,
		hkDevice: hkDevice{pin: pin, log: log},
	}

	return plug
}

func (p *HKPowerPlug) Start() error {
	sub, err := p.dev.NewSubscription()
	if err != nil {
		return err
	}
	p.sub = sub

	acc := accessory.NewSwitch(accessory.Info{
		Name:         fmt.Sprintf("Power Plug %d", p.dev.ID()),
		Manufacturer: "MiiO",
		Model:        "PowerPlug",
		SerialNumber: fmt.Sprintf("%d", p.dev.ID()),
	})

	acc.Switch.On.OnValueRemoteUpdate(func(on bool) {
		if on {
			p.dev.SetPower(common.PowerStateOn)
		} else {
			p.dev.SetPower(common.PowerStateOff)
		}
	})

	go p.StartTransport(p.dev.ID(), acc.Accessory)
	go func() {
		for event := range sub.Events() {
			p.log.Infof("Handling event: %T", event)
			switch event.(type) {
			case common.EventUpdatePower:
				powerState := event.(common.EventUpdatePower).PowerState
				if powerState == common.PowerStateOn {
					acc.Switch.On.SetValue(true)
				} else if powerState == common.PowerStateOff {
					acc.Switch.On.SetValue(false)
				}
			}
		}
	}()

	return nil
}

func (p *HKPowerPlug) Stop() error {
	p.sub.Close()
	p.StopTransport()
	return nil
}

type HKYeelight struct {
	hkDevice
	dev *device.Yeelight
	sub subscription.Subscription
}

func NewHKYeelight(dev *device.Yeelight, pin string) *HKYeelight {
	log := log.WithField("device", "HKYeelight").WithField("id", dev.ID())
	yeelight := &HKYeelight{
		dev:      dev,
		hkDevice: hkDevice{pin: pin, log: log},
	}

	return yeelight
}

func (y *HKYeelight) Start() error {
	sub, err := y.dev.NewSubscription()
	if err != nil {
		return err
	}
	y.sub = sub

	acc := accessory.NewLightbulb(accessory.Info{
		Name:         fmt.Sprintf("Yeelight %d", y.dev.ID()),
		Manufacturer: "MiiO",
		Model:        "Yeelight",
		SerialNumber: fmt.Sprintf("%d", y.dev.ID()),
	})

	acc.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		if on {
			y.dev.SetPower(common.PowerStateOn)
		} else {
			y.dev.SetPower(common.PowerStateOff)
		}
	})

	acc.Lightbulb.Brightness.OnValueRemoteUpdate(func(brightness int) {
		y.dev.SetBrightness(brightness)
	})

	acc.Lightbulb.Hue.OnValueRemoteUpdate(func(hue float64) {
		y.dev.SetHSV(int(hue), int(acc.Lightbulb.Saturation.GetValue()))
	})

	acc.Lightbulb.Saturation.OnValueRemoteUpdate(func(sat float64) {
		y.dev.SetHSV(int(acc.Lightbulb.Hue.GetValue()), int(sat))
	})

	go y.StartTransport(y.dev.ID(), acc.Accessory)
	go func() {
		for event := range sub.Events() {
			y.log.Infof("Handling event: %T", event)
			switch event.(type) {
			case common.EventUpdatePower:
				powerState := event.(common.EventUpdatePower).PowerState
				if powerState == common.PowerStateOn {
					acc.Lightbulb.On.SetValue(true)
				} else if powerState == common.PowerStateOff {
					acc.Lightbulb.On.SetValue(false)
				}
			case common.EventUpdateLight:
				ev := event.(common.EventUpdateLight)
				acc.Lightbulb.Hue.SetValue(float64(ev.Hue))
				acc.Lightbulb.Brightness.SetValue(ev.Brightness)
				acc.Lightbulb.Saturation.SetValue(float64(ev.Saturation))
			}
		}
	}()

	return nil
}
func (y *HKYeelight) Stop() error {
	y.sub.Close()
	y.StopTransport()
	return nil
}
