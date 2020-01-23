package lib

type UserAssociation struct {
	MemberGroups []string
}

func NewUserAssociation(memberGroups []string) *UserAssociation {
	return &UserAssociation{
		MemberGroups: memberGroups,
	}
}

func (ua *UserAssociation) IsCurrentFaculty() bool {
	return stringsContains(ua.MemberGroups, "CN=Williams-Faculty,OU=Groups-williams,DC=ad,DC=williams,DC=edu")
}

func (ua *UserAssociation) IsEmeritus() bool {
	return stringsContains(ua.MemberGroups, "CN=Williams-Emeriti,OU=Groups-williams,DC=ad,DC=williams,DC=edu")
}

func (ua *UserAssociation) IsCurrentStudent() bool {
	return stringsContains(ua.MemberGroups, "CN=Williams-Student,OU=Groups-williams,DC=ad,DC=williams,DC=edu")
}

func (ua *UserAssociation) IsContractor() bool {
	return stringsContains(ua.MemberGroups, "CN=Williams-Contractor,OU=Groups-williams,DC=ad,DC=williams,DC=edu")
}

func (ua *UserAssociation) IsStaffEmployee() bool {
	return stringsContains(ua.MemberGroups, "CN=Williams-Staff,OU=Groups-williams,DC=ad,DC=williams,DC=edu")
}

func (ua *UserAssociation) IsTemp() bool {
	return stringsContains(ua.MemberGroups, "CN=Williams-Temp,OU=Groups-williams,DC=ad,DC=williams,DC=edu")
}

func (ua *UserAssociation) IsAffiliate() bool {
	return stringsContains(ua.MemberGroups, "CN=Williams-Affiliate,OU=Groups-williams,DC=ad,DC=williams,DC=edu")
}

func (ua *UserAssociation) IsResearcher() bool {
	return stringsContains(ua.MemberGroups, "CN=Williams-Resa,OU=Groups-williams,DC=ad,DC=williams,DC=edu")
}

func (ua *UserAssociation) IsGradStudent() bool {
	return stringsContains(ua.MemberGroups, "CN=GEStudents,OU=williams,DC=ad,DC=williams,DC=edu")
}

func (ua *UserAssociation) IsStaff() bool {
	return ua.IsStaffEmployee() || ua.IsContractor() || ua.IsTemp() || ua.IsResearcher() || ua.IsAffiliate()
}

func (ua *UserAssociation) IsFaculty() bool {
	return ua.IsCurrentFaculty() || ua.IsEmeritus()
}

func (ua *UserAssociation) IsStudent() bool {
	return ua.IsCurrentStudent() || ua.IsGradStudent()
}

func stringsContains(slice []string, str string) bool {
	for _, val := range slice {
		if val == str {
			return true
		}
	}

	return false
}
