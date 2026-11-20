package daily_messages_update

import (
	"bytes"
	"encoding/json"
)

type rawMessage struct {
	ID             looseInt        `json:"ID"`
	Type           looseString     `json:"type"`
	Title          looseString     `json:"title"`
	Content        looseString     `json:"content"`
	DmText         looseString     `json:"dm_text"`
	Permalink      looseString     `json:"permalink"`
	WebsiteURL     looseString     `json:"website_url"`
	TicketURL      looseString     `json:"ticket_url"`
	ThumbURLFull   looseString     `json:"thumb_url_full"`
	Author         looseString     `json:"author"`
	AuthorEmail    looseString     `json:"author_email"`
	LdapDepartment looseString     `json:"ldap_department"`
	DeptsArr       looseMap        `json:"depts_arr"`
	Orgs           looseString     `json:"orgs"`
	Venue          looseString     `json:"venue"`
	VenueRoom      looseString     `json:"venue_room"`
	AllDay         json.RawMessage `json:"all_day"`
	TimeFormatted  looseString     `json:"time_formatted"`
	Terms          looseTerms      `json:"terms"`
	Meta           looseMeta       `json:"meta"`
}

func isObject(b []byte) bool {
	return bytes.HasPrefix(bytes.TrimSpace(b), []byte("{"))
}

type looseInt int

func (i *looseInt) UnmarshalJSON(b []byte) error {
	var n json.Number
	if err := json.Unmarshal(bytes.Trim(b, `"`), &n); err != nil {
		return err
	}
	v, err := n.Int64()
	if err != nil {
		return err
	}
	*i = looseInt(v)
	return nil
}

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

type looseMap map[string]string

func (m *looseMap) UnmarshalJSON(b []byte) error {
	*m = nil
	if !isObject(b) {
		return nil
	}
	var v map[string]looseString
	if err := json.Unmarshal(b, &v); err != nil {
		return nil
	}
	out := make(looseMap, len(v))
	for k, s := range v {
		out[k] = string(s)
	}
	*m = out
	return nil
}

type looseTerms struct {
	Audience []string
}

func (t *looseTerms) UnmarshalJSON(b []byte) error {
	*t = looseTerms{}
	if !isObject(b) {
		return nil
	}
	var v struct {
		Audience json.RawMessage `json:"daily_message_audience"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return nil
	}
	var terms []struct {
		Slug looseString `json:"slug"`
	}
	if err := json.Unmarshal(v.Audience, &terms); err != nil {
		return nil
	}
	for _, term := range terms {
		if term.Slug != "" {
			t.Audience = append(t.Audience, string(term.Slug))
		}
	}
	return nil
}

type looseMeta map[string]string

func (m *looseMeta) UnmarshalJSON(b []byte) error {
	*m = nil
	if !isObject(b) {
		return nil
	}
	var v map[string]json.RawMessage
	if err := json.Unmarshal(b, &v); err != nil {
		return nil
	}
	out := make(looseMeta, len(v))
	for k, raw := range v {
		var vals []struct {
			Value looseString `json:"value"`
		}
		if json.Unmarshal(raw, &vals) == nil && len(vals) > 0 {
			out[k] = string(vals[0].Value)
		} else {
			out[k] = ""
		}
	}
	*m = out
	return nil
}
