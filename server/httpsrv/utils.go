package httpsrv

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	admissionv "k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/yaml"
)

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Error("http server: failed to build json response: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func decodeAdmissionReview(data []byte) (*admissionv.AdmissionReview, error) {
	var admissionReview admissionv.AdmissionReview
	err := yaml.Unmarshal(data, &admissionReview)
	return &admissionReview, err
}
