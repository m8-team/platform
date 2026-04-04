package grpc

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestRewriteFormRequestMultipartToJSON(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("request_id", "2eddfsdff"); err != nil {
		t.Fatalf("write request_id: %v", err)
	}
	if err := writer.WriteField("reason", "242sdf"); err != nil {
		t.Fatalf("write reason: %v", err)
	}
	if err := writer.WriteField("performed_by", "dfsdfsf"); err != nil {
		t.Fatalf("write performed_by: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "http://example.test", &body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rewritten, err := rewriteFormRequest(req)
	if err != nil {
		t.Fatalf("rewrite form request: %v", err)
	}

	assertJSONBody(t, rewritten, map[string]any{
		"request_id":   "2eddfsdff",
		"reason":       "242sdf",
		"performed_by": "dfsdfsf",
	})
}

func TestRewriteFormRequestURLEncodedToJSON(t *testing.T) {
	form := url.Values{}
	form.Set("request_id", "2eddfsdff")
	form.Set("reason", "242sdf")
	form.Set("performed_by", "dfsdfsf")

	req, err := http.NewRequest(http.MethodPost, "http://example.test", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rewritten, err := rewriteFormRequest(req)
	if err != nil {
		t.Fatalf("rewrite form request: %v", err)
	}

	assertJSONBody(t, rewritten, map[string]any{
		"request_id":   "2eddfsdff",
		"reason":       "242sdf",
		"performed_by": "dfsdfsf",
	})
}

func TestRewriteFormRequestRejectsMultipartFiles(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "payload.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("payload")); err != nil {
		t.Fatalf("write file payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "http://example.test", &body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, err = rewriteFormRequest(req)
	if err == nil {
		t.Fatal("expected multipart file upload to be rejected")
	}
	if !strings.Contains(err.Error(), "multipart file uploads are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertJSONBody(t *testing.T, req *http.Request, want map[string]any) {
	t.Helper()

	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("content-type = %q, want application/json", req.Header.Get("Content-Type"))
	}

	var got map[string]any
	if err := json.NewDecoder(req.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("body field count = %d, want %d", len(got), len(want))
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("body[%q] = %v, want %v", key, got[key], wantValue)
		}
	}
}
