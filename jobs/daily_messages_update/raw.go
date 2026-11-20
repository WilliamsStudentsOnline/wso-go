package daily_messages_update

import (
	"bytes"
	"encoding/json"
)

type rawMessage struct {
	ID             int             `json:"ID"`
	Type           string          `json:"type"`
	Title          string          `json:"title"`
	Content        string          `json:"content"`
	DmText         string          `json:"dm_text"`
	Permalink      string          `json:"permalink"`
	WebsiteURL     looseString     `json:"website_url"`
	TicketURL      looseString     `json:"ticket_url"`
	ThumbURLFull   looseString     `json:"thumb_url_full"`
	Author         string          `json:"author"`
	AuthorEmail    string          `json:"author_email"`
	LdapDepartment string          `json:"ldap_department"`
	DeptsArr       looseMap        `json:"depts_arr"`
	Orgs           string          `json:"orgs"`
	Venue          looseString     `json:"venue"`
	VenueRoom      looseString     `json:"venue_room"`
	AllDay         json.RawMessage `json:"all_day"`
	TimeFormatted  string          `json:"time_formatted"`
	Terms          struct {
		Audience []struct {
			Slug string `json:"slug"`
		} `json:"daily_message_audience"`
	} `json:"terms"`
	Meta map[string][]struct {
		Value looseString `json:"value"`
	} `json:"meta"`
}

func (r rawMessage) metaValue(key string) string {
	if v := r.Meta[key]; len(v) > 0 {
		return string(v[0].Value)
	}
	return ""
}

// upstream sends false or null in place of empty strings
type looseString string

func (s *looseString) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		*s = ""
		return nil
	}
	*s = looseString(v)
	return nil
}

// PHP serializes an empty map as []
type looseMap map[string]string

func (m *looseMap) UnmarshalJSON(b []byte) error {
	if bytes.HasPrefix(bytes.TrimSpace(b), []byte("[")) {
		*m = nil
		return nil
	}
	var v map[string]string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*m = v
	return nil
}
