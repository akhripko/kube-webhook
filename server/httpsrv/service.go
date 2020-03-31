package httpsrv

import (
	admission "k8s.io/api/admission/v1beta1"
)

type Service interface {
	ValidateMutation(*admission.AdmissionRequest) (*admission.AdmissionResponse, error)
}
