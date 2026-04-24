from decimal import Decimal

import pytest

from src.metrics.churn import calculate_churn_rate


def test_churn_rate_basic() -> None:
    rate = calculate_churn_rate(churned=2, total_at_start=10)
    assert rate == Decimal("0.20")


def test_churn_rate_zero_customers() -> None:
    rate = calculate_churn_rate(churned=0, total_at_start=0)
    assert rate == Decimal("0")


def test_churn_rate_full_churn() -> None:
    rate = calculate_churn_rate(churned=5, total_at_start=5)
    assert rate == Decimal("1.00")


def test_churn_rate_no_churn() -> None:
    rate = calculate_churn_rate(churned=0, total_at_start=100)
    assert rate == Decimal("0")


def test_churn_rate_invalid_raises() -> None:
    with pytest.raises(ValueError, match="churned cannot exceed total_at_start"):
        calculate_churn_rate(churned=5, total_at_start=3)
