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

	time.Sleep(5 * time.Second)
}
