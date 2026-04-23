use thiserror::Error;

/// All errors that can occur in the billing engine.
#[derive(Debug, Error, PartialEq)]
pub enum BillingError {
    #[error("division by zero in billing calculation")]
    DivisionByZero,

    #[error("invalid period: days_remaining ({0}) exceeds days_in_period ({1})")]
    InvalidPeriod(u32, u32),

    #[error("currency mismatch: expected {expected}, got {got}")]
    CurrencyMismatch { expected: String, got: String },

    #[error("amount overflow")]
    Overflow,
}
