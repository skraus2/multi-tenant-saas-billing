-- Migration 004: Enable Row Level Security + create app role
--
-- Two DB roles:
--   billing     — superuser, used only for migrations, bypasses RLS by design
--   billing_app — non-superuser, used by the Go API, subject to RLS
--
-- billing_app has no password here — set via ALTER ROLE in deployment/local setup.
-- For local dev: docker-compose sets POSTGRES_APP_PASSWORD; see .env.example
-- For production: use a secrets manager, never hardcode passwords in migrations.
--
-- current_setting('app.current_tenant_id', true) returns NULL when not set (missing_ok=true).
-- NULL causes USING/WITH CHECK to evaluate false → zero rows returned, no error.
-- This is the correct silent-deny behavior for unauthenticated or background connections.
--
-- ALTER DEFAULT PRIVILEGES applies to objects created by the 'billing' role.
-- All migrations must be run as 'billing' to ensure billing_app inherits grants on new tables.

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'billing_app') THEN
        CREATE ROLE billing_app LOGIN;
    END IF;
END
$$;

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO billing_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO billing_app;
ALTER DEFAULT PRIVILEGES FOR ROLE billing IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO billing_app;
ALTER DEFAULT PRIVILEGES FOR ROLE billing IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO billing_app;

-- tenants: SELECT own row only (no tenant_id column — use id instead)
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenants_self_access ON tenants
    USING (id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (id = current_setting('app.current_tenant_id', true)::UUID);

ALTER TABLE tenant_settings ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_settings_tenant_isolation ON tenant_settings
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

ALTER TABLE subscription_plans ENABLE ROW LEVEL SECURITY;
CREATE POLICY subscription_plans_tenant_isolation ON subscription_plans
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
CREATE POLICY customers_tenant_isolation ON customers
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
CREATE POLICY subscriptions_tenant_isolation ON subscriptions
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
CREATE POLICY invoices_tenant_isolation ON invoices
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

ALTER TABLE stripe_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY stripe_events_tenant_isolation ON stripe_events
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);
