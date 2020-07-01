package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	admission "k8s.io/api/admission/v1"
	authentication "k8s.io/api/authentication/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestService_isAdmin(t *testing.T) {
	s := Service{
		adminUsers: map[string]struct{}{"userA": {}, "userB": {}},
	}
	assert.True(t, s.isAdminUser(authentication.UserInfo{
		Username: "userA",
	}))
	assert.False(t, s.isAdminUser(authentication.UserInfo{
		Username: "userC",
		Groups:   nil,
	}))
}

func TestService_isSystemUser(t *testing.T) {
	s := Service{
		systemUsers: map[string]struct{}{"userA": {}, "userB": {}},
	}
	assert.True(t, s.isSystemUser(authentication.UserInfo{
		Username: "userA",
	}))
	assert.False(t, s.isSystemUser(authentication.UserInfo{
		Username: "userC",
		Groups:   nil,
	}))
}

func TestService_BuildAdmissionResponse(t *testing.T) {
	req := admission.AdmissionRequest{
		UID: "reqID",
	}

	expect := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: false,
		Result: &meta.Status{
			Message: "some error",
		},
	}

	s := &Service{}
	resp := s.BuildAdmissionResponse(&req, errors.New("some error"))

	assert.Equal(t, expect, *resp)
}
