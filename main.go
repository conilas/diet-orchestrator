package main

import (
	processors "diet-scheduler/processors"
	"flag"
	"github.com/carlescere/scheduler"
)

func main() {

	server := flag.String("server", "localhost:9000", "the food service endpoint (ip:port)")
	certificate := flag.String("certificate", "../localhost.crt", "the certificate path")
	interval := flag.Int("interval", 45, "amount of seconds between each run for checking orders")

	flag.Parse()

	scheduler.Every(*interval).Seconds().Run(func() {
		processors.ScheduledJob(*server, *certificate)
	})

	select {}
}
