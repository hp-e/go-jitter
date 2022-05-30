package main_test

import (
	"fmt"
	"math"
	"testing"

	ping "github.com/go-ping/ping"
)

func TestJittering(t *testing.T) {
	pinger, err := ping.NewPinger("www.google.com")
	pinger.SetPrivileged(true)
	if err != nil {
		panic(err)
	}
	pinger.Count = 10
	err = pinger.Run() // Blocks until finished.
	if err != nil {
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

	fmt.Println(avgJitter / float64(measures))

	// fmt.Println(stats)
}
