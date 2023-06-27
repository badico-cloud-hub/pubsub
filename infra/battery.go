package infra

import (
	"time"

	batterygo "github.com/badico-cloud-hub/battery-go/battery"
	"github.com/badico-cloud-hub/battery-go/storages"
)

type Battery struct {
	storage *storages.GoCacheStorage
}

func NewBattery() *Battery {
	storage := storages.NewGoCacheStorage()
	return &Battery{storage: storage}
}

func (b *Battery) Init(interval time.Duration, updater func() []batterygo.BatteryArgument) error {
	battery := batterygo.NewBattery(b.storage, interval)
	go battery.Init(updater)
	return nil
}

func (b *Battery) Set(key string, value interface{}) error {
	err := b.storage.Set(key, value)
	if err != nil {
		return err
	}
	return nil
}

func (b *Battery) Get(key string) (interface{}, error) {
	value, err := b.storage.Get(key)
	if err != nil {
		return "", err
	}
	return value, nil
}
