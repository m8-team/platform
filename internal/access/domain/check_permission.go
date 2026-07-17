package domain

import "strings"

type CheckPermissionRequest struct {
	subject           Subject
	permission        Permission
	resource          Resource
	modelRevision     ModelRevision
	failMode          FailMode
	critical          bool
	failOpenReference string
}

type CheckPermissionInput struct {
	Subject           Subject
	Permission        Permission
	Resource          Resource
	ModelRevision     ModelRevision
	FailMode          FailMode
	Critical          bool
	FailOpenReference string
}

func NewCheckPermissionRequest(input CheckPermissionInput) (CheckPermissionRequest, error) {
	request := CheckPermissionRequest{
		subject:           input.Subject,
		permission:        input.Permission,
		resource:          input.Resource,
		modelRevision:     input.ModelRevision,
		failMode:          input.FailMode.WithDefault(),
		critical:          input.Critical,
		failOpenReference: strings.TrimSpace(input.FailOpenReference),
	}
	if err := request.Validate(); err != nil {
		return CheckPermissionRequest{}, err
	}

	return request, nil
}

func (r CheckPermissionRequest) Validate() error {
	if err := r.subject.Validate(); err != nil {
		return err
	}
	if err := r.permission.Validate(); err != nil {
		return err
	}
	if err := r.resource.Validate(); err != nil {
		return err
	}
	if err := r.modelRevision.Validate(); err != nil {
		return err
	}
	if !r.failMode.IsValid() {
		return ErrInvalidFailMode
	}
	if r.failMode == FailModeAllow {
		if r.critical {
			return ErrFailOpenForCriticalCheck
		}
		if r.failOpenReference == "" {
			return ErrFailOpenReferenceRequired
		}
	}

	return nil
}

func (r CheckPermissionRequest) Subject() Subject {
	return r.subject
}

func (r CheckPermissionRequest) Permission() Permission {
	return r.permission
}

func (r CheckPermissionRequest) Resource() Resource {
	return r.resource
}

func (r CheckPermissionRequest) ModelRevision() ModelRevision {
	return r.modelRevision
}

func (r CheckPermissionRequest) FailMode() FailMode {
	return r.failMode
}

func (r CheckPermissionRequest) Critical() bool {
	return r.critical
}

func (r CheckPermissionRequest) FailOpenReference() string {
	return r.failOpenReference
}
