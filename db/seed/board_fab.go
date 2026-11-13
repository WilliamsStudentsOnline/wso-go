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

var campusPlaces = []string{
	"Sawyer", "Schow", "Paresky", "Baxter", "Science Quad", "Hopkins Gate",
	"Mission", "Driscoll", "Goodrich", "Lasell Gym", "Williams Inn", "Route 2",
}

var rideHubs = []string{
	"Williamstown", "Albany", "Boston", "NYC", "Springfield", "Pittsfield",
	"Burlington", "Hartford", "Logan", "Albany Airport",
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
		src, dst := rideEndpoints()
		seed.Ride = &models.BoardRideInformation{
			OfferingRide: &offering,
			Source:       src,
			Destination:  dst,
		}
		label := "Looking for a ride"
		if offering {
			label = "Offering a ride"
		}
		seed.Thread.Title = fmt.Sprintf("%s: %s → %s", label, src, dst)
		depart := time.Now().Add(time.Duration(gofakeit.Number(12, 14*24)) * time.Hour)
		seed.Thread.StartsAt = &depart
		seed.Body = fmt.Sprintf(
			"Leaving %s for %s. Flexible on exact time — text if you want a seat.\n\n(Local seed data — not a real trip.)",
			src, dst,
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

// GenerateBoardReplyBody returns a short synthetic reply for local feeds.
func GenerateBoardReplyBody() string {
	replies := []string{
		"Interested — can you send details?",
		"I can help with this this weekend.",
		"Same, also looking.",
		"Thanks for posting — just messaged you.",
		"Still available?",
		"Count me in if you have room.",
	}
	return replies[gofakeit.Number(0, len(replies)-1)]
}

func pick(ss []string) string {
	return ss[gofakeit.Number(0, len(ss)-1)]
}

func rideEndpoints() (string, string) {
	a, b := pick(rideHubs), pick(rideHubs)
	for a == b {
		b = pick(rideHubs)
	}
	return a, b
}

func boardCopy(typ string) (title, body string) {
	place := pick(campusPlaces)
	switch typ {
	case models.BoardThreadTypeAnnouncement:
		titles := []string{
			"Mountain Day update — dining hours this week",
			"Reminder: register for spring housing",
			"Library extended hours before finals",
			"Campus shuttle schedule change starting Monday",
			"WSO maintenance window this Sunday night",
		}
		bodies := []string{
			"Quick heads-up for anyone planning around campus this week. Details are on the usual channels; reply here if something looks wrong.",
			"Please read before the deadline. If you already submitted, you can ignore this.",
			"Posting here so it’s findable later. Ping an admin if you need this pinned longer.",
		}
		return pick(titles), pick(bodies)
	case models.BoardThreadTypeJob:
		titles := []string{
			"Research assistant — short spring term gig",
			"Looking for a Paresky event staffer",
			"Paid note-taker for a large lecture",
			"Help needed setting up a club website",
			"Weekend dining shift coverage",
		}
		bodies := []string{
			fmt.Sprintf("Looking for a student available a few hours a week near %s. Fake seed posting — hours/pay are placeholders for local UI testing.", place),
			"Flexible schedule, on-campus only. This is synthetic seed data for local development.",
		}
		return pick(titles), pick(bodies)
	case models.BoardThreadTypeExchange:
		titles := []string{
			"Selling a mini fridge (good condition)",
			"Free: desk lamp by Sawyer entrance",
			"ISO winter coat, size M",
			"Trading textbooks — calc / chem",
			"Giving away a bike pump",
		}
		bodies := []string{
			fmt.Sprintf("Can meet near %s. Cash or Venmo. Seed listing — not a real sale.", place),
			"Message if you want it; first come first served. (Local seed data.)",
		}
		return pick(titles), pick(bodies)
	case models.BoardThreadTypeLostAndFound:
		titles := []string{
			"Lost: black North Face gloves",
			"Found: AirPods case near Paresky",
			"Lost water bottle — blue Hydro Flask",
			"Found student ID outside Schow",
			"Lost scarf on the Science Quad path",
		}
		bodies := []string{
			fmt.Sprintf("Last seen around %s. Text if you’ve seen it — thanks. (Seed post for local feeds.)", place),
			fmt.Sprintf("Picked this up near %s; describe it and it’s yours. Synthetic seed data.", place),
		}
		return pick(titles), pick(bodies)
	case models.BoardThreadTypeRide:
		return "Ride", "See ride details."
	default:
		titles := []string{
			"Anyone else confused about the new meal plan options?",
			"Best quiet study spots on a Sunday?",
			"Recommendations for a cheap printer on campus",
			"Is the gym crowded after 5pm?",
			"What’s the move for Thanksgiving travel?",
		}
		bodies := []string{
			"Curious what people are doing this year. Looking for practical tips, not hot takes.",
			fmt.Sprintf("Been hanging around %s and figured I’d ask here. Seed discussion for local UI.", place),
			"Reply with whatever worked for you — trying to collect a few options.",
		}
		return pick(titles), pick(bodies)
	}
}
