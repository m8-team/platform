# @m8/runtime

React/Next/json-render orchestration for the M8 platform UI. It composes the
module registry, TanStack Query, safe operation execution, runtime context,
Gravity UI theming and component registry extensions. Business modules must not
depend on this package.

`createM8Runtime` derives its query registry, operation registry, Next app spec,
navigation model and route access/query bridge from one `ModuleRegistry`.
Route query state is exposed to json-render under `/queries/{binding}`.
