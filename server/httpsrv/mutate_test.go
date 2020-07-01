package httpsrv

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mock "github.com/akhripko/kube-webhook/k8s-acl-sv/server/httpsrv/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admission "k8s.io/api/admission/v1"
)

func TestHTTPSrv_mutateHandleFunc(t *testing.T) {
	// prepare request data
	admissionRequest := admission.AdmissionRequest{
		UID: "abc",
	}
	admissionReview := admission.AdmissionReview{
		Request: &admissionRequest,
	}
	admissionReviewJSON, err := json.Marshal(admissionReview)
	require.NoError(t, err)

	admissionResponse := admission.AdmissionResponse{}

	// prepare expected response
	admissionReview.Response = &admissionResponse
	expectedResponseJSON, err := json.Marshal(admissionReview)
	require.NoError(t, err)
	admissionReview.Response = nil

	c := gomock.NewController(t)

	service := mock.NewMockService(c)
	//DecodeAdmissionReview(data []byte) (*admission.AdmissionReview, error)
	service.EXPECT().DecodeAdmissionReview(admissionReviewJSON).DoAndReturn(func(data []byte) (*admission.AdmissionReview, error) {
		return &admissionReview, nil
	})
	//Mutate(*admission.AdmissionRequest) *admission.AdmissionResponse
	service.EXPECT().Mutate(&admissionRequest).DoAndReturn(func(*admission.AdmissionRequest) *admission.AdmissionResponse {
		return &admissionResponse
	})

	s := HTTPSrv{
		service: service,
	}
	h := s.buildHandler()

	rr := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/mutate", bytes.NewReader(admissionReviewJSON))
	require.NoError(t, err)

	h.ServeHTTP(rr, req)

	require.Equal(t, 200, rr.Code)
	assert.Equal(t, string(expectedResponseJSON)+"\n", rr.Body.String())
}

func TestHTTPSrv_mutateHandleFunc_empty(t *testing.T) {
	// prepare request data
	admissionReviewJSON := []byte{}

	s := HTTPSrv{}
	h := s.buildHandler()

	rr := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/mutate", bytes.NewReader(admissionReviewJSON))
	require.NoError(t, err)

	h.ServeHTTP(rr, req)

	require.Equal(t, 400, rr.Code)
	assert.Equal(t, "payload is empty\n", rr.Body.String())
}

func TestHTTPSrv_mutateHandleFunc_wrongJson(t *testing.T) {
	// prepare request data
	admissionReviewJSON := []byte("json string with error")

	c := gomock.NewController(t)

	service := mock.NewMockService(c)
	//DecodeAdmissionReview(data []byte) (*admission.AdmissionReview, error)
	service.EXPECT().DecodeAdmissionReview(admissionReviewJSON).DoAndReturn(func(data []byte) (*admission.AdmissionReview, error) {
		return nil, errors.New("json error")
	})

	s := HTTPSrv{
		service: service,
	}
	h := s.buildHandler()

	rr := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/mutate", bytes.NewReader(admissionReviewJSON))
	require.NoError(t, err)

	h.ServeHTTP(rr, req)

	require.Equal(t, 500, rr.Code)
	assert.Equal(t, "failed to decode admission review: json error\n", rr.Body.String())
}

func TestHTTPSrv_mutateHandleFunc_emptyRequest(t *testing.T) {
	// prepare request data
	admissionReview := admission.AdmissionReview{}
	admissionReviewJSON, err := json.Marshal(admissionReview)
	require.NoError(t, err)

	c := gomock.NewController(t)

	service := mock.NewMockService(c)
	//DecodeAdmissionReview(data []byte) (*admission.AdmissionReview, error)
	service.EXPECT().DecodeAdmissionReview(admissionReviewJSON).DoAndReturn(func(data []byte) (*admission.AdmissionReview, error) {
		return &admissionReview, nil
	})

	s := HTTPSrv{
		service: service,
	}
	h := s.buildHandler()

	rr := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/mutate", bytes.NewReader(admissionReviewJSON))
	require.NoError(t, err)

	h.ServeHTTP(rr, req)

	require.Equal(t, 400, rr.Code)
	assert.Equal(t, "admission request is empty\n", rr.Body.String())
}
