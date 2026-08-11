import type {RuntimeContext} from '@m8/core';
import type {z} from 'zod';

export interface OperationExecutionContext<TInput> {
  readonly input: TInput;
  readonly signal: AbortSignal;
  readonly context: RuntimeContext;
}

export interface OperationDefinition<
  TInputSchema extends z.ZodType = z.ZodType,
  TOutputSchema extends z.ZodType = z.ZodType,
> {
  readonly id: string;
  readonly mode?: 'immediate' | 'long-running';
  readonly input: TInputSchema;
  readonly output: TOutputSchema;
  readonly requiredPermission?: string;
  readonly invalidate?: readonly string[];
  readonly execute: (
    context: OperationExecutionContext<z.output<TInputSchema>>,
  ) => Promise<z.input<TOutputSchema>>;
}

export function defineOperation<
  const TInput extends z.ZodType,
  const TOutput extends z.ZodType,
>(definition: OperationDefinition<TInput, TOutput>): OperationDefinition<TInput, TOutput> {
  return Object.freeze(definition);
}

export type OperationInput<TDefinition extends OperationDefinition> =
  z.input<TDefinition['input']>;
export type OperationOutput<TDefinition extends OperationDefinition> =
  z.output<TDefinition['output']>;
