// Package policy holds mandatory authorization restrictions. No policy creates
// entitlement on its own; a valid role assignment or grant is also required.
package policy

import (
	"errors"
	"net/netip"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/access/domain/authorization"
)
var (
	ErrInvalidCondition = errors.New("invalid access policy condition")
	ErrMissingTrustedContext = errors.New("missing verified policy input")
)
type ConditionKind string
const (
	ConditionIPInCIDR ConditionKind = "network.ip_in_cidr"
	ConditionMFA ConditionKind = "authentication.mfa"
	ConditionBefore ConditionKind = "time.before"
	ConditionResourceEquals ConditionKind = "resource.attribute_equals"
)
// Condition uses trusted, typed, immutable operands. It never evaluates arbitrary
// CEL or arbitrary caller-supplied programs. Support for CEL requires a separate
// restricted compiler and trusted evaluation context.
type Condition struct{
	kind ConditionKind
	prefixes []netip.Prefix
	before time.Time
	attribute string
	value string
}
func NewIPInCIDR(cidrs []string)(Condition,error){
	if len(cidrs)==0||len(cidrs)>64{return Condition{},ErrInvalidCondition}
	c:=Condition{kind:ConditionIPInCIDR}
	seen:=map[string]bool{}
	for _,cidr:=range cidrs {
		p,err:=netip.ParsePrefix(cidr)
		if err!=nil||p.Addr().Is4In6(){return Condition{},ErrInvalidCondition}
		p=p.Masked()
		if seen[p.String()]{return Condition{},ErrInvalidCondition}
		seen[p.String()]=true;c.prefixes=append(c.prefixes,p)
	}
	return c,nil
}
func RequireMFA()Condition{return Condition{kind:ConditionMFA}}
func Before(at time.Time)(Condition,error){
	if at.IsZero(){return Condition{},ErrInvalidCondition}
	return Condition{kind:ConditionBefore,before:at.UTC()},nil
}
func ResourceAttributeEquals(name,value string)(Condition,error){
	c:=Condition{kind:ConditionResourceEquals,attribute:strings.TrimSpace(name),value:value}
	return c,c.Validate()
}
func (c Condition) Validate()error{
	switch c.kind {
	case ConditionIPInCIDR:
		if len(c.prefixes)==0||len(c.prefixes)>64{return ErrInvalidCondition}
		for _,p:=range c.prefixes{if !p.IsValid(){return ErrInvalidCondition}}
	case ConditionMFA:
		// No caller-configurable bypass or optional MFA flags.
	case ConditionBefore:
		if c.before.IsZero(){return ErrInvalidCondition}
	case ConditionResourceEquals:
		if c.attribute==""||len(c.attribute)>128||len(c.value)>1024{return ErrInvalidCondition}
	default:
		return ErrInvalidCondition
	}
	return nil
}
func (c Condition) Kind()ConditionKind{return c.kind}
func (c Condition) Evaluate(ctx authorization.EvaluationContext)(bool,error){
	if err:=c.Validate();err!=nil{return false,err}
	if err:=ctx.Validate();err!=nil{return false,err}
	switch c.kind {
	case ConditionIPInCIDR:
		ip,ok:=ctx.ClientIP();if !ok{return false,ErrMissingTrustedContext}
		ip=ip.Unmap()
		for _,p:=range c.prefixes{if p.Contains(ip){return true,nil}}
		return false,nil
	case ConditionMFA:
		mfa,known:=ctx.MFA();if !known{return false,ErrMissingTrustedContext}
		return mfa,nil
	case ConditionBefore:
		return ctx.EvaluatedAt().Before(c.before),nil
	case ConditionResourceEquals:
		if ctx.ResourceRevision()<1{return false,ErrMissingTrustedContext}
		value,ok:=ctx.ResourceAttribute(c.attribute)
		if !ok{return false,ErrMissingTrustedContext}
		return value==c.value,nil
	}
	return false,ErrInvalidCondition
}
