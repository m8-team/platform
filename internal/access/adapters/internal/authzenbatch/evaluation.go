// Package authzenbatch resolves the field defaults shared by AuthZEN transports.
package authzenbatch

import (
	"cmp"

	accesspb "github.com/m8-team/platform/internal/proto/m8/platform/access/v1"
)

// Evaluation applies whole-field overrides. A nil item selects the batch defaults.
func Evaluation(batch *accesspb.EvaluateAccessBatchRequest, item *accesspb.Evaluation) *accesspb.EvaluateAccessRequest {
	return &accesspb.EvaluateAccessRequest{
		Subject:  cmp.Or(item.GetSubject(), batch.Subject),
		Action:   cmp.Or(item.GetAction(), batch.Action),
		Resource: cmp.Or(item.GetResource(), batch.Resource),
		Context:  cmp.Or(item.GetContext(), batch.Context),
	}
}
