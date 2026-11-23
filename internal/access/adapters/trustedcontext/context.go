// Package trustedcontext translates verified gateway/IdP and resource owner
// evidence into the typed Access policy evaluation context.
package trustedcontext

import (
	"context"
	"errors"
	"math"
	"net/netip"
	"time"

	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/domain"
	"github.com/m8-team/platform/internal/access/domain/authorization"
)

var ErrUnverifiedEvidence = errors.New("trusted authorization context cannot be verified")

// PeerAddress MUST come from an authenticated gateway peer/context established
// after proxy trust checks, never X-Forwarded-For or an arbitrary request field.
type PeerAddress interface {
	VerifiedClientIP(context.Context) (netip.Addr, error)
}
// SessionProof MUST check actual IdP assurance/session evidence. Caller-supplied
// MFA flags or unverified JWT custom properties are NOT acceptable.
type SessionProof interface {
	VerifiedMFA(context.Context, domain.TenantID, domain.Subject) (verified bool, known bool, err error)
}
// ResourceProof MUST obtain owner-authoritative attributes and their revision.
// No direct dependency on Resource Manager or a particular business module.
type ResourceProof interface {
	VerifiedResourceAttributes(context.Context, domain.TenantID, string, string) (uint64, map[string]string, error)
}

type Source struct {
	network PeerAddress
	session SessionProof
	resource ResourceProof
	now func() time.Time
}
var _ consistency.TrustedContextSource = (*Source)(nil)

func New(network PeerAddress, session SessionProof, resource ResourceProof, now func()time.Time)(*Source,error){
	if network==nil||session==nil||resource==nil||now==nil{return nil,ErrUnverifiedEvidence}
	return &Source{network:network,session:session,resource:resource,now:now},nil
}
func(s *Source) ResolveTrustedContext(ctx context.Context, tenant domain.TenantID, subject domain.Subject, kind,id string)(authorization.EvaluationContext,error){
	if s==nil||s.network==nil||s.session==nil||s.resource==nil||s.now==nil||
		tenant.Validate()!=nil||subject.Validate()!=nil||kind==""||id=="" {
		return authorization.EvaluationContext{},ErrUnverifiedEvidence
	}
	if err:=ctx.Err();err!=nil{return authorization.EvaluationContext{},err}
	ip,err:=s.network.VerifiedClientIP(ctx)
	if err!=nil||!ip.IsValid(){return authorization.EvaluationContext{},ErrUnverifiedEvidence}
	mfa,known,err:=s.session.VerifiedMFA(ctx,tenant,subject)
	if err!=nil||mfa&&!known{return authorization.EvaluationContext{},ErrUnverifiedEvidence}
	revision,attrs,err:=s.resource.VerifiedResourceAttributes(ctx,tenant,kind,id)
	if err!=nil||revision==0||revision>math.MaxInt64{return authorization.EvaluationContext{},ErrUnverifiedEvidence}
	input:=authorization.ContextInput{
		ClientIP:ip.Unmap(),MFAKnown:known,MFAVerified:mfa,
		EvaluatedAt:s.now().UTC(),ResourceRevision:int64(revision),
		ResourceAttributes:attrs,
	}
	c,err:=authorization.NewEvaluationContext(input)
	if err!=nil{return authorization.EvaluationContext{},ErrUnverifiedEvidence}
	return c,nil
}
