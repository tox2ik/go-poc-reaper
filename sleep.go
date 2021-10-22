package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {

	intervalMs := "1866"
	lifespanS := "6"
	if len(os.Args) >= 2 {
		intervalMs = os.Args[1]
	}
	if len(os.Args) >= 3 {
		lifespanS = os.Args[2]
	}

	heartbeat, e := strconv.Atoi(intervalMs)
	if e != nil {
		heartbeat = 3
	}

	lifespan, e := strconv.Atoi(lifespanS)
	if e != nil {
		lifespan = 6
	}

	end := time.Now().Add(time.Duration(lifespan) * time.Second)

	for ;; {
		fmt.Printf("%d\n", time.Now().Unix())
		time.Sleep(time.Millisecond * time.Duration(heartbeat))
		if time.Now().After(end) {
			fmt.Println("done.")
			os.Exit(0)
		}
	}
}
