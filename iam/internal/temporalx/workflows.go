package temporalx

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	CreateServiceAccountWorkflowName    = "iam.CreateServiceAccountWorkflow"
	RotateClientSecretWorkflowName      = "iam.RotateClientSecretWorkflow"
	GrantSupportAccessWorkflowName      = "iam.GrantTemporarySupportAccessWorkflow"
	SyncRelationshipsWorkflowName       = "iam.SyncRelationshipsToSpiceDBWorkflow"
	RebuildAccessReadModelsWorkflowName = "iam.RebuildAccessReadModelsWorkflow"
	ImportFederatedUserWorkflowName     = "iam.ImportFederatedUserWorkflow"
)

type CreateServiceAccountInput struct {
	ServiceAccountID string
	TenantID         string
	DisplayName      string
	Description      string
	RequestedBy      string
}

type RotateClientSecretInput struct {
	OAuthClientID string
	RequestedBy   string
	Reason        string
}

type GrantTemporarySupportAccessInput struct {
	SupportGrantID string
	TenantID       string
	RequestedBy    string
	Reason         string
	TTL            time.Duration
}

type SyncRelationshipsInput struct {
	BatchSize int
}

type RebuildAccessReadModelsInput struct {
	Projection string
}

type ImportFederatedUserInput struct {
	TenantID string
	Provider string
	Subject  string
}

func CreateServiceAccountWorkflow(ctx workflow.Context, input CreateServiceAccountInput) (string, error) {
	opts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, opts)
	if err := workflow.ExecuteActivity(ctx, "CreateServiceAccountMetadata", input).Get(ctx, nil); err != nil {
		return "", err
	}
	if err := workflow.ExecuteActivity(ctx, "CreateKeycloakServiceAccount", input).Get(ctx, nil); err != nil {
		return "", err
	}
	if err := workflow.ExecuteActivity(ctx, "SyncServiceAccountBindings", input).Get(ctx, nil); err != nil {
		return "", err
	}
	if err := workflow.ExecuteActivity(ctx, "WriteServiceAccountAudit", input).Get(ctx, nil); err != nil {
		return "", err
	}
	return input.ServiceAccountID, nil
}

func RotateClientSecretWorkflow(ctx workflow.Context, input RotateClientSecretInput) (string, error) {
	opts := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Second}
	ctx = workflow.WithActivityOptions(ctx, opts)
	var secretRef string
	if err := workflow.ExecuteActivity(ctx, "RotateKeycloakClientSecret", input).Get(ctx, &secretRef); err != nil {
		return "", err
	}
	if err := workflow.ExecuteActivity(ctx, "WriteSecretRotationAudit", input).Get(ctx, nil); err != nil {
		return "", err
	}
	return secretRef, nil
}

func GrantTemporarySupportAccessWorkflow(ctx workflow.Context, input GrantTemporarySupportAccessInput) error {
	opts := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Second}
	ctx = workflow.WithActivityOptions(ctx, opts)
	if err := workflow.ExecuteActivity(ctx, "ActivateSupportGrant", input).Get(ctx, nil); err != nil {
		return err
	}
	if err := workflow.Sleep(ctx, input.TTL); err != nil {
		return err
	}
	return workflow.ExecuteActivity(ctx, "ExpireSupportGrant", input).Get(ctx, nil)
}

func SyncRelationshipsToSpiceDBWorkflow(ctx workflow.Context, input SyncRelationshipsInput) error {
	opts := workflow.ActivityOptions{StartToCloseTimeout: time.Minute}
	ctx = workflow.WithActivityOptions(ctx, opts)
	return workflow.ExecuteActivity(ctx, "SyncRelationshipBatch", input).Get(ctx, nil)
}

func RebuildAccessReadModelsWorkflow(ctx workflow.Context, input RebuildAccessReadModelsInput) error {
	opts := workflow.ActivityOptions{StartToCloseTimeout: 5 * time.Minute}
	ctx = workflow.WithActivityOptions(ctx, opts)
	return workflow.ExecuteActivity(ctx, "RebuildProjection", input).Get(ctx, nil)
}

func ImportFederatedUserWorkflow(ctx workflow.Context, input ImportFederatedUserInput) error {
	opts := workflow.ActivityOptions{StartToCloseTimeout: time.Minute}
	ctx = workflow.WithActivityOptions(ctx, opts)
	return workflow.ExecuteActivity(ctx, "ImportFederatedUser", input).Get(ctx, nil)
}
