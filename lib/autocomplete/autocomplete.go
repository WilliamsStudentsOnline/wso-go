package autocomplete

type Autocomplete interface {
	AreaOfStudy(q string) ([]ACEntry, error)
	Course(q string) ([]ACEntry, error)
	Professor(q string) ([]ACEntry, error)
	Tag(q string) ([]ACEntry, error)
	Factrak(q string) ([]ACEntry, error)
}

type ACEntry struct {
	ID    uint   `json:"id"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

const (
	ACTypeCourse    = "course"
	ACTypeArea      = "area"
	ACTypeProfessor = "professor"
	ACTypeTag       = "tag"
)
