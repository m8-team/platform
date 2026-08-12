# UI orchestration packages

```text
packages/
├── core/       module identity, enablement, dependencies and ownership
├── query/      declarative remote data and TanStack Query integration
├── operation/  safe registered mutation execution
└── runtime/    thin json-render/query/operation wiring
```

```text
              @m8/core
             ▲        ▲
            /          \
     @m8/query      @m8/operation
            \          /
             \        /
              @m8/runtime
```

Module packages normally depend on `@m8/core`, plus `@m8/query` and
`@m8/operation` when they contribute remote data or mutations. They never
depend on `@m8/runtime`. Native json-render `Spec` is the only component-tree
language and state engine. `@json-render/next` remains responsible for routes,
layouts, metadata, matching, navigation and SSR.
