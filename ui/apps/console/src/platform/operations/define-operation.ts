import type {z} from 'zod';

interface OperationContext<TInput> {
  input: TInput;
  signal: AbortSignal;
}

export interface OperationDefinition<
  TInput extends z.ZodType,
  TOutput extends z.ZodType,
> {
  id: string;
  input: TInput;
  output: TOutput;
  requiredPermission?: string;
  destructive?: boolean;
  confirmation?: {
    title: string;
    description?: string;
    confirmLabel?: string;
  };
  invalidate?: string[];
  execute: (
    context: OperationContext<z.infer<TInput>>,
  ) => Promise<z.infer<TOutput>>;
}

export function defineOperation<
  TInput extends z.ZodType,
  TOutput extends z.ZodType,
>(definition: OperationDefinition<TInput, TOutput>) {
  return definition;
}
