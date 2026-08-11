import type {z} from 'zod';
import type {M8RuntimeContext} from '@m8/core';

import type {M8QueryDefinition} from './definition';
import {DuplicateQueryError, UnknownQueryError} from './errors';

type QueryDefinition = M8QueryDefinition<z.ZodType, z.ZodType>;

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

export class QueryRegistry {
  private readonly definitions = new Map<string, QueryDefinition>();

  constructor(definitions: readonly QueryDefinition[] = []) {
    for (const definition of definitions) this.register(definition);
  }

  register<const TDefinition extends QueryDefinition>(definition: TDefinition): TDefinition {
    if (this.definitions.has(definition.id)) throw new DuplicateQueryError(definition.id);
    this.definitions.set(definition.id, definition);
    return definition;
  }

  has(queryId: string): boolean {
    return this.definitions.has(queryId);
  }

  get(queryId: string): QueryDefinition | undefined {
    return this.definitions.get(queryId);
  }

  require(queryId: string): QueryDefinition {
    const definition = this.get(queryId);
    if (!definition) throw new UnknownQueryError(queryId);
    return definition;
  }

  parseInput<TDefinition extends QueryDefinition>(
    definition: TDefinition,
    input: unknown,
  ): z.output<TDefinition['input']> {
    return definition.input.parse(normalizeQueryInput(definition.input, input)) as z.output<TDefinition['input']>;
  }

  queryKey(queryId: string, input: unknown): readonly unknown[] {
    const definition = this.require(queryId);
    return [definition.id, ...definition.queryKey(this.parseInput(definition, input))];
  }

  async execute(queryId: string, input: unknown, signal: AbortSignal, context: M8RuntimeContext = {}): Promise<unknown> {
    const definition = this.require(queryId);
    const parsedInput = this.parseInput(definition, input);
    const output = await definition.execute({input: parsedInput, signal, context});
    return definition.output.parse(output);
  }
}

export class M8QueryRuntime {
  constructor(readonly registry: QueryRegistry) {}

  queryKey(queryId: string, input: unknown): readonly unknown[] {
    return this.registry.queryKey(queryId, input);
  }

  execute(queryId: string, input: unknown, signal: AbortSignal, context: M8RuntimeContext = {}): Promise<unknown> {
    return this.registry.execute(queryId, input, signal, context);
  }
}
