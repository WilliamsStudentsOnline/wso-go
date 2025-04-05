package search

import (
	"fmt"
	"strings"
)

type MetaMenu interface {
	Menus() []*Menu
	GetName() string

	StringWithIdent(ident string) string
	StringBuilder(builder *strings.Builder)
	StringBuilderWithIdent(builder *strings.Builder, ident string)
	String() string
}

type Menu map[string][]string

func (m *Menu) StringWithIdent(ident string) string {
	var b strings.Builder
	m.StringBuilderWithIdent(&b, ident)
	return b.String()
}

func (m *Menu) StringBuilder(builder *strings.Builder) {
	m.StringBuilderWithIdent(builder, "")
}

func (m *Menu) StringBuilderWithIdent(builder *strings.Builder, ident string) {
	for subHeader, items := range *m {
		builder.WriteString(ident + subHeader + ":\n")
		for idx, item := range items {
			builder.WriteString(fmt.Sprintf("%s  %d. %s\n", ident, idx, item))
		}
	}
}

func (m *Menu) String() string {
	return m.StringWithIdent("")
}

type ConstantMenu struct {
	Name  string `json:"name"`
	*Menu `json:"menu"`
}

func (m *ConstantMenu) Menus() []*Menu {
	return []*Menu{m.Menu}
}

func (m *ConstantMenu) GetName() string {
	return m.Name
}

func (m *ConstantMenu) StringWithIdent(ident string) string {
	var b strings.Builder
	m.StringBuilderWithIdent(&b, ident)
	return b.String()
}

func (m *ConstantMenu) StringBuilder(builder *strings.Builder) {
	m.StringBuilderWithIdent(builder, "")
}

func (m *ConstantMenu) StringBuilderWithIdent(builder *strings.Builder, ident string) {
	builder.WriteString(ident + m.Name + ":\n")
	m.Menu.StringBuilderWithIdent(builder, ident+"  ")
}

func (m *ConstantMenu) String() string {
	return m.StringWithIdent("")
}

type DayMeals map[string]*Menu

type DailyMenu struct {
	Name string
	Days map[string]DayMeals
}

func (m *DailyMenu) GetName() string {
	return m.Name
}

func (m *DailyMenu) Menus() []*Menu {
	var menus []*Menu

	for _, dayMeals := range m.Days {
		for _, menu := range dayMeals {
			menus = append(menus, menu)
		}
	}

	return menus
}

func (m *DailyMenu) StringWithIdent(ident string) string {
	var b strings.Builder
	m.StringBuilderWithIdent(&b, ident)
	return b.String()
}

func (m *DailyMenu) StringBuilder(builder *strings.Builder) {
	m.StringBuilderWithIdent(builder, "")
}

func (m *DailyMenu) StringBuilderWithIdent(builder *strings.Builder, ident string) {
	builder.WriteString(ident + m.Name + ":\n")
	for day, dayMeals := range m.Days {
		builder.WriteString(ident + "  " + day + ":\n")
		for label, menu := range dayMeals {
			builder.WriteString(ident + "    " + label + "\n")
			menu.StringBuilderWithIdent(builder, ident+"      ")
		}
	}
}

func (m *DailyMenu) String() string {
	return m.StringWithIdent("")
}
