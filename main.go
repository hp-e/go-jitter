package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	ping "github.com/go-ping/ping"
)

func main() {

	host := flag.String("host", "www.google.com", "Sets the host to measure jitter on")
	fmt.Printf("Starts measuring jitter against %s... \nctrl+c to terminate", *host)
	go measureJitter(*host)

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
}

func measureJitter(host string) {

	for {
		pinger, err := ping.NewPinger(host)
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

		fmt.Printf("\nResult: %vms", avgJitter/float64(measures))
		time.Sleep(time.Second * 10)
	}
}
