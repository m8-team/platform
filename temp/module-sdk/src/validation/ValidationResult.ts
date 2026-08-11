export type M8ValidationSeverity = 'error' | 'warning';

export type M8ValidationIssue = {
  severity: M8ValidationSeverity;
  code: string;
  message: string;
  path?: string;
};

export type M8ValidationResult = {
  valid: boolean;
  issues: M8ValidationIssue[];
  errors: M8ValidationIssue[];
  warnings: M8ValidationIssue[];
};

export function createValidationResult(issues: M8ValidationIssue[]): M8ValidationResult {
  const errors = issues.filter((issue) => issue.severity === 'error');
  const warnings = issues.filter((issue) => issue.severity === 'warning');

  return {
    valid: errors.length === 0,
    issues,
    errors,
    warnings,
  };
}

export function errorIssue(
  code: string,
  message: string,
  path?: string,
): M8ValidationIssue {
  return {
    severity: 'error',
    code,
    message,
    ...(path ? {path} : {}),
  };
}

export function warningIssue(
  code: string,
  message: string,
  path?: string,
): M8ValidationIssue {
  return {
    severity: 'warning',
    code,
    message,
    ...(path ? {path} : {}),
  };
}
