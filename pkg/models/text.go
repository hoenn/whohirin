package models

import (
	"regexp"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"jaytaylor.com/html2text"
)

type formatter struct {
	Width int
}

func (f *formatter) Text(s string) string {
	ss := convertHTML(s)
	ss = highlightWords(ss)
	return wordwrap.String(ss, f.Width)
}

type hl struct {
	r *regexp.Regexp
	c string
}

var mapping = map[string]*hl{
	"go": &hl{
		r: regexp.MustCompile("(?i)\\b(golang|gopher|go)\\b"),
		c: "12",
	},
	"scala": &hl{
		r: regexp.MustCompile("(?i)\\bscala\\b"),
		c: "6",
	},
	"c": &hl{
		r: regexp.MustCompile("(?i)\\bc\\b"),
		c: "22",
	},
	"python": &hl{
		r: regexp.MustCompile("(?i)\\b(python|django|flask)\\b"),
		c: "90",
	},
	"ruby": &hl{
		r: regexp.MustCompile("(?i)\\b(ruby|rails)\\b"),
		c: "100",
	},
	"typescript": &hl{
		r: regexp.MustCompile("(?i)\\b(typescript)\\b"),
		c: "120",
	},
	"javascript": &hl{
		r: regexp.MustCompile("(?i)\\b(node.js|node|javascript|js)\\b"),
		c: "85",
	},
	"jvm": &hl{
		r: regexp.MustCompile("(?i)\\b(java|groovy|kotlin|jvm)\\b"),
		c: "143",
	},
	"remote": &hl{
		r: regexp.MustCompile("(?i)remote"),
		c: "45",
	},
	"backend": &hl{
		r: regexp.MustCompile("(?i)\\b(backend|back-end|back end)\\b"),
		c: "160",
	},
}

func highlightWords(s string) string {
	for _, highlighter := range mapping {
		s = highlighter.r.ReplaceAllStringFunc(s, func(match string) string {
			return lipgloss.NewStyle().Foreground(lipgloss.Color(highlighter.c)).Render(match)
		})
	}
	return s
}

func convertHTML(s string) string {
	text, err := html2text.FromString(s, html2text.Options{TextOnly: true})
	if err != nil {
		// log error
		return ""
	}
	return text
}
