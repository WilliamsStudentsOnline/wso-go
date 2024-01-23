package models

import (
	"encoding/json"
	"errors"
	"time"
)

type Location string

const (
	LocationUndefined  Location = ""
	LocationWilliams   Location = "WILLIAMS"
	LocationAlbany     Location = "ALBANY"
	LocationNYC        Location = "NYC"
	LocationBoston     Location = "BOSTON"
	LocationPittsfield Location = "PITTSFIELD"
)

// BulletinRide Schema
type BulletinRide struct {
	BaseSchema

	Body           string    `gorm:"size:65535" json:"body"`
	Date           time.Time `json:"date"`
	Offer          *bool     `gorm:"not null" json:"offer"`
	Source         Location  `json:"source" binding:"required" enums:",WILLIAMS,ALBANY,NYC,BOSTON,PITTSFIELD"`
	Destination    Location  `json:"destination" binding:"required" enums:",WILLIAMS,ALBANY,NYC,BOSTON,PITTSFIELD"`
	AvailableSeats uint      `gorm:"not null" json:"availableSeats"`
	Price          float64   `json:"price"`

	// Belongs to user
	UserID uint  `json:"userID"`
	User   *User `json:"user,omitempty"`
}

func (*BulletinRide) TableName() string {
	return "bulletin_rides"
}

func (location *Location) UnmarshalJSON(b []byte) error {
	// Define a secondary type to avoid ending up with a recursive call to json.Unmarshal
	type L Location
	var r = (*L)(location)
	err := json.Unmarshal(b, &r)
	if err != nil {
		panic(err)
	}
	switch *location {
	case LocationUndefined, LocationAlbany, LocationBoston, LocationNYC, LocationPittsfield, LocationWilliams:
		return nil
	}
	return errors.New("invalid location")
}
