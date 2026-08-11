import {describe, expect, it} from 'vitest';
import type {ModuleDefinition} from '@m8/core';

import {selectEnabledModules} from './module-enablement';

const page = {root: 'root', elements: {root: {type: 'Text', props: {}, children: []}}};
const modules: readonly ModuleDefinition[] = [
  {id: 'base', title: 'Base', routes: {'/base': {page}}},
  {
    id: 'feature', title: 'Feature', routes: {'/feature': {page}},
    dependencies: {required: ['base']},
  },
];

describe('selectEnabledModules', () => {
  it('filters by installed module IDs and dependencies', () => {
    expect(selectEnabledModules(modules, {
    }).map(module => module.id)).toEqual(['base', 'feature']);

    expect(selectEnabledModules(modules, {
      enabledModuleIds: ['feature'],
    })).toEqual([]);
  });
});
