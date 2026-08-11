import type {z} from 'zod';

import type {M8OperationDefinition} from './definition';
import {DuplicateOperationError, UnknownOperationError} from './errors';

export type OperationDefinition = M8OperationDefinition<z.ZodType, z.ZodType>;

export class OperationRegistry {
  private readonly definitions = new Map<string, OperationDefinition>();

  constructor(definitions: readonly OperationDefinition[] = []) {
    for (const definition of definitions) this.register(definition);
  }

  register<const TDefinition extends OperationDefinition>(definition: TDefinition): TDefinition {
    if (this.definitions.has(definition.id)) throw new DuplicateOperationError(definition.id);
    this.definitions.set(definition.id, definition);
    return definition;
  }

  get(operationId: string): OperationDefinition | undefined {
    return this.definitions.get(operationId);
  }

  require(operationId: string): OperationDefinition {
    const definition = this.get(operationId);
    if (!definition) throw new UnknownOperationError(operationId);
    return definition;
  }
}
