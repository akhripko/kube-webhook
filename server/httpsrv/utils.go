package httpsrv

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Error("http server: failed to build json response: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
