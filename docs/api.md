# AnyVali HTTP API Reference

The AnyVali CLI includes an HTTP server (`anyvali serve`) that exposes a REST API for validating JSON data against AnyVali schemas. This document covers every endpoint, request format, and deployment pattern.

---

## Starting the Server

```bash
anyvali serve [flags]
```

### Flags

| Flag | Short | Default | Environment Variable | Description |
|------|-------|---------|---------------------|-------------|
| `--port` | `-p` | `8080` | `ANYVALI_PORT` | Port to listen on |
| `--host` | | `0.0.0.0` | `ANYVALI_HOST` | Host/address to bind to |
| `--schemas` | `-s` | | `ANYVALI_SCHEMAS_DIR` | Directory of schema files to load at startup |
| `--cors` | | `false` | `ANYVALI_CORS` | Enable CORS headers |
| `--cors-origins` | | `*` | `ANYVALI_CORS_ORIGINS` | Allowed CORS origins (comma-separated) |
| `--read-timeout` | | `30s` | `ANYVALI_READ_TIMEOUT` | HTTP read timeout |
| `--write-timeout` | | `30s` | `ANYVALI_WRITE_TIMEOUT` | HTTP write timeout |
| `--max-body` | | `1MB` | `ANYVALI_MAX_BODY` | Maximum request body size |

Environment variables take effect when the corresponding flag is not provided. Flags always take precedence.

```bash
# Using flags
anyvali serve -s ./schemas -p 3000

# Using environment variables
export ANYVALI_SCHEMAS_DIR=./schemas
export ANYVALI_PORT=3000
anyvali serve
```

---

## Authentication

The AnyVali server does not include built-in authentication or authorization. For production deployments, place the server behind a reverse proxy that handles auth. See [Reverse Proxy for Auth](#reverse-proxy-for-auth) for Nginx and Caddy examples.

---

## Endpoints

### `POST /validate`

Validate JSON data against an inline schema document. The request body contains both the schema and the input data.

#### Request format (inline schema)

```json
{
  "schema": {
    "anyvaliVersion": "1.0",
    "schemaVersion": "1",
    "root": {
      "kind": "object",
      "properties": {
        "name": { "kind": "string", "minLength": 1 },
        "email": { "kind": "string", "format": "email" }
      },
      "required": ["name", "email"],
      "unknownKeys": "reject"
    },
    "definitions": {},
    "extensions": {}
  },
  "data": {
    "name": "Alice",
    "email": "alice@example.com"
  }
}
```

#### Request format (schema reference)

Reference a pre-loaded schema by name instead of embedding it:

```json
{
  "schemaRef": "user",
  "data": {
    "name": "Alice",
    "email": "alice@example.com"
  }
}
```

The `schemaRef` value corresponds to a schema file loaded from the `--schemas` directory. File names map to schema names by stripping the `.schema.json` or `.json` extension.

#### Success response

**Status:** `200 OK`

```json
{
  "valid": true,
  "issues": []
}
```

#### Failure response (validation failed)

**Status:** `200 OK`

The HTTP status is still 200 because the request was processed successfully -- the data simply did not pass validation.

```json
{
  "valid": false,
  "issues": [
    {
      "code": "too_small",
      "message": "String must have at least 1 character",
      "path": ["name"],
      "expected": 1,
      "received": 0
    },
    {
      "code": "invalid_string",
      "message": "Invalid email format",
      "path": ["email"],
      "expected": "email",
      "received": "not-an-email"
    }
  ]
}
```

#### Error response (bad request)

**Status:** `400 Bad Request`

```json
{
  "error": "invalid_request",
  "message": "Request body must contain either 'schema' or 'schemaRef', and 'data'."
}
```

#### Curl examples

Inline schema:

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "schema": {
      "anyvaliVersion": "1.0",
      "schemaVersion": "1",
      "root": {
        "kind": "string",
        "format": "email"
      },
      "definitions": {},
      "extensions": {}
    },
    "data": "alice@example.com"
  }'
```

Schema reference:

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "schemaRef": "user",
    "data": {"name": "Alice", "email": "alice@example.com"}
  }'
```

---

### `POST /validate/:name`

Validate JSON data against a pre-loaded schema. The request body is just the data to validate -- no schema wrapper needed.

#### How schema names map to files

When the server starts with `--schemas ./schemas`, it loads every `.json` and `.schema.json` file in that directory:

| File | Schema Name | Endpoint |
|------|-------------|----------|
| `user.schema.json` | `user` | `POST /validate/user` |
| `order.schema.json` | `order` | `POST /validate/order` |
| `app-config.json` | `app-config` | `POST /validate/app-config` |

#### Request format

The entire request body is the data to validate:

```json
{
  "name": "Alice",
  "email": "alice@example.com",
  "age": 30
}
```

#### Success response

**Status:** `200 OK`

```json
{
  "valid": true,
  "issues": []
}
```

#### Failure response

**Status:** `200 OK`

```json
{
  "valid": false,
  "issues": [
    {
      "code": "required",
      "message": "Required field is missing",
      "path": ["email"]
    }
  ]
}
```

#### Schema not found

**Status:** `404 Not Found`

```json
{
  "error": "schema_not_found",
  "message": "No schema loaded with name 'unknown-schema'."
}
```

#### Examples

```bash
# Validate a user
curl -X POST http://localhost:8080/validate/user \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}'

# Validate an order
curl -X POST http://localhost:8080/validate/order \
  -H "Content-Type: application/json" \
  -d '{"items": [{"sku": "ABC-123", "qty": 2}], "total": 29.99}'

# Validate from a file
curl -X POST http://localhost:8080/validate/config \
  -H "Content-Type: application/json" \
  -d @config.json
```

---

### `GET /schemas`

List all pre-loaded schemas.

#### Response format

**Status:** `200 OK`

```json
{
  "schemas": [
    {
      "name": "user",
      "rootKind": "object",
      "file": "user.schema.json"
    },
    {
      "name": "order",
      "rootKind": "object",
      "file": "order.schema.json"
    },
    {
      "name": "config",
      "rootKind": "object",
      "file": "config.json"
    }
  ]
}
```

#### Example

```bash
curl http://localhost:8080/schemas
```

If no schemas directory was provided at startup, the list is empty:

```json
{
  "schemas": []
}
```

---

### `GET /health`

Health check endpoint for container orchestration and load balancers.

#### Response format

**Status:** `200 OK`

```json
{
  "status": "ok"
}
```

#### Use for Docker and Kubernetes health checks

Docker Compose:

```yaml
healthcheck:
  test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
  interval: 10s
  timeout: 5s
  retries: 3
```

Kubernetes liveness and readiness probes:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 2
  periodSeconds: 5
```

---

## Error Handling

All error responses follow the same shape:

```json
{
  "error": "<error_code>",
  "message": "<human-readable description>"
}
```

### Error codes and status codes

| HTTP Status | Error Code | When |
|-------------|-----------|------|
| `400` | `invalid_request` | Request body is missing, malformed JSON, or missing required fields |
| `400` | `invalid_schema` | Inline schema document is not a valid AnyVali schema |
| `400` | `schema_import_failed` | Schema document failed to import (unsupported kind, missing extension) |
| `404` | `schema_not_found` | Referenced schema name does not exist in loaded schemas |
| `405` | `method_not_allowed` | Wrong HTTP method for the endpoint |
| `413` | `body_too_large` | Request body exceeds `--max-body` limit |
| `500` | `internal_error` | Unexpected server error |

### Example error responses

Malformed JSON:

```bash
curl -X POST http://localhost:8080/validate/user \
  -H "Content-Type: application/json" \
  -d 'not valid json'
```

```json
HTTP/1.1 400 Bad Request

{
  "error": "invalid_request",
  "message": "Request body contains malformed JSON."
}
```

Invalid schema:

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"schema": {"root": {"kind": "bogus"}}, "data": 42}'
```

```json
HTTP/1.1 400 Bad Request

{
  "error": "invalid_schema",
  "message": "Unknown schema kind 'bogus'."
}
```

---

## Docker Deployment

### Basic usage

```bash
docker run --rm \
  -p 8080:8080 \
  -v $(pwd)/schemas:/schemas:ro \
  anyvali/cli serve -s /schemas
```

### Docker Compose with schemas volume

```yaml
version: "3.8"

services:
  anyvali:
    image: anyvali/cli
    command: serve -s /schemas -p 8080
    ports:
      - "8080:8080"
    volumes:
      - ./schemas:/schemas:ro
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
    deploy:
      resources:
        limits:
          memory: 128M
          cpus: "0.5"
```

### Kubernetes deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: anyvali
spec:
  replicas: 2
  selector:
    matchLabels:
      app: anyvali
  template:
    metadata:
      labels:
        app: anyvali
    spec:
      containers:
        - name: anyvali
          image: anyvali/cli:latest
          command: ["anyvali", "serve", "-s", "/schemas", "-p", "8080"]
          ports:
            - containerPort: 8080
          volumeMounts:
            - name: schemas
              mountPath: /schemas
              readOnly: true
          resources:
            requests:
              memory: "64Mi"
              cpu: "100m"
            limits:
              memory: "128Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 2
            periodSeconds: 5
      volumes:
        - name: schemas
          configMap:
            name: anyvali-schemas
---
apiVersion: v1
kind: Service
metadata:
  name: anyvali
spec:
  selector:
    app: anyvali
  ports:
    - port: 8080
      targetPort: 8080
```

Load schemas from a ConfigMap:

```bash
kubectl create configmap anyvali-schemas \
  --from-file=user.schema.json=schemas/user.schema.json \
  --from-file=order.schema.json=schemas/order.schema.json
```

---

## Performance

### Schema caching

Schemas loaded from the `--schemas` directory are parsed and compiled once at startup. Subsequent validation requests reuse the compiled schema -- there is no per-request parsing overhead for pre-loaded schemas.

Inline schemas sent via `POST /validate` are parsed on each request. For high-throughput use cases, prefer loading schemas from the directory and using `POST /validate/:name` or `schemaRef`.

### Concurrency

The server uses Go's standard HTTP server, which spawns a goroutine per incoming connection. Validation is CPU-bound and runs concurrently without locks on read-only compiled schemas. The server handles many concurrent requests efficiently with minimal memory overhead.

### Recommended resource limits

| Scenario | Memory | CPU |
|----------|--------|-----|
| Small schemas, low traffic | 32-64 MB | 0.1 CPU |
| Medium schemas, moderate traffic | 64-128 MB | 0.25-0.5 CPU |
| Large schemas or high concurrency | 128-256 MB | 0.5-1.0 CPU |

Memory usage scales with the number and size of loaded schemas. Each concurrent validation allocates temporarily for input parsing and issue collection, then releases.

---

## Integration Examples

### Curl

```bash
# Validate inline
curl -s -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "schema": {
      "anyvaliVersion": "1.0",
      "schemaVersion": "1",
      "root": { "kind": "string", "minLength": 1 },
      "definitions": {},
      "extensions": {}
    },
    "data": "hello"
  }' | jq .

# Validate against a named schema
curl -s -X POST http://localhost:8080/validate/user \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}' | jq .

# Check if valid (scripting)
if curl -s -X POST http://localhost:8080/validate/user \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}' | jq -e '.valid' > /dev/null; then
  echo "Valid"
else
  echo "Invalid"
fi
```

### JavaScript (fetch)

```javascript
async function validate(schemaName, data) {
  const response = await fetch(`http://localhost:8080/validate/${schemaName}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });

  const result = await response.json();

  if (!result.valid) {
    console.error("Validation failed:");
    for (const issue of result.issues) {
      console.error(`  ${issue.path.join(".")}: [${issue.code}] ${issue.message}`);
    }
  }

  return result;
}

// Usage
const result = await validate("user", {
  name: "Alice",
  email: "alice@example.com",
});
```

### Python (requests)

```python
import requests

ANYVALI_URL = "http://localhost:8080"

def validate(schema_name: str, data: dict) -> dict:
    """Validate data against a named schema."""
    resp = requests.post(
        f"{ANYVALI_URL}/validate/{schema_name}",
        json=data,
    )
    resp.raise_for_status()
    return resp.json()

def validate_or_raise(schema_name: str, data: dict) -> None:
    """Validate and raise ValueError if invalid."""
    result = validate(schema_name, data)
    if not result["valid"]:
        messages = [
            f"{'.'.join(str(p) for p in i['path'])}: [{i['code']}] {i['message']}"
            for i in result["issues"]
        ]
        raise ValueError(f"Validation failed:\n" + "\n".join(messages))

# Usage
result = validate("user", {"name": "Alice", "email": "alice@example.com"})
print(result)  # {"valid": true, "issues": []}

# Or raise on failure
validate_or_raise("order", {"items": [], "total": -5})
# ValueError: Validation failed:
#   items: [too_small] Array must have at least 1 item
#   total: [too_small] Number must be >= 0
```

### Go

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ValidationResult struct {
	Valid  bool    `json:"valid"`
	Issues []Issue `json:"issues"`
}

type Issue struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Path     []any  `json:"path"`
	Expected any    `json:"expected"`
	Received any    `json:"received"`
}

func validate(schemaName string, data any) (*ValidationResult, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:8080/validate/%s", schemaName),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ValidationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
```

### Using as a sidecar in microservices

The AnyVali server works well as a sidecar container. Each service gets its own validation instance with only the schemas it needs.

```yaml
# Kubernetes pod with sidecar
apiVersion: v1
kind: Pod
metadata:
  name: api-with-validator
spec:
  containers:
    - name: api
      image: myapp/api:latest
      env:
        - name: VALIDATOR_URL
          value: "http://localhost:8080"
      ports:
        - containerPort: 3000

    - name: validator
      image: anyvali/cli:latest
      command: ["anyvali", "serve", "-s", "/schemas", "-p", "8080"]
      volumeMounts:
        - name: schemas
          mountPath: /schemas
          readOnly: true
      resources:
        requests:
          memory: "32Mi"
          cpu: "50m"
        limits:
          memory: "64Mi"
          cpu: "200m"

  volumes:
    - name: schemas
      configMap:
        name: api-schemas
```

Your application code calls `localhost:8080` for validation, which never leaves the pod network:

```python
# In your application
import requests

def validate_request(data: dict) -> bool:
    resp = requests.post("http://localhost:8080/validate/create-user", json=data)
    return resp.json()["valid"]
```

### Reverse proxy for auth

#### Nginx

```nginx
upstream anyvali {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl;
    server_name validator.example.com;

    ssl_certificate     /etc/ssl/certs/validator.pem;
    ssl_certificate_key /etc/ssl/private/validator.key;

    location / {
        # Basic auth
        auth_basic "AnyVali API";
        auth_basic_user_file /etc/nginx/.htpasswd;

        proxy_pass http://anyvali;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        # Health check without auth
        proxy_pass http://anyvali;
    }
}
```

#### Caddy

```
validator.example.com {
    basicauth / {
        admin $2a$14$...hashed_password...
    }

    @health path /health
    handle @health {
        reverse_proxy localhost:8080
    }

    handle {
        reverse_proxy localhost:8080
    }
}
```

---

## Further Reading

- **[CLI Reference](https://docs.anyvali.com/docs/cli)** -- all CLI commands and flags
- **[Portability Guide](https://docs.anyvali.com/docs/portability-guide)** -- design schemas that work across all 10 languages
- **[Product Overview](https://docs.anyvali.com/docs/overview)** -- architecture and design decisions
- **[Canonical Spec](https://docs.anyvali.com/spec/spec)** -- the normative specification
