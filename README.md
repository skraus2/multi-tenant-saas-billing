# BillFlow

Multi-Tenant SaaS Billing Platform. SaaS-Unternehmen registrieren sich, verwalten ihre Kunden, definieren Subscription-Pläne und wickeln Zahlungen über Stripe ab — vollständig isoliert voneinander.

## Was ist BillFlow?

BillFlow ist eine Billing-Infrastruktur für SaaS-Unternehmen. Statt selbst eine Billing-Integration zu bauen, nutzen sie BillFlow als fertige Lösung:

- **Kunden verwalten** — Endkunden anlegen, bearbeiten, Stripe-Synchronisation automatisch
- **Pläne definieren** — Basic 9€/Mo, Pro 29€/Mo, Enterprise — beliebig konfigurierbar
- **Subscriptions starten** — Kunde einem Plan zuordnen, Stripe übernimmt die Abrechnung
- **Zahlungen überwachen** — fehlgeschlagene Zahlungen erkennen, Dunning-Prozess automatisch
- **Umsatz verstehen** — MRR, Churn-Rate, aktive Subscriptions auf einen Blick

Jedes Unternehmen hat einen vollständig isolierten Bereich. Kein Tenant sieht Daten eines anderen.

## Stack

| Service          | Technologie        | Aufgabe                                              |
| ---------------- | ------------------ | ---------------------------------------------------- |
| `billing-engine` | Rust               | Proration, Rechnungsberechnung, Zahlungsverarbeitung |
| `api`            | Go                 | REST API, Stripe Webhooks, Background Jobs           |
| `dashboard`      | TypeScript + React | Frontend für SaaS-Unternehmen                        |
| `reporting`      | Python             | MRR-Analyse, Churn-Reports, Audits                   |
| Datenbank        | PostgreSQL         | Row Level Security für Tenant-Isolation              |
| Zahlungen        | Stripe             | Subscriptions, Invoices, Webhooks                    |

## Lizenz

Proprietär. Alle Rechte vorbehalten.
