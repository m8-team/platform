# @m8/runtime

React/Next/json-render orchestration. It composes the
module registry, TanStack Query, safe operation execution, runtime context,
Gravity UI theming and component registry extensions. Business modules must not
depend on this package.

`createRuntime` derives query and operation registries plus navigation from a
`ModuleRegistry`; application specs remain a build/server concern through
`buildNextAppSpec`. Route query state is projected under
`/__runtime/queries/{binding}`.

## Public API

- `createRuntime`, `Runtime`, `RuntimeProvider`, `useRuntime`
- `buildNextAppSpec`
- `createRuntimeState`, `withRuntimeState`
- `createActionHandlers` (only `executeOperation`)
- module enablement and component-registry composition helpers
