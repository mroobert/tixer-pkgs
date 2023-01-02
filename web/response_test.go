package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mroobert/tixer-pkgs/web"
)

func TestWriteJson_SetsCorrectlyTheBodyWithStatusAndHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		status      int
		data        web.Envelope
		headers     http.Header
		wantStatus  int
		wantData    []byte
		wantHeaders http.Header
	}{
		{
			name:   "Valid JSON with additional headers",
			status: http.StatusOK,
			data:   web.Envelope{"message": "valid JSON with headers"},
			headers: http.Header{
				"Header1": []string{"value1"},
			},
			wantStatus: http.StatusOK,
			wantData:   []byte(`{"message":"valid JSON with headers"}`),
			wantHeaders: http.Header{
				"Header1":      []string{"value1"},
				"Content-Type": []string{"application/json"},
			},
		},
		{
			name:       "Valid JSON with no additional headers",
			status:     http.StatusOK,
			data:       web.Envelope{"message": "valid JSON with no headers"},
			headers:    nil,
			wantStatus: http.StatusOK,
			wantData:   []byte(`{"message":"valid JSON with no headers"}`),
			wantHeaders: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
		{
			name:   "Valid JSON with no data and additional headers",
			status: http.StatusNotFound,
			data:   web.Envelope{},
			headers: http.Header{
				"Header1": []string{"value1"},
			},
			wantStatus: http.StatusNotFound,
			wantData:   []byte(`{}`),
			wantHeaders: http.Header{
				"Header1":      []string{"value1"},
				"Content-Type": []string{"application/json"},
			},
		},
		{
			name:       "Valid JSON with no data and no headers",
			status:     http.StatusNotFound,
			data:       web.Envelope{},
			headers:    nil,
			wantStatus: http.StatusNotFound,
			wantData:   []byte(`{}`),
			wantHeaders: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			web.WriteJSON(recorder, tt.status, tt.data, tt.headers)

			gotStatus := recorder.Result().StatusCode
			if gotStatus != tt.wantStatus {
				t.Errorf("Got status code %d, want %d", gotStatus, tt.wantStatus)
			}

			gotData := recorder.Body.Bytes()
			if !cmp.Equal(gotData, tt.wantData) {
				t.Errorf("Differences in the data: %s", cmp.Diff(tt.wantData, gotData))
			}

			gotHeaders := recorder.Header()
			if !cmp.Equal(gotHeaders, tt.wantHeaders) {
				t.Errorf("Differences in the headers: %s", cmp.Diff(tt.wantHeaders, gotHeaders))
			}

		})
	}
}

func TestWriteJson_ReturnsAMarshallingErrForInvalidData(t *testing.T) {
	t.Parallel()

	data := web.Envelope{"key": make(chan int)}

	recorder := httptest.NewRecorder()
	err := web.WriteJSON(recorder, http.StatusOK, data, nil)

	if err == nil {
		t.Error("Expected a marshalling error, but got nil")
	}
}
