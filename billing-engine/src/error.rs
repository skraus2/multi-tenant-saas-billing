use thiserror::Error;

/// All errors that can occur in the billing engine.
#[derive(Debug, Error, PartialEq)]
pub enum AppError {
    /// A billing calculation attempted to divide by zero.
    #[error("division by zero in billing calculation")]
    DivisionByZero,

    /// A monetary amount was invalid (e.g. NaN, infinite, or out of range).
    #[error("invalid amount: {0}")]
    InvalidAmount(String),

    /// A monetary amount was negative where only non-negative values are valid.
    /// Distinct from `InvalidAmount` so callers can match on this case programmatically.
    #[error("amount must not be negative: {0}")]
    NegativeAmount(String),

    /// The billing period parameters are inconsistent.
    #[error("invalid period: days_remaining ({0}) exceeds days_in_period ({1})")]
    InvalidPeriod(u32, u32),

    /// Two monetary values with different currencies were combined.
    #[error("currency mismatch: expected {expected}, got {got}")]
    CurrencyMismatch { expected: String, got: String },

    /// A monetary amount overflowed the representable range.
    #[error("amount overflow: value exceeds i64 range")]
    Overflow,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn division_by_zero_is_constructible() {
        let e = AppError::DivisionByZero;
        assert_eq!(e.to_string(), "division by zero in billing calculation");
    }

    #[test]
    fn invalid_amount_carries_reason() {
        let e = AppError::InvalidAmount("amount must be >= 0".to_string());
        assert_eq!(e.to_string(), "invalid amount: amount must be >= 0");
    }

    #[test]
    fn invalid_period_shows_both_values() {
        let e = AppError::InvalidPeriod(15, 10);
        assert_eq!(
            e.to_string(),
            "invalid period: days_remaining (15) exceeds days_in_period (10)"
        );
    }

    #[test]
    fn currency_mismatch_shows_both_currencies() {
        let e = AppError::CurrencyMismatch {
            expected: "EUR".to_string(),
            got: "USD".to_string(),
        };
        assert_eq!(e.to_string(), "currency mismatch: expected EUR, got USD");
    }

    #[test]
    fn overflow_is_constructible() {
        let e = AppError::Overflow;
        assert_eq!(e.to_string(), "amount overflow: value exceeds i64 range");
    }

    #[test]
    fn negative_amount_carries_reason() {
        let e = AppError::NegativeAmount("plan price must be >= 0".to_string());
        assert_eq!(
            e.to_string(),
            "amount must not be negative: plan price must be >= 0"
        );
    }

    #[test]
    fn errors_are_comparable() {
        assert_eq!(AppError::DivisionByZero, AppError::DivisionByZero);
        assert_ne!(AppError::DivisionByZero, AppError::Overflow);
        assert_ne!(
            AppError::InvalidAmount("x".to_string()),
            AppError::NegativeAmount("x".to_string())
        );
    }
}
