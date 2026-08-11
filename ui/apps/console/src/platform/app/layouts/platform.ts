import type {NextAppSpec} from '@json-render/next';

export const platformLayout: NonNullable<NextAppSpec['layouts']>[string] = {
  root: 'layout',
  elements: {
    layout: {
      type: 'Page',
      props: {width: 'wide'},
      children: ['slot'],
    },
    slot: {
      type: 'Slot',
      props: {},
      children: [],
    },
  },
};
