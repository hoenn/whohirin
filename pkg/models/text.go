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

type highlight struct {
	r     *regexp.Regexp
	color string
}

var mapping = map[string]*highlight{
	"go": &highlight{
		r:     regexp.MustCompile("(?i)\\b(golang|gopher|go)\\b"),
		color: "12",
	},
	"scala": &highlight{
		r:     regexp.MustCompile("(?i)\\bscala\\b"),
		color: "6",
	},
	"c": &highlight{
		r:     regexp.MustCompile("(?i)\\bc\\b"),
		color: "22",
	},
	"python": &highlight{
		r:     regexp.MustCompile("(?i)\\b(python|django|flask)\\b"),
		color: "90",
	},
	"ruby": &highlight{
		r:     regexp.MustCompile("(?i)\\b(ruby|rails)\\b"),
		color: "100",
	},
	"typescript": &highlight{
		r:     regexp.MustCompile("(?i)\\b(typescript)\\b"),
		color: "120",
	},
	"javascript": &highlight{
		r:     regexp.MustCompile("(?i)\\b(node.js|node|javascript|js)\\b"),
		color: "85",
	},
	"jvm": &highlight{
		r:     regexp.MustCompile("(?i)\\b(java|groovy|kotlin|jvm)\\b"),
		color: "143",
	},
	"remote": &highlight{
		r:     regexp.MustCompile("(?i)remote"),
		color: "45",
	},
	"backend": &highlight{
		r:     regexp.MustCompile("(?i)\\b(backend|back-end|back end)\\b"),
		color: "160",
	},
}

func highlightWords(s string) string {
	for _, highlighter := range mapping {
		s = highlighter.r.ReplaceAllStringFunc(s, func(match string) string {
			return lipgloss.NewStyle().Foreground(lipgloss.Color(highlighter.color)).Render(match)
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
