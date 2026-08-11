import type {z} from 'zod';

export interface M8QueryExecutionContext<TInput> {
  readonly input: TInput;
  readonly signal: AbortSignal;
}

export interface M8QueryDefinition<
  TInputSchema extends z.ZodType = z.ZodType,
  TOutputSchema extends z.ZodType = z.ZodType,
> {
  readonly id: string;
  readonly input: TInputSchema;
  readonly output: TOutputSchema;
  readonly queryKey: (input: z.output<TInputSchema>) => readonly unknown[];
  readonly execute: (
    context: M8QueryExecutionContext<z.output<TInputSchema>>,
  ) => Promise<z.input<TOutputSchema>>;
}

export function defineQuery<
  const TInput extends z.ZodType,
  const TOutput extends z.ZodType,
>(definition: M8QueryDefinition<TInput, TOutput>): M8QueryDefinition<TInput, TOutput> {
  return Object.freeze(definition);
}

export type M8QueryInput<TDefinition extends M8QueryDefinition> =
  z.input<TDefinition['input']>;
export type M8QueryOutput<TDefinition extends M8QueryDefinition> =
  z.output<TDefinition['output']>;
