import type {NextAppSpec} from '@json-render/next';
import type {ModuleRegistry, M8ModuleDefinition} from '@m8/core';

export function buildNextAppSpec(
  modules: ModuleRegistry<readonly M8ModuleDefinition[]>,
  options: {
    metadata?: NextAppSpec['metadata'];
    layouts?: NextAppSpec['layouts'];
    state?: NextAppSpec['state'];
  } = {},
): NextAppSpec {
  return modules.buildNextAppSpec(options);
}
