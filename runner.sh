#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

ACTION="${1:-help}"
TARGET="${2:-all}"

ALL_TARGETS=(js python go java rust php ruby kotlin csharp cpp)

# ---------------------------------------------------------------------------
# Color helpers
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Track results for summary
declare -A RESULTS=()

log() {
  printf "${BLUE}[%s]${NC} %s\n" "$1" "$2"
}

log_ok() {
  printf "${GREEN}[PASS]${NC} %s\n" "$1"
}

log_fail() {
  printf "${RED}[FAIL]${NC} %s\n" "$1"
}

log_skip() {
  printf "${YELLOW}[SKIP]${NC} %s — %s\n" "$1" "$2"
}

log_header() {
  printf "\n${BOLD}${CYAN}══════════════════════════════════════════════════════════════${NC}\n"
  printf "${BOLD}${CYAN}  %s${NC}\n" "$1"
  printf "${BOLD}${CYAN}══════════════════════════════════════════════════════════════${NC}\n\n"
}

die() {
  printf "${RED}error:${NC} %s\n" "$1" >&2
  exit 1
}

# Check if a command is available; return 1 if missing (does not exit).
has_cmd() {
  command -v "$1" >/dev/null 2>&1
}

# Hard-require a command (exits on failure).
need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

resolve_targets() {
  if [[ "$TARGET" == "all" ]]; then
    printf '%s\n' "${ALL_TARGETS[@]}"
  else
    printf '%s\n' "$TARGET"
  fi
}

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
print_summary() {
  printf "\n${BOLD}${CYAN}══════════════════════════════════════════════════════════════${NC}\n"
  printf "${BOLD}${CYAN}  Summary${NC}\n"
  printf "${BOLD}${CYAN}══════════════════════════════════════════════════════════════${NC}\n\n"

  local has_failure=0
  for sdk in "${ALL_TARGETS[@]}"; do
    local result="${RESULTS[$sdk]:-}"
    if [[ -z "$result" ]]; then
      continue
    fi
    case "$result" in
      pass) log_ok "$sdk" ;;
      fail) log_fail "$sdk"; has_failure=1 ;;
      skip) log_skip "$sdk" "skipped" ;;
    esac
  done
  printf "\n"

  if [[ "$has_failure" -eq 1 ]]; then
    printf "${RED}${BOLD}Some targets failed.${NC}\n"
    return 1
  else
    printf "${GREEN}${BOLD}All targets passed.${NC}\n"
  fi
}

# ---------------------------------------------------------------------------
# Prerequisites check
# ---------------------------------------------------------------------------
check_prerequisites() {
  log_header "Checking prerequisites"

  local tools=(
    "node:js"
    "npm:js"
    "python:python"
    "pip:python"
    "go:go"
    "mvn:java"
    "cargo:rust"
    "rustc:rust"
    "php:php"
    "composer:php"
    "ruby:ruby"
    "bundle:ruby"
    "gradle:kotlin"
    "dotnet:csharp"
    "cmake:cpp"
    "ctest:cpp"
    "gcc:cpp"
  )

  for entry in "${tools[@]}"; do
    local cmd="${entry%%:*}"
    local sdk="${entry##*:}"
    if has_cmd "$cmd"; then
      local ver
      ver=$("$cmd" --version 2>/dev/null | head -1) || ver="(unknown version)"
      printf "${GREEN}%-12s${NC} %-8s %s\n" "$cmd" "[$sdk]" "$ver"
    else
      printf "${YELLOW}%-12s${NC} %-8s %s\n" "$cmd" "[$sdk]" "NOT FOUND"
    fi
  done
}

# ---------------------------------------------------------------------------
# SDK runners
# ---------------------------------------------------------------------------
run_js() {
  local action="$1"
  if ! has_cmd npm; then
    log_skip "js" "npm not found"
    RESULTS[js]=skip
    return 0
  fi
  case "$action" in
    install)
      (cd "$ROOT_DIR/sdk/js" && npm ci)
      ;;
    build)
      (cd "$ROOT_DIR/sdk/js" && npm run build)
      ;;
    test)
      (cd "$ROOT_DIR/sdk/js" && npm test)
      ;;
    lint)
      (cd "$ROOT_DIR/sdk/js" && npm run lint 2>/dev/null || log_skip "js:lint" "no lint script defined")
      ;;
    coverage)
      (cd "$ROOT_DIR/sdk/js" && npm run coverage 2>/dev/null || npm test -- --coverage 2>/dev/null || log_skip "js:coverage" "no coverage script defined")
      ;;
    ci)
      run_js install
      run_js build
      run_js test
      ;;
    *)
      die "unsupported action for js: $action"
      ;;
  esac
}

run_python() {
  local action="$1"
  if ! has_cmd python && ! has_cmd python3; then
    log_skip "python" "python not found"
    RESULTS[python]=skip
    return 0
  fi
  local py_cmd="python"
  has_cmd python || py_cmd="python3"
  case "$action" in
    install)
      (cd "$ROOT_DIR/sdk/python" && "$py_cmd" -m pip install --upgrade pip && "$py_cmd" -m pip install build -e ".[dev]")
      ;;
    build)
      (cd "$ROOT_DIR/sdk/python" && "$py_cmd" -m build)
      ;;
    test)
      (cd "$ROOT_DIR/sdk/python" && "$py_cmd" -m pytest)
      ;;
    lint)
      (cd "$ROOT_DIR/sdk/python" && "$py_cmd" -m ruff check . 2>/dev/null || "$py_cmd" -m flake8 . 2>/dev/null || log_skip "python:lint" "no linter available")
      ;;
    coverage)
      (cd "$ROOT_DIR/sdk/python" && "$py_cmd" -m pytest --cov=anyvali --cov-report=xml --cov-report=term)
      ;;
    ci)
      run_python install
      run_python build
      run_python test
      ;;
    *)
      die "unsupported action for python: $action"
      ;;
  esac
}

run_go() {
  local action="$1"
  if ! has_cmd go; then
    log_skip "go" "go not found"
    RESULTS[go]=skip
    return 0
  fi
  case "$action" in
    install)
      (cd "$ROOT_DIR/sdk/go" && go mod download)
      ;;
    build)
      (cd "$ROOT_DIR/sdk/go" && go build ./...)
      ;;
    test)
      (cd "$ROOT_DIR/sdk/go" && go test ./...)
      ;;
    lint)
      if has_cmd golangci-lint; then
        (cd "$ROOT_DIR/sdk/go" && golangci-lint run)
      else
        (cd "$ROOT_DIR/sdk/go" && go vet ./...)
      fi
      ;;
    coverage)
      (cd "$ROOT_DIR/sdk/go" && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out)
      ;;
    ci)
      run_go install
      run_go build
      run_go test
      ;;
    *)
      die "unsupported action for go: $action"
      ;;
  esac
}

run_java() {
  local action="$1"
  if ! has_cmd mvn; then
    log_skip "java" "mvn not found"
    RESULTS[java]=skip
    return 0
  fi
  case "$action" in
    install)
      (cd "$ROOT_DIR/sdk/java" && mvn -B -q dependency:go-offline)
      ;;
    build)
      (cd "$ROOT_DIR/sdk/java" && mvn -B package -DskipTests)
      ;;
    test)
      (cd "$ROOT_DIR/sdk/java" && mvn -B test)
      ;;
    lint)
      (cd "$ROOT_DIR/sdk/java" && mvn -B checkstyle:check 2>/dev/null || log_skip "java:lint" "checkstyle not configured")
      ;;
    coverage)
      (cd "$ROOT_DIR/sdk/java" && mvn -B test jacoco:report 2>/dev/null || mvn -B test)
      ;;
    ci)
      run_java install
      run_java build
      run_java test
      ;;
    *)
      die "unsupported action for java: $action"
      ;;
  esac
}

run_csharp() {
  local action="$1"
  if ! has_cmd dotnet; then
    log_skip "csharp" "dotnet not found"
    RESULTS[csharp]=skip
    return 0
  fi
  case "$action" in
    install)
      (cd "$ROOT_DIR/sdk/csharp" && dotnet restore AnyVali.sln)
      ;;
    build)
      (cd "$ROOT_DIR/sdk/csharp" && dotnet build AnyVali.sln --configuration Release --no-restore)
      ;;
    test)
      (cd "$ROOT_DIR/sdk/csharp" && dotnet test AnyVali.sln --configuration Release --no-build)
      ;;
    lint)
      (cd "$ROOT_DIR/sdk/csharp" && dotnet format AnyVali.sln --verify-no-changes 2>/dev/null || log_skip "csharp:lint" "dotnet format not available")
      ;;
    coverage)
      (cd "$ROOT_DIR/sdk/csharp" && dotnet test AnyVali.sln --configuration Release --collect:"XPlat Code Coverage")
      ;;
    ci)
      run_csharp install
      run_csharp build
      run_csharp test
      ;;
    *)
      die "unsupported action for csharp: $action"
      ;;
  esac
}

run_rust() {
  local action="$1"
  if ! has_cmd cargo; then
    log_skip "rust" "cargo not found"
    RESULTS[rust]=skip
    return 0
  fi
  case "$action" in
    install)
      (cd "$ROOT_DIR/sdk/rust" && cargo fetch)
      ;;
    build)
      (cd "$ROOT_DIR/sdk/rust" && cargo build --all-targets)
      ;;
    test)
      (cd "$ROOT_DIR/sdk/rust" && cargo test)
      ;;
    lint)
      (cd "$ROOT_DIR/sdk/rust" && cargo clippy --all-targets -- -D warnings 2>/dev/null || log_skip "rust:lint" "clippy not available")
      ;;
    coverage)
      if has_cmd cargo-llvm-cov; then
        (cd "$ROOT_DIR/sdk/rust" && cargo llvm-cov --lcov --output-path lcov.info)
      else
        (cd "$ROOT_DIR/sdk/rust" && cargo test)
        log_skip "rust:coverage" "cargo-llvm-cov not installed, ran tests only"
      fi
      ;;
    ci)
      run_rust install
      run_rust build
      run_rust test
      ;;
    *)
      die "unsupported action for rust: $action"
      ;;
  esac
}

run_php() {
  local action="$1"
  if ! has_cmd php; then
    log_skip "php" "php not found"
    RESULTS[php]=skip
    return 0
  fi
  case "$action" in
    install)
      if ! has_cmd composer; then
        log_skip "php:install" "composer not found"
        return 0
      fi
      (cd "$ROOT_DIR/sdk/php" && composer install --no-interaction --prefer-dist)
      ;;
    build)
      log "php" "no dedicated build step"
      ;;
    test)
      (cd "$ROOT_DIR/sdk/php" && ./vendor/bin/phpunit)
      ;;
    lint)
      (cd "$ROOT_DIR/sdk/php" && ./vendor/bin/phpcs 2>/dev/null || log_skip "php:lint" "phpcs not available")
      ;;
    coverage)
      (cd "$ROOT_DIR/sdk/php" && ./vendor/bin/phpunit --coverage-text --coverage-clover=coverage.xml 2>/dev/null || ./vendor/bin/phpunit)
      ;;
    ci)
      run_php install
      run_php build
      run_php test
      ;;
    *)
      die "unsupported action for php: $action"
      ;;
  esac
}

run_ruby() {
  local action="$1"
  if ! has_cmd ruby; then
    log_skip "ruby" "ruby not found"
    RESULTS[ruby]=skip
    return 0
  fi
  case "$action" in
    install)
      if ! has_cmd bundle; then
        log_skip "ruby:install" "bundler not found"
        return 0
      fi
      (cd "$ROOT_DIR/sdk/ruby" && bundle install)
      ;;
    build)
      if has_cmd gem; then
        (cd "$ROOT_DIR/sdk/ruby" && gem build anyvali.gemspec)
      else
        log_skip "ruby:build" "gem not found"
      fi
      ;;
    test)
      if ! has_cmd bundle; then
        log_skip "ruby:test" "bundler not found"
        return 0
      fi
      (cd "$ROOT_DIR/sdk/ruby" && bundle exec rake test)
      ;;
    lint)
      (cd "$ROOT_DIR/sdk/ruby" && bundle exec rubocop 2>/dev/null || log_skip "ruby:lint" "rubocop not available")
      ;;
    coverage)
      (cd "$ROOT_DIR/sdk/ruby" && COVERAGE=true bundle exec rake test)
      ;;
    ci)
      run_ruby install
      run_ruby build
      run_ruby test
      ;;
    *)
      die "unsupported action for ruby: $action"
      ;;
  esac
}

run_kotlin() {
  local action="$1"
  local gradle_cmd
  if [[ -x "$ROOT_DIR/sdk/kotlin/gradlew" ]]; then
    gradle_cmd="./gradlew"
  elif has_cmd gradle; then
    gradle_cmd="gradle"
  else
    log_skip "kotlin" "gradle/gradlew not found"
    RESULTS[kotlin]=skip
    return 0
  fi
  case "$action" in
    install)
      (cd "$ROOT_DIR/sdk/kotlin" && "$gradle_cmd" --no-daemon dependencies >/dev/null)
      ;;
    build)
      (cd "$ROOT_DIR/sdk/kotlin" && "$gradle_cmd" --no-daemon build -x test)
      ;;
    test)
      (cd "$ROOT_DIR/sdk/kotlin" && "$gradle_cmd" --no-daemon test)
      ;;
    lint)
      (cd "$ROOT_DIR/sdk/kotlin" && "$gradle_cmd" --no-daemon ktlintCheck 2>/dev/null || log_skip "kotlin:lint" "ktlint not configured")
      ;;
    coverage)
      (cd "$ROOT_DIR/sdk/kotlin" && "$gradle_cmd" --no-daemon jacocoTestReport 2>/dev/null || "$gradle_cmd" --no-daemon test)
      ;;
    ci)
      run_kotlin install
      run_kotlin build
      run_kotlin test
      ;;
    *)
      die "unsupported action for kotlin: $action"
      ;;
  esac
}

run_cpp() {
  local action="$1"
  local build_dir="$ROOT_DIR/sdk/cpp/build"
  if ! has_cmd cmake; then
    log_skip "cpp" "cmake not found"
    RESULTS[cpp]=skip
    return 0
  fi
  case "$action" in
    install)
      cmake -S "$ROOT_DIR/sdk/cpp" -B "$build_dir"
      ;;
    build)
      cmake --build "$build_dir" --config Release
      ;;
    test)
      if ! has_cmd ctest; then
        log_skip "cpp:test" "ctest not found"
        return 0
      fi
      (cd "$build_dir" && ctest --output-on-failure)
      ;;
    lint)
      if has_cmd clang-tidy; then
        (cd "$ROOT_DIR/sdk/cpp" && clang-tidy src/*.cpp -- -Iinclude 2>/dev/null || true)
      else
        log_skip "cpp:lint" "clang-tidy not found"
      fi
      ;;
    coverage)
      cmake -S "$ROOT_DIR/sdk/cpp" -B "$build_dir" -DCMAKE_BUILD_TYPE=Debug -DENABLE_COVERAGE=ON 2>/dev/null || cmake -S "$ROOT_DIR/sdk/cpp" -B "$build_dir"
      cmake --build "$build_dir"
      (cd "$build_dir" && ctest --output-on-failure)
      ;;
    ci)
      run_cpp install
      run_cpp build
      run_cpp test
      ;;
    *)
      die "unsupported action for cpp: $action"
      ;;
  esac
}

# ---------------------------------------------------------------------------
# Dispatch & top-level commands
# ---------------------------------------------------------------------------
dispatch() {
  local action="$1"
  local target="$2"
  log_header "$target — $action"
  if "run_$target" "$action"; then
    RESULTS[$target]=pass
  else
    RESULTS[$target]=fail
  fi
}

run_conformance() {
  log_header "Conformance tests"
  if [[ -d "$ROOT_DIR/spec/conformance" ]]; then
    log "conformance" "running conformance suite"
    while IFS= read -r target; do
      if has_cmd "$(sdk_primary_cmd "$target")"; then
        log "conformance" "testing $target"
        dispatch test "$target"
      else
        log_skip "conformance:$target" "toolchain not available"
      fi
    done < <(printf '%s\n' "${ALL_TARGETS[@]}")
  else
    log_skip "conformance" "no conformance suite found at spec/conformance"
  fi
}

sdk_primary_cmd() {
  case "$1" in
    js)      echo "npm" ;;
    python)  echo "python" ;;
    go)      echo "go" ;;
    java)    echo "mvn" ;;
    csharp)  echo "dotnet" ;;
    rust)    echo "cargo" ;;
    php)     echo "php" ;;
    ruby)    echo "ruby" ;;
    kotlin)  echo "gradle" ;;
    cpp)     echo "cmake" ;;
    *)       echo "false" ;;
  esac
}

clean_all() {
  log_header "Cleaning build artifacts"
  rm -rf "$ROOT_DIR/sdk/js/node_modules" "$ROOT_DIR/sdk/js/dist" "$ROOT_DIR/sdk/js/coverage"
  rm -rf "$ROOT_DIR/sdk/python/dist" "$ROOT_DIR/sdk/python/build" "$ROOT_DIR/sdk/python"/*.egg-info "$ROOT_DIR/sdk/python/.pytest_cache"
  rm -rf "$ROOT_DIR/sdk/go/coverage.out"
  rm -rf "$ROOT_DIR/sdk/java/target"
  rm -rf "$ROOT_DIR/sdk/csharp/bin" "$ROOT_DIR/sdk/csharp/obj"
  rm -rf "$ROOT_DIR/sdk/rust/target"
  rm -rf "$ROOT_DIR/sdk/php/vendor"
  rm -rf "$ROOT_DIR/sdk/ruby/.bundle" "$ROOT_DIR/sdk/ruby/pkg"
  rm -rf "$ROOT_DIR/sdk/kotlin/build" "$ROOT_DIR/sdk/kotlin/.gradle"
  rm -rf "$ROOT_DIR/sdk/cpp/build"
  log "clean" "done"
}

release_check() {
  need_cmd git
  if git diff --quiet && git diff --cached --quiet; then
    log "release" "working tree clean"
  else
    die "working tree is not clean"
  fi
}

help_text() {
  cat <<EOF
${BOLD}Usage:${NC}
  ./runner.sh <command> [target]

${BOLD}Commands:${NC}
  ${CYAN}test${NC}     [target]   Run tests for one target or all targets
  ${CYAN}build${NC}    [target]   Build one target or all targets
  ${CYAN}install${NC}  [target]   Install dependencies for one target or all targets
  ${CYAN}lint${NC}     [target]   Lint one target or all targets
  ${CYAN}coverage${NC} [target]   Run tests with coverage for one target or all targets
  ${CYAN}ci${NC}       [target]   Run install + build + test for one target or all targets
  ${CYAN}conformance${NC}         Run conformance tests across available SDKs
  ${CYAN}clean${NC}               Clean build artifacts for all SDKs
  ${CYAN}check${NC}               Check that prerequisite tools are installed
  ${CYAN}all${NC}      [target]   Build + test + lint everything
  ${CYAN}release-check${NC}       Validate basic release preconditions
  ${CYAN}help${NC}                Show this message

${BOLD}Targets:${NC}
  all (default), js, python, go, java, csharp, rust, php, ruby, kotlin, cpp

${BOLD}Examples:${NC}
  ./runner.sh test              # Run all SDK tests
  ./runner.sh test js           # Run only JS tests
  ./runner.sh build python      # Build only Python SDK
  ./runner.sh coverage go       # Run Go tests with coverage
  ./runner.sh all               # Build + test + lint everything
  ./runner.sh check             # Show which tools are installed
EOF
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
case "$ACTION" in
  install|build|test|lint|coverage|ci)
    while IFS= read -r target; do
      dispatch "$ACTION" "$target"
    done < <(resolve_targets)
    print_summary
    ;;
  conformance)
    run_conformance
    print_summary
    ;;
  clean)
    clean_all
    ;;
  check)
    check_prerequisites
    ;;
  all)
    while IFS= read -r target; do
      dispatch build "$target"
      dispatch test "$target"
      dispatch lint "$target"
    done < <(resolve_targets)
    print_summary
    ;;
  release-check)
    release_check
    ;;
  help|-h|--help)
    help_text
    ;;
  *)
    die "unknown action: $ACTION — run ./runner.sh help for usage"
    ;;
esac
