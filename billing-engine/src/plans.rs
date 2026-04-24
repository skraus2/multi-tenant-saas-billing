use crate::money::Money;

/// Billing interval for a subscription plan.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum BillingInterval {
    Monthly,
    Yearly,
}

impl BillingInterval {
    /// Returns the number of days in this billing period (approximate for yearly).
    pub fn days_in_period(&self) -> u32 {
        match self {
            BillingInterval::Monthly => 30,
            BillingInterval::Yearly => 365,
        }
    }
}

/// A subscription plan offered by a tenant.
#[derive(Debug, Clone, PartialEq)]
pub struct Plan {
    pub id: String,
    pub name: String,
    pub price: Money,
    pub interval: BillingInterval,
}

impl Plan {
    /// Creates a new subscription plan.
    pub fn new(
        id: impl Into<String>,
        name: impl Into<String>,
        price: Money,
        interval: BillingInterval,
    ) -> Self {
        Self {
            id: id.into(),
            name: name.into(),
            price,
            interval,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::money::Currency;

    #[test]
    fn plan_monthly_days() {
        assert_eq!(BillingInterval::Monthly.days_in_period(), 30);
    }

    #[test]
    fn plan_yearly_days() {
        assert_eq!(BillingInterval::Yearly.days_in_period(), 365);
    }

    #[test]
    fn plan_creation() {
        let price = Money::from_cents(2900, Currency::EUR);
        let plan = Plan::new("pro", "Pro Plan", price.clone(), BillingInterval::Monthly);
        assert_eq!(plan.name, "Pro Plan");
        assert_eq!(plan.price, price);
        assert_eq!(plan.interval, BillingInterval::Monthly);
    }
}
