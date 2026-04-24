from decimal import Decimal

from pydantic import BaseModel, Field


class TenantMetrics(BaseModel):
    """Aggregated metrics for a single tenant."""

    tenant_id: str
    mrr: Decimal
    churn_rate: Decimal
    active_subscriptions: int = Field(ge=0)
    failed_payments_30d: int = Field(ge=0)
