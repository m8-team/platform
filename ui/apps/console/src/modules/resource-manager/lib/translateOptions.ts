import type {Translate, TranslationKey} from '../../../i18n'

export function translateOptions<T extends string>(
  options: Array<{value: T; content?: string; titleKey?: TranslationKey}>,
  t: Translate,
) {
  return options.map((option) => ({
    value: option.value,
    content: option.titleKey ? t(option.titleKey) : option.content ?? option.value,
  }))
}
