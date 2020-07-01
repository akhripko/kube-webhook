package infosrv

import "net/http"

func serveVersion(w http.ResponseWriter, _ *http.Request) {
	writeFile("version", w)
}
