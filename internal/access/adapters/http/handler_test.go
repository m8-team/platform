package http

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	pb "github.com/m8-team/platform/internal/proto/m8/platform/access/v1"
	"google.golang.org/protobuf/proto"
)

type testServer struct {
	pb.UnimplementedAccessServiceServer
}

func (testServer) EvaluateAccess(context.Context, *pb.EvaluateAccessRequest) (*pb.EvaluateAccessResponse, error) {
	return &pb.EvaluateAccessResponse{}, nil
}

func (testServer) EvaluateAccessBatch(_ context.Context, req *pb.EvaluateAccessBatchRequest) (*pb.EvaluateAccessBatchResponse, error) {
	if len(req.Evaluations) == 0 {
		panic("empty batch must use EvaluateAccess")
	}
	return &pb.EvaluateAccessBatchResponse{Evaluations: []*pb.EvaluateAccessResponse{{}}}, nil
}

func (testServer) SearchActions(context.Context, *pb.SearchActionsRequest) (*pb.SearchActionsResponse, error) {
	return &pb.SearchActionsResponse{Page: &pb.PageResponse{}}, nil
}

func TestAuthZENResponses(t *testing.T) {
	handler, err := NewHandler(t.Context(), testServer{})
	if err != nil {
		t.Fatal(err)
	}
	const defaults = `"subject":{"type":"user","id":"u"},"action":{"name":"read"},"resource":{"type":"file","id":"f"}`
	for _, tc := range []struct{ path, body, want string }{
		{"evaluation", "{" + defaults + "}", `{"decision":false}`},
		{"evaluations", "{" + defaults + "}", `{"decision":false}`},
		{"evaluations", "{" + defaults + `,"evaluations":[]}`, `{"decision":false}`},
		{"evaluations", "{" + defaults + `,"evaluations":[{}]}`, `{"evaluations":[{"decision":false}]}`},
		{"search/action", `{"subject":{"type":"user","id":"u"},"resource":{"type":"file","id":"f"}}`, `{"results":[],"page":{"next_token":""}}`},
	} {
		req := httptest.NewRequest("POST", "/access/v1/"+tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		var got, want any
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &got) != nil || json.Unmarshal([]byte(tc.want), &want) != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %d %s, want %s", tc.path, w.Code, w.Body, tc.want)
		}
	}
}

func TestAuthZENValidation(t *testing.T) {
	subject := &pb.Subject{Type: "user", Id: "u"}
	action := &pb.Action{Name: "read"}
	resource := &pb.Resource{Type: "file", Id: "f"}
	for _, tc := range []struct {
		msg   proto.Message
		valid bool
	}{
		{&pb.EvaluateAccessBatchRequest{}, false},
		{&pb.EvaluateAccessBatchRequest{Evaluations: []*pb.Evaluation{{}}}, false},
		{&pb.EvaluateAccessBatchRequest{Subject: subject, Action: action, Resource: resource}, true},
		{&pb.EvaluateAccessBatchRequest{Subject: subject, Action: action, Evaluations: []*pb.Evaluation{{Resource: resource}}}, true},
		{&pb.EvaluateAccessBatchRequest{Subject: subject, Action: action, Evaluations: []*pb.Evaluation{{Resource: resource}, {}}}, false},
		{&pb.SubjectSelector{Type: "user", Id: strings.Repeat("x", 257)}, true},
		{&pb.ResourceSelector{Type: "file", Id: strings.Repeat("x", 257)}, true},
	} {
		var err error
		if batch, ok := tc.msg.(*pb.EvaluateAccessBatchRequest); ok {
			err = validateBatchRequest(batch)
		} else {
			err = protovalidate.Validate(tc.msg)
		}
		if (err == nil) != tc.valid {
			t.Errorf("%T: valid=%t, got %v", tc.msg, tc.valid, err)
		}
	}
	for identifier, valid := range map[string]bool{"https://example.com/tenant": true, "http://example.com": false, "https://example.com?q=x": false, "https://example.com/#f": false} {
		msg := &pb.ConfigurationResponse{PolicyDecisionPoint: identifier, AccessEvaluationEndpoint: "https://example.com/access/v1/evaluation"}
		if err := protovalidate.Validate(msg); (err == nil) != valid {
			t.Errorf("identifier %q: valid=%t, got %v", identifier, valid, err)
		}
	}
}
