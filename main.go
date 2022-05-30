package main

import (
	"fmt"
	"math"

	ping "github.com/go-ping/ping"
)

func main() {

	fmt.Println("Start measuring jitter...")
	pinger, err := ping.NewPinger("www.google.com")
	pinger.SetPrivileged(true)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	pinger.Count = 10
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	stats := pinger.Statistics()

	// j := []float64{}

	pr := stats.Rtts[0].Milliseconds()
	var avgJitter float64
	var measures int

	for i, v := range stats.Rtts {

		if i == 0 {
			continue
		}

		cur := math.Abs(float64(pr) - float64(v.Milliseconds()))

		avgJitter = avgJitter + cur

		// j = append(j, cur)
		pr = v.Milliseconds()
		measures++
	}

	fmt.Printf("Result: %vms", avgJitter/float64(measures))
}
