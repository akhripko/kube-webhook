package service

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	admission "k8s.io/api/admission/v1"
)

func (s *Service) connect(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	log.Debug("[connect] raw object: ", string(req.Object.Raw))
	switch {
	case s.isSystemUser(req.UserInfo):
		log.Debug("[connect] user `", req.UserInfo.Username, "` is system")
	case s.isAdminUser(req.UserInfo):
		log.Debug("[connect] user `", req.UserInfo.Username, "` is admin")
	default:
		pod, err := s.api.GetPod(req.Namespace, req.Name)
		if err != nil {
			return s.BuildAdmissionResponse(req, err)
		}
		owners := getOwnersFromAnnotations(pod.Annotations)
		log.Debug("[connect] owners: ", owners, " user:", req.UserInfo.Username)
		if !isOwner(req.UserInfo.Username, owners) {
			respMsg := req.UserInfo.Username + " is not allowed to connect to " + req.Namespace + "/" + req.Name
			log.Debug("[connect] resp msg: ", respMsg)
			// forbidden
			return s.BuildAdmissionResponse(req, ErrorForbidden(respMsg))
		}
		log.Debug("[connect] user: ", req.UserInfo.Username, " is owner")
	}
	return s.BuildAdmissionResponse(req, nil)
}

func (s *Service) Validate(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	if req.DryRun != nil && *req.DryRun {
		log.Debugf("[validate] dry run: %v", req)
		return s.BuildAdmissionResponse(req, nil)
	}
	log.Debug("[validate] req.Kind: ", req.Kind.String(), " operation:", req.Operation)
	log.Debug(fmt.Sprintf("[validate] req.UserInfo: %+v", req.UserInfo))
	// skip validation on CREATE
	if req.Operation == admission.Create {
		log.Debug("[validate] create operation: skip validation")
		return s.BuildAdmissionResponse(req, nil)
	}
	// custom logic for CONNECT
	if req.Operation == admission.Connect {
		return s.connect(req)
	}
	// unmarshal old object
	var oldObj *AdmissionRequestObject
	if len(req.OldObject.Raw) > 0 {
		if err := json.Unmarshal(req.OldObject.Raw, &oldObj); err != nil {
			log.Error("[validate] failed to unmarshal old object: " + err.Error())
			return s.BuildAdmissionResponse(req, errors.New("failed to unmarshal old object: "+err.Error()))
		}
	}
	// extract owners
	owners := s.extractOwners(oldObj)
	log.Debug("[validate] object owners: ", owners)
	// get user
	user := req.UserInfo.Username
	// validate access
	switch {
	case s.isSystemUser(req.UserInfo):
		log.Debug("[validate] user `", user, "` is system")
		// operation is allowed
		return s.BuildAdmissionResponse(req, nil)
	case s.isAdminUser(req.UserInfo):
		log.Debug("[validate] user `", user, "` is admin")
		// operation is allowed
		return s.BuildAdmissionResponse(req, nil)
	case !isOwner(user, owners):
		respMsg := user + " is not owner of " + oldObj.Namespace + "/" + oldObj.Name
		log.Debug("[validate] resp msg: ", respMsg)
		// forbidden
		return s.BuildAdmissionResponse(req, ErrorForbidden(respMsg))
	}
	// operation is allowed
	return s.BuildAdmissionResponse(req, nil)
}
