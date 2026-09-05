import * as fs from "node:fs";
import * as path from "node:path";
import assert from "node:assert/strict";
import { importSchema } from "../../src/interchange/importer.js";
import { exportSchema, encrypt, decrypt, safeParseEncrypted } from "../../src/index.js";
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
  roundtrip?: boolean;
  sensitivePaths?: (string | number)[][];
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
    let schema = importSchema(tc.schema);
    if (tc.roundtrip) {
      const exported = exportSchema(schema.describe("Imported contract"));
      assert.deepEqual(exported.definitions, tc.schema.definitions);
      schema = importSchema(JSON.parse(JSON.stringify(exported)));
    }
    if (tc.sensitivePaths) {
      for (const imported of [importSchema(tc.schema), schema]) {
        assert.equal(safeParseEncrypted(imported, tc.input).success, false);
        const paths: (string | number)[][] = [];
        const encrypted = encrypt(imported, tc.input, (path, value) => {
          paths.push([...path]);
          return `encrypted:${JSON.stringify(value)}`;
        });
        assert.deepEqual(paths, tc.sensitivePaths);
        assert.equal(safeParseEncrypted(imported, encrypted).success, true);
        paths.length = 0;
        const plaintext = decrypt(imported, encrypted, (path, value) => {
          paths.push([...path]);
          return JSON.parse((value as string).slice("encrypted:".length));
        });
        assert.deepEqual(paths, tc.sensitivePaths);
        assert.deepEqual(plaintext, tc.output);
      }
    }
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
