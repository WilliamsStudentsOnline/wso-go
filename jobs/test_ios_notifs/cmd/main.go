package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

func main() {
	var authKeyPath string
	var tokenKeyID string
	var tokenTeamID string
	var production bool
	var deviceToken string
	var topic string

	flag.StringVar(&authKeyPath, "authkey", "", "auth key token")
	flag.StringVar(&tokenKeyID, "token-key", "LBFFL2GS2N", "auth token key id")
	flag.StringVar(&tokenTeamID, "token-team", "99PB38A22W", "auth token team id")
	flag.BoolVar(&production, "production", false, "use production APN")
	flag.StringVar(&deviceToken, "device", "", "device token")
	flag.StringVar(&topic, "topic", "edu.williams.wso.WilliamsMobile", "notification topic (app bundle ID)")
	flag.Parse()

	authKeyPathFull, err := filepath.Abs(authKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	authKey, err := token.AuthKeyFromFile(authKeyPathFull)
	if err != nil {
		log.Fatal("token error:", err)
	}

	token := &token.Token{
		AuthKey: authKey,
		// KeyID from developer account (Certificates, Identifiers & Profiles -> Keys)
		KeyID: tokenKeyID,
		// TeamID from developer account (View Account -> Membership)
		TeamID: tokenTeamID,
	}

	// TODO: DEVELOPMENT HERE
	client := apns2.NewTokenClient(token)
	if production {
		client = client.Production()
	} else {
		client = client.Development()
	}

	notification := &apns2.Notification{}
	notification.DeviceToken = deviceToken
	notification.Topic = topic
	notification.PushType = apns2.PushTypeAlert
	notification.Payload = payload.NewPayload().AlertTitle("hello world").AlertBody("body of alert")

	res, err := client.Push(notification)

	if err != nil {
		log.Fatal("Error:", err)
	}

	fmt.Printf("%v %v %v\n", res.StatusCode, res.ApnsID, res.Reason)
}
