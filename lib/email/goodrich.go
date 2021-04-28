package email

import (
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/go-mail/mail"
	"go.uber.org/zap"
)

type GoodrichMailer struct {
	msgCh chan *mail.Message
	log   *zap.SugaredLogger
}

func NewGoodrichMailer(cfg *config.Config, log *zap.SugaredLogger) *GoodrichMailer {
	ch := make(chan *mail.Message)

	go func() {
		d := mail.NewDialer("wso.williams.edu", 465, "goodrich@wso.williams.edu", cfg.Secrets.GoodrichEmailPassword)
		d.StartTLSPolicy = mail.MandatoryStartTLS

		var s mail.SendCloser
		var err error
		open := false
		for {
			select {
			case m, ok := <-ch:
				if !ok {
					return
				}
				if !open {
					if s, err = d.Dial(); err != nil {
						log.Error(err)
						continue
					}
					open = true
				}
				if err := mail.Send(s, m); err != nil {
					log.Error(err)
					continue
				}

			// Close the connection to the SMTP server if no email was sent in
			// the last 30 seconds.
			case <-time.After(30 * time.Second):
				if open {
					if err := s.Close(); err != nil {
						log.Warn(err)
					}
					open = false
				}
			}
		}
	}()

	return &GoodrichMailer{
		msgCh: ch,
		log:   log,
	}
}

func (m *GoodrichMailer) SendMail(mail *mail.Message) {
	m.msgCh <- mail
}
