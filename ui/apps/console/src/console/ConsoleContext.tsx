import {createContext, useContext} from 'react'

import type {AppLanguage, Translate} from '../i18n'

export interface ConsoleSelection {
  organization: string
  workspace: string
  projectId: string
  projectOptions: Array<{value: string; content: string}>
  setWorkspace: (value: string) => void
  setProjectId: (value: string) => void
}

export interface ConsoleI18n {
  language: AppLanguage
  t: Translate
}

export const ConsoleSelectionContext = createContext<ConsoleSelection | null>(null)
export const ConsoleI18nContext = createContext<ConsoleI18n | null>(null)

export function useConsoleSelection() {
  const selection = useContext(ConsoleSelectionContext)

  if (!selection) {
    throw new Error('Console selection context is not available')
  }

  return selection
}

export function useConsoleI18n() {
  const i18n = useContext(ConsoleI18nContext)

  if (!i18n) {
    throw new Error('Console i18n context is not available')
  }

  return i18n
}
