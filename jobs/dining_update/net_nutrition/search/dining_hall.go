package search

import (
	"strings"
)

/*
Examples:
Mission Park
Paresky
Driscoll
Eco Cafe
*/

type DiningHall struct {
	Name  string              `json:"name"`
	Menus map[string]MetaMenu `json:"menus"`
}

func (dh *DiningHall) StringWithIdent(ident string) string {
	var b strings.Builder
	dh.StringBuilderWithIdent(&b, ident)
	return b.String()
}

func (dh *DiningHall) StringBuilder(builder *strings.Builder) {
	dh.StringBuilderWithIdent(builder, "")
}

func (dh *DiningHall) StringBuilderWithIdent(builder *strings.Builder, ident string) {
	builder.WriteString(ident + dh.Name + ":\n")
	for _, metaMenu := range dh.Menus {
		metaMenu.StringBuilderWithIdent(builder, ident+"  ")
	}
}

func (dh *DiningHall) String() string {
	return dh.StringWithIdent("")
}

func NewDiningHall(name string) *DiningHall {
	return &DiningHall{
		Name:  name,
		Menus: make(map[string]MetaMenu),
	}
}
