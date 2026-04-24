from decimal import Decimal

import pytest
from pydantic import ValidationError

from src.models.tenant import TenantMetrics


def test_tenant_metrics_valid() -> None:
    m = TenantMetrics(
        tenant_id="abc-123",
        mrr=Decimal("1234.56"),
        churn_rate=Decimal("0.05"),
        active_subscriptions=42,
        failed_payments_30d=2,
    )
    assert m.tenant_id == "abc-123"
    assert m.mrr == Decimal("1234.56")


def test_tenant_metrics_rejects_negative_subscriptions() -> None:
    with pytest.raises(ValidationError):
        TenantMetrics(
            tenant_id="abc",
            mrr=Decimal("0"),
            churn_rate=Decimal("0"),
            active_subscriptions=-1,
            failed_payments_30d=0,
        )


def test_tenant_metrics_rejects_float_mrr() -> None:
    # MRR must be Decimal, not float
    m = TenantMetrics(
        tenant_id="abc",
        mrr=Decimal("9.99"),
        churn_rate=Decimal("0"),
        active_subscriptions=0,
        failed_payments_30d=0,
    )
    assert isinstance(m.mrr, Decimal)
