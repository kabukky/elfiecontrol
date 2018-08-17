package main

import (
	"encoding/hex"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"
)

var (
	messageRate  = 25 * time.Millisecond
	power        = float32(0.3) // steering values will by multiplied by this
	stopSteering = false

	// All data without checksum. Will be calculated before sending
	steeringDummyData, _ = hex.DecodeString("ff080000000090101000")
	stopData, _          = hex.DecodeString("ff087e3f403f901010a0")
	controlsOnData, _    = hex.DecodeString("ff08003f403f10101000")
	hoverOnData, _       = hex.DecodeString("ff087e3f403f90101000")
	rotorOnData, _       = hex.DecodeString("ff087e3f403f90101040")
	flyUpStartData, _    = hex.DecodeString("ff08903b403f90101040")
	flyUpHighData, _     = hex.DecodeString("ff08c43b403f90101000")
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

func sendStopData(conn net.Conn) {
	log.Println("Sending stop data")
	sendMessageDuration(stopData, conn, 500*time.Millisecond)
}

func sendSteeringCommand(conn net.Conn) {
	if glfw.GetJoystickButtons(glfw.Joystick1)[0] == 1 {
		stopSteering = true
		sendStopData(conn)
		return
	}
	axes := glfw.GetJoystickAxes(glfw.Joystick1)
	log.Println("Axes: ", axes)
	// Fit -1 to 1 range of joystick into 0 to 255 range of drone
	steeringDummyData[2] = uint8((-axes[1] + 1) * 127.5)
	steeringDummyData[3] = uint8(((axes[0] + 1) * 127.5) / 2)
	steeringDummyData[4] = uint8(((axes[3] + 1) * 127.5) / 2)
	steeringDummyData[5] = uint8(((axes[2] + 1) * 127.5) / 2)
	sendMessage(steeringDummyData, conn)
}

func sendMessageDuration(message []byte, conn net.Conn, duration time.Duration) {
	ticker := time.NewTicker(messageRate)
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		time.Sleep(duration)
		done <- true
	}()
	for {
		select {
		case <-done:
			// ticker ended
			return
		case <-ticker.C:
			sendMessage(message, conn)
		}
	}
}

func sendMessage(message []byte, conn net.Conn) {
	// Add checksum to byte slice (JJRC chose a simple algorithm: substract all bytes from each other)
	_, err := conn.Write(append(message, message[0]-message[1]-message[2]-message[3]-message[4]-message[5]-message[6]-message[7]-message[8]-message[9]))
	if err != nil {
		log.Println("Error while sending data: ", err)
	}
}
