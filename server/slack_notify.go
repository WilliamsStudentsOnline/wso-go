package server

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

// If SlackWebhookURL is defined, we post errors to slack
func SlackRecovery(cfg *config.Config) func(c *gin.Context) {
	if cfg.SlackWebhookURL == "" {
		return func(c *gin.Context) {}
	}

	return func(c *gin.Context) {
		defer func() {
			// If we returned errors or if we aborted and status is internal server error, post error to slack
			if len(c.Errors) > 0 || (c.IsAborted() && c.Writer.Status() >= 500) {
				sb := new(strings.Builder)

				// The case if we have the errors
				if c.Errors.Last() != nil {
					sb.WriteString("*An Error has Occurred!*\n" +
						"*Errors:*\n")
					for _, err := range c.Errors {
						sb.WriteString("```")
						sb.WriteString(err.Error())
						sb.WriteString("```\n")
					}
				} else if c.IsAborted() && c.Writer.Status() >= 500 {
					// If we don't know the error
					sb.WriteString("*An Unknown Error has Occurred!*\n")
				}

				sb.WriteString("*Hostname*\n")
				sb.WriteString(c.Request.Host)
				sb.WriteString("\n")

				sb.WriteString("*Backtrace*\n" +
					"```")

				for i := len(c.HandlerNames()) - 1; i >= 0; i-- {
					sb.WriteString(c.HandlerNames()[i])
					sb.WriteString("\n")
				}
				sb.WriteString("```")

				attachment := slack.Attachment{
					Color:         "danger",
					AuthorName:    "WSO-Go",
					AuthorSubname: cfg.Env,
					Text:          sb.String(),
					Ts:            json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
				}

				err := slack.PostWebhook(cfg.SlackWebhookURL, &slack.WebhookMessage{
					Attachments: []slack.Attachment{attachment},
				})
				if err != nil {
					log.Error("Slack Error", err)
				}
			}

		}()
		c.Next()
	}

}
