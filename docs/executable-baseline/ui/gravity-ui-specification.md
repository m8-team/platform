---
title: "M8 Gravity UI Specification v1.0"
---

# M8 Gravity UI Specification v1.0

[Executable Baseline](../index.md) | [UI Baseline](index.md)

    | № | Экран | Назначение | Основные API |
    | ---: | --- | --- | --- |
    | 1 | Platform Overview | SLO, incidents, operations, projects, authentication and audit summary | ListOperations, SearchResources |
| 2 | Organizations | Organization lifecycle and resource hierarchy | CreateOrganization, UpdateOrganization, GetResource |
| 3 | Workspaces | Workspace lifecycle and child projects | CreateWorkspace, ListChildResources |
| 4 | Projects | Project lifecycle, labels, move/delete operations | CreateProject, MoveProject, DeleteProject |
| 5 | User Pools | Pools, status and identity policy | CreateUserPool, UpdateUserPool |
| 6 | Users | Profiles, external identities, status and merge | CreateUser, UpdateUserProfile, MergeUsers |
| 7 | Groups and Memberships | Group membership management | CreateGroup, AddMembership, RemoveMembership |
| 8 | Authentication Transactions | State, challenges, provider and operation | StartAuthentication, GetAuthentication, CancelAuthentication |
| 9 | Clients | Client registration and auth policy | RegisterClient, UpdateClientPolicy |
| 10 | Access Explorer | Permission check and decision explanation | CheckPermission, ExplainAccessDecision |
| 11 | Roles and Bindings | Roles, bindings and expiry | CreateRole, CreateRoleBinding, RevokeRoleBinding |
| 12 | Policy Simulator | Access impact simulation | SimulateAccessChange, CheckPolicyImpact |
| 13 | Risk Decisions | Assessment, signals and explanation | GetRiskAssessment, ExplainRiskDecision |
| 14 | Risk Policies | Policy lifecycle and simulation | CreateRiskPolicy, PublishRiskPolicy, SimulateRiskPolicy |
| 15 | Managed Resources | Desired/observed state, outputs and drift | CreateManagedResource, GetManagedResource, DetectDrift |
| 16 | Operations | Progress, wait, cancel and error | ListOperations, GetOperation, CancelOperation |
| 17 | Audit Log | Search, correlation chain and integrity | SearchAuditEvents, GetCorrelationChain, VerifyAuditIntegrity |

    ## Общие состояния

    Каждый экран обязан иметь loading, empty, partial, stale projection, permission denied,
    dependency unavailable и retry states. Mutation показывает idempotency-safe retry,
    Operation progress и ссылку на Audit correlation chain.

    ## Безопасность

    UI не является источником авторизации. Все действия повторно проверяются сервисом.
    Secret material и sensitive risk signals не отображаются без специального permission.
