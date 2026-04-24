-- Migration 002: Core billing tables
-- All tables have tenant_id for explicit tenant isolation (RLS added in 004)

CREATE TABLE IF NOT EXISTS subscription_plans (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name             TEXT        NOT NULL,
    price_cents      BIGINT      NOT NULL CHECK (price_cents >= 0),
    currency         TEXT        NOT NULL DEFAULT 'EUR',
    interval         TEXT        NOT NULL CHECK (interval IN ('month', 'year')),
    stripe_price_id  TEXT,
    active           BOOLEAN     NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customers (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email               TEXT        NOT NULL,
    name                TEXT,
    stripe_customer_id  TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, email)
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    customer_id             UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    plan_id                 UUID        NOT NULL REFERENCES subscription_plans(id) ON DELETE RESTRICT,
    status                  TEXT        NOT NULL CHECK (status IN ('active', 'trialing', 'canceled', 'past_due', 'unpaid', 'incomplete')),
    stripe_subscription_id  TEXT,
    current_period_start    TIMESTAMPTZ,
    current_period_end      TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscription_plans_tenant_id ON subscription_plans(tenant_id);
CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON customers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_id ON subscriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_subscription_id ON subscriptions(stripe_subscription_id);

-- Prevent cross-tenant FK references: customer and plan must belong to the same tenant as the subscription
CREATE OR REPLACE FUNCTION check_subscription_tenant_consistency()
RETURNS TRIGGER AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM customers WHERE id = NEW.customer_id AND tenant_id = NEW.tenant_id
    ) THEN
        RAISE EXCEPTION 'customer_id % does not belong to tenant %', NEW.customer_id, NEW.tenant_id;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM subscription_plans WHERE id = NEW.plan_id AND tenant_id = NEW.tenant_id
    ) THEN
        RAISE EXCEPTION 'plan_id % does not belong to tenant %', NEW.plan_id, NEW.tenant_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_subscription_tenant_check
    BEFORE INSERT OR UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION check_subscription_tenant_consistency();
