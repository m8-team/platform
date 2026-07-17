import type {M8MaybePromise} from '../primitives';
import type {M8ModuleRuntimeContext} from '../runtime/ModuleRuntimeContext';

export type M8BreadcrumbContribution = {
  resolve: (ctx: M8BreadcrumbContext) => M8MaybePromise<M8BreadcrumbItem[]>;
};

export type M8BreadcrumbContext = {
  runtime: M8ModuleRuntimeContext;
  params: Record<string, string>;
  pathname: string;
};

export type M8BreadcrumbItem = {
  title: string;
  to?: string;
};
