# Development Guide

How to build, test, and contribute to AnyVali locally.

## Prerequisites

The full repository spans multiple toolchains. Install what you need:

| SDK | Toolchain |
|-----|-----------|
| JS/TS | Node.js 18+ |
| Python | Python 3.10+ |
| Go | Go 1.21+ |
| Java | JDK 17+ |
| C# | .NET 8 SDK |
| Rust | Rust toolchain (stable) |
| PHP | PHP 8.1+ |
| Ruby | Ruby 3.0+ |
| Kotlin | JDK 17+ and Gradle |
| C++ | CMake 3.14+ and a C++17 compiler |

You don't need all of them. The runner skips nothing implicitly -- if a toolchain is missing, it fails loudly.

## Using the Runner

The root `runner.sh` is the main entry point for all operations:

```bash
chmod +x runner.sh
./runner.sh help
```

### Common commands

```bash
# Install dependencies for all SDKs
./runner.sh install

# Build all SDKs
./runner.sh build

# Test all SDKs
./runner.sh test

# Test a specific SDK
./runner.sh test js
./runner.sh test python
./runner.sh test go

# Run the full CI pipeline locally
./runner.sh ci

# Check release readiness
./runner.sh release-check
```

## Running Tests

### JavaScript / TypeScript

```bash
cd sdk/js
npm install
npm test

# With coverage
npx vitest run --coverage
```

### Python

```bash
cd sdk/python
pip install -e ".[dev]"
PYTHONPATH=src pytest tests/

# With coverage
PYTHONPATH=src pytest tests/ --cov=anyvali --cov-report=term-missing
```

### Go

```bash
cd sdk/go
go test ./...

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Java

```bash
cd sdk/java
mvn test
```

### C#

```bash
cd sdk/csharp
dotnet test
```

### Rust

```bash
cd sdk/rust
cargo test
```

### PHP

```bash
cd sdk/php
composer install
./vendor/bin/phpunit
```

### Ruby

```bash
cd sdk/ruby
bundle install
rake test
```

### Kotlin

```bash
cd sdk/kotlin
./gradlew test
```

### C++

```bash
cd sdk/cpp
mkdir build && cd build
cmake ..
cmake --build .
ctest
```

## CLI & HTTP API Server

The CLI and HTTP API server are in the `cli/` directory, written in Go.

### Build

```bash
cd cli
go build -o anyvali .
```

### Cross-compile

```bash
cd cli

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o anyvali-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o anyvali-linux-arm64 .

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o anyvali-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o anyvali-darwin-arm64 .

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o anyvali-windows-amd64.exe .
```

### Test

```bash
cd cli
go test ./...
```

### Docker

```bash
docker build -f cli/Dockerfile -t anyvali/cli .
docker run --rm anyvali/cli version
docker run --rm -p 8080:8080 -v $(pwd)/schemas:/schemas anyvali/cli serve --schemas /schemas
```

## Docs Site

The repository includes a static docs site powered by MkDocs Material and served by Angie in Docker.

### Local preview

```bash
python docs-site/build_docs.py
pip install -r docs-site/requirements.txt
mkdocs serve
```

### Docker build

```bash
docker build -t anyvali-docs .
docker run --rm -p 8080:80 anyvali-docs
```

Coolify can deploy the repository directly with the root `Dockerfile`.

## CI, Releases, and Changelogs

GitHub Actions workflows are included for:

- **ci.yml** -- Multi-SDK CI with matrix testing
- **release-please.yml** -- Automated release PRs and changelog generation

Release and changelog management use [Release Please](https://github.com/googleapis/release-please) with manifest-based versioning so each package can be versioned independently.

## Commit Messages

Use conventional commits:

```text
feat(js): add tuple schema export
fix(go): reject invalid uint32 coercion
docs: clarify numeric alias defaults
chore: update CI node version
```

## Adding a New Schema Kind

1. Add the kind to `spec/spec.md`
2. Add node shape to `spec/json-format.md`
3. Add test cases to `spec/corpus/`
4. Implement in all 10 SDKs
5. Update docs

## Adding a New SDK

See the [SDK Authors Guide](sdk-authors-guide.md) for implementation requirements.
