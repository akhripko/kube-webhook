package service

import (
	"encoding/json"
	"testing"

	mock "github.com/akhripko/kube-webhook/k8s-acl-sv/service/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admission "k8s.io/api/admission/v1"
	authentication "k8s.io/api/authentication/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestService_Validate_Delete_Allowed(t *testing.T) {
	oldObj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userA",
			},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Delete,
		UserInfo: authentication.UserInfo{
			Username: "userA",
		},
		OldObject: runtime.RawExtension{
			Raw: oldObj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Validate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Validate_Create(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userA",
			},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Create,
		UserInfo: authentication.UserInfo{
			Username: "userA",
		},
		Object: runtime.RawExtension{
			Raw: obj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Validate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Validate_Delete_Forbidden(t *testing.T) {
	oldObj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Name:      "objectName",
			Namespace: "namespace",
			Annotations: annotations{
				ownersKey: "userB",
			},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Delete,
		UserInfo: authentication.UserInfo{
			Username: "userA",
		},
		OldObject: runtime.RawExtension{
			Raw: oldObj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: false,
		Result: &meta.Status{
			Message: ErrorForbidden("userA is not owner of namespace/objectName").Error(),
		},
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Validate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Validate_Connect_System(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		Namespace: "namespace1",
		Name:      "pod1",
		UID:       "abc",
		Operation: admission.Connect,
		UserInfo: authentication.UserInfo{
			Username: "systemUser",
		},
		Object: runtime.RawExtension{
			Raw: obj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	c := gomock.NewController(t)
	api := mock.NewMockAPI(c)
	api.EXPECT().GetPod("namespace1", "pod1").Return(&core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Annotations: map[string]string{"owners": "userA,userB"},
			Namespace:   "namespace1",
			Name:        "pod1",
		},
	}, nil).Times(1)

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
		api:         api,
	}

	resp := s.Validate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Validate_Connect_Admin(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		Namespace: "namespace1",
		Name:      "pod1",
		UID:       "abc",
		Operation: admission.Connect,
		UserInfo: authentication.UserInfo{
			Username: "adminUser",
		},
		Object: runtime.RawExtension{
			Raw: obj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	c := gomock.NewController(t)
	api := mock.NewMockAPI(c)
	api.EXPECT().GetPod("namespace1", "pod1").Return(&core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Annotations: map[string]string{"owners": "userA,userB"},
			Namespace:   "namespace1",
			Name:        "pod1",
		},
	}, nil).Times(1)

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
		api:         api,
	}

	resp := s.Validate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Validate_Connect_Allowed(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		Namespace: "namespace1",
		Name:      "pod1",
		UID:       "abc",
		Operation: admission.Connect,
		UserInfo: authentication.UserInfo{
			Username: "userA",
		},
		Object: runtime.RawExtension{
			Raw: obj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	c := gomock.NewController(t)
	api := mock.NewMockAPI(c)
	api.EXPECT().GetPod("namespace1", "pod1").Return(&core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Annotations: map[string]string{"owners": "userA,userB"},
			Namespace:   "namespace1",
			Name:        "pod1",
		},
	}, nil).Times(1)

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
		api:         api,
	}

	resp := s.Validate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Validate_Connect_Forbidden(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		Namespace: "namespace1",
		Name:      "pod1",
		UID:       "abc",
		Operation: admission.Connect,
		UserInfo: authentication.UserInfo{
			Username: "userC",
		},
		Object: runtime.RawExtension{
			Raw: obj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: false,
		Result: &meta.Status{
			Message: "Forbidden: userC is not allowed to connect to namespace1/pod1",
		},
	}

	c := gomock.NewController(t)
	api := mock.NewMockAPI(c)
	api.EXPECT().GetPod("namespace1", "pod1").Return(&core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Annotations: map[string]string{"owners": "userA,userB"},
			Namespace:   "namespace1",
			Name:        "pod1",
		},
	}, nil).Times(1)

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
		api:         api,
	}

	resp := s.Validate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}
