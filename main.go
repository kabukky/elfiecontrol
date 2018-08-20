package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"
)

func main() {
	// Initialize glfw (need for joystick input)
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()
	// Create connection
	conn, err := net.Dial("udp", "172.16.10.1:8080")
	if err != nil {
		panic(err)
	}

	// Send Idle data
	log.Println("Sending prepare data")
	sendMessageDuration(gyroCalibData, conn, 2*time.Second)
	sendMessageDuration(controlsOnData, conn, 2*time.Second)
	sendMessageDuration(hoverOnData, conn, 2*time.Second)

	sendMessageDuration(rotorOnData, conn, 2*time.Second)

	// Catch SIGTERM
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// Send Stop data
		stopSteering = true
		sendStopData(conn)
		log.Println("Exiting")
		os.Exit(1)
	}()

	for !stopSteering {
		// Send joystick command
		sendSteeringCommand(conn)
		time.Sleep(messageRate)
	}
	sendStopData(conn)
}
