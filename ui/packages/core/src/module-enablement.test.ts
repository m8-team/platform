import {describe, expect, it} from 'vitest';

import {selectEnabledModules} from './module-enablement';
import type {ModuleDefinition} from './types';

const installedModules: readonly ModuleDefinition[] = [
  {id: 'base', title: 'Base'},
  {
    id: 'feature',
    title: 'Feature',
    dependencies: {required: ['base']},
  },
];

describe('selectEnabledModules', () => {
  it('enables all installed modules by default', () => {
    expect(selectEnabledModules(installedModules).map(module => module.id))
      .toEqual(['base', 'feature']);
  });

  it('removes enabled modules whose required dependency is disabled', () => {
    expect(selectEnabledModules(installedModules, {
      enabledModuleIds: ['feature'],
    })).toEqual([]);
  });

  it('keeps explicitly enabled modules with their dependency', () => {
    expect(selectEnabledModules(installedModules, {
      enabledModuleIds: ['feature', 'base'],
    }).map(module => module.id)).toEqual(['base', 'feature']);
  });
});
