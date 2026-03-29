# AnyVali CLI Reference

The AnyVali CLI validates data against AnyVali schemas from the command line. It reads exported AnyVali JSON schema documents and validates JSON input -- no SDK or programming language required.

---

## Installation

### Pre-built binary

Download the latest release for your platform from [GitHub Releases](https://github.com/anyvali/anyvali/releases):

```bash
# macOS (Apple Silicon)
curl -Lo anyvali https://github.com/anyvali/anyvali/releases/latest/download/anyvali-darwin-arm64
chmod +x anyvali
sudo mv anyvali /usr/local/bin/

# macOS (Intel)
curl -Lo anyvali https://github.com/anyvali/anyvali/releases/latest/download/anyvali-darwin-amd64
chmod +x anyvali
sudo mv anyvali /usr/local/bin/

# Linux (x86_64)
curl -Lo anyvali https://github.com/anyvali/anyvali/releases/latest/download/anyvali-linux-amd64
chmod +x anyvali
sudo mv anyvali /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/anyvali/anyvali/releases/latest/download/anyvali-windows-amd64.exe -OutFile anyvali.exe
```

### Build from source

Requires Go 1.21+.

```bash
cd cli
go build -o anyvali .
```

### Docker

```bash
docker pull anyvali/cli
docker run --rm anyvali/cli version
```

---

## Commands

### `anyvali validate`

Validate JSON input against an AnyVali schema document.

```
anyvali validate [flags]
```

#### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--schema` | `-s` | (required) | Path to the AnyVali schema document (JSON file) |
| `--input` | `-i` | stdin | Path to the JSON input file |
| `--data` | `-d` | | Inline JSON string to validate |
| `--format` | `-f` | `text` | Output format: `text`, `json`, or `quiet` |
| `--schema-name` | `-n` | | Name of a schema in the definitions (validates against a definition instead of root) |

#### Inline JSON input

```bash
anyvali validate -s user.schema.json -d '{"name": "Alice", "email": "alice@test.com"}'
```

Output on success:

```
OK: input is valid.
```

Output on failure:

```
INVALID: 2 issue(s) found.

  name: [too_small] String must have at least 1 character
  email: [invalid_string] Invalid email format
```

#### File input

```bash
anyvali validate -s user.schema.json -i request.json
```

#### Stdin piping

```bash
cat request.json | anyvali validate -s user.schema.json
```

```bash
echo '{"name": "Bob", "email": "bob@example.com"}' | anyvali validate -s user.schema.json
```

Useful for piping output from other tools:

```bash
curl -s https://api.example.com/user/42 | anyvali validate -s user.schema.json
```

#### JSON output format

Machine-readable output for scripting and CI:

```bash
anyvali validate -s user.schema.json -d '{"name": ""}' -f json
```

Success output:

```json
{
  "valid": true,
  "issues": []
}
```

Failure output:

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
    }
  ]
}
```

#### Text output format

The default. Human-readable, one issue per line:

```bash
anyvali validate -s user.schema.json -d '{"name": ""}' -f text
```

```
INVALID: 1 issue(s) found.

  name: [too_small] String must have at least 1 character
```

#### Quiet mode

No output. Only the exit code matters:

```bash
anyvali validate -s user.schema.json -d '{"name": "Alice", "email": "a@b.com"}' -f quiet
echo $?
# 0
```

```bash
anyvali validate -s user.schema.json -d '{"name": ""}' -f quiet
echo $?
# 1
```

#### Using in CI pipelines

```yaml
# GitHub Actions
- name: Validate config
  run: |
    anyvali validate -s schemas/config.schema.json -i config.json -f quiet
```

```yaml
# GitLab CI
validate-config:
  script:
    - anyvali validate -s schemas/config.schema.json -i config.json -f json
  allow_failure: false
```

#### Using in pre-commit hooks

Add to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: anyvali-validate
        name: Validate config files
        entry: anyvali validate -s schemas/config.schema.json -i
        files: 'config\.json$'
        language: system
```

Or as a Git hook directly (`.git/hooks/pre-commit`):

```bash
#!/bin/sh
set -e

for f in $(git diff --cached --name-only --diff-filter=ACM -- '*.json'); do
  if echo "$f" | grep -q 'config'; then
    anyvali validate -s schemas/config.schema.json -i "$f" -f quiet
  fi
done
```

#### Batch validation with shell loops

Validate every JSON file in a directory:

```bash
for f in data/*.json; do
  echo "--- $f ---"
  anyvali validate -s schema.json -i "$f"
done
```

Validate and collect failures:

```bash
failed=0
for f in data/*.json; do
  if ! anyvali validate -s schema.json -i "$f" -f quiet; then
    echo "FAIL: $f"
    failed=$((failed + 1))
  fi
done
echo "$failed file(s) failed validation."
exit $failed
```

---

### `anyvali serve`

Start an HTTP validation server that exposes a REST API for validating data against schemas.

```
anyvali serve [flags]
```

#### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--port` | `-p` | `8080` | Port to listen on |
| `--host` | | `0.0.0.0` | Host/address to bind to |
| `--schemas` | `-s` | | Directory of schema files to load at startup |
| `--cors` | | `false` | Enable CORS headers for browser access |
| `--cors-origins` | | `*` | Allowed CORS origins (comma-separated) |
| `--read-timeout` | | `30s` | HTTP read timeout |
| `--write-timeout` | | `30s` | HTTP write timeout |
| `--max-body` | | `1MB` | Maximum request body size |

#### Starting with a schemas directory

```bash
mkdir schemas
# Copy exported AnyVali schema documents into the directory
cp user.schema.json schemas/
cp order.schema.json schemas/

anyvali serve -s schemas/ -p 3000
```

Schema files are loaded at startup. The file name (without `.schema.json` or `.json` extension) becomes the schema name used in the `/validate/:name` endpoint:

```
schemas/
  user.schema.json    -> /validate/user
  order.schema.json   -> /validate/order
  config.json         -> /validate/config
```

#### Using with Docker

```bash
docker run --rm -p 8080:8080 -v $(pwd)/schemas:/schemas anyvali/cli serve -s /schemas
```

#### Docker Compose example

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
```

#### Health check configuration

The server exposes `GET /health` automatically. Use it for container orchestration:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

#### CORS for browser access

Enable CORS when calling the API from a browser:

```bash
anyvali serve -s schemas/ --cors --cors-origins "https://app.example.com,https://staging.example.com"
```

Allow all origins (development only):

```bash
anyvali serve -s schemas/ --cors
```

#### Using as a sidecar validation service

Run alongside your application in a pod or Docker Compose stack. Your application sends validation requests to `localhost:8080`:

```yaml
# docker-compose.yml
services:
  api:
    build: .
    environment:
      VALIDATOR_URL: http://validator:8080
    depends_on:
      validator:
        condition: service_healthy

  validator:
    image: anyvali/cli
    command: serve -s /schemas -p 8080
    volumes:
      - ./schemas:/schemas:ro
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
```

---

### `anyvali inspect`

Display detailed information about an AnyVali schema document.

```
anyvali inspect <schema-file>
```

#### Example output

```bash
anyvali inspect user.schema.json
```

```
Schema: user.schema.json
AnyVali Version: 1.0
Schema Version: 1

Root: object
  Properties:
    name       string  required  minLength=1, maxLength=100
    email      string  required  format=email
    age        int     optional  min=0, max=150
    tags       array   optional  items=string, maxItems=10

  Unknown Keys: reject

Definitions: (none)
Extensions: (none)
Portability: fully portable
```

For schemas with definitions:

```bash
anyvali inspect tree.schema.json
```

```
Schema: tree.schema.json
AnyVali Version: 1.0
Schema Version: 1

Root: ref -> #/definitions/TreeNode

Definitions:
  TreeNode (object)
    value      string  required
    children   array   optional  items=ref(#/definitions/TreeNode)

Extensions: (none)
Portability: fully portable
```

---

### `anyvali check`

Check whether a schema document is fully portable across all AnyVali SDKs.

```
anyvali check <schema-file>
```

Use this before distributing a schema to teams using different languages. It verifies that the schema only uses Tier 1 (portable core) features.

#### Fully portable

```bash
anyvali check user.schema.json
```

```
OK: schema is fully portable (Tier 1).
```

#### Has extensions

```bash
anyvali check extended-user.schema.json
```

```
WARNING: schema uses extensions (Tier 2).

  Extensions found:
    js: custom transform metadata
    go: struct tag hints

  This schema will only import correctly in SDKs that understand
  the "js" and "go" extension namespaces.
```

#### Non-portable

```bash
anyvali check local-schema.json
```

```
ERROR: schema is not portable.

  Issues:
    root.properties.name: uses unsupported schema kind "branded"
```

---

### `anyvali version`

Print the CLI version and the AnyVali spec version it supports.

```bash
anyvali version
```

```
anyvali v0.1.0
spec: AnyVali 1.0
go: go1.21.0
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Valid / success |
| `1` | Invalid (validation failed, schema not portable) |
| `2` | Error (bad schema, bad input, missing file, IO error) |

These exit codes are stable and safe for scripting:

```bash
anyvali validate -s schema.json -i data.json -f quiet
case $? in
  0) echo "Valid" ;;
  1) echo "Validation failed" ;;
  2) echo "Error (check schema or input file)" ;;
esac
```

---

## Examples

### CI pipeline validation

Validate all config files in a repository on every push:

```yaml
# .github/workflows/validate.yml
name: Validate configs
on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install anyvali
        run: |
          curl -Lo anyvali https://github.com/anyvali/anyvali/releases/latest/download/anyvali-linux-amd64
          chmod +x anyvali
          sudo mv anyvali /usr/local/bin/

      - name: Validate app config
        run: anyvali validate -s schemas/config.schema.json -i config.json

      - name: Validate feature flags
        run: anyvali validate -s schemas/features.schema.json -i features.json

      - name: Validate all data files
        run: |
          for f in data/*.json; do
            anyvali validate -s schemas/data.schema.json -i "$f" -f quiet || exit 1
          done
```

### Pre-commit hook

Reject commits that include invalid JSON configs:

```bash
#!/bin/sh
# .git/hooks/pre-commit
set -e

staged_configs=$(git diff --cached --name-only --diff-filter=ACM -- '*.config.json')

for f in $staged_configs; do
  if ! anyvali validate -s schemas/config.schema.json -i "$f" -f quiet; then
    echo "Validation failed for $f"
    exit 1
  fi
done
```

### Config file validation

Validate application configuration before starting a service:

```bash
#!/bin/sh
# start.sh
set -e

echo "Validating configuration..."
anyvali validate -s schemas/app-config.schema.json -i /etc/myapp/config.json

echo "Starting application..."
exec ./myapp serve
```

### API request validation sidecar

Use the AnyVali server as a validation sidecar. Your application sends request bodies to the sidecar before processing them:

```python
# Python application using anyvali sidecar
import requests

def validate_request(schema_name: str, data: dict) -> bool:
    resp = requests.post(
        f"http://localhost:8080/validate/{schema_name}",
        json=data,
    )
    result = resp.json()
    return result["valid"]

# In your request handler:
if not validate_request("create-user", request.json):
    return {"error": "Invalid request"}, 400
```

### Shell scripting with anyvali

Generate a validation report for a directory of files:

```bash
#!/bin/bash
# validate-report.sh
schema="$1"
dir="$2"

total=0
passed=0
failed=0

for f in "$dir"/*.json; do
  total=$((total + 1))
  result=$(anyvali validate -s "$schema" -i "$f" -f json 2>&1)

  if echo "$result" | grep -q '"valid": true'; then
    passed=$((passed + 1))
  else
    failed=$((failed + 1))
    echo "FAIL: $f"
    echo "$result" | head -20
    echo ""
  fi
done

echo "==========================="
echo "Total: $total  Passed: $passed  Failed: $failed"
exit $failed
```

Run it:

```bash
chmod +x validate-report.sh
./validate-report.sh schemas/event.schema.json events/
```

---

## Further Reading

- **[HTTP API Reference](https://docs.anyvali.com/docs/api)** -- full REST API documentation for `anyvali serve`
- **[Portability Guide](https://docs.anyvali.com/docs/portability-guide)** -- design schemas that work across all 10 languages
- **[Product Overview](https://docs.anyvali.com/docs/overview)** -- architecture and design decisions
- **[Canonical Spec](https://docs.anyvali.com/spec/spec)** -- the normative specification
