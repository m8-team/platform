import {Button, Icon, Select} from '@gravity-ui/uikit'
import {ActionBar} from '@gravity-ui/navigation'
import {ArrowRotateRight, Clock, Plus} from '@gravity-ui/icons'

export interface ConsoleActionBarOption {
  value: string
  content: string
}

interface ConsoleActionBarProps {
  organization: string
  workspace: string
  projectId: string
  organizationOptions: ConsoleActionBarOption[]
  workspaceOptions: ConsoleActionBarOption[]
  projectOptions: ConsoleActionBarOption[]
  onOrganizationUpdate: (value: string[]) => void
  onWorkspaceUpdate: (value: string[]) => void
  onProjectUpdate: (value: string[]) => void
}

export function ConsoleActionBar({
  organization,
  workspace,
  projectId,
  organizationOptions,
  workspaceOptions,
  projectOptions,
  onOrganizationUpdate,
  onWorkspaceUpdate,
  onProjectUpdate,
}: ConsoleActionBarProps) {
  return (
    <ActionBar aria-label="M8 Platform action bar" className="m8-actionbar">
      <ActionBar.Section>
        <ActionBar.Group>
          <ActionBar.Item>
            <Switcher
              label="Org"
              value={[organization]}
              options={organizationOptions}
              onUpdate={onOrganizationUpdate}
            />
          </ActionBar.Item>
          <ActionBar.Item>
            <Switcher
              label="Workspace"
              value={[workspace]}
              options={workspaceOptions}
              onUpdate={onWorkspaceUpdate}
            />
          </ActionBar.Item>
          <ActionBar.Item>
            <Switcher
              label="Project"
              value={[projectId]}
              options={projectOptions}
              onUpdate={onProjectUpdate}
            />
          </ActionBar.Item>
        </ActionBar.Group>
        <ActionBar.Group pull="right">
          <ActionBar.Item>
            <Button view="normal">
              <Icon data={ArrowRotateRight} size={14} />
              Refresh
            </Button>
          </ActionBar.Item>
          <ActionBar.Item>
            <Button view="outlined">
              <Icon data={Clock} size={14} />
              Open operation
            </Button>
          </ActionBar.Item>
          <ActionBar.Item>
            <Button view="action">
              <Icon data={Plus} size={14} />
              New project
            </Button>
          </ActionBar.Item>
        </ActionBar.Group>
      </ActionBar.Section>
    </ActionBar>
  )
}

interface SwitcherProps {
  label: string
  value: string[]
  options: ConsoleActionBarOption[]
  onUpdate: (value: string[]) => void
}

function Switcher({label, value, options, onUpdate}: SwitcherProps) {
  return (
    <div className="m8-field m8-switcher">
      <Select aria-label={label} value={value} options={options} width="max" onUpdate={onUpdate} />
    </div>
  )
}
