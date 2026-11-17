package catalog_update

import (
	"slices"
	"strings"
	"unicode"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Deliberately ignores at_williams, since past instructors must match too
type InstructorMatcher struct {
	byUnix map[string]uint
	byName map[string][]uint // normalized name/nickname keys → candidate user IDs
	users  map[uint]models.User
}

type instructorMatchStats struct {
	UIDHits   int
	NameHits  int
	Unmatched int
	Ambiguous int
}

func NewInstructorMatcher(db *gorm.DB) (*InstructorMatcher, error) {
	var users []models.User
	if err := db.Select("id, unix_id, name, nickname, type").Find(&users).Error; err != nil {
		return nil, err
	}

	m := &InstructorMatcher{
		byUnix: make(map[string]uint, len(users)),
		byName: make(map[string][]uint),
		users:  make(map[uint]models.User, len(users)),
	}
	for i := range users {
		u := users[i]
		m.users[u.ID] = u
		if u.UnixID != "" {
			m.byUnix[strings.ToLower(u.UnixID)] = u.ID
		}
		for _, key := range userNameKeys(u) {
			m.byName[key] = appendUnique(m.byName[key], u.ID)
		}
	}
	return m, nil
}

func (m *InstructorMatcher) Match(unixID, first, middle, last string) (userID *uint, via string) {
	unixID = strings.TrimSpace(unixID)
	first = strings.TrimSpace(first)
	middle = strings.TrimSpace(middle)
	last = strings.TrimSpace(last)

	if unixID != "" {
		if id, ok := m.byUnix[strings.ToLower(unixID)]; ok {
			return uintPtr(id), "unix_id"
		}
	}

	keys := catalogNameKeys(first, middle, last)
	if len(keys) == 0 {
		return nil, ""
	}

	candidates := make(map[uint]struct{})
	for _, key := range keys {
		for _, id := range m.byName[key] {
			candidates[id] = struct{}{}
		}
	}
	if len(candidates) == 0 {
		return nil, ""
	}

	id, ok := pickUniqueMatch(candidates, m.users)
	if !ok {
		return nil, "ambiguous"
	}
	return uintPtr(id), "name"
}

func RematchOfferingInstructors(db *gorm.DB, log *zap.SugaredLogger) (instructorMatchStats, error) {
	stats := instructorMatchStats{}
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	matcher, err := NewInstructorMatcher(db)
	if err != nil {
		return stats, err
	}

	var rows []models.CourseOfferingInstructor
	if err := db.Where("user_id IS NULL").Find(&rows).Error; err != nil {
		return stats, err
	}

	for i := range rows {
		row := &rows[i]
		first, middle, last := splitListedName(row.NameAsListed)
		userID, via := matcher.Match(row.UnixID, first, middle, last)
		switch via {
		case "unix_id":
			stats.UIDHits++
		case "name":
			stats.NameHits++
		case "ambiguous":
			stats.Ambiguous++
			continue
		default:
			stats.Unmatched++
			continue
		}
		if err := db.Model(row).Update("user_id", *userID).Error; err != nil {
			return stats, err
		}
	}
	log.Infof("Rematch instructors: uid=%d name=%d unmatched=%d ambiguous=%d",
		stats.UIDHits, stats.NameHits, stats.Unmatched, stats.Ambiguous)
	return stats, nil
}

func NormalizePersonName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	folded, _, err := transform.String(t, s)
	if err != nil {
		folded = s
	}
	var b strings.Builder
	b.Grow(len(folded))
	prevSpace := false
	for _, r := range strings.ToLower(folded) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevSpace = false
		case r == '\'' || r == '.' || r == '`':
			prevSpace = false
		default:
			if !prevSpace && b.Len() > 0 {
				b.WriteByte(' ')
				prevSpace = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func catalogNameKeys(first, middle, last string) []string {
	var keys []string
	add := func(parts ...string) {
		filtered := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				filtered = append(filtered, p)
			}
		}
		if len(filtered) == 0 {
			return
		}
		key := NormalizePersonName(strings.Join(filtered, " "))
		if key != "" {
			keys = appendUnique(keys, key)
		}
	}
	add(first, middle, last)
	add(first, last)
	return keys
}

func userNameKeys(u models.User) []string {
	var keys []string
	add := func(s string) {
		key := NormalizePersonName(s)
		if key != "" {
			keys = appendUnique(keys, key)
		}
	}

	name := strings.TrimSpace(u.Name)
	add(name)

	parts := strings.Fields(name)
	if len(parts) >= 3 {
		add(parts[0] + " " + parts[len(parts)-1])
	}

	if u.Nickname != nil {
		nick := strings.TrimSpace(*u.Nickname)
		if nick != "" && len(parts) >= 1 {
			last := parts[len(parts)-1]
			add(nick + " " + last)
			if len(parts) >= 2 {
				// Nickname + everything after the first name, for multi-word surnames
				add(nick + " " + strings.Join(parts[1:], " "))
			}
		}
	}
	return keys
}

func splitListedName(name string) (first, middle, last string) {
	parts := strings.Fields(strings.TrimSpace(name))
	switch len(parts) {
	case 0:
		return "", "", ""
	case 1:
		return parts[0], "", ""
	case 2:
		return parts[0], "", parts[1]
	default:
		return parts[0], strings.Join(parts[1:len(parts)-1], " "), parts[len(parts)-1]
	}
}

func pickUniqueMatch(candidates map[uint]struct{}, users map[uint]models.User) (uint, bool) {
	if len(candidates) == 1 {
		for id := range candidates {
			return id, true
		}
	}

	var profs []uint
	for id := range candidates {
		if users[id].Type == models.UserTypeProfessor {
			profs = append(profs, id)
		}
	}
	if len(profs) == 1 {
		return profs[0], true
	}
	return 0, false
}

func appendUnique[T comparable](xs []T, x T) []T {
	if slices.Contains(xs, x) {
		return xs
	}
	return append(xs, x)
}

func uintPtr(id uint) *uint {
	return &id
}
