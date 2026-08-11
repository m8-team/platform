# M8 declarative UI packages

```text
packages/
├── core/       contracts, modules, registry and NextAppSpec composition
├── query/      declarative remote data and TanStack Query integration
├── operation/  safe registered mutation execution
└── runtime/    React, Next, json-render and Gravity UI orchestration
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
language; `NextAppSpec` / `NextRouteSpec` remain responsible for Next routes,
layouts, metadata, loaders and SSR.
