import type {z} from 'zod';

interface QueryContext<TInput> {
  input: TInput;
  signal: AbortSignal;
}

export interface QueryDefinition<
  TInputSchema extends z.ZodType,
  TOutputSchema extends z.ZodType,
> {
  id: string;
  input: TInputSchema;
  output: TOutputSchema;
  queryKey: (input: z.infer<TInputSchema>) => readonly unknown[];
  execute: (
    context: QueryContext<z.infer<TInputSchema>>,
  ) => Promise<z.infer<TOutputSchema>>;
}

export function defineQuery<
  TInput extends z.ZodType,
  TOutput extends z.ZodType,
>(definition: QueryDefinition<TInput, TOutput>) {
  return definition;
}
