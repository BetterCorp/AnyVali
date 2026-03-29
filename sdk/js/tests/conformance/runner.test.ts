import { describe, it, expect } from "vitest";
import * as path from "node:path";
import { loadCorpus, runTestCase } from "./runner.js";

const corpusDir = path.resolve(__dirname, "../../../../spec/corpus");

const suites = loadCorpus(corpusDir);

if (suites.length === 0) {
  describe("Conformance corpus", () => {
    it("no corpus files found (skipping)", () => {
      expect(true).toBe(true);
    });
  });
} else {
  for (const suite of suites) {
    describe(`Conformance: ${suite.suite}`, () => {
      for (const tc of suite.cases) {
        it(tc.description, () => {
          const result = runTestCase(tc);
          if (!result.passed) {
            throw new Error(result.error);
          }
        });
      }
    });
  }
}
