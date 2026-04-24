use rust_decimal::prelude::{RoundingStrategy, ToPrimitive};
use rust_decimal::Decimal;
use rust_decimal_macros::dec;

use crate::error::BillingError;

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
    ///
    /// Uses `ROUND_HALF_UP` (e.g. 0.5 → 1, 1.5 → 2) — the standard for
    /// financial calculations in most jurisdictions.
    ///
    /// # Errors
    /// Returns `BillingError::Overflow` if the amount exceeds `i64` range
    /// (approximately ±92 trillion cents / ±920 billion USD). This should
    /// never occur with amounts created via `from_cents`, but can occur if
    /// `Money::new` is called with an extreme `Decimal` value.
    pub fn to_cents(&self) -> Result<i64, BillingError> {
        (self.amount * dec!(100))
            .round_dp_with_strategy(0, RoundingStrategy::MidpointAwayFromZero)
            .to_i64()
            .ok_or(BillingError::Overflow)
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
        assert_eq!(m.to_cents().unwrap(), cents);
    }

    #[test]
    fn to_cents_negative_roundtrip() {
        let cents = -500i64;
        let m = Money::from_cents(cents, Currency::EUR);
        assert_eq!(m.to_cents().unwrap(), cents);
    }

    #[test]
    fn to_cents_rounds_half_up() {
        // 9.995 → 999.5 cents → rounds to 1000 (ROUND_HALF_UP)
        let m = Money::new(dec!(9.995), Currency::EUR);
        assert_eq!(m.to_cents().unwrap(), 1000);
    }

    #[test]
    fn to_cents_rounds_half_up_even_base() {
        // 10.005 → 1000.5 cents → rounds to 1001 (ROUND_HALF_UP, not Banker's)
        let m = Money::new(dec!(10.005), Currency::EUR);
        assert_eq!(m.to_cents().unwrap(), 1001);
    }

    #[test]
    fn to_cents_overflow_returns_error() {
        // Amount far exceeds i64 range → must return Err, not silent 0
        let huge = Decimal::from(i64::MAX) + Decimal::ONE;
        let m = Money::new(huge, Currency::EUR);
        assert!(m.to_cents().is_err());
    }

    #[test]
    fn display_formats_correctly() {
        let m = Money::from_cents(999, Currency::EUR);
        assert_eq!(m.to_string(), "9.99 EUR");
    }

    // Property-based tests — invariants that must hold for all inputs
    mod proptest_suite {
        use super::*;
        use proptest::prelude::*;

        proptest! {
            /// `from_cents(x).to_cents() == x` for all practical financial amounts.
            /// Range bounded to ±i64::MAX/100 to avoid overflow when multiplying by 100
            /// inside `to_cents`. Covers ±92 trillion cents (≈ ±920 billion USD).
            #[test]
            fn from_cents_to_cents_roundtrip(cents in i64::MIN/100..=i64::MAX/100) {
                let m = Money::from_cents(cents, Currency::EUR);
                prop_assert_eq!(m.to_cents().unwrap(), cents);
            }

            /// Currency is preserved through from_cents (no mutation).
            #[test]
            fn currency_preserved(cents in -1_000_000_00i64..=1_000_000_00i64) {
                let m = Money::from_cents(cents, Currency::USD);
                prop_assert_eq!(m.currency, Currency::USD);
            }

            /// Amount is never negative for non-negative cent inputs.
            #[test]
            fn non_negative_cents_produce_non_negative_amount(cents in 0i64..=i64::MAX/100) {
                let m = Money::from_cents(cents, Currency::EUR);
                prop_assert!(m.amount >= rust_decimal::Decimal::ZERO);
            }
        }
    }
}
