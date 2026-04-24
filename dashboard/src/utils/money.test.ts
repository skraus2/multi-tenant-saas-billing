import { describe, expect, it } from "vitest";
import { formatMoney } from "./money";

describe("formatMoney", () => {
  it("formats euros correctly", () => {
    expect(formatMoney(999, "EUR")).toBe("9,99\u00a0€");
  });

  it("formats zero correctly", () => {
    expect(formatMoney(0, "EUR")).toBe("0,00\u00a0€");
  });

  it("formats large amounts correctly", () => {
    expect(formatMoney(100000, "EUR")).toBe("1.000,00\u00a0€");
  });

  it("formats negative amounts (credits)", () => {
    expect(formatMoney(-500, "EUR")).toBe("-5,00\u00a0€");
  });

  it("formats USD with custom locale", () => {
    const result = formatMoney(2999, "USD", "en-US");
    expect(result).toBe("$29.99");
  });
});
