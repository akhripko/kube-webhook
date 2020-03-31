package service

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	admission "k8s.io/api/admission/v1beta1"
	apitypes "k8s.io/api/apps/v1beta1"
)

func New() (*Service, error) {
	return &Service{}, nil
}

type Service struct{}

func (s *Service) ValidateMutation(req *admission.AdmissionRequest) (*admission.AdmissionResponse, error) {
	var deployment *apitypes.Deployment
	if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
		return nil, fmt.Errorf("unable unmarshal deployment json object %v", err)
	}

	log.Debug(fmt.Sprintf("req.UserInfo: %+v", req.UserInfo))
	log.Debug(fmt.Sprintf("deployment: %+v", deployment))
	log.Debug(fmt.Sprintf("req: %+v", req))

	admissionResponse := &admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	return admissionResponse, nil
}
