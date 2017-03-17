package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bgentry/heliospectra"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devices, err := heliospectra.ScanUDP(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println("Got devices from scan:")
	for _, d := range devices {
		fmt.Printf("heliospectra.DeviceInfo%+v\n", d)
	}

	if len(devices) > 0 {
		demoDevice(ctx, devices[0])
	}
}

func demoDevice(ctx context.Context, di heliospectra.DeviceInfo) {
	device := heliospectra.NewDevice(di.IPAddr, nil)

	diagCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	diag, err := device.Diagnostic(diagCtx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("diagnostic data: %+v\n", diag)

	intensities := make([]int, len(diag.Wavelengths))

	lightshow(ctx, device, intensities, 0)
	lightshow(ctx, device, intensities, 1)
	lightshow(ctx, device, intensities, 2)
	lightshow(ctx, device, intensities, 3)

	for _ = range diag.Wavelengths {
		intensities = append(intensities, 0)
	}
}

const (
	lightshowSteps    = 50
	lightshowStepSize = 2
)

func lightshow(ctx context.Context, device *heliospectra.Device, intensities []int, idx int) {
	// turn all off
	device.SetIntensities(ctx, intensities...)

	for i := 1; i < lightshowSteps; i++ {
		time.Sleep(30 * time.Millisecond)
		intensities[idx] = lightshowStepSize * i
		device.SetIntensities(ctx, intensities...)
	}

	for i := lightshowSteps; i >= 0; i-- {
		time.Sleep(30 * time.Millisecond)
		intensities[idx] = lightshowStepSize * i
		device.SetIntensities(ctx, intensities...)
	}

	for i := range intensities {
		intensities[i] = 0
	}
	device.SetIntensities(ctx, intensities...)
}
