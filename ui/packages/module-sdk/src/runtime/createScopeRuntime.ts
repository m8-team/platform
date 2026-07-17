import type {M8ScopeId} from '../primitives';
import type {M8ScopeRuntime} from './ModuleRuntimeContext';

export function createScopeRuntime(
  current: Record<M8ScopeId, string | undefined>,
): M8ScopeRuntime {
  return {
    current,

    get(scopeId) {
      return current[scopeId];
    },

    has(scopeId) {
      return Boolean(current[scopeId]);
    },

    require(scopeId) {
      const value = current[scopeId];

      if (!value) {
        throw new Error(`Scope value is required: ${scopeId}`);
      }

      return value;
    },
  };
}
