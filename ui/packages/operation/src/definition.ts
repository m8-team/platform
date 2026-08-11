import type {M8RuntimeContext} from '@m8/core';
import type {z} from 'zod';

export interface M8OperationConfirmation {
  readonly title: string;
  readonly description?: string;
  readonly confirmLabel?: string;
}

export interface M8OperationExecutionContext<TInput> {
  readonly input: TInput;
  readonly signal: AbortSignal;
  readonly context: M8RuntimeContext;
}

export interface M8OperationDefinition<
  TInputSchema extends z.ZodType = z.ZodType,
  TOutputSchema extends z.ZodType = z.ZodType,
> {
  readonly id: string;
  readonly mode?: 'immediate' | 'long-running';
  readonly input: TInputSchema;
  readonly output: TOutputSchema;
  readonly requiredPermission?: string;
  readonly destructive?: boolean;
  readonly confirmation?: M8OperationConfirmation;
  readonly invalidate?: readonly string[];
  readonly completion?: {readonly invalidate?: readonly string[]};
  readonly execute: (
    context: M8OperationExecutionContext<z.output<TInputSchema>>,
  ) => Promise<z.input<TOutputSchema>>;
}

export function defineOperation<
  const TInput extends z.ZodType,
  const TOutput extends z.ZodType,
>(definition: M8OperationDefinition<TInput, TOutput>): M8OperationDefinition<TInput, TOutput> {
  return Object.freeze(definition);
}

export type M8OperationInput<TDefinition extends M8OperationDefinition> =
  z.input<TDefinition['input']>;
export type M8OperationOutput<TDefinition extends M8OperationDefinition> =
  z.output<TDefinition['output']>;
