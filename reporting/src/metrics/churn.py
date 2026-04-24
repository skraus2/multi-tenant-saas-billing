from decimal import ROUND_HALF_UP, Decimal


def calculate_churn_rate(churned: int, total_at_start: int) -> Decimal:
    """Calculate churn rate as a decimal fraction between 0 and 1.

    Args:
        churned: Number of customers who churned in the period.
        total_at_start: Total customers at the start of the period.

    Returns:
        Churn rate as Decimal (e.g. Decimal("0.20") for 20%).

    Raises:
        ValueError: If churned exceeds total_at_start.
    """
    if churned > total_at_start:
        raise ValueError(
            f"churned cannot exceed total_at_start: {churned} > {total_at_start}"
        )
    if total_at_start == 0:
        return Decimal("0")
    rate = Decimal(churned) / Decimal(total_at_start)
    return rate.quantize(Decimal("0.01"), rounding=ROUND_HALF_UP)
