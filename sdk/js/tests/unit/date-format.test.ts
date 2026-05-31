import { describe, it, expect } from "vitest";
import { string } from "../../src/index.js";

const date = () => string().format("date");

describe("date format validation", () => {
  it("accepts valid ISO dates including early years (0001-0099)", () => {
    for (const d of ["0001-01-01", "0050-01-01", "0099-12-31", "2021-06-15", "9999-12-31"]) {
      expect(date().safeParse(d).success).toBe(true);
    }
  });

  it("rejects impossible calendar dates", () => {
    for (const d of ["2021-02-30", "2021-04-31", "2021-13-01", "2021-00-10"]) {
      expect(date().safeParse(d).success).toBe(false);
    }
  });
});
