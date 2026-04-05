# IAM Admin UI

Gravity UI admin console for the M8 Platform IAM control plane.

## Local run

```bash
npm install
npm run dev
```

By default the UI now runs in `live` mode and reads tenants and other entities from the current IAM API.

- Dev proxy: `/api` and `/openapi` are forwarded to `http://127.0.0.1:8082`
- Default API mode: `live`
- Default mock fallback in live mode: `off`

## Environment

- `VITE_IAM_API_MODE=live|mock`
- `VITE_IAM_API_BASE_URL=`
- `VITE_IAM_FALLBACK_TO_MOCK=false`
- `VITE_IAM_DEFAULT_TENANT_ID=tenant-demo`

Leave `VITE_IAM_API_BASE_URL` empty for local development so browser requests stay same-origin and go through the Vite proxy to `http://127.0.0.1:8082`. Set an absolute API URL only when that backend is configured for CORS.

To force demo data instead of the real backend:

```bash
VITE_IAM_API_MODE=mock npm run dev
```
