import type {z} from 'zod';

import type {OperationDefinition} from './definition';
import {DuplicateOperationError, UnknownOperationError} from './errors';

export type RegisteredOperationDefinition = OperationDefinition<z.ZodType, z.ZodType>;

export class OperationRegistry {
  private readonly definitions = new Map<string, RegisteredOperationDefinition>();

  constructor(definitions: readonly RegisteredOperationDefinition[] = []) {
    for (const definition of definitions) this.register(definition);
  }

  register<const TDefinition extends RegisteredOperationDefinition>(definition: TDefinition): TDefinition {
    if (this.definitions.has(definition.id)) throw new DuplicateOperationError(definition.id);
    this.definitions.set(definition.id, definition);
    return definition;
  }

  get(operationId: string): RegisteredOperationDefinition | undefined {
    return this.definitions.get(operationId);
  }

  require(operationId: string): RegisteredOperationDefinition {
    const definition = this.get(operationId);
    if (!definition) throw new UnknownOperationError(operationId);
    return definition;
  }
}
