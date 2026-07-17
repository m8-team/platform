---
title: "M8 Platform API Contract Catalog"
description: "Каталог API contract candidates."
keywords:
  - "M8 Platform"
  - "contracts"
---

# M8 Platform API Contract Catalog {#api-contract-catalog}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [Contracts](index.md) | [Requirements Catalog](../../architecture/requirements/index.md) | `api-catalog.yaml`

{% endnote %}

_M8-API-000 · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `M8-API-000` |
| Версия | `0.1` |
| Статус | Базовая проектная редакция |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | Protobuf/ConnectRPC API всех ограниченных контекстов |

> Каталог назначает устойчивые идентификаторы API-контрактам и связывает их с требованиями. Запись `proposed` не заменяет `.proto`: перед реализацией должны быть утверждены поля, validation rules, permissions, error mapping и compatibility baseline.

# 1. Нормативные правила

- Публичные контракты MUST описываться Protobuf и проходить `buf lint`/`buf breaking`.
- Mutation MUST принимать idempotency key; update MUST поддерживать revision/ETag и FieldMask, где применимо.
- Длительная команда MUST возвращать `google.longrunning.Operation`.
- RequestContext, resource scope и actor MUST передаваться по правилам PADS.
- Внутренний RPC не становится публичным без отдельного review и permission.

# 2. Пакеты и сервисы

| Контекст | Package | Service |
| --- | --- | --- |
| Resource Manager | m8.resource_manager.v1 | ResourceManagerService |
| Identity | m8.identity.v1 | IdentityService |
| Authentication | m8.authentication.v1 | AuthenticationService |
| Access | m8.access.v1 | AccessService |
| Risk Decision | m8.risk_decision.v1 | RiskDecisionService |
| Provisioning | m8.provisioning.v1 | ProvisioningService |
| Audit | m8.audit.v1 | AuditService |
| Common Operation | m8.operations.v1 | OperationsService |

# 3. Реестр API

## 3.1. Resource Manager

| ID | RPC method | Видимость | Тип | Ответ | Требования |
| --- | --- | --- | --- | --- | --- |
| `API-RM-001` | `CreateOrganization` | public | command | `CreateOrganizationResponse` | `RM-FR-001` |
| `API-RM-002` | `UpdateOrganization` | public | command | `UpdateOrganizationResponse` | `RM-FR-002` |
| `API-RM-003` | `SuspendOrganization` | public | command | `SuspendOrganizationResponse` | `RM-FR-003` |
| `API-RM-004` | `RestoreOrganization` | public | command | `RestoreOrganizationResponse` | `RM-FR-004` |
| `API-RM-005` | `ArchiveOrganization` | public | command | `ArchiveOrganizationResponse` | `RM-FR-005` |
| `API-RM-010` | `CreateWorkspace` | public | command | `CreateWorkspaceResponse` | `RM-FR-010` |
| `API-RM-011` | `UpdateWorkspace` | public | command | `UpdateWorkspaceResponse` | `RM-FR-011` |
| `API-RM-012` | `SetWorkspaceSuspension` | public | command | `SetWorkspaceSuspensionResponse` | `RM-FR-012` |
| `API-RM-020` | `CreateProject` | public | command | `CreateProjectResponse` | `RM-FR-020` |
| `API-RM-021` | `MoveProject` | public | command | `google.longrunning.Operation` | `RM-FR-021` |
| `API-RM-022` | `DeleteProject` | public | command | `google.longrunning.Operation` | `RM-FR-022` |
| `API-RM-023` | `SuspendProject` | public | command | `SuspendProjectResponse` | `RM-FR-023` |
| `API-RM-024` | `RestoreProject` | public | command | `RestoreProjectResponse` | `RM-FR-024` |
| `API-RM-030` | `RegisterService` | public | command | `RegisterServiceResponse` | `RM-FR-030` |
| `API-RM-031` | `UpdateServiceRegistration` | public | command | `UpdateServiceRegistrationResponse` | `RM-FR-031` |
| `API-RM-032` | `UnregisterService` | public | command | `UnregisterServiceResponse` | `RM-FR-032` |
| `API-RM-040` | `GetResource` | public | query | `GetResourceResponse` | `RM-FR-040` |
| `API-RM-041` | `ListChildResources` | public | query | `ListChildResourcesResponse` | `RM-FR-041` |
| `API-RM-042` | `GetResourcePath` | public | query | `GetResourcePathResponse` | `RM-FR-042` |
| `API-RM-043` | `SearchResources` | public | query | `SearchResourcesResponse` | `RM-FR-043` |
| `API-RM-044` | `UpdateResourceLabels` | public | command | `UpdateResourceLabelsResponse` | `RM-FR-044` |

## 3.2. Identity

| ID | RPC method | Видимость | Тип | Ответ | Требования |
| --- | --- | --- | --- | --- | --- |
| `API-ID-001` | `CreateUserPool` | public | command | `CreateUserPoolResponse` | `ID-FR-001` |
| `API-ID-002` | `UpdateUserPool` | public | command | `UpdateUserPoolResponse` | `ID-FR-002` |
| `API-ID-003` | `SuspendUserPool` | public | command | `SuspendUserPoolResponse` | `ID-FR-003` |
| `API-ID-010` | `CreateUser` | public | command | `CreateUserResponse` | `ID-FR-010` |
| `API-ID-011` | `UpdateUserProfile` | public | command | `UpdateUserProfileResponse` | `ID-FR-011` |
| `API-ID-012` | `DisableUser` | public | command | `DisableUserResponse` | `ID-FR-012` |
| `API-ID-013` | `RestoreUser` | public | command | `RestoreUserResponse` | `ID-FR-013` |
| `API-ID-014` | `SecurityBlockUser` | public | command | `SecurityBlockUserResponse` | `ID-FR-014` |
| `API-ID-015` | `GetUser` | public | query | `GetUserResponse` | `ID-FR-015` |
| `API-ID-016` | `SearchUsers` | public | query | `SearchUsersResponse` | `ID-FR-016` |
| `API-ID-020` | `LinkExternalIdentity` | public | command | `LinkExternalIdentityResponse` | `ID-FR-020` |
| `API-ID-021` | `DetectExternalIdentityConflict` | public | command | `DetectExternalIdentityConflictResponse` | `ID-FR-021` |
| `API-ID-022` | `UnlinkExternalIdentity` | public | command | `UnlinkExternalIdentityResponse` | `ID-FR-022` |
| `API-ID-023` | `ResolveSubjectByExternalIdentity` | public | query | `ResolveSubjectByExternalIdentityResponse` | `ID-FR-023` |
| `API-ID-024` | `ResolveSubjectByContact` | public | query | `ResolveSubjectByContactResponse` | `ID-FR-024` |
| `API-ID-030` | `CreateGroup` | public | command | `CreateGroupResponse` | `ID-FR-030` |
| `API-ID-031` | `UpdateGroup` | public | command | `UpdateGroupResponse` | `ID-FR-031` |
| `API-ID-032` | `AddMembership` | public | command | `AddMembershipResponse` | `ID-FR-032` |
| `API-ID-033` | `RemoveMembership` | public | command | `RemoveMembershipResponse` | `ID-FR-033` |
| `API-ID-034` | `ListMemberships` | public | query | `ListMembershipsResponse` | `ID-FR-034` |
| `API-ID-040` | `MergeUsers` | public | command | `google.longrunning.Operation` | `ID-FR-040` |
| `API-ID-041` | `DetectPotentialDuplicates` | public | command | `DetectPotentialDuplicatesResponse` | `ID-FR-041` |
| `API-ID-050` | `DeleteOrAnonymizeUser` | public | command | `google.longrunning.Operation` | `ID-FR-050` |

## 3.3. Authentication

| ID | RPC method | Видимость | Тип | Ответ | Требования |
| --- | --- | --- | --- | --- | --- |
| `API-AUTH-001` | `StartAuthentication` | public | command | `StartAuthenticationResponse` | `AUTH-FR-001`, `AUTH-FR-017` |
| `API-AUTH-002` | `SelectAuthenticationProvider` | internal | command | `SelectAuthenticationProviderResponse` | `AUTH-FR-002` |
| `API-AUTH-003` | `CreateAuthenticationChallenge` | internal | command | `CreateAuthenticationChallengeResponse` | `AUTH-FR-003` |
| `API-AUTH-004` | `GetAuthentication` | public | query | `GetAuthenticationResponse` | `AUTH-FR-004` |
| `API-AUTH-005` | `WaitAuthentication` | public | command | `WaitAuthenticationResponse` | `AUTH-FR-005` |
| `API-AUTH-006` | `CancelAuthentication` | public | command | `CancelAuthenticationResponse` | `AUTH-FR-006` |
| `API-AUTH-007` | `HandleProviderCallback` | internal | command | `HandleProviderCallbackResponse` | `AUTH-FR-007` |
| `API-AUTH-008` | `ExpireAuthentication` | admin | command | `ExpireAuthenticationResponse` | `AUTH-FR-008` |
| `API-AUTH-009` | `FailAuthentication` | admin | command | `FailAuthenticationResponse` | `AUTH-FR-009` |
| `API-AUTH-010` | `CreateAuthenticationHandoff` | public | command | `CreateAuthenticationHandoffResponse` | `AUTH-FR-010` |
| `API-AUTH-011` | `RedeemAuthenticationHandoff` | public | command | `RedeemAuthenticationHandoffResponse` | `AUTH-FR-011` |
| `API-AUTH-012` | `CreateAuthenticationSessionReference` | internal | command | `CreateAuthenticationSessionReferenceResponse` | `AUTH-FR-012` |
| `API-AUTH-013` | `RevokeAuthenticationSession` | public | command | `google.longrunning.Operation` | `AUTH-FR-013` |
| `API-AUTH-014` | `ValidateClient` | public | query | `ValidateClientResponse` | `AUTH-FR-014` |
| `API-AUTH-015` | `RegisterClient` | public | command | `RegisterClientResponse` | `AUTH-FR-015` |
| `API-AUTH-016` | `UpdateClientPolicy` | public | command | `UpdateClientPolicyResponse` | `AUTH-FR-016` |
| `API-AUTH-018` | `StartCibaAuthentication` | public | command | `google.longrunning.Operation` | `AUTH-FR-018` |
| `API-AUTH-019` | `HandleCibaDecision` | public | command | `HandleCibaDecisionResponse` | `AUTH-FR-019` |
| `API-AUTH-020` | `StartStepUp` | public | command | `google.longrunning.Operation` | `AUTH-FR-020` |
| `API-AUTH-021` | `SelectStepUpChallenge` | public | command | `SelectStepUpChallengeResponse` | `AUTH-FR-021` |
| `API-AUTH-022` | `SendOtpChallenge` | public | command | `SendOtpChallengeResponse` | `AUTH-FR-022` |
| `API-AUTH-023` | `VerifyOtp` | public | query | `VerifyOtpResponse` | `AUTH-FR-023` |
| `API-AUTH-024` | `ResendAuthenticationChallenge` | public | command | `ResendAuthenticationChallengeResponse` | `AUTH-FR-024` |
| `API-AUTH-025` | `VerifyWebAuthnAssertion` | public | query | `VerifyWebAuthnAssertionResponse` | `AUTH-FR-025` |
| `API-AUTH-026` | `RegisterWebAuthnCredential` | public | command | `google.longrunning.Operation` | `AUTH-FR-026` |
| `API-AUTH-027` | `StartFederatedAuthentication` | public | command | `google.longrunning.Operation` | `AUTH-FR-027` |
| `API-AUTH-028` | `HandleFederatedCallback` | public | command | `HandleFederatedCallbackResponse` | `AUTH-FR-028` |
| `API-AUTH-029` | `RevokeClientAccess` | public | command | `google.longrunning.Operation` | `AUTH-FR-029` |

## 3.4. Access

| ID | RPC method | Видимость | Тип | Ответ | Требования |
| --- | --- | --- | --- | --- | --- |
| `API-ACC-001` | `CheckPermission` | public | query | `CheckPermissionResponse` | `ACC-FR-001` |
| `API-ACC-002` | `BatchCheckPermissions` | public | query | `BatchCheckPermissionsResponse` | `ACC-FR-002` |
| `API-ACC-003` | `ExplainAccessDecision` | public | query | `ExplainAccessDecisionResponse` | `ACC-FR-003` |
| `API-ACC-004` | `ListEffectivePermissions` | public | query | `ListEffectivePermissionsResponse` | `ACC-FR-004` |
| `API-ACC-005` | `ListSubjectsWithAccess` | public | query | `ListSubjectsWithAccessResponse` | `ACC-FR-005` |
| `API-ACC-010` | `CreatePermissionDefinition` | public | command | `CreatePermissionDefinitionResponse` | `ACC-FR-010` |
| `API-ACC-011` | `CreateRole` | public | command | `CreateRoleResponse` | `ACC-FR-011` |
| `API-ACC-012` | `UpdateRole` | public | command | `UpdateRoleResponse` | `ACC-FR-012` |
| `API-ACC-013` | `DeleteRole` | public | command | `DeleteRoleResponse` | `ACC-FR-013` |
| `API-ACC-014` | `CreateRoleBinding` | public | command | `CreateRoleBindingResponse` | `ACC-FR-014` |
| `API-ACC-015` | `RevokeRoleBinding` | public | command | `RevokeRoleBindingResponse` | `ACC-FR-015` |
| `API-ACC-016` | `SetRoleBindingExpiration` | public | command | `SetRoleBindingExpirationResponse` | `ACC-FR-016` |
| `API-ACC-020` | `CreateRelationship` | public | command | `CreateRelationshipResponse` | `ACC-FR-020` |
| `API-ACC-021` | `DeleteRelationship` | public | command | `DeleteRelationshipResponse` | `ACC-FR-021` |
| `API-ACC-022` | `BatchUpdateRelationships` | public | command | `BatchUpdateRelationshipsResponse` | `ACC-FR-022` |
| `API-ACC-030` | `SimulateAccessChange` | public | query | `SimulateAccessChangeResponse` | `ACC-FR-030` |
| `API-ACC-031` | `CheckPolicyImpact` | public | query | `CheckPolicyImpactResponse` | `ACC-FR-031` |
| `API-ACC-040` | `CreateAccessReview` | public | command | `google.longrunning.Operation` | `ACC-FR-040` |
| `API-ACC-041` | `CompleteAccessReview` | public | command | `CompleteAccessReviewResponse` | `ACC-FR-041` |
| `API-ACC-050` | `PublishAuthorizationModel` | admin | command | `google.longrunning.Operation` | `ACC-FR-050` |
| `API-ACC-051` | `SynchronizeSpiceDB` | admin | command | `google.longrunning.Operation` | `ACC-FR-051` |

## 3.5. Risk Decision

| ID | RPC method | Видимость | Тип | Ответ | Требования |
| --- | --- | --- | --- | --- | --- |
| `API-RISK-001` | `EvaluateAuthenticationRisk` | public | command | `EvaluateAuthenticationRiskResponse` | `RISK-FR-001` |
| `API-RISK-002` | `EvaluatePrivilegedActionRisk` | public | command | `EvaluatePrivilegedActionRiskResponse` | `RISK-FR-002` |
| `API-RISK-003` | `ResolveRiskDecision` | internal | command | `ResolveRiskDecisionResponse` | `RISK-FR-003` |
| `API-RISK-004` | `GetRiskAssessment` | public | query | `GetRiskAssessmentResponse` | `RISK-FR-004` |
| `API-RISK-005` | `ExpireRiskDecision` | public | command | `ExpireRiskDecisionResponse` | `RISK-FR-005` |
| `API-RISK-010` | `CreateRiskPolicy` | public | command | `CreateRiskPolicyResponse` | `RISK-FR-010` |
| `API-RISK-011` | `PublishRiskPolicy` | admin | command | `PublishRiskPolicyResponse` | `RISK-FR-011` |
| `API-RISK-012` | `RollbackRiskPolicy` | admin | command | `RollbackRiskPolicyResponse` | `RISK-FR-012` |
| `API-RISK-013` | `SimulateRiskPolicy` | public | query | `SimulateRiskPolicyResponse` | `RISK-FR-013` |
| `API-RISK-014` | `ExplainRiskDecision` | public | query | `ExplainRiskDecisionResponse` | `RISK-FR-014` |
| `API-RISK-020` | `IngestDeviceSignals` | public | command | `IngestDeviceSignalsResponse` | `RISK-FR-020` |
| `API-RISK-021` | `CheckVelocity` | public | query | `CheckVelocityResponse` | `RISK-FR-021` |
| `API-RISK-022` | `IngestExternalRiskSignal` | public | command | `IngestExternalRiskSignalResponse` | `RISK-FR-022` |
| `API-RISK-030` | `CreateManualReview` | public | command | `google.longrunning.Operation` | `RISK-FR-030` |
| `API-RISK-031` | `CompleteManualReview` | public | command | `CompleteManualReviewResponse` | `RISK-FR-031` |
| `API-RISK-040` | `CreateRiskFeedback` | public | command | `CreateRiskFeedbackResponse` | `RISK-FR-040` |

## 3.6. Provisioning

| ID | RPC method | Видимость | Тип | Ответ | Требования |
| --- | --- | --- | --- | --- | --- |
| `API-PROV-001` | `RegisterResourceDefinition` | public | command | `RegisterResourceDefinitionResponse` | `PROV-FR-001` |
| `API-PROV-002` | `PublishResourceDefinition` | public | command | `PublishResourceDefinitionResponse` | `PROV-FR-002` |
| `API-PROV-003` | `RegisterDriver` | admin | command | `RegisterDriverResponse` | `PROV-FR-003` |
| `API-PROV-004` | `DisableDriver` | admin | command | `DisableDriverResponse` | `PROV-FR-004` |
| `API-PROV-010` | `CreateManagedResource` | public | command | `google.longrunning.Operation` | `PROV-FR-010` |
| `API-PROV-011` | `UpdateDesiredState` | public | command | `google.longrunning.Operation` | `PROV-FR-011` |
| `API-PROV-012` | `DeleteManagedResource` | public | command | `google.longrunning.Operation` | `PROV-FR-012` |
| `API-PROV-013` | `PauseReconciliation` | public | command | `PauseReconciliationResponse` | `PROV-FR-013` |
| `API-PROV-014` | `ResumeReconciliation` | public | command | `ResumeReconciliationResponse` | `PROV-FR-014` |
| `API-PROV-020` | `SelectPlacement` | internal | command | `SelectPlacementResponse` | `PROV-FR-020` |
| `API-PROV-021` | `PinPlacement` | internal | command | `PinPlacementResponse` | `PROV-FR-021` |
| `API-PROV-030` | `ReconcileManagedResource` | internal | command | `google.longrunning.Operation` | `PROV-FR-030` |
| `API-PROV-031` | `GetObservedState` | public | query | `GetObservedStateResponse` | `PROV-FR-031` |
| `API-PROV-032` | `DetectDrift` | internal | command | `DetectDriftResponse` | `PROV-FR-032` |
| `API-PROV-033` | `RemediateDrift` | public | command | `google.longrunning.Operation` | `PROV-FR-033` |
| `API-PROV-034` | `ImportManagedResource` | public | command | `google.longrunning.Operation` | `PROV-FR-034` |
| `API-PROV-040` | `RetryProvisioningStep` | admin | command | `RetryProvisioningStepResponse` | `PROV-FR-040` |
| `API-PROV-041` | `RunCompensation` | admin | command | `google.longrunning.Operation` | `PROV-FR-041` |
| `API-PROV-042` | `CreateManualRemediation` | admin | command | `google.longrunning.Operation` | `PROV-FR-042` |
| `API-PROV-050` | `GetManagedResourceOutputs` | public | query | `GetManagedResourceOutputsResponse` | `PROV-FR-050` |
| `API-PROV-051` | `GetManagedResource` | public | query | `GetManagedResourceResponse` | `PROV-FR-051` |

## 3.7. Audit

| ID | RPC method | Видимость | Тип | Ответ | Требования |
| --- | --- | --- | --- | --- | --- |
| `API-AUD-001` | `IngestAuditEvent` | public | command | `IngestAuditEventResponse` | `AUD-FR-001` |
| `API-AUD-002` | `ValidateAuditEventProvenance` | internal | query | `ValidateAuditEventProvenanceResponse` | `AUD-FR-002` |
| `API-AUD-003` | `ValidateAuditDataMinimization` | internal | query | `ValidateAuditDataMinimizationResponse` | `AUD-FR-003` |
| `API-AUD-004` | `PersistAuditEvent` | internal | command | `PersistAuditEventResponse` | `AUD-FR-004` |
| `API-AUD-005` | `AcknowledgeRequiredAuditEvent` | internal | command | `AcknowledgeRequiredAuditEventResponse` | `AUD-FR-005` |
| `API-AUD-010` | `SearchAuditEvents` | public | query | `SearchAuditEventsResponse` | `AUD-FR-010` |
| `API-AUD-011` | `GetAuditEvent` | public | query | `GetAuditEventResponse` | `AUD-FR-011` |
| `API-AUD-012` | `GetCorrelationChain` | public | query | `GetCorrelationChainResponse` | `AUD-FR-012` |
| `API-AUD-013` | `GetResourceHistory` | public | query | `GetResourceHistoryResponse` | `AUD-FR-013` |
| `API-AUD-020` | `CreateAuditExport` | public | command | `google.longrunning.Operation` | `AUD-FR-020` |
| `API-AUD-021` | `GetAuditExport` | public | query | `GetAuditExportResponse` | `AUD-FR-021` |
| `API-AUD-030` | `ApplyRetentionPolicy` | admin | command | `google.longrunning.Operation` | `AUD-FR-030` |
| `API-AUD-031` | `SetLegalHold` | admin | command | `SetLegalHoldResponse` | `AUD-FR-031` |
| `API-AUD-040` | `VerifyAuditIntegrity` | admin | query | `google.longrunning.Operation` | `AUD-FR-040` |
| `API-AUD-041` | `GetIntegrityProof` | admin | query | `GetIntegrityProofResponse` | `AUD-FR-041` |
| `API-AUD-050` | `AuditAuditAccess` | admin | command | `AuditAuditAccessResponse` | `AUD-FR-050` |

## 3.8. Common Operation

| ID | RPC method | Видимость | Тип | Ответ | Требования |
| --- | --- | --- | --- | --- | --- |
| `API-OPS-001` | `GetOperation` | public | query | `GetOperationResponse` | `OPS-FR-001` |
| `API-OPS-002` | `ListOperations` | public | query | `ListOperationsResponse` | `OPS-FR-002` |
| `API-OPS-003` | `WaitOperation` | public | command | `WaitOperationResponse` | `OPS-FR-003` |
| `API-OPS-004` | `CancelOperation` | public | command | `CancelOperationResponse` | `OPS-FR-004` |
| `API-OPS-005` | `DeleteOperation` | public | command | `DeleteOperationResponse` | `OPS-FR-005` |
| `API-OPS-006` | `UpdateOperationProgress` | internal | command | `UpdateOperationProgressResponse` | `OPS-FR-006` |
| `API-OPS-007` | `CompleteOperation` | internal | command | `CompleteOperationResponse` | `OPS-FR-007` |
| `API-OPS-008` | `FailOperation` | internal | command | `FailOperationResponse` | `OPS-FR-008` |
| `API-OPS-009` | `LinkOperationWorkflow` | internal | command | `LinkOperationWorkflowResponse` | `OPS-FR-009` |
| `API-OPS-010` | `StartIdempotentOperation` | internal | command | `StartIdempotentOperationResponse` | `OPS-FR-010` |

# 4. Обязательные поля публичного mutation request

```proto
message MutationContext {
  m8.common.v1.RequestContext request_context = 1;
  string idempotency_key = 2;
  string expected_revision = 3;
  google.protobuf.FieldMask update_mask = 4;
}
```

# 5. Design gate

Перед переводом API в `approved` должны быть определены request/response messages, field behavior, protovalidate, permissions, canonical errors, timeout/retry policy, LRO metadata/result, examples и compatibility test baseline.
