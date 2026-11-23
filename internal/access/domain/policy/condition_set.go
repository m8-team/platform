package policy

import "github.com/m8-team/platform/internal/access/domain/authorization"

// ConditionSet requires ALL of its 'all' conditions and at least one of its
// 'any' conditions (when present). Missing trusted input fails closed.
type ConditionSet struct{
	all []Condition
	any []Condition
}
func NewConditionSet(all,any []Condition)(ConditionSet,error){
	s:=ConditionSet{all:append([]Condition(nil),all...),any:append([]Condition(nil),any...)}
	return s,s.Validate()
}
func (s ConditionSet) Validate()error{
	if len(s.all)+len(s.any)==0||len(s.all)+len(s.any)>64{return ErrInvalidCondition}
	for _,c:=range s.all{if err:=c.Validate();err!=nil{return err}}
	for _,c:=range s.any{if err:=c.Validate();err!=nil{return err}}
	return nil
}
func (s ConditionSet) All()[]Condition{return append([]Condition(nil),s.all...)}
func (s ConditionSet) Any()[]Condition{return append([]Condition(nil),s.any...)}
func (s ConditionSet) Evaluate(ctx authorization.EvaluationContext)(bool,error){
	if err:=s.Validate();err!=nil{return false,err}
	allOK:=true
	for _,c:=range s.all{
		ok,err:=c.Evaluate(ctx);if err!=nil{return false,err}
		if !ok{allOK=false}
	}
	anyOK:=len(s.any)==0
	for _,c:=range s.any{
		ok,err:=c.Evaluate(ctx);if err!=nil{return false,err}
		if ok{anyOK=true}
	}
	return allOK&&anyOK,nil
}
