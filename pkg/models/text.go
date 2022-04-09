package models

import (
	"github.com/muesli/reflow/wordwrap"
	"jaytaylor.com/html2text"
)

type formatter struct {
	Width int
}

func (f *formatter) Text(s string) string {
	ss := convertHTML(s)
	return wordwrap.String(ss, f.Width)
}

func convertHTML(s string) string {
	text, err := html2text.FromString(s, html2text.Options{TextOnly: true})
	if err != nil {
		// log error
		return ""
	}
	return text
}
