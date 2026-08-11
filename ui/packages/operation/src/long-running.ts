export type LongRunningOperationStatus =
  | 'PENDING'
  | 'RUNNING'
  | 'SUCCEEDED'
  | 'FAILED'
  | 'CANCELLED';

export interface LongRunningOperation<TProgress = unknown, TResult = unknown> {
  readonly id: string;
  readonly status: LongRunningOperationStatus;
  readonly progress?: TProgress;
  readonly result?: TResult;
  readonly error?: Readonly<{code?: string; message: string}>;
}

export interface LongRunningOperationAdapter {
  wait(operationId: string, options?: {signal?: AbortSignal}): Promise<LongRunningOperation>;
  cancel?(operationId: string, signal: AbortSignal): Promise<void>;
}
