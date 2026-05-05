# OIDC Federation Trust Chain Resolver Browser

A web application for resolving and inspecting OIDC federation trust chains.

## Features

- **Trust Chain Resolution**: Enter a trust anchor and subject entity to resolve the complete trust chain
- **JWT Inspection**: View JWT headers, payloads, and raw tokens for each statement in the chain
- **Metadata Display**: Inspect resolved federation metadata in a user-friendly format
- **Preview Edits**: Edit trust chain statements and preview the resulting metadata
- **Diff View**: Compare original vs. modified metadata side-by-side
- **Collapsible Cards**: Expand/collapse trust chain nodes for easy navigation

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 18+
- npm

### Build

1. **Build the frontend:**
   ```bash
   cd frontend
   npm install
   npm run build
   cd ..
   ```

2. **Build the backend:**
   ```bash
   go build ./cmd/server
   ```

### Run

```bash
./server
```

The server will start on `http://localhost:8080`

## Usage

1. Open your browser to `http://localhost:8080`

2. Enter the required information:
   - **Trust Anchor URL**: The federation trust anchor (e.g., `https://ta.example.com`)
   - **Subject Entity ID**: The entity you want to resolve (e.g., `https://rp.example.com`)
   - **Entity Types** (optional): Filter by specific entity types

3. Click **Resolve** to fetch and display the trust chain

4. Inspect the results:
   - **Left column**: Trust chain visualization with expandable JWT details
   - **Right column**: Resolved metadata

5. Use **Preview Edits** to modify statements and see how changes affect metadata resolution

6. Toggle **Show Diff** to compare original vs. modified metadata

## API Endpoints

### POST /api/resolve

Resolve a trust chain from a subject to a trust anchor.

**Request:**
```json
{
  "sub": "https://rp.example.com",
  "trust_anchor": ["https://ta.example.com"],
  "entity_types": ["openid_relying_party"]
}
```

**Response:**
```json
{
  "metadata": { ... },
  "trust_chain": [
    {
      "index": 0,
      "type": "trust_anchor_entity_configuration",
      "issuer": "https://ta.example.com",
      "subject": "https://ta.example.com",
      "jwt_header": { ... },
      "jwt_payload": { ... },
      "raw_jwt": "...",
      "iat": 1234567890,
      "exp": 1234567890
    }
  ],
  "expires_at": 1234567890
}
```

### POST /api/resolve/preview

Preview metadata resolution with edited trust chain (no signature verification).

**Request:**
```json
{
  "trust_anchor": "https://ta.example.com",
  "trust_chain": [
    {
      "issuer": "https://ta.example.com",
      "subject": "https://ta.example.com",
      "jwks": { "keys": [...] },
      "metadata": { ... },
      "constraints": { ... }
    }
  ]
}
```

**Response:**
```json
{
  "metadata": { ... },
  "valid": true
}
```

## Project Structure

```
resolve-browser/
├── cmd/server/main.go          # Server entry point
├── internal/
│   ├── api/types.go            # API request/response types
│   └── handlers/
│       ├── resolve.go          # Resolve endpoint handler
│       └── preview.go          # Preview endpoint handler
├── frontend/
│   ├── src/
│   │   ├── components/         # Svelte components
│   │   ├── lib/                # Utilities, stores, API client
│   │   ├── App.svelte          # Root component
│   │   └── main.js             # Entry point
│   ├── static/                 # Built assets (generated)
│   ├── package.json
│   └── vite.config.js
├── go.mod
└── IMPLEMENTATION_PLAN.md
```

## Development

### Frontend Development

Run the Vite dev server with hot reload:

```bash
cd frontend
npm run dev
```

Then proxy API requests to the Go backend (running on :8080) by configuring Vite:

```javascript
// vite.config.js
export default defineConfig({
  plugins: [svelte()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080'
    }
  }
})
```

### Backend Development

Run with auto-reload using `air` or similar:

```bash
go run ./cmd/server
```

## Technology Stack

- **Backend**: Go Fiber v2
- **Frontend**: Svelte 5 + Vite
- **Styling**: Tailwind CSS
- **JSON Diff**: jsondiffpatch
- **OIDC Federation**: github.com/go-oidfed/lib

## License

Same as the go-oidfed project.
