package httpsrv

import (
	admission "k8s.io/api/admission/v1"
)

type Service interface {
	DecodeAdmissionReview(data []byte) (*admission.AdmissionReview, error)
	Validate(*admission.AdmissionRequest) *admission.AdmissionResponse
	Mutate(*admission.AdmissionRequest) *admission.AdmissionResponse
	AddOwners(*admission.AdmissionRequest) *admission.AdmissionResponse
}
