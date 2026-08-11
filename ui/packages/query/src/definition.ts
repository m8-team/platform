import type {z} from 'zod';
import type {RuntimeContext} from '@m8/core';

export interface QueryExecutionContext<TInput> {
  readonly input: TInput;
  readonly signal: AbortSignal;
  readonly context: RuntimeContext;
}

export interface QueryDefinition<
  TInputSchema extends z.ZodType = z.ZodType,
  TOutputSchema extends z.ZodType = z.ZodType,
> {
  readonly id: string;
  readonly input: TInputSchema;
  readonly output: TOutputSchema;
  readonly queryKey?: (
    input: z.output<TInputSchema>,
    context: RuntimeContext,
  ) => readonly unknown[];
  readonly execute: (
    context: QueryExecutionContext<z.output<TInputSchema>>,
  ) => Promise<z.input<TOutputSchema>>;
}

export function defineQuery<
  const TInput extends z.ZodType,
  const TOutput extends z.ZodType,
>(definition: QueryDefinition<TInput, TOutput>): QueryDefinition<TInput, TOutput> {
  return Object.freeze(definition);
}

export type QueryInput<TDefinition extends QueryDefinition> =
  z.input<TDefinition['input']>;
export type QueryOutput<TDefinition extends QueryDefinition> =
  z.output<TDefinition['output']>;
