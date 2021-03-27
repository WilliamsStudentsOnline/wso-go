package notification

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/sideshow/apns2/payload"
)

type Notification struct {
	Title string
	Body  string
}

type Client struct {
	apns *APNS
}

func NewClient(cfg *config.Config) (*Client, error) {
	apns, err := NewAPNS(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{apns: apns}, nil
}

func (c *Client) NotifyIOS(notif *Notification, token string) (deleteToken bool, err error) {
	pl := payload.NewPayload().AlertTitle(notif.Title).AlertBody(notif.Body)
	return c.apns.Alert(pl, token)
}
