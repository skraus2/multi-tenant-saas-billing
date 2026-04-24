-- Migration 003: Transaction tables
-- invoices: nullable subscription_id (manual invoices have no subscription)
-- stripe_events: TEXT PK (Stripe's own event ID, e.g. evt_1AbCdEf), no updated_at (immutable)

CREATE TABLE IF NOT EXISTS invoices (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    customer_id        UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    subscription_id    UUID        REFERENCES subscriptions(id) ON DELETE SET NULL,
    amount_cents       BIGINT      NOT NULL CHECK (amount_cents >= 0),
    currency           TEXT        NOT NULL DEFAULT 'EUR',
    status             TEXT        NOT NULL CHECK (status IN ('draft', 'open', 'paid', 'void', 'uncollectible')),
    stripe_invoice_id  TEXT,
    due_date           DATE,
    paid_at            TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stripe_events (
    event_id      TEXT        PRIMARY KEY,
    tenant_id     UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_type    TEXT        NOT NULL,
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoices_tenant_id ON invoices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invoices_subscription_id ON invoices(subscription_id);
CREATE INDEX IF NOT EXISTS idx_invoices_stripe_invoice_id ON invoices(stripe_invoice_id);
CREATE INDEX IF NOT EXISTS idx_stripe_events_tenant_id ON stripe_events(tenant_id);

-- Prevent cross-tenant FK references: customer and subscription must belong to the same tenant
CREATE OR REPLACE FUNCTION check_invoice_tenant_consistency()
RETURNS TRIGGER AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM customers WHERE id = NEW.customer_id AND tenant_id = NEW.tenant_id
    ) THEN
        RAISE EXCEPTION 'customer_id % does not belong to tenant %', NEW.customer_id, NEW.tenant_id;
    END IF;
    IF NEW.subscription_id IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM subscriptions WHERE id = NEW.subscription_id AND tenant_id = NEW.tenant_id
    ) THEN
        RAISE EXCEPTION 'subscription_id % does not belong to tenant %', NEW.subscription_id, NEW.tenant_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_invoice_tenant_check
    BEFORE INSERT OR UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION check_invoice_tenant_consistency();
