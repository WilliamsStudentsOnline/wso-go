package seed

import (
	"fmt"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/brianvoe/gofakeit"
)

// BoardThreadSeed is a generated board thread plus OP body and optional ride info.
type BoardThreadSeed struct {
	Thread *models.BoardThread
	Body   string
	Ride   *models.BoardRideInformation
}

var boardTypeCycle = []string{
	models.BoardThreadTypeDiscussion,
	models.BoardThreadTypeAnnouncement,
	models.BoardThreadTypeJob,
	models.BoardThreadTypeExchange,
	models.BoardThreadTypeLostAndFound,
	models.BoardThreadTypeRide,
}

// GenerateBoardThread builds a synthetic board thread of the given type.
func GenerateBoardThread(typ string, userID uint, userName string) BoardThreadSeed {
	if !models.IsValidBoardThreadType(typ) {
		typ = models.BoardThreadTypeDiscussion
	}

	// Spread activity across ~6 months so local feeds aren't all "today"
	now := time.Now().Add(-time.Duration(gofakeit.Number(1, 180*24)) * time.Hour)
	title, body := boardCopy(typ)
	seed := BoardThreadSeed{
		Thread: &models.BoardThread{
			Type:           typ,
			Title:          title,
			UserID:         userID,
			ExUserName:     userName,
			RepliesEnabled: lib.TruePtr(),
			LastActive:     now,
		},
		Body: body,
	}

	if typ == models.BoardThreadTypeRide {
		offering := gofakeit.Bool()
		seed.Ride = &models.BoardRideInformation{
			OfferingRide: &offering,
			Source:       gofakeit.City(),
			Destination:  gofakeit.City(),
		}
		label := "request"
		if offering {
			label = "offer"
		}
		seed.Thread.Title = fmt.Sprintf(
			"Seed ride %s: %s → %s",
			label,
			seed.Ride.Source,
			seed.Ride.Destination,
		)
	}

	if typ == models.BoardThreadTypeExchange || typ == models.BoardThreadTypeLostAndFound {
		if gofakeit.Bool() {
			seed.Thread.Resolved = lib.BoolToPtr(false)
		}
	}

	return seed
}

// GenerateBoardThreadCycled picks type by index so small n still covers every type.
func GenerateBoardThreadCycled(i int, userID uint, userName string) BoardThreadSeed {
	typ := boardTypeCycle[i%len(boardTypeCycle)]
	return GenerateBoardThread(typ, userID, userName)
}

// GenerateBoardReplyBody returns a longer synthetic reply for local feeds.
func GenerateBoardReplyBody() string {
	return strings.TrimSpace(fmt.Sprintf(
		"%s\n\n%s",
		gofakeit.Sentence(12),
		gofakeit.Paragraph(1, 3, 8, "\n\n"),
	))
}

func boardCopy(typ string) (title, body string) {
	para := func() string {
		return gofakeit.Paragraph(2, 4, 12, "\n\n")
	}
	switch typ {
	case models.BoardThreadTypeAnnouncement:
		return "Seed announcement: " + gofakeit.Sentence(6),
			para()
	case models.BoardThreadTypeJob:
		return "Seed job: " + gofakeit.JobTitle(),
			fmt.Sprintf("%s\n\nHours / pay / contact are fake fixtures for local development only.\n\n%s", gofakeit.Sentence(10), para())
	case models.BoardThreadTypeExchange:
		return fmt.Sprintf("Seed exchange: %s", gofakeit.HipsterWord()),
			para()
	case models.BoardThreadTypeLostAndFound:
		return fmt.Sprintf("Seed lost & found: %s", gofakeit.Word()),
			fmt.Sprintf("Last seen near %s.\n\n%s", gofakeit.StreetName(), para())
	case models.BoardThreadTypeRide:
		return "Seed ride",
			para()
	default:
		return fmt.Sprintf("Seed discussion: %s", gofakeit.Sentence(5)),
			para()
	}
}
