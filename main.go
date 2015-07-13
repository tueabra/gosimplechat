package main

import (
	"fmt"
	"github.com/fatih/color"
	"net"
)

func Log(v ...interface{}) {
	fmt.Println(v...)
}

var fmtInfo = color.New(color.FgGreen)

func LogInfo(v ...interface{}) {
	fmtInfo.Println(v...)
}

var fmtMessage = color.New(color.FgYellow)

func LogMessage(v ...interface{}) {
	fmtMessage.Println(v...)
}

var fmtError = color.New(color.FgRed)

func LogError(v ...interface{}) {
	fmtError.Println(v...)
}

func main() {
	LogInfo("Starting server...")

	server := NewServer()
	go server.Serve()

	service := "localhost:9988"
	tcpAddr, error := net.ResolveTCPAddr("tcp", service)
	if error != nil {
		LogError("Error: Could not resolve address")
	} else {
		netListen, error := net.Listen(tcpAddr.Network(), tcpAddr.String())
		if error != nil {
			LogError(error)
		} else {
			defer netListen.Close()
			LogInfo("Waiting for clients")
			for {
				connection, error := netListen.Accept()
				if error != nil {
					LogError("Client error: ", error)
				} else {
					go server.HandleClient(connection)
				}
			}
		}
	}
}
