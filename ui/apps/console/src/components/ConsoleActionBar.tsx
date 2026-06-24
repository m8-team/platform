import {Button, Icon, Select} from '@gravity-ui/uikit'
import {ActionBar} from '@gravity-ui/navigation'
import {ArrowRotateRight, Clock, Globe, Plus} from '@gravity-ui/icons'

export interface ConsoleActionBarOption {
  value: string
  content: string
}

interface ConsoleActionBarProps {
  language: string
  organization: string
  workspace: string
  projectId: string
  languageOptions: ConsoleActionBarOption[]
  organizationOptions: ConsoleActionBarOption[]
  workspaceOptions: ConsoleActionBarOption[]
  projectOptions: ConsoleActionBarOption[]
  labels: {
    organization: string
    workspace: string
    project: string
    language: string
    refresh: string
    openOperation: string
    newProject: string
  }
  onLanguageUpdate: (value: string[]) => void
  onOrganizationUpdate: (value: string[]) => void
  onWorkspaceUpdate: (value: string[]) => void
  onProjectUpdate: (value: string[]) => void
}

export function ConsoleActionBar({
  language,
  organization,
  workspace,
  projectId,
  languageOptions,
  organizationOptions,
  workspaceOptions,
  projectOptions,
  labels,
  onLanguageUpdate,
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
              label={labels.organization}
              value={[organization]}
              options={organizationOptions}
              onUpdate={onOrganizationUpdate}
            />
          </ActionBar.Item>
          <ActionBar.Item>
            <Switcher
              label={labels.workspace}
              value={[workspace]}
              options={workspaceOptions}
              onUpdate={onWorkspaceUpdate}
            />
          </ActionBar.Item>
          <ActionBar.Item>
            <Switcher
              label={labels.project}
              value={[projectId]}
              options={projectOptions}
              onUpdate={onProjectUpdate}
            />
          </ActionBar.Item>
        </ActionBar.Group>
        <ActionBar.Group pull="right">
          <ActionBar.Item>
            <div className="m8-language-switcher">
              <Icon data={Globe} size={14} />
              <Switcher
                label={labels.language}
                value={[language]}
                options={languageOptions}
                onUpdate={onLanguageUpdate}
              />
            </div>
          </ActionBar.Item>
          <ActionBar.Item>
            <Button view="normal">
              <Icon data={ArrowRotateRight} size={14} />
              {labels.refresh}
            </Button>
          </ActionBar.Item>
          <ActionBar.Item>
            <Button view="outlined">
              <Icon data={Clock} size={14} />
              {labels.openOperation}
            </Button>
          </ActionBar.Item>
          <ActionBar.Item>
            <Button view="action">
              <Icon data={Plus} size={14} />
              {labels.newProject}
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
