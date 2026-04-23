use rust_decimal::Decimal;
use rust_decimal_macros::dec;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Currency {
    EUR,
    USD,
    GBP,
}

impl std::fmt::Display for Currency {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Currency::EUR => write!(f, "EUR"),
            Currency::USD => write!(f, "USD"),
            Currency::GBP => write!(f, "GBP"),
        }
    }
}

/// Represents a monetary amount with currency.
/// Amount is always stored as `Decimal` — never `f64`/`f32`.
#[derive(Debug, Clone, PartialEq)]
pub struct Money {
    pub amount: Decimal,
    pub currency: Currency,
}

impl Money {
    /// Creates a `Money` value from a decimal amount.
    pub fn new(amount: Decimal, currency: Currency) -> Self {
        Self { amount, currency }
    }

    /// Creates a `Money` value from an integer cent amount (e.g. 999 → 9.99).
    pub fn from_cents(cents: i64, currency: Currency) -> Self {
        Self {
            amount: Decimal::new(cents, 2),
            currency,
        }
    }

    /// Returns the amount as integer cents (e.g. 9.99 → 999).
    pub fn to_cents(&self) -> i64 {
        (self.amount * dec!(100))
            .round_dp(0)
            .to_string()
            .parse::<i64>()
            .unwrap_or(0)
    }
}

impl std::fmt::Display for Money {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{} {}", self.amount, self.currency)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use rust_decimal_macros::dec;

    #[test]
    fn from_cents_positive() {
        let m = Money::from_cents(999, Currency::EUR);
        assert_eq!(m.amount, dec!(9.99));
        assert_eq!(m.currency, Currency::EUR);
    }

    #[test]
    fn from_cents_zero() {
        let m = Money::from_cents(0, Currency::EUR);
        assert_eq!(m.amount, dec!(0));
    }

    #[test]
    fn from_cents_negative() {
        // Negative = credit
        let m = Money::from_cents(-1000, Currency::EUR);
        assert_eq!(m.amount, dec!(-10.00));
    }

    #[test]
    fn to_cents_roundtrip() {
        let cents = 2900i64;
        let m = Money::from_cents(cents, Currency::EUR);
        assert_eq!(m.to_cents(), cents);
    }

    #[test]
    fn to_cents_negative_roundtrip() {
        let cents = -500i64;
        let m = Money::from_cents(cents, Currency::EUR);
        assert_eq!(m.to_cents(), cents);
    }

    #[test]
    fn to_cents_rounds_correctly() {
        // 9.995 should round to 1000 cents
        let m = Money::new(dec!(9.995), Currency::EUR);
        assert_eq!(m.to_cents(), 1000);
    }

    #[test]
    fn display_formats_correctly() {
        let m = Money::from_cents(999, Currency::EUR);
        assert_eq!(m.to_string(), "9.99 EUR");
    }
}
