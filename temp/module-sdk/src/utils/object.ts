export function compact<TValue>(items: Array<TValue | undefined | null | false>): TValue[] {
  return items.filter(Boolean) as TValue[];
}

export function unique<TValue>(items: TValue[]): TValue[] {
  return Array.from(new Set(items));
}
