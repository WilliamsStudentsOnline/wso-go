package seed

import (
	"fmt"
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
// Titles/bodies are clearly fake fixtures — not production content.
func GenerateBoardThread(typ string, userID uint, userName string) BoardThreadSeed {
	if !models.IsValidBoardThreadType(typ) {
		typ = models.BoardThreadTypeDiscussion
	}

	now := time.Now().Add(-time.Duration(gofakeit.Number(0, 72)) * time.Hour)
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

func boardCopy(typ string) (title, body string) {
	switch typ {
	case models.BoardThreadTypeAnnouncement:
		return "Seed announcement: campus notice",
			"Synthetic announcement body for local development."
	case models.BoardThreadTypeJob:
		return "Seed job: campus helper wanted",
			"Synthetic job listing. No real employer or contact."
	case models.BoardThreadTypeExchange:
		return fmt.Sprintf("Seed exchange: %s", gofakeit.Word()),
			"Synthetic exchange post. Item does not exist."
	case models.BoardThreadTypeLostAndFound:
		return fmt.Sprintf("Seed lost & found: %s", gofakeit.Word()),
			"Synthetic lost-and-found post for local development."
	case models.BoardThreadTypeRide:
		return "Seed ride",
			"Synthetic ride post for local development."
	default:
		return fmt.Sprintf("Seed discussion: %s", gofakeit.Sentence(4)),
			"Synthetic discussion body for local development."
	}
}
