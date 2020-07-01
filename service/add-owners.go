package service

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	admission "k8s.io/api/admission/v1"
)

func (s *Service) AddOwners(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	if req.DryRun != nil && *req.DryRun {
		log.Debugf("[add owners] dry run: %v", req)
		return s.BuildAdmissionResponse(req, nil)
	}
	log.Debug("[add owners] req.Kind: ", req.Kind.String())
	log.Debug(fmt.Sprintf("[add owners] req.UserInfo: %+v", req.UserInfo))
	// for DELETE operation the new object is empty. No need add patch.
	if req.Operation == admission.Delete {
		log.Debug("[add owners] delete operation: no need add patch, user:", req.UserInfo.Username)
		return s.BuildAdmissionResponse(req, nil)
	}
	// log connect operation & stop
	if req.Operation == admission.Connect {
		log.Debug("[add owners] connect operation: no need add patch, user:", req.UserInfo.Username)
		return s.BuildAdmissionResponse(req, nil)
	}
	// unmarshal old object
	var oldObj *AdmissionRequestObject
	if req.Operation != admission.Create {
		if len(req.OldObject.Raw) > 0 {
			if err := json.Unmarshal(req.OldObject.Raw, &oldObj); err != nil {
				log.Error("[add owners] failed to unmarshal old object: " + err.Error())
				return s.BuildAdmissionResponse(req, nil)
			}
		}
	}
	// unmarshal new object
	var newObj *AdmissionRequestObject
	if err := json.Unmarshal(req.Object.Raw, &newObj); err != nil {
		log.Error("[add owners] failed to unmarshal new object: " + err.Error())
		return s.BuildAdmissionResponse(req, nil)
	}
	// extract owners
	owners := s.extractOwners(oldObj)
	log.Debug("[add owners] old object owners: ", owners)
	newOwners := s.extractOwners(newObj)
	log.Debug("[add owners] new object owners: ", newOwners)
	// get user
	user := req.UserInfo.Username
	switch {
	case s.isSystemUser(req.UserInfo):
		log.Debug("[add owners] user `", user, "` is system")
		if len(newOwners) == 0 {
			newOwners = owners
		}
		newOwners = removeItemFromItems(user, newOwners)
	case s.isAdminUser(req.UserInfo):
		log.Debug("[add owners] user `", user, "` is admin")
		// admin didn't set owners, let's use from old object
		if len(newOwners) == 0 {
			newOwners = owners
		}
		// object without owners, so admin is owner
		if len(newOwners) == 0 {
			newOwners = append(newOwners, user)
		}
	default:
		if req.Operation == admission.Create {
			// owners list was provided, current user should be added
			newOwners = append(newOwners, user)
			log.Debug("[add owners] add user `", user, "` to owners: ", owners)
			break
		}
		if isOwner(user, owners) {
			log.Debug("[add owners] user `", user, "` is owner (", owners, ")")
			// new owners list was not provided
			if len(newOwners) == 0 {
				newOwners = owners
			}
			break
		}
		// add current user to owners
		newOwners = append(owners, user) // nolint
		log.Debug("[add owners] add user `", user, "` to owners: ", owners)
	}
	// normalize new owners list
	newOwners = removeRepeatedItems(newOwners)
	// add required patch for mutation
	// build patch object
	log.Debug("[add owners] build patch...")
	annots := newAnnotations(withOwners(newOwners))
	patchObj := newPatchObject(
		withCap(2*len(annots)),
		withAnnotations(newObj, annots),
		withSpecTemplateAnnotations(newObj, annots),
	)
	// make patch json
	patch, err := json.Marshal(patchObj)
	if err != nil {
		log.Error("[add owners] failed to marshal patch: ", err.Error())
		return s.BuildAdmissionResponse(req, nil)
	}
	log.Debug("[add owners] patch response: ", string(patch))
	// build admission response with patch
	res := s.BuildAdmissionResponse(req, nil)
	res.Patch = patch
	res.PatchType = &jsonPatchType
	return res
}
