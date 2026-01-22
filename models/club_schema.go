package models

type Category string

const (
	//Using the same categories as defined in interal spreadsheet of RSOs
	CategoryClubSports                      Category = "club sport"
	CategoryDance                           Category = "dance performance"
	CategoryAcademicAndHonors               Category = "academic and honors"
	CategoryAdvocacyDebatePolitical         Category = "advocacy, debate, and political"
	CategoryAffinityCulturallyBasedMinco    Category = "affinity, culterally based, and MiNCO"
	CategoryArtsEntertainment               Category = "arts and entertainment"
	CategoryCommunitySupportServiceLearning Category = "community Support and/or Service Learning"
	CategoryEnvironmentSustainability       Category = "environmental and sustainability"
	CategoryHealthWellness                  Category = "health and wellness"
	CategoryProfessionalCareer              Category = "professional and career"
	CategoryRecreationSports                Category = "recreation and sports"
	CategoryReligiousSpiritual              Category = "religious and spiritual"
)

// Clubtrak Schema
type Club struct {
	BaseSchema
	Name               string   `gnorm:"not null" json:"name"`
	Category           Category `gnorm:"type=ENUM('club sport', 'dance performance', 'academic and honors', 'advocacy, debate, and political', 'affinity, culterally based, and MiNCO', 'arts and entertainment', 'community support and/or service learning', 'environmental and sustainability', 'health and wellness', 'professional and career', 'recreation and sports', 'religious and spiritual');not null" json:"category"`
	Subscribers        int      `json:"subscribers"`
	MeetingDescription string   `json:"meetingDescription"`
	ClubDescription    string   `gnorm:"not null" json:"clubDescription"`
	ClubPhotoFilePath  string   `json:"clubPhotoFilePath"`
	ContactEmail       string   `json:"contactEmail"`
	ContactPhoneNumber string   `json:"contactPhoneNumber"`
	Website            string   `json:"website"`

	// Club leader's DB ID
	ClubAdminID uint `gnorm:"not null" json:"clubAdminID"`
}

func (*Club) TableName() string {
	return "clubs"
}
