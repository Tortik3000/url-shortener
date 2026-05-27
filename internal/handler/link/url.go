package link

import "strings"

func buildShortURL(baseURL, code string) string {
	return strings.TrimRight(baseURL, "/") + "/" + code
}
