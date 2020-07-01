package httpsrv

import (
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// /add-owners
func (s *HTTPSrv) addOwnersHandleFunc(w http.ResponseWriter, r *http.Request) {
	log.Debug("http server 'add owners': mutate request > ", r.Method)
	// read request payload
	var (
		payload []byte
		err     error
	)
	if r.Body != nil {
		payload, err = ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error("http server 'add owners': failed to read payload: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if len(payload) == 0 {
		log.Error("http server 'add owners': received empty payload")
		http.Error(w, "payload is empty", http.StatusBadRequest)
		return
	}
	// decode payload
	admissionReview, err := s.service.DecodeAdmissionReview(payload)
	if err != nil {
		log.Error("http server 'add owners': failed to decode admission review: ", err.Error())
		http.Error(w, "failed to decode admission review: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// check if request is not empty
	if admissionReview.Request == nil {
		log.Error("http server 'add owners': received admission with empty request")
		http.Error(w, "admission request is empty", http.StatusBadRequest)
		return
	}
	log.Debug("http server 'add owners': received admission review request UID=", admissionReview.Request.UID)
	// handle request
	admissionReview.Response = s.service.AddOwners(admissionReview.Request)
	// write response
	writeJSON(w, admissionReview)
}
