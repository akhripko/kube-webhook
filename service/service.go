package service

import (
	admission "k8s.io/api/admission/v1"
	authentication "k8s.io/api/authentication/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type API interface {
	GetPod(namespace string, name string) (*core.Pod, error)
}

func New(systemUsers []string, adminUsers []string, api API) *Service {
	return &Service{
		adminUsers:  mapFromSlice(adminUsers),
		systemUsers: mapFromSlice(systemUsers),
		api:         api,
	}
}

type Service struct {
	adminUsers  map[string]struct{}
	systemUsers map[string]struct{}
	api         API
}

func (s *Service) DecodeAdmissionReview(data []byte) (*admission.AdmissionReview, error) {
	var admissionReview admission.AdmissionReview
	err := yaml.Unmarshal(data, &admissionReview)
	return &admissionReview, err
}

func (s *Service) BuildAdmissionResponse(req *admission.AdmissionRequest, err error) *admission.AdmissionResponse {
	res := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}
	if err != nil {
		res.Allowed = false
		res.Result = &meta.Status{
			Message: err.Error(),
		}
	}
	return &res
}

func (s *Service) isSystemUser(userInfo authentication.UserInfo) bool {
	_, ok := s.systemUsers[userInfo.Username]
	return ok
}

func (s *Service) isAdminUser(userInfo authentication.UserInfo) bool {
	_, ok := s.adminUsers[userInfo.Username]
	return ok
}

func (s *Service) extractOwners(reqObj *AdmissionRequestObject) []string {
	if reqObj == nil || reqObj.Annotations == nil {
		return nil
	}
	return getOwnersFromAnnotations(reqObj.Annotations)
}
