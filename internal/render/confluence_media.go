package render

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

var mediaSize = regexp.MustCompile(`^([0-9]+)(?:x([0-9]+))?$`)

// renderConfluenceMedia maps DokuWiki media options to Confluence image
// attributes while ignoring transport/display flags such as direct and
// nolink. DokuWiki's numeric size syntax is width-oriented; x-separated
// dimensions preserve both width and height.
func renderConfluenceMedia(filename, options string) string {
	attrs := ""
	for _, option := range strings.Split(options, "&") {
		option = strings.TrimSpace(option)
		m := mediaSize.FindStringSubmatch(option)
		if m == nil {
			continue
		}
		if m[2] == "" {
			attrs = fmt.Sprintf(` ac:width="%s"`, html.EscapeString(m[1]))
		} else {
			attrs = fmt.Sprintf(` ac:width="%s" ac:height="%s"`, html.EscapeString(m[1]), html.EscapeString(m[2]))
		}
		break
	}
	return fmt.Sprintf(`<ac:image%s><ri:attachment ri:filename="%s"/></ac:image>`, attrs, html.EscapeString(filename))
}
