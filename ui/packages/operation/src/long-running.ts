export type M8LongRunningOperationStatus =
  | 'PENDING'
  | 'RUNNING'
  | 'SUCCEEDED'
  | 'FAILED'
  | 'CANCELLED';

export interface M8LongRunningOperation<TProgress = unknown, TResult = unknown> {
  readonly id: string;
  readonly status: M8LongRunningOperationStatus;
  readonly progress?: TProgress;
  readonly result?: TResult;
  readonly error?: Readonly<{code?: string; message: string}>;
}

export interface LongRunningOperationAdapter {
  get(operationId: string, signal: AbortSignal): Promise<M8LongRunningOperation>;
  cancel?(operationId: string, signal: AbortSignal): Promise<void>;
}
