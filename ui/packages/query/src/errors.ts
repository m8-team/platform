export class QueryRegistryError extends Error {
  override readonly name: string = 'QueryRegistryError';
}

export class DuplicateQueryError extends QueryRegistryError {
  override readonly name = 'DuplicateQueryError';
  constructor(readonly queryId: string) {
    super(`Duplicate query id: "${queryId}".`);
  }
}

export class UnknownQueryError extends QueryRegistryError {
  override readonly name = 'UnknownQueryError';
  constructor(readonly queryId: string) {
    super(`Unknown query id: "${queryId}".`);
  }
}
