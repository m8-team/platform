import * as React from 'react';

import {env} from '@/shared/config/env';
import type {AppContextSelection} from '@/shared/types/iam';

export type AppUIContextValue = {
  context: AppContextSelection;
  globalSearch: string;
  navCompact: boolean;
  setContext: React.Dispatch<React.SetStateAction<AppContextSelection>>;
  setGlobalSearch: React.Dispatch<React.SetStateAction<string>>;
  setNavCompact: React.Dispatch<React.SetStateAction<boolean>>;
};

export const AppUIContext = React.createContext<AppUIContextValue | null>(null);

export const defaultAppContext: AppContextSelection = {
  tenantId: env.defaultTenantId,
  organizationId: 'org-1',
  environment: 'prod',
  region: 'eu-central',
};

export function useAppUI() {
  const context = React.useContext(AppUIContext);

  if (!context) {
    throw new Error('useAppUI must be used within AppProviders');
  }

  return context;
}
