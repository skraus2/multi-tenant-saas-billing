/**
 * Formats a monetary amount from integer cents to a locale-aware currency string.
 *
 * @param cents - Amount in integer cents (e.g. 999 = 9.99)
 * @param currency - ISO 4217 currency code (e.g. "EUR", "USD")
 * @param locale - BCP 47 locale string (default: "de-DE")
 * @returns Formatted currency string (e.g. "9,99 €")
 */
export function formatMoney(
  cents: number,
  currency: string,
  locale = "de-DE",
): string {
  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
  }).format(cents / 100);
}
