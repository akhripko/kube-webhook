package httpsrv

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_writeJSON(t *testing.T) {
	expected := "{\"abc\":123}\n"
	data := struct {
		Value int `json:"abc"`
	}{
		Value: 123,
	}
	rr := httptest.NewRecorder()
	writeJSON(rr, data)
	assert.Equal(t, expected, rr.Body.String())
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}
