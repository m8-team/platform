export type M8Id = string;

export type M8ModuleId = string;
export type M8ScopeId = string;
export type M8MountPointId = string;
export type M8SlotId = string;
export type M8Permission = string;
export type M8Capability = string;
export type M8FeatureFlag = string;
export type M8ModuleKind = string;
export type M8ModuleLifecycle = string;

export type M8Metadata = Record<string, unknown>;

export type M8Dictionary<TValue = unknown> = Record<string, TValue>;

export type M8MaybePromise<T> = T | Promise<T>;

export type M8ComponentIconProps = {
  className?: string;
  width?: number | string;
  height?: number | string;
};
