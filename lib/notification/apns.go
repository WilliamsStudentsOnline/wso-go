package notification

import (
	"path/filepath"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

type APNS struct {
	Client *apns2.Client
	Topic  string
}

func NewAPNS(cfg *config.Config) (*APNS, error) {
	authKeyPathFull, err := filepath.Abs(cfg.APNSAuthKey)
	if err != nil {
		return nil, err
	}

	authKey, err := token.AuthKeyFromFile(authKeyPathFull)
	if err != nil {
		return nil, err
	}

	tk := &token.Token{
		AuthKey: authKey,
		KeyID:   cfg.APNSKeyID,
		TeamID:  cfg.APNSTeamID,
	}

	client := apns2.NewTokenClient(tk)
	if cfg.APNSProduction {
		client = client.Production()
	} else {
		client = client.Development()
	}

	return &APNS{
		Client: client,
		Topic:  cfg.APNSTopic,
	}, nil
}

func (ns *APNS) Alert(payload *payload.Payload, token string) (deleteToken bool, err error) {
	notif := &apns2.Notification{
		DeviceToken: token,
		Topic:       ns.Topic,
		PushType:    apns2.PushTypeAlert,
		Payload:     payload,
	}

	res, err := ns.Client.Push(notif)
	if res != nil {
		deleteToken = res.Reason == apns2.ReasonUnregistered || res.Reason == apns2.ReasonTopicDisallowed ||
			res.Reason == apns2.ReasonDeviceTokenNotForTopic || res.Reason == apns2.ReasonBadDeviceToken
	}
	return
}
