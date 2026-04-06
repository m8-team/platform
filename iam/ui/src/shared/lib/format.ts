import {dateTime, guessUserTimeZone} from '@gravity-ui/date-utils';

import type {EntityStatus, Severity} from '@/shared/types/iam';

export const defaultTimeZone = guessUserTimeZone() || 'Europe/Vienna';

export function formatDateTime(input?: string | null): string {
  if (!input) {
    return '—';
  }
  return dateTime({input, timeZone: defaultTimeZone}).format('YYYY-MM-DD HH:mm');
}

export function formatShortDate(input?: string | null): string {
  if (!input) {
    return '—';
  }
  return dateTime({input, timeZone: defaultTimeZone}).format('YYYY-MM-DD');
}

export function formatCount(value: number): string {
  return new Intl.NumberFormat('ru-RU').format(value);
}

export function humanizeStatus(status: EntityStatus): string {
  const dictionary: Record<EntityStatus, string> = {
    active: 'Active',
    paused: 'Paused',
    suspended: 'Suspended',
    disabled: 'Disabled',
    pending: 'Pending',
    running: 'Running',
    done: 'Done',
    failed: 'Failed',
    expired: 'Expired',
    revoked: 'Revoked',
    trial: 'Trial',
  };

  return dictionary[status];
}

export function statusToSeverity(status: EntityStatus): Severity {
  switch (status) {
    case 'active':
    case 'done':
      return 'success';
    case 'pending':
    case 'paused':
    case 'trial':
      return 'warning';
    case 'failed':
    case 'revoked':
    case 'suspended':
    case 'disabled':
      return 'danger';
    default:
      return 'info';
  }
}

export function titleFromId(value: string): string {
  return value
    .replace(/[-_.]/g, ' ')
    .replace(/\b\w/g, (char) => char.toUpperCase());
}
