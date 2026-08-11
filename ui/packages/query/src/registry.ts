import type {z} from 'zod';
import type {RuntimeContext} from '@m8/core';

import type {QueryDefinition} from './definition';
import {DuplicateQueryError, UnknownQueryError} from './errors';

type RegisteredQueryDefinition = QueryDefinition<z.ZodType, z.ZodType>;

function normalizeObjectInput(schema: z.ZodType, input: unknown): unknown {
  if (input === null || typeof input !== 'object' || Array.isArray(input)) return input;
  const candidate = schema as z.ZodType & {shape?: Record<string, z.ZodType>};
  if (!candidate.shape) return input;
  return Object.fromEntries(Object.entries(input).flatMap(([key, value]) => {
    if (value !== null && value !== undefined) return [[key, value]];
    const field = candidate.shape?.[key];
    if (value === null && field?.safeParse(null).success) return [[key, null]];
    return [];
  }));
}

export function normalizeQueryInput(schema: z.ZodType, input: unknown): unknown {
  return normalizeObjectInput(schema, input);
}

export function runtimeScopeKey(context: RuntimeContext): readonly unknown[] {
  return [
    context.tenantId ?? null,
    context.organizationId ?? null,
    context.workspaceId ?? null,
    context.projectId ?? null,
    context.actor?.id ?? null,
  ];
}

export class QueryRegistry {
  private readonly definitions = new Map<string, RegisteredQueryDefinition>();

  constructor(definitions: readonly RegisteredQueryDefinition[] = []) {
    for (const definition of definitions) this.register(definition);
  }

  register<const TDefinition extends RegisteredQueryDefinition>(definition: TDefinition): TDefinition {
    if (this.definitions.has(definition.id)) throw new DuplicateQueryError(definition.id);
    this.definitions.set(definition.id, definition);
    return definition;
  }

  has(queryId: string): boolean {
    return this.definitions.has(queryId);
  }

  get(queryId: string): RegisteredQueryDefinition | undefined {
    return this.definitions.get(queryId);
  }

  require(queryId: string): RegisteredQueryDefinition {
    const definition = this.get(queryId);
    if (!definition) throw new UnknownQueryError(queryId);
    return definition;
  }

  parseInput<TDefinition extends RegisteredQueryDefinition>(
    definition: TDefinition,
    input: unknown,
  ): z.output<TDefinition['input']> {
    return definition.input.parse(normalizeQueryInput(definition.input, input)) as z.output<TDefinition['input']>;
  }

  queryKey(queryId: string, input: unknown, context: RuntimeContext = {}): readonly unknown[] {
    const definition = this.require(queryId);
    const parsedInput = this.parseInput(definition, input);
    return [
      'query',
      definition.id,
      ...runtimeScopeKey(context),
      ...(definition.queryKey?.(parsedInput, context) ?? [parsedInput]),
    ];
  }

  async execute(queryId: string, input: unknown, signal: AbortSignal, context: RuntimeContext = {}): Promise<unknown> {
    const definition = this.require(queryId);
    const parsedInput = this.parseInput(definition, input);
    const output = await definition.execute({input: parsedInput, signal, context});
    return definition.output.parse(output);
  }
}

export class QueryRuntime {
  constructor(readonly registry: QueryRegistry) {}

  queryKey(queryId: string, input: unknown, context: RuntimeContext = {}): readonly unknown[] {
    return this.registry.queryKey(queryId, input, context);
  }

  execute(queryId: string, input: unknown, signal: AbortSignal, context: RuntimeContext = {}): Promise<unknown> {
    return this.registry.execute(queryId, input, signal, context);
  }
}
