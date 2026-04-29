# Backend Load Tests

Phase 12 load coverage targets:

- catalog list
- product detail
- search
- cart
- checkout
- notification fanout

Run examples (requires `k6`):

```bash
k6 run backend/loadtests/catalog.js
k6 run backend/loadtests/search.js
```
