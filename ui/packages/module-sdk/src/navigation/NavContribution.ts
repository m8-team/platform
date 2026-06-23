import type {ComponentType} from 'react';
import type {
  M8FeatureFlag,
  M8Metadata,
  M8MountPointId,
  M8Permission,
} from '../primitives';
import type {M8ComponentIconProps} from '../primitives';

export type M8NavigationContribution = {
  id: string;
  parentId?: string;
  title: string;
  description?: string;
  to: string;
  mountPointId?: M8MountPointId;
  icon?: ComponentType<M8ComponentIconProps>;
  order?: number;
  requiredPermissions?: M8Permission[];
  requiredFeatureFlags?: M8FeatureFlag[];
  badge?: M8NavigationBadge;
  exact?: boolean;
  children?: M8NavigationContribution[];
  metadata?: M8Metadata;
};

export type M8NavigationBadge = {
  text: string;
  tone?: string;
};
