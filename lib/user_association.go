package lib

import "strings"

type UserAssociation struct {
	Affiliation string
	Family string
}

func NewUserAssociation(affiliation string, family string) *UserAssociation {
	return &UserAssociation{
		Affiliation: affiliation,
		Family: family,
	}
}

func (ua *UserAssociation) IsUndergrad() bool {
	return strings.Contains(ua.Affiliation, "UGRD")
}

func (ua *UserAssociation) IsGradStudent() bool {
	return strings.Contains(ua.Affiliation, "GRAD")
}

func (ua *UserAssociation) IsAlum() bool {
	return strings.Contains(ua.Affiliation, "ALUM")
}

func (ua *UserAssociation) IsFaculty() bool {
	return ua.IsCurrentFaculty() || ua.IsEmeritus()
}

func (ua *UserAssociation) IsCurrentFaculty() bool {
	return strings.Contains(ua.Affiliation, "EMPF")
}

func (ua *UserAssociation) IsEmeritus() bool {
	return strings.Contains(ua.Affiliation, "EMER")
}

func (ua *UserAssociation) IsEmployee() bool {
	return strings.Contains(ua.Affiliation, "EMPL")
}

// Contracted employees are usually artist associates, and considered staff
func (ua *UserAssociation) IsContractedEmployee() bool {
	return strings.Contains(ua.Affiliation, "CONT")
}

func (ua *UserAssociation) IsResearchAssociate() bool {
	return strings.Contains(ua.Affiliation, "RESA")
}

func (ua *UserAssociation) IsClassEmployee() bool {
	return ua.Family == "E"
}

func (ua *UserAssociation) IsClassStudent() bool {
	return ua.Family == "S"
}

func (ua *UserAssociation) IsClassOther() bool {
	return ua.Family == "O"
}

func (ua *UserAssociation) IsCurrentStudent() bool {
	return ua.IsUndergrad() || ua.IsGradStudent()
}

func (ua *UserAssociation) IsStaff() bool {
	return ua.IsEmployee() || ua.IsContractedEmployee() || ua.IsResearchAssociate()
}

func (ua *UserAssociation) IsOnlyAlum() bool {
	return ua.Affiliation == "ALUM~"
}