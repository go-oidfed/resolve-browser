# OpenID Federation Trust Chain Resolver Browser

A web application for resolving and inspecting OpenID Federation trust chains.

## Features

- **Trust Chain Resolution**: Enter a trust anchor and subject entity to resolve the complete trust chain
- **JWT Inspection**: View JWT headers, payloads, and raw tokens for each statement in the chain
- **Metadata Display**: Inspect resolved federation metadata in a user-friendly format
- **Preview Edits**: Edit trust chain statements and preview the resulting metadata
- **Diff View**: Compare original vs. modified metadata side-by-side
- **Collapsible Cards**: Expand/collapse trust chain nodes for easy navigation

## Quick Start

### Prerequisites

- Go 1.25+
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

## Environment Variables

| Variable             | Description                                   | Format                                          | Default  |
|----------------------|-----------------------------------------------|-------------------------------------------------|----------|
| `PROXY_HEADER`       | Header name for client IP behind proxy        | String (e.g., `X-Forwarded-For`)                | Empty    |
| `CACHE_MAX_LIFETIME` | Maximum lifetime for cached entity statements | Go duration string (e.g., `1h`, `3600s`, `24h`) | No limit |

**Example:**
```bash
export CACHE_MAX_LIFETIME=1h
./server
```

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

Preview metadata resolution with edited trust chain (no signature verification). The trust chain should be provided as a JSON array in order from leaf entity to trust anchor.

Trust chain elements can be provided in three formats:
1. **JSON object**: Plain JSON with standard OIDC Federation fields
2. **Base64-encoded JSON**: Base64-encoded string of a JSON object
3. **JWT**: Raw JWT string (signature is not verified, only payload is used)

**Request:**
```json
[
  {
    "iss": "https://rp.example.com",
    "sub": "https://rp.example.com",
    "iat": 1234567890,
    "exp": 1234567890,
    "jwks": { "keys": [...] },
    "metadata": { ... },
    "authority_hints": ["https://intermediate.example.com"]
  },
  {
    "iss": "https://intermediate.example.com",
    "sub": "https://rp.example.com",
    "jwks": { "keys": [...] },
    "metadata_policy": { ... }
  },
  "eyJhbGciOiJFUzI1NiIsInR5cCI6ImVudGl0eS1zdGF0ZW1lbnQrand0In0..."
]
```

**Response:**
```json
{
  "metadata": { ... },
  "valid": true
}
```
