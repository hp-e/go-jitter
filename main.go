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

const pingCount = 10

func main() {

	host := flag.String("host", "www.google.com", "Sets the host to measure jitter on")
	showPings := flag.Bool("l", false, "If true, shows the all the ping stats")
	pingCount := flag.Int("c", 10, "The number of pings to perform")

	flag.Parse()
	fmt.Printf("\nPING and JITTER Measurement")
	fmt.Printf("\nThis tool will measure the jitter by pinging %s %d of times", *host, pingCount)
	fmt.Printf("\nYou can end measuring by using ctrl+c")
	fmt.Printf("\n\nWhat is a good jitter?")
	fmt.Printf("\n\nA variance below 15%% of the average ping (status = 'avg round trip' / 'jitter')")

	go measureJitter(*host, *showPings, *pingCount)

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
}

func measureJitter(host string, showPings bool, pingCount int) {
	measureCounter := 0

	for {
		measureCounter++

		fmt.Printf("\n\nMeasurement #%d\n", measureCounter)

		pinger, err := ping.NewPinger(host)
		pinger.SetPrivileged(true)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		pinger.Count = pingCount
		err = pinger.Run() // Blocks until finished.
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		stats := pinger.Statistics()

		pr := stats.Rtts[0].Milliseconds()
		var jtr float64
		var measures int

		for i, v := range stats.Rtts {

			if i == 0 {
				continue
			}

			if showPings {
				fmt.Printf("\nReply from %s: bytes=%d time=%vms TTL=%d", stats.IPAddr.String(), pinger.Size, v.Milliseconds(), pinger.TTL)
			}
			cur := math.Abs(float64(pr) - float64(v.Milliseconds()))

			jtr = jtr + cur
			// avg = avg + float64(v.Milliseconds())

			pr = v.Milliseconds()
			measures++
		}

		avgJitter := jtr / float64(measures)
		threshold := float64(stats.AvgRtt.Milliseconds()) / avgJitter
		varianceStatus := "good"

		if threshold > 15 {
			varianceStatus = "bad"
		}

		fmt.Printf("\n\nPing stats for %v:", stats.IPAddr.String())
		fmt.Printf("\n\tPackets: Sent = %v, Received = %v, Lost = %v", stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		fmt.Printf("\nApproximate round trip times in milli-seconds:")
		fmt.Printf("\n\tPackets: Minimum = %vms, Maximum = %vms, Average = %vms", stats.MinRtt.Milliseconds(), stats.MaxRtt.Milliseconds(), stats.AvgRtt.Milliseconds())

		fmt.Printf("\nJitter: %vms, Variance: %v%% [variance is %s]", math.Round(avgJitter), math.Round(threshold), varianceStatus)

		time.Sleep(time.Second * 5)
	}
}
