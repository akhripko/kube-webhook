package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admission "k8s.io/api/admission/v1"
	authentication "k8s.io/api/authentication/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestService_Mutate_Create(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				"abc": "123",
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
		Patch:   []byte(`[{"op":"add","path":"/metadata/annotations/` + ownersKey + `","value":"userA"}]`),
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Create_WithOwners(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userB, userC, userC, userC, userB",
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
		Patch:   []byte(`[{"op":"replace","path":"/metadata/annotations/` + ownersKey + `","value":"userB,userC,userA"}]`),
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Update_Owner(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				"abc": "123",
			},
		},
	})
	require.NoError(t, err)

	oldObj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userA, userB",
			},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Update,
		UserInfo: authentication.UserInfo{
			Username: "userA",
		},
		Object: runtime.RawExtension{
			Raw: obj,
		},
		OldObject: runtime.RawExtension{
			Raw: oldObj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
		Patch:   []byte(`[{"op":"add","path":"/metadata/annotations/` + ownersKey + `","value":"userA,userB"}]`),
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Update_Owner_WithOwners(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userA, userB, userC",
			},
		},
	})
	require.NoError(t, err)

	oldObj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userA, userB",
			},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Update,
		UserInfo: authentication.UserInfo{
			Username: "userA",
		},
		Object: runtime.RawExtension{
			Raw: obj,
		},
		OldObject: runtime.RawExtension{
			Raw: oldObj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
		Patch:   []byte(`[{"op":"replace","path":"/metadata/annotations/` + ownersKey + `","value":"userA,userB,userC"}]`),
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Update_Admin(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				"abc": "123",
			},
		},
	})

	require.NoError(t, err)
	oldObj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userB",
			},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Update,
		UserInfo: authentication.UserInfo{
			Username: "adminUser",
		},
		Object: runtime.RawExtension{
			Raw: obj,
		},
		OldObject: runtime.RawExtension{
			Raw: oldObj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
		Patch:   []byte(`[{"op":"add","path":"/metadata/annotations/` + ownersKey + `","value":"userB"}]`),
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Create_Admin(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				"abc": "123",
			},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Create,
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
		Patch:   []byte(`[{"op":"add","path":"/metadata/annotations/` + ownersKey + `","value":"adminUser"}]`),
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Create_Admin_WithNewOwners(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userA, userB, userC",
			},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Create,
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
		Patch:   []byte(`[{"op":"replace","path":"/metadata/annotations/` + ownersKey + `","value":"userA,userB,userC"}]`),
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Update_Admin_WithNewOwners(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userA, userB, userC",
			},
		},
	})

	require.NoError(t, err)
	oldObj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userA, userB",
			},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Update,
		UserInfo: authentication.UserInfo{
			Username: "adminUser",
		},
		Object: runtime.RawExtension{
			Raw: obj,
		},
		OldObject: runtime.RawExtension{
			Raw: oldObj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
		Patch:   []byte(`[{"op":"replace","path":"/metadata/annotations/` + ownersKey + `","value":"userA,userB,userC"}]`),
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Create_System(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userA, userB",
			},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Create,
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
		Patch:   []byte(`[{"op":"replace","path":"/metadata/annotations/` + ownersKey + `","value":"userA,userB"}]`),
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Update_System(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				"abc": "123",
			},
		},
	})

	require.NoError(t, err)
	oldObj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				ownersKey: "userA, userB, systemUser",
			},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Update,
		UserInfo: authentication.UserInfo{
			Username: "systemUser",
		},
		Object: runtime.RawExtension{
			Raw: obj,
		},
		OldObject: runtime.RawExtension{
			Raw: oldObj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
		Patch:   []byte(`[{"op":"add","path":"/metadata/annotations/` + ownersKey + `","value":"userA,userB"}]`),
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Update_System_EmptyOwners(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{
				"abc": "123",
			},
		},
	})

	require.NoError(t, err)
	oldObj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{},
		},
	})
	require.NoError(t, err)

	// prepare request data
	req := admission.AdmissionRequest{
		UID:       "abc",
		Operation: admission.Update,
		UserInfo: authentication.UserInfo{
			Username: "systemUser",
		},
		Object: runtime.RawExtension{
			Raw: obj,
		},
		OldObject: runtime.RawExtension{
			Raw: oldObj,
		},
	}

	expected := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
		Patch:   []byte(`[{"op":"add","path":"/metadata/annotations/` + ownersKey + `","value":""}]`),
	}

	s := Service{
		adminUsers:  mapFromSlice([]string{"adminUser"}),
		systemUsers: mapFromSlice([]string{"systemUser"}),
	}

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Delete_Allowed(t *testing.T) {
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

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Delete_Forbidden(t *testing.T) {
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

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}

func TestService_Mutate_Forbidden(t *testing.T) {
	obj, err := json.Marshal(AdmissionRequestObject{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Annotations: annotations{},
		},
	})

	require.NoError(t, err)
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
		Operation: admission.Update,
		UserInfo: authentication.UserInfo{
			Username: "userA",
			Groups:   []string{"groupA"},
		},
		Object: runtime.RawExtension{
			Raw: obj,
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

	resp := s.Mutate(&req)

	assert.EqualValues(t, expected.UID, resp.UID)
	assert.EqualValues(t, expected.Allowed, resp.Allowed)
	assert.EqualValues(t, expected.Result, resp.Result)
	assert.EqualValues(t, expected.Patch, resp.Patch)
}
