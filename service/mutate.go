package service

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	admission "k8s.io/api/admission/v1"
)

func (s *Service) Mutate(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	if req.DryRun != nil && *req.DryRun {
		log.Debugf("[mutate] dry run: %v", req)
		return s.BuildAdmissionResponse(req, nil)
	}
	log.Debug("[mutate] req.Kind: ", req.Kind.String(), " operation:", req.Operation)
	log.Debug(fmt.Sprintf("[mutate] req.UserInfo: %+v", req.UserInfo))
	// unmarshal old object
	var oldObj *AdmissionRequestObject
	if req.Operation != admission.Create {
		if len(req.OldObject.Raw) > 0 {
			if err := json.Unmarshal(req.OldObject.Raw, &oldObj); err != nil {
				log.Error("[mutate] failed to unmarshal old object: " + err.Error())
				return s.BuildAdmissionResponse(req, errors.New("failed to unmarshal old object: "+err.Error()))
			}
		}
	}
	// unmarshal new object
	var newObj *AdmissionRequestObject
	if req.Operation != admission.Delete {
		if err := json.Unmarshal(req.Object.Raw, &newObj); err != nil {
			log.Error("[mutate] failed to unmarshal new object: " + err.Error())
			return s.BuildAdmissionResponse(req, errors.New("failed to unmarshal new object: "+err.Error()))
		}
	}
	// extract owners
	owners := s.extractOwners(oldObj)
	log.Debug("[mutate] old object owners: ", owners)
	newOwners := s.extractOwners(newObj)
	log.Debug("[mutate] new object owners: ", newOwners)
	// get user
	user := req.UserInfo.Username
	// validate access & update list of owners if required
	switch {
	case s.isSystemUser(req.UserInfo):
		log.Debug("[mutate] user `", user, "` is system")
		if len(newOwners) == 0 {
			newOwners = owners
		}
		newOwners = removeItemFromItems(user, newOwners)
	case s.isAdminUser(req.UserInfo):
		log.Debug("[mutate] user `", user, "` is admin")
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
			log.Debug("[mutate] add user `", user, "` to owners: ", owners)
			break
		}
		// return Forbidden if user is not owner
		if !isOwner(user, owners) {
			respMsg := user + " is not owner of " + oldObj.Namespace + "/" + oldObj.Name
			log.Debug("[mutate] resp msg: ", respMsg)
			return s.BuildAdmissionResponse(req, ErrorForbidden(respMsg))
		}
		log.Debug("[mutate] user `", user, "` is owner")
		// new list was not provided
		if len(newOwners) == 0 {
			newOwners = owners
		}
	}
	// for DELETE operation the new object is empty. No need add patch.
	if req.Operation == admission.Delete {
		log.Debug("[mutate] delete operation: no need add patch")
		return s.BuildAdmissionResponse(req, nil)
	}
	// normalize new owners list
	newOwners = removeRepeatedItems(newOwners)
	// add required patch for mutation
	// build patch object
	log.Debug("[mutate] build patch...")
	annots := newAnnotations(withOwners(newOwners))
	patchObj := newPatchObject(
		withCap(2*len(annots)),
		withAnnotations(newObj, annots),
		withSpecTemplateAnnotations(newObj, annots),
	)
	// make patch json
	patch, err := json.Marshal(patchObj)
	if err != nil {
		log.Error("[mutate] failed to marshal patch: ", err.Error())
		return s.BuildAdmissionResponse(req, errors.New("failed to build patch object: "+err.Error()))
	}
	log.Debug("[mutate] patch response: ", string(patch))
	// build admission response with patch
	res := s.BuildAdmissionResponse(req, nil)
	res.Patch = patch
	res.PatchType = &jsonPatchType
	return res
}
