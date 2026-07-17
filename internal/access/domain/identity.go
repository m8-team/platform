package domain

import "strings"

type Subject struct {
	subjectType string
	id          string
}

func NewSubject(subjectType string, id string) (Subject, error) {
	subject := Subject{
		subjectType: strings.TrimSpace(subjectType),
		id:          strings.TrimSpace(id),
	}
	if err := subject.Validate(); err != nil {
		return Subject{}, err
	}

	return subject, nil
}

func (s Subject) Validate() error {
	if s.subjectType == "" {
		return ErrSubjectTypeRequired
	}
	if s.id == "" {
		return ErrSubjectIDRequired
	}

	return nil
}

func (s Subject) Type() string {
	return s.subjectType
}

func (s Subject) ID() string {
	return s.id
}

type Resource struct {
	resourceType string
	id           string
}

func NewResource(resourceType string, id string) (Resource, error) {
	resource := Resource{
		resourceType: strings.TrimSpace(resourceType),
		id:           strings.TrimSpace(id),
	}
	if err := resource.Validate(); err != nil {
		return Resource{}, err
	}

	return resource, nil
}

func (r Resource) Validate() error {
	if r.resourceType == "" {
		return ErrResourceTypeRequired
	}
	if r.id == "" {
		return ErrResourceIDRequired
	}

	return nil
}

func (r Resource) Type() string {
	return r.resourceType
}

func (r Resource) ID() string {
	return r.id
}
