# Modules

Feature modules live here as vertical slices: auth, users, catalog, cart,
checkout, orders, payments, shipping, sellers, admin, reviews, notifications,
chat, and analytics.

Each module should own its HTTP handlers, service logic, repository interfaces,
and module-local request/response DTOs.
