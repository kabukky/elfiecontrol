package main

import (
	"encoding/hex"
	"log"
	"net"
	"time"
)

var (
	messageRate = 25 * time.Millisecond

	// All data without checksum. Will be calculated before sending
	stopData, _       = hex.DecodeString("ff087e3f403f901010a0")
	controlsOnData, _ = hex.DecodeString("ff08003f403f10101000")
	hoverOnData, _    = hex.DecodeString("ff087e3f403f90101000")
	rotorOnData, _    = hex.DecodeString("ff087e3f403f90101040")
	flyUpStartData, _ = hex.DecodeString("ff08903b403f90101040")
	flyUpHighData, _  = hex.DecodeString("ff08c43b403f90101000")
)

func main() {
	// Create connection
	conn, err := net.Dial("udp", "172.16.10.1:8080")
	if err != nil {
		panic(err)
	}

	// Send Idle data
	log.Println("Sending prepare data")
	sendMessageDuration(controlsOnData, conn, 2*time.Second)
	sendMessageDuration(hoverOnData, conn, 2*time.Second)

	// Send rotor on data
	log.Println("Sending rotor on data")
	sendMessageDuration(rotorOnData, conn, 4*time.Second)

	// Send fly up on data
	log.Println("Sending fly up data")
	//flyUp(conn)
	sendMessageDuration(flyUpHighData, conn, 4*time.Second)

	// Send Stop data
	log.Println("Sending stop data")
	sendMessageDuration(stopData, conn, 2*time.Second)

	log.Println("Done sending data")
}

func flyUp(conn net.Conn) {
	// Increase speed 36 times. Each time by 2.
	accelerationTimes := 24
	for i := 0; i < accelerationTimes; i++ {
		if i != 0 {
			flyUpStartData[2] = flyUpStartData[2] + 2
			flyUpStartData[10] = flyUpStartData[10] - 2
		}
		log.Println("Sending loop ", i)
		duration := 50 * time.Millisecond
		if i == (accelerationTimes - 1) {
			duration = 5000 * time.Millisecond
		}
		sendMessageDuration(flyUpStartData, conn, duration)
	}
	return
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
