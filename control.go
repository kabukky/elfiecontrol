package main

import (
	"encoding/hex"
	"log"
	"net"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"
)

var (
	messageRate  = 25 * time.Millisecond
	stopSteering = false

	// All data without checksum. Will be calculated before sending
	steeringDummyData, _ = hex.DecodeString("ff080000000090101000")
	stopData, _          = hex.DecodeString("ff087e3f403f901010a0")
	controlsOnData, _    = hex.DecodeString("ff08003f403f10101000")
	hoverOnData, _       = hex.DecodeString("ff087e3f403f90101000")
	rotorOnData, _       = hex.DecodeString("ff087e3f403f90101040")
	gyroCalibData, _     = hex.DecodeString("ff08003f403f50101000")
	compassActiveData, _ = hex.DecodeString("ff087e3f403f90101010") // No idea what this is for
)

func sendStopData(conn net.Conn) {
	log.Println("Sending stop data")
	sendMessageDuration(stopData, conn, 500*time.Millisecond)
}

func sendSteeringCommand(conn net.Conn) {
	if !glfw.JoystickPresent(glfw.Joystick1) {
		log.Println("No gamepad present.")
		return
	}
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
