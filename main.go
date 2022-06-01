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

type Color string

const (
	ColorRed    Color = "\u001b[31m"
	ColorGreen        = "\u001b[32m"
	ColorYellow       = "\u001b[33m"
	ColorReset        = "\u001b[0m"
)

func main() {

	host := flag.String("host", "www.google.com", "Sets the host to measure jitter on")
	pingCount := flag.Int("c", 20, "The number of pings to perform")

	flag.Parse()

	fmt.Printf("\nPING and JITTER")
	fmt.Printf("\nThis tool will measure the jitter by pinging %s, %d times", *host, *pingCount)
	fmt.Printf("\nYou can end measuring by using ctrl+c")

	go measureJitter(*host, *pingCount)

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
}

func measureJitter(host string, pingCount int) {
	measureCounter := 0

	for {
		measureCounter++

		fmt.Printf("\n\nMeasurement #%d\n", measureCounter)

		pinger, err := ping.NewPinger(host)

		// variables for the onRecv
		var previousRtt, accumulatedDiff float64
		var measurementCount int
		isFirst := true

		// this function executes when we receive data from each ping
		// it prints a new line when data is received but also accumulate the diff and number of measurements
		// that will be used in the final jitter calculation
		pinger.OnRecv = func(p *ping.Packet) {

			if isFirst {
				// if this is the first ping we have to use that as the first observation
				// the rttValue is set at the end of the function and will be used in the first calculation
				fmt.Printf("\nReply from %s: bytes=%d, time=%3vms", p.IPAddr.String(), p.Nbytes, p.Rtt.Milliseconds())

			} else {
				measurementCount++

				diff := math.Abs(previousRtt - float64(p.Rtt.Milliseconds()))
				accumulatedDiff = accumulatedDiff + diff
				currentJitter := math.Round(accumulatedDiff / float64(measurementCount))

				fmt.Printf("\nReply from %s: bytes=%d, time=%3vms, diff=%3v, acc=%3v, jtr=%3v",
					p.IPAddr.String(),
					p.Nbytes,
					p.Rtt.Milliseconds(),
					diff,
					accumulatedDiff,
					currentJitter,
				)
			}

			previousRtt = float64(p.Rtt.Milliseconds())
			isFirst = false
		}

		pinger.Count = pingCount

		// this must be set to true
		pinger.SetPrivileged(true)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		// Starts the pinging,bBlocks until finished.
		err = pinger.Run()
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		stats := pinger.Statistics()

		// the final jitter calculation based on the accumulated diff
		// divided by the number of measurements.
		avgJitter := accumulatedDiff / float64(measurementCount)

		// prints the result of ping statistics, packets sent/received and the jitter
		reportPingStats(stats)
		reportJitter(avgJitter, float64(stats.AvgRtt.Milliseconds()))

		// time to sleep before exeuting a new round of pings
		time.Sleep(time.Second * 5)
	}
}

func reportPingStats(stats *ping.Statistics) {
	fmt.Printf("\n\nPing stats for %v:", stats.IPAddr.String())
	fmt.Printf("\n\tPackets: Sent = %v, Received = %v, Lost = %v", stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
	fmt.Printf("\n\tPacket Loss Score: %s", packetLossScore(stats.PacketsRecv, stats.PacketLoss))

	fmt.Printf("\nApproximate round trip times in milli-seconds:")
	fmt.Printf("\n\tPackets: Minimum = %vms, Maximum = %vms, Average = %vms", stats.MinRtt.Milliseconds(), stats.MaxRtt.Milliseconds(), stats.AvgRtt.Milliseconds())
	fmt.Printf("\n\tRtt Score: %s", packetRttScore(stats.AvgRtt.Milliseconds()))

}

func reportJitter(avgJitter, avgPing float64) {
	jitterScore := jitterVarianceScore(avgJitter, avgPing)
	fmt.Printf("\nJitter")
	fmt.Printf("\n\tAverage = %vms", math.Round(avgJitter))
	fmt.Printf("\n\tJitter score: %s\n", jitterScore)

}

func jitterVarianceScore(jitter, avgPing float64) string {

	// if the jitter is below 15% of the average latency
	// that's very good
	// e.g average latency is 100ms, 15% is 15ms.
	// if the jitter is <= 15ms the score will be 'very good'
	veryGood := avgPing * 0.15

	if jitter <= veryGood {
		return colorText(ColorGreen, "very good")
	}

	// if the jitter is below 30ms we can consider it to be good
	if jitter <= 30.0 {
		return colorText(ColorGreen, "good")
	}

	// we do not want to be here. Still for VOD and Live broadcast could work in a fair manner
	// but this could also indicate some issues with the network
	if jitter <= 50 {
		return colorText(ColorYellow, "fair")
	}

	// we definatly do not want to be here.
	// this could mean something is wrong with the network
	return colorText(ColorRed, "bad")
}

func packetLossScore(packetsReceived int, packetsLost float64) string {

	if packetsLost == 0 {
		return colorText(ColorGreen, "good")
	}

	if (packetsLost / float64(packetsReceived)) <= 1 {
		return colorText(ColorYellow, "fair")
	}

	return colorText(ColorRed, "bad")

}

func packetRttScore(avgRtt int64) string {

	if avgRtt <= 50 {
		return colorText(ColorGreen, "very, very good")
	}
	if avgRtt <= 100 {
		return colorText(ColorGreen, "very good")
	}
	if avgRtt <= 150 {
		return colorText(ColorGreen, "good")
	}
	if avgRtt <= 300 {
		return colorText(ColorYellow, "ok")
	}
	if avgRtt <= 400 {
		return colorText(ColorYellow, "fair")
	}

	return colorText(ColorRed, "bad")

}

func colorText(color Color, input string) string {
	return fmt.Sprintf("%s%s%s", string(color), input, string(ColorReset))
}
