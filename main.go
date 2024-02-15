//HBXchat/main.go

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"HBX-CRYPTO-CHAT/src"

	"github.com/sirupsen/logrus"
)

const figlet = `
HBX CRYPTO CHAT 
`

func init() {
	// Log as Text with color
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
	})

	// Log to stdout
	logrus.SetOutput(os.Stdout)
}

func main() {
	username, chatroom, loglevel, discovery := parseInputFlags()
	setLogLevel(loglevel)
	displayWelcomeMessage()
	chatapp := initializeChat(username, chatroom, discovery)
	runChatUI(chatapp)
}

func parseInputFlags() (username, chatroom, loglevel, discovery *string) {
	username = flag.String("user", "", "username to use in the chatroom.")
	chatroom = flag.String("room", "", "chatroom to join.")
	loglevel = flag.String("log", "", "level of logs to print.")
	discovery = flag.String("discover", "", "method to use for discovery.")
	flag.Parse()
	return
}

func setLogLevel(loglevel *string) {
	switch *loglevel {
	case "panic", "PANIC":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal", "FATAL":
		logrus.SetLevel(logrus.FatalLevel)
	case "error", "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn", "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	case "info", "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug", "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "trace", "TRACE":
		logrus.SetLevel(logrus.TraceLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func displayWelcomeMessage() {
	fmt.Println(figlet)
	fmt.Println("The HBXchat Application is starting.")
	fmt.Println("This may take upto 30 seconds.")
	fmt.Println()
}

// Adjusted to return *src.ChatRoom instead of *src.ChatApp
func initializeChat(username, chatroom, discovery *string) *src.ChatRoom {
	p2phost := src.NewP2P()
	logrus.Infoln("Completed P2P Setup")

	switch *discovery {
	case "announce":
		p2phost.AnnounceConnect()
	case "advertise":
		p2phost.AdvertiseConnect()
	default:
		p2phost.AdvertiseConnect()
	}
	logrus.Infoln("Connected to Service Peers")

	chatroomInstance, _ := src.JoinChatRoom(p2phost, *username, *chatroom)
	logrus.Infof("Joined the '%s' chatroom as '%s'", chatroomInstance.RoomName, chatroomInstance.UserName)
	time.Sleep(time.Second * 5) // Wait for network setup to complete

	return chatroomInstance
}

// Adjusted to accept *src.ChatRoom instead of *src.ChatApp
func runChatUI(chatroom *src.ChatRoom) {
	ui := src.NewUI(chatroom)
	ui.Run()
}
