package board

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

const (
	minPollOptions      = 2
	maxPollOptions      = 4
	maxPollOptionLength = 255
)

type createThreadPollParams struct {
	Options  []string   `json:"options"`
	ClosesAt *time.Time `json:"closesAt"`
}

func newPoll(params *createThreadPollParams) (*models.BoardPoll, bool) {
	if len(params.Options) < minPollOptions || len(params.Options) > maxPollOptions {
		return nil, false
	}
	if params.ClosesAt != nil && !params.ClosesAt.After(time.Now()) {
		return nil, false
	}

	poll := &models.BoardPoll{ClosesAt: params.ClosesAt}
	seen := map[string]bool{}
	for i, label := range params.Options {
		label = strings.TrimSpace(label)
		if label == "" || utf8.RuneCountInString(label) > maxPollOptionLength {
			return nil, false
		}
		key := strings.ToLower(label)
		if seen[key] {
			return nil, false
		}
		seen[key] = true
		poll.Options = append(poll.Options, &models.BoardPollOption{Position: i, Label: label})
	}
	return poll, true
}

func (t *Controller) fillPollResults(c *gin.Context, poll *models.BoardPoll) error {
	if poll == nil {
		return nil
	}

	if userID := services.GetUserID(c); userID != 0 {
		optionID, err := t.pollModel.GetVoteOptionID(poll.ID, userID)
		if err != nil {
			return err
		}
		poll.MyVote = optionID
	}
	if poll.MyVote == nil && !poll.Closed {
		return nil
	}

	counts, err := t.pollModel.CountVotes(poll.ID)
	if err != nil {
		return err
	}
	total := 0
	for _, o := range poll.Options {
		votes := counts[o.ID]
		o.Votes = &votes
		total += votes
	}
	poll.TotalVotes = &total
	return nil
}
