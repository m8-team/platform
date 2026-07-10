---
title: "M8 Platform Event Contract Catalog"
description: "Каталог event contract candidates."
keywords:
  - "M8 Platform"
  - "contracts"
---

# M8 Platform Event Contract Catalog {#event-contract-catalog}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [Contracts](index.md) | [Requirements Catalog](../../architecture/requirements/index.md) | `event-catalog.yaml`

{% endnote %}

_M8-EVT-000 · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `M8-EVT-000` |
| Версия | `0.1` |
| Статус | Базовая проектная редакция |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | Domain Events и Integration Events |

# 1. Нормативная модель

- Domain Event существует внутри владельца агрегата и MAY не покидать процесс.
- Integration Event является опубликованным фактом и MUST фиксироваться через Transactional Outbox.
- Delivery — at-least-once; consumer MUST дедуплицировать `event_id`.
- Ordering гарантируется только в пределах заявленного partition key и aggregate revision.
- В событии запрещены секреты, OTP, токены и необработанные credential.

# 2. Канонический envelope

```yaml
event_id: uuid
event_type: m8.authentication.authentication_started.v1
occurred_at: RFC3339 UTC
producer: m8-authentication
aggregate_id: authentication_id
aggregate_revision: 1
correlation_id: uuid
causation_id: uuid
actor_ref: minimal trusted reference
resource_scope: Organization/Workspace/Project reference
schema_version: 1
payload: typed protobuf message
```

# 3. Реестр событий

## 3.1. Resource Manager

| ID | Событие | Topic | Partition key | Требования |
| --- | --- | --- | --- | --- |
| `EVT-RM-001` | `OrganizationCreated` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-001` |
| `EVT-RM-002` | `OrganizationUpdated` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-002` |
| `EVT-RM-003` | `OrganizationSuspended` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-003` |
| `EVT-RM-004` | `OrganizationRestored` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-004` |
| `EVT-RM-005` | `OrganizationArchived` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-005` |
| `EVT-RM-010` | `WorkspaceCreated` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-010` |
| `EVT-RM-011` | `WorkspaceUpdated` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-011` |
| `EVT-RM-012` | `WorkspaceSuspensionChanged` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-012` |
| `EVT-RM-020` | `ProjectCreated` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-020` |
| `EVT-RM-021` | `ProjectMoved` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-021` |
| `EVT-RM-022` | `ProjectDeletionRequested` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-022` |
| `EVT-RM-023` | `ProjectSuspended` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-023` |
| `EVT-RM-024` | `ProjectRestored` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-024` |
| `EVT-RM-030` | `ServiceRegistered` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-030` |
| `EVT-RM-031` | `ServiceRegistrationUpdated` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-031` |
| `EVT-RM-032` | `ServiceUnregistered` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-032` |
| `EVT-RM-044` | `UpdateResourceLabelsCompleted` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-044` |
| `EVT-RM-050` | `ResourceLifecycleChanged` | `m8.resource_manager.events.v1` | `resource_id` | `RM-FR-050` |

## 3.2. Identity

| ID | Событие | Topic | Partition key | Требования |
| --- | --- | --- | --- | --- |
| `EVT-ID-001` | `UserPoolCreated` | `m8.identity.events.v1` | `subject_id` | `ID-FR-001` |
| `EVT-ID-002` | `UserPoolUpdated` | `m8.identity.events.v1` | `subject_id` | `ID-FR-002` |
| `EVT-ID-003` | `UserPoolSuspended` | `m8.identity.events.v1` | `subject_id` | `ID-FR-003` |
| `EVT-ID-010` | `UserCreated` | `m8.identity.events.v1` | `subject_id` | `ID-FR-010` |
| `EVT-ID-011` | `UserProfileUpdated` | `m8.identity.events.v1` | `subject_id` | `ID-FR-011` |
| `EVT-ID-012` | `UserDisabled` | `m8.identity.events.v1` | `subject_id` | `ID-FR-012` |
| `EVT-ID-013` | `UserRestored` | `m8.identity.events.v1` | `subject_id` | `ID-FR-013` |
| `EVT-ID-020` | `ExternalIdentityLinked` | `m8.identity.events.v1` | `subject_id` | `ID-FR-020` |
| `EVT-ID-021` | `DetectExternalIdentityConflictCompleted` | `m8.identity.events.v1` | `subject_id` | `ID-FR-021` |
| `EVT-ID-022` | `ExternalIdentityUnlinked` | `m8.identity.events.v1` | `subject_id` | `ID-FR-022` |
| `EVT-ID-030` | `CreateGroupCompleted` | `m8.identity.events.v1` | `subject_id` | `ID-FR-030` |
| `EVT-ID-031` | `UpdateGroupCompleted` | `m8.identity.events.v1` | `subject_id` | `ID-FR-031` |
| `EVT-ID-032` | `MembershipAdded` | `m8.identity.events.v1` | `subject_id` | `ID-FR-032` |
| `EVT-ID-033` | `MembershipRemoved` | `m8.identity.events.v1` | `subject_id` | `ID-FR-033` |
| `EVT-ID-040` | `UsersMerged` | `m8.identity.events.v1` | `subject_id` | `ID-FR-040` |
| `EVT-ID-041` | `DetectPotentialDuplicatesCompleted` | `m8.identity.events.v1` | `subject_id` | `ID-FR-041` |
| `EVT-ID-050` | `UserPrivacyActionCompleted` | `m8.identity.events.v1` | `subject_id` | `ID-FR-050` |
| `EVT-ID-060` | `IdentityLifecycleChanged` | `m8.identity.events.v1` | `subject_id` | `ID-FR-060` |

## 3.3. Authentication

| ID | Событие | Topic | Partition key | Требования |
| --- | --- | --- | --- | --- |
| `EVT-AUTH-001` | `AuthenticationStarted` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-001`, `AUTH-FR-017` |
| `EVT-AUTH-002` | `SelectAuthenticationProviderCompleted` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-002` |
| `EVT-AUTH-003` | `AuthenticationChallengeCreated` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-003` |
| `EVT-AUTH-006` | `AuthenticationCancelled` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-006` |
| `EVT-AUTH-007` | `AuthenticationProviderCallbackProcessed` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-007` |
| `EVT-AUTH-008` | `AuthenticationExpired` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-008` |
| `EVT-AUTH-009` | `AuthenticationFailed` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-009` |
| `EVT-AUTH-010` | `AuthenticationHandoffCreated` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-010` |
| `EVT-AUTH-011` | `AuthenticationHandoffRedeemed` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-011` |
| `EVT-AUTH-012` | `AuthenticationSessionCreated` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-012` |
| `EVT-AUTH-013` | `AuthenticationSessionRevoked` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-013` |
| `EVT-AUTH-015` | `AuthenticationClientRegistered` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-015` |
| `EVT-AUTH-016` | `AuthenticationClientPolicyUpdated` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-016` |
| `EVT-AUTH-018` | `CibaAuthenticationStarted` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-018` |
| `EVT-AUTH-019` | `CibaDecisionProcessed` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-019` |
| `EVT-AUTH-020` | `StepUpStarted` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-020` |
| `EVT-AUTH-021` | `SelectStepUpChallengeCompleted` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-021` |
| `EVT-AUTH-022` | `OtpChallengeSent` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-022` |
| `EVT-AUTH-023` | `OtpVerified` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-023` |
| `EVT-AUTH-024` | `AuthenticationChallengeResent` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-024` |
| `EVT-AUTH-025` | `WebAuthnAssertionVerified` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-025` |
| `EVT-AUTH-026` | `WebAuthnCredentialRegistered` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-026` |
| `EVT-AUTH-027` | `FederatedAuthenticationStarted` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-027` |
| `EVT-AUTH-028` | `FederatedCallbackProcessed` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-028` |
| `EVT-AUTH-029` | `ClientAccessRevoked` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-029` |
| `EVT-AUTH-030` | `AuthenticationLifecycleChanged` | `m8.authentication.events.v1` | `authentication_id` | `AUTH-FR-030` |

## 3.4. Access

| ID | Событие | Topic | Partition key | Требования |
| --- | --- | --- | --- | --- |
| `EVT-ACC-010` | `CreatePermissionDefinitionCompleted` | `m8.access.events.v1` | `resource_id` | `ACC-FR-010` |
| `EVT-ACC-011` | `RoleCreated` | `m8.access.events.v1` | `resource_id` | `ACC-FR-011` |
| `EVT-ACC-012` | `RoleUpdated` | `m8.access.events.v1` | `resource_id` | `ACC-FR-012` |
| `EVT-ACC-013` | `RoleDeleted` | `m8.access.events.v1` | `resource_id` | `ACC-FR-013` |
| `EVT-ACC-014` | `RoleBindingCreated` | `m8.access.events.v1` | `resource_id` | `ACC-FR-014` |
| `EVT-ACC-015` | `RoleBindingRevoked` | `m8.access.events.v1` | `resource_id` | `ACC-FR-015` |
| `EVT-ACC-016` | `RoleBindingExpirationChanged` | `m8.access.events.v1` | `resource_id` | `ACC-FR-016` |
| `EVT-ACC-020` | `AccessRelationshipCreated` | `m8.access.events.v1` | `resource_id` | `ACC-FR-020` |
| `EVT-ACC-021` | `AccessRelationshipDeleted` | `m8.access.events.v1` | `resource_id` | `ACC-FR-021` |
| `EVT-ACC-040` | `AccessReviewCreated` | `m8.access.events.v1` | `resource_id` | `ACC-FR-040` |
| `EVT-ACC-041` | `AccessReviewCompleted` | `m8.access.events.v1` | `resource_id` | `ACC-FR-041` |
| `EVT-ACC-050` | `AuthorizationModelPublished` | `m8.access.events.v1` | `resource_id` | `ACC-FR-050` |
| `EVT-ACC-060` | `AccessRelationshipChanged` | `m8.access.events.v1` | `resource_id` | `ACC-FR-060` |

## 3.5. Risk Decision

| ID | Событие | Topic | Partition key | Требования |
| --- | --- | --- | --- | --- |
| `EVT-RISK-005` | `RiskDecisionExpired` | `m8.risk_decision.events.v1` | `assessment_id` | `RISK-FR-005` |
| `EVT-RISK-010` | `RiskPolicyCreated` | `m8.risk_decision.events.v1` | `assessment_id` | `RISK-FR-010` |
| `EVT-RISK-011` | `RiskPolicyPublished` | `m8.risk_decision.events.v1` | `assessment_id` | `RISK-FR-011` |
| `EVT-RISK-012` | `RiskPolicyRolledBack` | `m8.risk_decision.events.v1` | `assessment_id` | `RISK-FR-012` |
| `EVT-RISK-020` | `DeviceSignalsIngested` | `m8.risk_decision.events.v1` | `assessment_id` | `RISK-FR-020` |
| `EVT-RISK-022` | `ExternalRiskSignalIngested` | `m8.risk_decision.events.v1` | `assessment_id` | `RISK-FR-022` |
| `EVT-RISK-030` | `ManualRiskReviewCreated` | `m8.risk_decision.events.v1` | `assessment_id` | `RISK-FR-030` |
| `EVT-RISK-031` | `ManualRiskReviewCompleted` | `m8.risk_decision.events.v1` | `assessment_id` | `RISK-FR-031` |
| `EVT-RISK-040` | `RiskFeedbackCreated` | `m8.risk_decision.events.v1` | `assessment_id` | `RISK-FR-040` |
| `EVT-RISK-050` | `RiskDecisionRecorded` | `m8.risk_decision.events.v1` | `assessment_id` | `RISK-FR-050` |

## 3.6. Provisioning

| ID | Событие | Topic | Partition key | Требования |
| --- | --- | --- | --- | --- |
| `EVT-PROV-001` | `ResourceDefinitionRegistered` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-001` |
| `EVT-PROV-002` | `ResourceDefinitionPublished` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-002` |
| `EVT-PROV-003` | `ProvisioningDriverRegistered` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-003` |
| `EVT-PROV-004` | `ProvisioningDriverDisabled` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-004` |
| `EVT-PROV-010` | `ManagedResourceCreated` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-010` |
| `EVT-PROV-011` | `ManagedResourceDesiredStateUpdated` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-011` |
| `EVT-PROV-012` | `ManagedResourceDeletionRequested` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-012` |
| `EVT-PROV-020` | `SelectPlacementCompleted` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-020` |
| `EVT-PROV-021` | `PlacementPinned` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-021` |
| `EVT-PROV-030` | `ManagedResourceReconciled` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-030` |
| `EVT-PROV-032` | `ManagedResourceDriftDetected` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-032` |
| `EVT-PROV-033` | `ManagedResourceDriftRemediated` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-033` |
| `EVT-PROV-034` | `ManagedResourceImported` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-034` |
| `EVT-PROV-040` | `ProvisioningStepRetried` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-040` |
| `EVT-PROV-041` | `ProvisioningCompensationCompleted` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-041` |
| `EVT-PROV-042` | `ManualRemediationCreated` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-042` |
| `EVT-PROV-060` | `ProvisioningLifecycleChanged` | `m8.provisioning.events.v1` | `managed_resource_id` | `PROV-FR-060` |

## 3.7. Audit

| ID | Событие | Topic | Partition key | Требования |
| --- | --- | --- | --- | --- |
| `EVT-AUD-001` | `AuditEventAccepted` | `m8.audit.events.v1` | `audit_event_id` | `AUD-FR-001` |
| `EVT-AUD-004` | `PersistAuditEventCompleted` | `m8.audit.events.v1` | `audit_event_id` | `AUD-FR-004` |
| `EVT-AUD-005` | `AcknowledgeRequiredAuditEventCompleted` | `m8.audit.events.v1` | `audit_event_id` | `AUD-FR-005` |
| `EVT-AUD-020` | `AuditExportCreated` | `m8.audit.events.v1` | `audit_event_id` | `AUD-FR-020` |
| `EVT-AUD-030` | `AuditRetentionApplied` | `m8.audit.events.v1` | `audit_event_id` | `AUD-FR-030` |
| `EVT-AUD-031` | `AuditLegalHoldChanged` | `m8.audit.events.v1` | `audit_event_id` | `AUD-FR-031` |
| `EVT-AUD-040` | `AuditIntegrityVerified` | `m8.audit.events.v1` | `audit_event_id` | `AUD-FR-040` |

## 3.8. Common Operation

| ID | Событие | Topic | Partition key | Требования |
| --- | --- | --- | --- | --- |
| `EVT-OPS-004` | `OperationCancellationRequested` | `m8.operations.events.v1` | `operation_id` | `OPS-FR-004` |
| `EVT-OPS-005` | `OperationDeleted` | `m8.operations.events.v1` | `operation_id` | `OPS-FR-005` |
| `EVT-OPS-006` | `OperationProgressUpdated` | `m8.operations.events.v1` | `operation_id` | `OPS-FR-006` |
| `EVT-OPS-007` | `OperationCompleted` | `m8.operations.events.v1` | `operation_id` | `OPS-FR-007` |
| `EVT-OPS-008` | `OperationFailed` | `m8.operations.events.v1` | `operation_id` | `OPS-FR-008` |
| `EVT-OPS-009` | `OperationWorkflowLinked` | `m8.operations.events.v1` | `operation_id` | `OPS-FR-009` |
| `EVT-OPS-010` | `OperationStarted` | `m8.operations.events.v1` | `operation_id` | `OPS-FR-010` |

# 4. Совместимость

Добавление optional-поля совместимо; удаление, изменение номера/смысла поля, смена partition key или расширение PII считается breaking change. Несовместимая семантика выпускается новым event type/version.
