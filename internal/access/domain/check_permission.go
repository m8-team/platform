package domain

import "strings"

type CheckPermissionRequest struct {
	tenantID          TenantID
	minZedToken       string
	subject           Subject
	permission        Permission
	resource          Resource
	modelRevision     ModelRevision
	failMode          FailMode
	critical          bool
	failOpenReference string
}

type CheckPermissionInput struct {
	TenantID          TenantID
	MinZedToken       string
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
		tenantID:          input.TenantID,
		minZedToken:       strings.TrimSpace(input.MinZedToken),
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
	if err := r.tenantID.Validate(); err != nil {
		return err
	}
	if len(r.minZedToken) > 4096 {
		return ErrInvalidZedToken
	}
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

// TenantID is the trusted tenant partition of this check.
func (r CheckPermissionRequest) TenantID() TenantID {
	return r.tenantID
}

// MinZedToken pins the minimum graph snapshot required by the caller.
func (r CheckPermissionRequest) MinZedToken() string {
	return r.minZedToken
}
