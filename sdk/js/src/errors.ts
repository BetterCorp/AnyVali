import type { ValidationIssue } from "./types.js";

export class ValidationError extends Error {
  public readonly issues: ValidationIssue[];

  constructor(issues: ValidationIssue[]) {
    const message = issues
      .map(
        (i) =>
          `[${i.code}] ${i.path.length > 0 ? i.path.join(".") + ": " : ""}${i.message}`
      )
      .join("\n");
    super(message);
    this.name = "ValidationError";
    this.issues = issues;
  }
}
