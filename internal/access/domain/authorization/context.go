package authorization

import (
	"errors"
	"net/netip"
	"strings"
	"time"
)

var ErrInvalidEvaluationContext = errors.New("invalid verified evaluation context")

// EvaluationContext is a snapshot resolved by trusted service middleware and
// the resource owner. Domain constructors do NOT authenticate any client input.
type EvaluationContext struct {
	clientIP netip.Addr
	hasIP bool
	mfaKnown bool
	mfaVerified bool
	now time.Time
	attributes map[string]string
	resourceRevision int64
}
type ContextInput struct {
	ClientIP netip.Addr
	MFAKnown bool
	MFAVerified bool
	EvaluatedAt time.Time
	ResourceAttributes map[string]string
	ResourceRevision int64
}
func NewEvaluationContext(input ContextInput) (EvaluationContext,error) {
	c:=EvaluationContext{clientIP:input.ClientIP,hasIP:input.ClientIP.IsValid(),mfaKnown:input.MFAKnown,
	mfaVerified:input.MFAVerified,now:input.EvaluatedAt.UTC(),resourceRevision:input.ResourceRevision,
	attributes:make(map[string]string,len(input.ResourceAttributes))}
	for k,v:=range input.ResourceAttributes {c.attributes[k]=v}
	return c,c.Validate()
}
func (c EvaluationContext) Validate() error {
	if c.now.IsZero()||c.mfaVerified&&!c.mfaKnown||len(c.attributes)>64{return ErrInvalidEvaluationContext}
	if len(c.attributes)>0&&c.resourceRevision<1{return ErrInvalidEvaluationContext}
	if c.resourceRevision<0{return ErrInvalidEvaluationContext}
	for k,v:=range c.attributes {
		if k==""||strings.TrimSpace(k)!=k||len(k)>128||len(v)>1024{return ErrInvalidEvaluationContext}
	}
	return nil
}
func (c EvaluationContext) ClientIP() (netip.Addr,bool){return c.clientIP,c.hasIP}
func (c EvaluationContext) MFA() (bool,bool){return c.mfaVerified,c.mfaKnown}
func (c EvaluationContext) EvaluatedAt() time.Time{return c.now}
func (c EvaluationContext) ResourceAttribute(name string)(string,bool){v,ok:=c.attributes[name];return v,ok}
func (c EvaluationContext) ResourceRevision() int64{return c.resourceRevision}
