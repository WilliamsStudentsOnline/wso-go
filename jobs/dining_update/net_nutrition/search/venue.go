package search

import (
	"strings"
)

type Venue struct {
	Name string `json:"name"`
	DiningHalls []*DiningHall `json:"dining_halls"`
}

func (v *Venue) StringWithIdent(ident string) string {
	var b strings.Builder
	v.StringBuilderWithIdent(&b, ident)
	return b.String()
}

func (v *Venue) StringBuilder(builder *strings.Builder) {
	v.StringBuilderWithIdent(builder, "")
}

func (v *Venue) StringBuilderWithIdent(builder *strings.Builder, ident string) {
	builder.WriteString(ident + v.Name + ":\n")
	for _, dh := range v.DiningHalls {
		dh.StringBuilderWithIdent(builder, ident + "  ")
	}
}

func (v *Venue) String() string {
	return v.StringWithIdent("")
}