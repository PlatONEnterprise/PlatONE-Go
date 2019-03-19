package main

import (
	"os"
	"os/signal"
)

func ReadChan(outChan chan int) int {
	length := len(outChan)
	for i := 0; i <= length; i++ {
		<-outChan
	}
	return length
}

func Trap(ch chan int) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	<-c
	ch <- 1
}
