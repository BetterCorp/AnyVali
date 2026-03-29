import * as fs from "node:fs";
import * as path from "node:path";
import { importSchema } from "../../src/interchange/importer.js";
import type { AnyValiDocument } from "../../src/types.js";

export interface CorpusFile {
  suite: string;
  cases: CorpusTestCase[];
}

export interface CorpusTestCase {
  description: string;
  schema: AnyValiDocument;
  input: unknown;
  valid: boolean;
  output: unknown;
  issues: Array<{
    code: string;
    path: (string | number)[];
    expected?: string;
    received?: string;
  }>;
}

export interface CorpusTestResult {
  description: string;
  passed: boolean;
  error?: string;
}

/**
 * Load all corpus test files from the given directory (recursively).
 * Returns an array of { suite, cases } objects.
 */
export function loadCorpus(corpusDir: string): CorpusFile[] {
  const suites: CorpusFile[] = [];

  if (!fs.existsSync(corpusDir)) {
    return suites;
  }

  function walk(dir: string): void {
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        walk(fullPath);
      } else if (entry.name.endsWith(".json")) {
        const content = fs.readFileSync(fullPath, "utf-8");
        const parsed = JSON.parse(content) as CorpusFile;
        suites.push(parsed);
      }
    }
  }

  walk(corpusDir);
  return suites;
}

/**
 * Run a single corpus test case.
 */
export function runTestCase(tc: CorpusTestCase): CorpusTestResult {
  try {
    const schema = importSchema(tc.schema);
    const result = schema.safeParse(tc.input);

    if (tc.valid) {
      // Expect success
      if (!result.success) {
        return {
          description: tc.description,
          passed: false,
          error: `Expected success but got failure: ${JSON.stringify(result.issues)}`,
        };
      }
      // Compare output
      const expectedJson = JSON.stringify(tc.output);
      const actualJson = JSON.stringify(result.data);
      if (expectedJson !== actualJson) {
        return {
          description: tc.description,
          passed: false,
          error: `Output mismatch: expected ${expectedJson}, got ${actualJson}`,
        };
      }
      return { description: tc.description, passed: true };
    } else {
      // Expect failure
      if (result.success) {
        return {
          description: tc.description,
          passed: false,
          error: `Expected failure but got success: ${JSON.stringify(result.data)}`,
        };
      }
      // Check expected issue codes and paths
      for (const expectedIssue of tc.issues) {
        const found = result.issues.some((actual) => {
          if (actual.code !== expectedIssue.code) return false;
          if (
            JSON.stringify(actual.path) !== JSON.stringify(expectedIssue.path)
          ) {
            return false;
          }
          if (
            expectedIssue.expected !== undefined &&
            actual.expected !== expectedIssue.expected
          ) {
            return false;
          }
          if (
            expectedIssue.received !== undefined &&
            actual.received !== expectedIssue.received
          ) {
            return false;
          }
          return true;
        });
        if (!found) {
          return {
            description: tc.description,
            passed: false,
            error: `Expected issue ${JSON.stringify(expectedIssue)} not found in actual issues: ${JSON.stringify(result.issues)}`,
          };
        }
      }
      return { description: tc.description, passed: true };
    }
  } catch (err: any) {
    return {
      description: tc.description,
      passed: false,
      error: `Exception: ${err.message}`,
    };
  }
}
