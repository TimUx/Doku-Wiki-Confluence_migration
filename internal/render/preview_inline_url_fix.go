package render

import "regexp"

func init() {
	// Keep the DokuWiki italic matcher from crossing generated HTML tags.
	italic = regexp.MustCompile(`//([^/<\n]+?)//`)
}
