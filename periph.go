// Copyright 2021 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

/*
func periphLED() error {
	// Initialize the device.
	if _, err := host.Init(); err != nil {
		return err
	}
	devs := ftdi.All()
	if len(devs) == 0 {
		return errors.New("connect a FTDI device")
	}
	if false {
		if len(devs) > 1 {
			return fmt.Errorf("found a surprising number of FTDI devices (%d), please disconnect a few", len(devs))
		}
	}
	ft, ok := devs[0].(*ftdi.FT232H)
	if !ok {
		return fmt.Errorf("need a FT232H, got %T: %s", devs[0], devs[0])
	}
	defer ft.Halt()

	if false {
		f := 800 * physic.KiloHertz
		d, err := nrzled.NewStream(ft.D1, &nrzled.Opts{NumPixels: *p, Channels: 3, Freq: f})
		if err != nil {
			return err
		}
		defer d.Halt()

		b := d.Bounds()
		img := image.NewNRGBA(b)
		for i := 0; i < b.Max.Y; i++ {
			img.SetNRGBA(i, 0, color.NRGBA{uint8(255 - 2*i), uint8(2 * i), 0, 255})
		}
		for ctx.Err() == nil {
			if err := d.Draw(b, img, image.Point{}); err != nil {
				return err
			}
			// To loop only once.
			cancel()
		}
	}

	//b := []byte{0b10101010}
	b := []byte{0xFF, 0xFF}
	return ft.D1.StreamOut(&gpiostream.BitStream{Bits: b, Freq: 10 * physic.KiloHertz, LSBF: false})
	return nil
}
*/
