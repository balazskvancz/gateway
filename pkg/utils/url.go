package utils

import "strings"

const (
	urlDelimiter = "/"
)

func GetUrlParts(url string) []string {
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}

	queryStart := strings.Index(url, "?")

	if queryStart > 0 {
		url = url[:queryStart]
	}

	return strings.Split(url, urlDelimiter)[1:]
}
