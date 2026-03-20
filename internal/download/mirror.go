package download

import "net/http"

func ResolveURL(primary string, mirrors []string) string {
	if checkURL(primary) {
		return primary
	}
	for _, m := range mirrors {
		if checkURL(m) {
			return m
		}
	}
	return primary
}

func checkURL(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
