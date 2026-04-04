import * as React from 'react';

import {DynamicView, SpecTypes, dynamicViewConfig} from '@gravity-ui/dynamic-forms';
import {DatePicker} from '@gravity-ui/date-components';
import {dateTime} from '@gravity-ui/date-utils';
import {Button, Dialog, Drawer, Select, Text, TextArea, TextInput} from '@gravity-ui/uikit';

import {appToaster} from '@/app/providers/app-toaster';
import {
  useCreateTenantMutation,
  useCreateServiceAccountMutation,
  useCreateSupportGrantMutation,
  useDeleteTenantMutation,
  useGrantAccessMutation,
  useGroupsQuery,
  useRolesQuery,
  useServiceAccountsQuery,
  useTenantsQuery,
  useUpdateTenantMutation,
  useUsersQuery,
  useRotateSecretMutation,
} from '@/entities/queries';
import {formatDateTime} from '@/shared/lib/format';
import {FieldHint, HighlightAlert, SectionCard} from '@/shared/ui/blocks';
import type {ResourceType} from '@/shared/types/iam';

function closeDialog(onClose: () => void) {
  return () => onClose();
}

function syncOverlayOpen(nextOpen: boolean, onClose: () => void) {
  if (!nextOpen) {
    onClose();
  }
}

function createDefaultSupportGrantExpiry() {
  return dateTime({input: new Date(Date.now() + 4 * 60 * 60 * 1000)});
}

const serviceAccountReviewSpec = {
  type: SpecTypes.Object,
  properties: {
    tenantId: {
      type: SpecTypes.String,
      viewSpec: {type: 'base', layout: 'row', layoutTitle: 'Tenant'},
    },
    displayName: {
      type: SpecTypes.String,
      viewSpec: {type: 'base', layout: 'row', layoutTitle: 'Display name'},
    },
    description: {
      type: SpecTypes.String,
      viewSpec: {type: 'textarea', layout: 'row', layoutTitle: 'Description'},
    },
  },
  viewSpec: {type: 'base', layout: 'section', layoutTitle: 'Preview'},
} as const;

const grantReviewSpec = {
  type: SpecTypes.Object,
  properties: {
    subject: {type: SpecTypes.String, viewSpec: {type: 'base', layout: 'row', layoutTitle: 'Subject'}},
    roleId: {type: SpecTypes.String, viewSpec: {type: 'base', layout: 'row', layoutTitle: 'Role'}},
    resourceId: {
      type: SpecTypes.String,
      viewSpec: {type: 'base', layout: 'row', layoutTitle: 'Resource'},
    },
    expiresAt: {
      type: SpecTypes.String,
      viewSpec: {type: 'base', layout: 'row', layoutTitle: 'Expires at'},
    },
  },
  viewSpec: {type: 'base', layout: 'section', layoutTitle: 'Binding review'},
} as const;

const supportGrantReviewSpec = {
  type: SpecTypes.Object,
  properties: {
    tenantId: {type: SpecTypes.String, viewSpec: {type: 'base', layout: 'row', layoutTitle: 'Tenant'}},
    subjectId: {type: SpecTypes.String, viewSpec: {type: 'base', layout: 'row', layoutTitle: 'Support subject'}},
    roleId: {type: SpecTypes.String, viewSpec: {type: 'base', layout: 'row', layoutTitle: 'Role'}},
    incidentId: {type: SpecTypes.String, viewSpec: {type: 'base', layout: 'row', layoutTitle: 'Incident'}},
    expiresAt: {type: SpecTypes.String, viewSpec: {type: 'base', layout: 'row', layoutTitle: 'Expires at'}},
    reason: {type: SpecTypes.String, viewSpec: {type: 'textarea', layout: 'row', layoutTitle: 'Reason'}},
  },
  viewSpec: {type: 'base', layout: 'section', layoutTitle: 'Support grant review'},
} as const;

const tenantIdPattern = /^[a-z][a-z0-9-]{2,127}$/;
const tenantTierOptions = [
  {value: 'active', content: 'Active / Enterprise'},
  {value: 'trial', content: 'Trial'},
];
const tenantRegionOptions = [
  {value: 'eu-central', content: 'eu-central'},
  {value: 'us-east', content: 'us-east'},
  {value: 'eu-west', content: 'eu-west'},
  {value: 'ap-south', content: 'ap-south'},
];

function TenantEditorDialog({
  open,
  mode,
  tenant,
  onClose,
  onSuccess,
}: {
  open: boolean;
  mode: 'create' | 'edit';
  tenant?: {
    tenantId: string;
    name: string;
    externalRef?: string;
    region: string;
    status: 'active' | 'trial' | string;
  };
  onClose: () => void;
  onSuccess?: (tenantId: string) => void;
}) {
  const createMutation = useCreateTenantMutation();
  const updateMutation = useUpdateTenantMutation();
  const mutation = mode === 'create' ? createMutation : updateMutation;
  const [tenantId, setTenantId] = React.useState(tenant?.tenantId ?? '');
  const [displayName, setDisplayName] = React.useState(tenant?.name ?? '');
  const [externalRef, setExternalRef] = React.useState(tenant?.externalRef ?? '');
  const [region, setRegion] = React.useState(tenant?.region ?? 'eu-central');
  const [tier, setTier] = React.useState<'active' | 'trial'>(tenant?.status === 'trial' ? 'trial' : 'active');

  React.useEffect(() => {
    if (open) {
      setTenantId(tenant?.tenantId ?? '');
      setDisplayName(tenant?.name ?? '');
      setExternalRef(tenant?.externalRef ?? '');
      setRegion(tenant?.region ?? 'eu-central');
      setTier(tenant?.status === 'trial' ? 'trial' : 'active');
    }
  }, [open, tenant]);

  const canSubmit = tenantIdPattern.test(tenantId.trim()) && displayName.trim().length > 1;

  return (
    <Dialog
      open={open}
      onClose={closeDialog(onClose)}
      onOpenChange={(nextOpen) => syncOverlayOpen(nextOpen, onClose)}
      size="m"
    >
      <Dialog.Header caption={mode === 'create' ? 'Create Tenant' : 'Edit Tenant'} />
      <Dialog.Body>
        <div className="form-grid">
          <TextInput
            label="Tenant ID"
            placeholder="tenant-acme-prod"
            value={tenantId}
            disabled={mode === 'edit'}
            onUpdate={setTenantId}
          />
          <TextInput
            label="Display name"
            placeholder="Acme Production"
            value={displayName}
            onUpdate={setDisplayName}
          />
          <TextInput
            label="External ref"
            placeholder="crm-acme-001"
            value={externalRef}
            onUpdate={setExternalRef}
          />
          <Select
            label="Region"
            value={[region]}
            options={tenantRegionOptions}
            onUpdate={(value) => setRegion(value[0] ?? 'eu-central')}
          />
          <Select
            label="Tier"
            value={[tier]}
            options={tenantTierOptions}
            onUpdate={(value) => setTier((value[0] ?? 'active') as 'active' | 'trial')}
          />
          <div className="form-grid__full">
            <FieldHint>
              {mode === 'create'
                ? 'Tenant metadata is created through the IAM identity facade.'
                : 'Display name, external reference and labels are updated via the tenant API.'}
            </FieldHint>
          </div>
        </div>
      </Dialog.Body>
      <Dialog.Footer
        textButtonCancel="Cancel"
        textButtonApply={mode === 'create' ? 'Create' : 'Save'}
        onClickButtonCancel={onClose}
        onClickButtonApply={() => {
          mutation.mutate(
            {
              tenantId: tenantId.trim(),
              displayName: displayName.trim(),
              externalRef: externalRef.trim(),
              region,
              tier,
            },
            {
              onSuccess: (savedTenant) => {
                appToaster.add({
                  name: `${mode}-tenant-${savedTenant.tenantId}`,
                  title: mode === 'create' ? 'Tenant created' : 'Tenant updated',
                  content: `${savedTenant.name} (${savedTenant.tenantId})`,
                  theme: 'success',
                });
                onSuccess?.(savedTenant.tenantId);
                onClose();
              },
            },
          );
        }}
        propsButtonApply={{view: 'action', disabled: !canSubmit}}
        loading={mutation.isPending}
      />
    </Dialog>
  );
}

export function CreateTenantDialog({
  open,
  onClose,
  onSuccess,
}: {
  open: boolean;
  onClose: () => void;
  onSuccess?: (tenantId: string) => void;
}) {
  return <TenantEditorDialog open={open} mode="create" onClose={onClose} onSuccess={onSuccess} />;
}

export function EditTenantDialog({
  open,
  tenant,
  onClose,
  onSuccess,
}: {
  open: boolean;
  tenant: {
    tenantId: string;
    name: string;
    externalRef?: string;
    region: string;
    status: 'active' | 'trial' | string;
  };
  onClose: () => void;
  onSuccess?: (tenantId: string) => void;
}) {
  return (
    <TenantEditorDialog
      open={open}
      mode="edit"
      tenant={tenant}
      onClose={onClose}
      onSuccess={onSuccess}
    />
  );
}

export function DeleteTenantDialog({
  open,
  tenantId,
  tenantName,
  onClose,
  onSuccess,
}: {
  open: boolean;
  tenantId: string;
  tenantName: string;
  onClose: () => void;
  onSuccess?: (tenantId: string) => void;
}) {
  const mutation = useDeleteTenantMutation();
  const [reason, setReason] = React.useState('');

  React.useEffect(() => {
    if (open) {
      setReason('');
    }
  }, [open]);

  return (
    <Dialog
      open={open}
      onClose={closeDialog(onClose)}
      onOpenChange={(nextOpen) => syncOverlayOpen(nextOpen, onClose)}
      size="s"
    >
      <Dialog.Header caption="Delete Tenant" />
      <Dialog.Body>
        <HighlightAlert
          title={`${tenantName} (${tenantId})`}
          message="Tenant metadata will be removed from IAM. Use a clear reason so the audit trail stays useful."
          theme="danger"
        />
        <div className="form-grid">
          <div className="form-grid__full">
            <Text variant="body-2">Reason</Text>
            <TextArea
              rows={4}
              value={reason}
              onUpdate={setReason}
              placeholder="Tenant decommissioned after migration to the new organization."
            />
          </div>
        </div>
      </Dialog.Body>
      <Dialog.Footer
        textButtonCancel="Cancel"
        textButtonApply="Delete"
        onClickButtonCancel={onClose}
        onClickButtonApply={() => {
          mutation.mutate(
            {tenantId, reason: reason.trim()},
            {
              onSuccess: () => {
                appToaster.add({
                  name: `delete-tenant-${tenantId}`,
                  title: 'Tenant deleted',
                  content: `${tenantName} removed from IAM`,
                  theme: 'success',
                });
                onSuccess?.(tenantId);
                onClose();
              },
            },
          );
        }}
        propsButtonApply={{view: 'outlined-danger', disabled: reason.trim().length < 3}}
        loading={mutation.isPending}
      />
    </Dialog>
  );
}

export function CreateServiceAccountWizard({
  open,
  defaultTenantId,
  onClose,
}: {
  open: boolean;
  defaultTenantId?: string;
  onClose: () => void;
}) {
  const tenantsQuery = useTenantsQuery();
  const mutation = useCreateServiceAccountMutation();
  const [step, setStep] = React.useState<'form' | 'review'>('form');
  const [tenantId, setTenantId] = React.useState(defaultTenantId ?? '');
  const [displayName, setDisplayName] = React.useState('');
  const [description, setDescription] = React.useState('');

  React.useEffect(() => {
    if (open) {
      setStep('form');
      setTenantId(defaultTenantId ?? '');
      setDisplayName('');
      setDescription('');
    }
  }, [defaultTenantId, open]);

  const tenantOptions = (tenantsQuery.data?.items ?? []).map((tenant) => ({
    value: tenant.tenantId,
    content: `${tenant.name} (${tenant.tenantId})`,
  }));

  const canContinue = tenantId.trim().length > 0 && displayName.trim().length > 1;

  return (
    <Dialog
      open={open}
      onClose={closeDialog(onClose)}
      onOpenChange={(nextOpen) => syncOverlayOpen(nextOpen, onClose)}
      size="m"
    >
      <Dialog.Header caption="Create Service Account" />
      <Dialog.Body>
        <div className="wizard-stepper">
          <span className={`wizard-stepper__step${step === 'form' ? ' wizard-stepper__step_active' : ''}`}>1. Details</span>
          <span className={`wizard-stepper__step${step === 'review' ? ' wizard-stepper__step_active' : ''}`}>2. Review</span>
        </div>
        {step === 'form' ? (
          <div className="form-grid">
            <Select
              label="Tenant"
              value={tenantId ? [tenantId] : []}
              options={tenantOptions}
              placeholder="Choose tenant"
              onUpdate={(value) => setTenantId(value[0] ?? '')}
            />
            <TextInput
              label="Display name"
              placeholder="billing-worker"
              value={displayName}
              onUpdate={setDisplayName}
            />
            <div className="form-grid__full">
              <Text variant="body-2">Description</Text>
              <TextArea rows={4} value={description} onUpdate={setDescription} placeholder="Worker for invoice sync and exports" />
            </div>
          </div>
        ) : (
          <div className="review-panel">
            <DynamicView
              value={{tenantId, displayName, description}}
              spec={serviceAccountReviewSpec}
              config={dynamicViewConfig}
            />
          </div>
        )}
      </Dialog.Body>
      <Dialog.Footer
        textButtonCancel={step === 'review' ? 'Back' : 'Close'}
        textButtonApply={step === 'review' ? 'Create' : 'Review'}
        onClickButtonCancel={() => {
          if (step === 'review') {
            setStep('form');
            return;
          }
          onClose();
        }}
        onClickButtonApply={() => {
          if (step === 'form') {
            setStep('review');
            return;
          }

          mutation.mutate(
            {tenantId, displayName, description},
            {
              onSuccess: (account) => {
                appToaster.add({
                  name: `sa-created-${account.serviceAccountId}`,
                  title: 'Service account created',
                  content: `${account.name} for ${account.tenantName}`,
                  theme: 'success',
                });
                onClose();
              },
            },
          );
        }}
        propsButtonApply={{
          disabled: step === 'form' ? !canContinue : mutation.isPending,
          view: 'action',
        }}
        loading={mutation.isPending}
      />
    </Dialog>
  );
}

export function RotateSecretDialog({
  open,
  clientId,
  clientName,
  onClose,
}: {
  open: boolean;
  clientId: string;
  clientName: string;
  onClose: () => void;
}) {
  const [note, setNote] = React.useState('Routine rotation');
  const mutation = useRotateSecretMutation();

  React.useEffect(() => {
    if (open) {
      setNote('Routine rotation');
    }
  }, [open]);

  return (
    <Dialog
      open={open}
      onClose={closeDialog(onClose)}
      onOpenChange={(nextOpen) => syncOverlayOpen(nextOpen, onClose)}
      size="s"
    >
      <Dialog.Header caption="Rotate OAuth Secret" />
      <Dialog.Body>
        <HighlightAlert
          title={clientName}
          message="Current secret stays visible only once. Capture the new value in your secure vault during the rotation window."
          theme="warning"
        />
        <div className="form-grid">
          <div className="form-grid__full">
            <Text variant="body-2">Rotation note</Text>
            <TextArea rows={4} value={note} onUpdate={setNote} placeholder="Routine rotation for production client" />
          </div>
        </div>
      </Dialog.Body>
      <Dialog.Footer
        textButtonCancel="Cancel"
        textButtonApply="Rotate"
        onClickButtonCancel={onClose}
        onClickButtonApply={() => {
          mutation.mutate(
            {clientId, note},
            {
              onSuccess: () => {
                appToaster.add({
                  name: `rotate-${clientId}`,
                  title: 'Secret rotated',
                  content: `OAuth client ${clientName} updated`,
                  theme: 'success',
                });
                onClose();
              },
            },
          );
        }}
        propsButtonApply={{view: 'action', disabled: note.trim().length < 3}}
        loading={mutation.isPending}
      />
    </Dialog>
  );
}

export function GrantAccessDrawer({
  open,
  onClose,
  tenantId,
  resourceType,
  resourceId,
}: {
  open: boolean;
  onClose: () => void;
  tenantId: string;
  resourceType: ResourceType;
  resourceId: string;
}) {
  const usersQuery = useUsersQuery();
  const groupsQuery = useGroupsQuery();
  const serviceAccountsQuery = useServiceAccountsQuery();
  const rolesQuery = useRolesQuery();
  const mutation = useGrantAccessMutation();
  const [subjectType, setSubjectType] = React.useState<'userAccount' | 'group' | 'serviceAccount'>('userAccount');
  const [subjectId, setSubjectId] = React.useState('');
  const [roleId, setRoleId] = React.useState('');
  const [expiresAt, setExpiresAt] = React.useState<string>('');

  React.useEffect(() => {
    if (open) {
      setSubjectType('userAccount');
      setSubjectId('');
      setRoleId('');
      setExpiresAt('');
    }
  }, [open]);

  const userOptions = (usersQuery.data?.items ?? [])
    .filter((user) => user.tenantIds.includes(tenantId))
    .map((user) => ({value: user.userId, content: `${user.name} (${user.userId})`}));
  const groupOptions = (groupsQuery.data?.items ?? [])
    .filter((group) => group.tenantId === tenantId)
    .map((group) => ({value: group.groupId, content: `${group.name} (${group.groupId})`}));
  const serviceAccountOptions = (serviceAccountsQuery.data?.items ?? [])
    .filter((account) => account.tenantId === tenantId)
    .map((account) => ({
      value: account.serviceAccountId,
      content: `${account.name} (${account.serviceAccountId})`,
    }));
  const roleOptions = (rolesQuery.data?.items ?? []).map((role) => ({
    value: role.roleId,
    content: `${role.name} (${role.roleId})`,
  }));

  const subjectOptions =
    subjectType === 'group'
      ? groupOptions
      : subjectType === 'serviceAccount'
        ? serviceAccountOptions
        : userOptions;

  const subjectLabel =
    subjectOptions.find((option) => option.value === subjectId)?.content ?? subjectId;

  return (
    <Drawer open={open} onOpenChange={(nextOpen) => syncOverlayOpen(nextOpen, onClose)} size={520}>
      <div className="drawer-panel">
        <div className="drawer-panel__header">
          <Text variant="subheader-3">Grant Access</Text>
          <Text color="secondary">
            {resourceType}/{resourceId}
          </Text>
        </div>
        <div className="drawer-panel__body">
          <div className="form-grid">
            <Select
              label="Subject type"
              value={[subjectType]}
              options={[
                {value: 'userAccount', content: 'User'},
                {value: 'group', content: 'Group'},
                {value: 'serviceAccount', content: 'Service account'},
              ]}
              onUpdate={(value) => {
                const next = (value[0] ?? 'userAccount') as 'userAccount' | 'group' | 'serviceAccount';
                setSubjectType(next);
                setSubjectId('');
              }}
            />
            <Select
              label="Subject"
              value={subjectId ? [subjectId] : []}
              options={subjectOptions}
              placeholder="Choose subject"
              onUpdate={(value) => setSubjectId(value[0] ?? '')}
            />
            <Select
              label="Role"
              value={roleId ? [roleId] : []}
              options={roleOptions}
              placeholder="Choose role"
              onUpdate={(value) => setRoleId(value[0] ?? '')}
            />
            <TextInput
              label="Expires at"
              value={expiresAt}
              onUpdate={setExpiresAt}
              placeholder="Optional ISO timestamp"
            />
          </div>
          <SectionCard title="Binding review" description="Request will be sent through the authz facade.">
            <DynamicView
              value={{
                subject: subjectLabel,
                roleId,
                resourceId,
                expiresAt: expiresAt || 'No expiration',
              }}
              spec={grantReviewSpec}
              config={dynamicViewConfig}
            />
          </SectionCard>
        </div>
        <div className="drawer-panel__footer">
          <Button view="flat" onClick={onClose}>
            Cancel
          </Button>
          <Button
            view="action"
            loading={mutation.isPending}
            disabled={!subjectId || !roleId}
            onClick={() => {
              mutation.mutate(
                {
                  tenantId,
                  resourceType,
                  resourceId,
                  subjectType,
                  subjectId,
                  subjectName: String(subjectLabel),
                  roleId,
                  source: 'direct',
                  expiresAt: expiresAt || undefined,
                },
                {
                  onSuccess: () => {
                    appToaster.add({
                      name: `binding-${resourceId}-${subjectId}`,
                      title: 'Access granted',
                      content: `${subjectLabel} -> ${roleId}`,
                      theme: 'success',
                    });
                    onClose();
                  },
                },
              );
            }}
          >
            Grant Access
          </Button>
        </div>
      </div>
    </Drawer>
  );
}

export function SupportGrantWizard({
  open,
  defaultTenantId,
  onClose,
}: {
  open: boolean;
  defaultTenantId?: string;
  onClose: () => void;
}) {
  const tenantsQuery = useTenantsQuery();
  const usersQuery = useUsersQuery();
  const rolesQuery = useRolesQuery();
  const mutation = useCreateSupportGrantMutation();
  const [step, setStep] = React.useState<'form' | 'review'>('form');
  const [tenantId, setTenantId] = React.useState(defaultTenantId ?? '');
  const [subjectId, setSubjectId] = React.useState('');
  const [roleId, setRoleId] = React.useState('support-operator');
  const [incidentId, setIncidentId] = React.useState('INC-');
  const [reason, setReason] = React.useState('');
  const [expiresAt, setExpiresAt] = React.useState<ReturnType<typeof dateTime> | null>(
    createDefaultSupportGrantExpiry,
  );

  React.useEffect(() => {
    if (open) {
      setStep('form');
      setTenantId(defaultTenantId ?? '');
      setSubjectId('');
      setRoleId('support-operator');
      setIncidentId('INC-');
      setReason('');
      setExpiresAt(createDefaultSupportGrantExpiry());
    }
  }, [defaultTenantId, open]);

  const tenantOptions = (tenantsQuery.data?.items ?? []).map((tenant) => ({
    value: tenant.tenantId,
    content: `${tenant.name} (${tenant.tenantId})`,
  }));
  const subjectOptions = (usersQuery.data?.items ?? [])
    .filter((user) => user.tenantIds.includes(tenantId))
    .map((user) => ({value: user.userId, content: `${user.name} (${user.userId})`}));
  const roleOptions = (rolesQuery.data?.items ?? []).map((role) => ({
    value: role.roleId,
    content: `${role.name} (${role.roleId})`,
  }));

  return (
    <Dialog
      open={open}
      onClose={closeDialog(onClose)}
      onOpenChange={(nextOpen) => syncOverlayOpen(nextOpen, onClose)}
      size="l"
    >
      <Dialog.Header caption="Grant Temporary Support Access" />
      <Dialog.Body>
        <div className="wizard-stepper">
          <span className={`wizard-stepper__step${step === 'form' ? ' wizard-stepper__step_active' : ''}`}>1. Scope</span>
          <span className={`wizard-stepper__step${step === 'review' ? ' wizard-stepper__step_active' : ''}`}>2. Review</span>
        </div>
        {step === 'form' ? (
          <div className="form-grid">
            <Select
              label="Tenant"
              value={tenantId ? [tenantId] : []}
              options={tenantOptions}
              onUpdate={(value) => setTenantId(value[0] ?? '')}
            />
            <Select
              label="Support subject"
              value={subjectId ? [subjectId] : []}
              options={subjectOptions}
              onUpdate={(value) => setSubjectId(value[0] ?? '')}
            />
            <Select
              label="Role"
              value={roleId ? [roleId] : []}
              options={roleOptions}
              onUpdate={(value) => setRoleId(value[0] ?? '')}
            />
            <TextInput
              label="Incident ID"
              value={incidentId}
              onUpdate={setIncidentId}
              placeholder="INC-1901"
            />
            <div className="form-grid__full">
              <Text variant="body-2">Expiration</Text>
              <DatePicker value={expiresAt} onUpdate={setExpiresAt} />
              <FieldHint>Local timezone value: {formatDateTime(expiresAt?.toISOString())}</FieldHint>
            </div>
            <div className="form-grid__full">
              <Text variant="body-2">Reason</Text>
              <TextArea rows={4} value={reason} onUpdate={setReason} placeholder="Customer reports billing issue and needs temporary guided investigation." />
            </div>
          </div>
        ) : (
          <div className="review-panel">
            <DynamicView
              value={{
                tenantId,
                subjectId,
                roleId,
                incidentId,
                expiresAt: expiresAt?.toISOString() ?? '',
                reason,
              }}
              spec={supportGrantReviewSpec}
              config={dynamicViewConfig}
            />
          </div>
        )}
      </Dialog.Body>
      <Dialog.Footer
        textButtonCancel={step === 'review' ? 'Back' : 'Close'}
        textButtonApply={step === 'review' ? 'Grant Access' : 'Review'}
        onClickButtonCancel={() => {
          if (step === 'review') {
            setStep('form');
            return;
          }
          onClose();
        }}
        onClickButtonApply={() => {
          if (step === 'form') {
            setStep('review');
            return;
          }

          mutation.mutate(
            {
              tenantId,
              subjectId,
              roleId,
              incidentId,
              reason,
              expiresAt: expiresAt?.toISOString() ?? new Date().toISOString(),
            },
            {
              onSuccess: (grant) => {
                appToaster.add({
                  name: `support-${grant.grantId}`,
                  title: 'Support grant submitted',
                  content: `${grant.subjectName} / ${grant.tenantName}`,
                  theme: 'success',
                });
                onClose();
              },
            },
          );
        }}
        propsButtonApply={{
          view: 'action',
          disabled:
            step === 'form'
              ? !(tenantId && subjectId && roleId && incidentId.trim() && reason.trim())
              : mutation.isPending,
        }}
        loading={mutation.isPending}
      />
    </Dialog>
  );
}
