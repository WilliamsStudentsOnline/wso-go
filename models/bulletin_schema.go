package models

import "time"

// Bulletin Types
const (
	BulletinTypeLostAndFound = "lostAndFound"
	BulletinTypeJob          = "job"
	BulletinTypeExchange     = "exchange"
	BulletinTypeAnnouncement = "announcement"
)

// Bulletin Model Schema
type Bulletin struct {
	BaseSchema
	Type      string     `gorm:"index:index_bulletins_on_type;not null;" json:"type"`
	Title     string     `gorm:"not null;" json:"title"`
	Body      string     `gorm:"size:65535" json:"body"`
	StartDate time.Time  `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
	Offer     *bool      `json:"offer"`

	// Author information
	UserID uint  `json:"userID"`
	User   *User `json:"user,omitempty"`
}

// TableName returns the name of the bulletins table
func (*Bulletin) TableName() string {
	return "bulletins"
}

// NewBulletinWithID creates a new bulletin with ID
func NewBulletinWithID(bulletinID uint) Bulletin {
	return Bulletin{
		BaseSchema: BaseSchema{
			ID: bulletinID,
		},
	}
}

// Default start date to now if it is empty.
func (b *Bulletin) BeforeCreate() (err error) {
	if b.StartDate.IsZero() {
		b.StartDate = time.Now()
	}
	return
}

// IsLostAndFound checks if the bulletin is a lost and found posting
func (b *Bulletin) IsLostAndFound() bool {
	return b.Type == BulletinTypeLostAndFound
}

// IsJob checks if the bulletin is a job posting
func (b *Bulletin) IsJob() bool {
	return b.Type == BulletinTypeJob
}

// IsExchange checks if the the bulletin is an exchange
func (b *Bulletin) IsExchange() bool {
	return b.Type == BulletinTypeExchange
}

// IsAnnouncement checks if the bulletin is an announcement
func (b *Bulletin) IsAnnouncement() bool {
	return b.Type == BulletinTypeAnnouncement
}
