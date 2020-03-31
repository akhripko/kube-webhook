package httpsrv

import (
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	admission "k8s.io/api/admission/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// /mutate
func (s *HTTPSrv) mutateHandleFunc(w http.ResponseWriter, r *http.Request) {
	log.Debug("http server: mutate request")

	// read request payload
	var (
		data []byte
		err  error
	)
	if r.Body != nil {
		data, err = ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error("http server: failed to read payload: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if len(data) == 0 {
		log.Warn("http server: received empty payload")
		return
	}

	// decode payload
	admissionReview, err := decodeAdmissionReview(data)
	if err != nil {
		log.Error("http server: failed to decode data: ", err)
		admissionReview.Response = &admission.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: false,
			Result: &meta.Status{
				Message: err.Error(),
			},
		}
		writeJSON(w, admissionReview)
		return
	}
	log.Info("http server: received admission review request ", admissionReview.Request.UID)
	log.Debug(fmt.Sprintf("http server: admission request: %+v", admissionReview.Request))

	// handle request
	res, err := s.service.ValidateMutation(admissionReview.Request)
	if err != nil {
		log.Error("http server: failed to handle mutation request: ", err)
		res = &admission.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: false,
			Result: &meta.Status{
				Message: err.Error(),
			},
		}
	}
	admissionReview.Response = res
	writeJSON(w, admissionReview)
}
