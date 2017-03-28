// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package spitool implements tools for with with SPI.
package spitool

import (
	"errors"
	"sync"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/spi"
)

// MakeConn create a Conn out of a Bus with a manually managed CS line.
func MakeConn(b spi.Bus, p gpio.PinOut) (*ConnGPIO, error) {
	l, ok := b.(sync.Locker)
	if !ok {
		return nil, errors.New("spi: MakeConn can only work with SPI Bus that implement sync.Locker")
	}
	pins, _ := b.(spi.Pins)
	c, err := b.CS(-1)
	if err != nil {
		return nil, err
	}
	return &ConnGPIO{c: c, p: p, l: l, pins: pins}, nil
}

// ConnGPIO is a SPI ConnCloser that uses an arbitrary GPIO pin as the chip
// select line.
type ConnGPIO struct {
	c      spi.ConnCloser
	p      gpio.PinOut
	l      sync.Locker
	pins   spi.Pins
	active gpio.Level
}

func (c *ConnGPIO) Close() error {
	return c.c.Close()
}

// DevParams implements Conn.
func (c *ConnGPIO) DevParams(highHz int64, mode spi.Mode, bits int) error {
	c.active = gpio.Level(mode&spi.Mode2 == 0)
	if err := c.p.Out(!c.active); err != nil {
		return err
	}
	return c.c.DevParams(highHz, mode, bits)
}

// Tx implements Conn.
func (c *ConnGPIO) Tx(w, r []byte) error {
	c.l.Lock()
	defer c.l.Unlock()
	if err := c.p.Out(c.active); err != nil {
		return err
	}
	defer c.p.Out(!c.active)
	return c.c.Tx(w, r)
}

// Duplex implements Conn.
func (c *ConnGPIO) Duplex() conn.Duplex {
	return c.c.Duplex()
}

//

var _ spi.ConnCloser = &ConnGPIO{}
