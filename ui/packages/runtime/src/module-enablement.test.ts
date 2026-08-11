import {describe, expect, it} from 'vitest';
import type {M8ModuleDefinition} from '@m8/core';

import {selectEnabledModules} from './module-enablement';

const page = {root: 'root', elements: {root: {type: 'Text', props: {}, children: []}}};
const modules: readonly M8ModuleDefinition[] = [
  {id: 'base', title: 'Base', basePath: '/base', routes: {'/': {page}}},
  {
    id: 'feature', title: 'Feature', basePath: '/feature', routes: {'/': {page}},
    dependencies: {required: ['base']}, availability: {feature: 'feature'},
  },
];

describe('selectEnabledModules', () => {
  it('filters by availability and dependencies', () => {
    expect(selectEnabledModules(modules, {
      context: {features: ['feature'], permissions: ['read']},
    }).map(module => module.id)).toEqual(['base', 'feature']);

    expect(selectEnabledModules(modules, {
      context: {features: ['feature'], permissions: ['read']},
      enabledModuleIds: ['feature'],
    })).toEqual([]);
  });
});
