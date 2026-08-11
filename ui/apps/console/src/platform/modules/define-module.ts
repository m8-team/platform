import {ModuleDefinition} from "@/platform/modules/types";
import {moduleSchema} from "@/platform/modules/module-schema";

export function defineModule<const TModule extends ModuleDefinition>(module: TModule): TModule {
  moduleSchema.parse(module);
  
  return module;
}
