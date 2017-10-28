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
}

func NewHKYeelight(dev *device.Yeelight) *HKYeelight {
	return &HKYeelight{
		dev: dev,
	}
}

func (y *HKYeelight) Start() error {
	return nil
}
func (y *HKYeelight) Stop() error {
	return nil
}
