#!/usr/bin/env bash
# Validate the supported Composer path installation from a source checkout.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
source_dir="${1:-$repo_root/sdk/php}"
version="${2:-0.0.0}"
source_dir="$(cd "$source_dir" && pwd)"
consumer_dir="$(mktemp -d)"
trap 'rm -rf "$consumer_dir"' EXIT

php -- "$source_dir" "$consumer_dir" "$version" <<'PHP'
<?php
[$script, $source, $consumer, $version] = $argv;
$manifest = [
    'name' => 'anyvali/install-smoke',
    'repositories' => [
        ['type' => 'path', 'url' => $source, 'options' => [
            'symlink' => false, 'versions' => ['anyvali/anyvali' => $version],
        ]],
        ['packagist.org' => false],
    ],
    'require' => ['anyvali/anyvali' => $version],
];
file_put_contents($consumer . '/composer.json', json_encode($manifest, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES | JSON_THROW_ON_ERROR));
PHP

composer install --working-dir="$consumer_dir" --no-interaction --no-dev --prefer-dist --no-progress
composer check-platform-reqs --working-dir="$consumer_dir" --no-dev
php -- "$consumer_dir" "$version" <<'PHP'
<?php
[$script, $consumer, $version] = $argv;
require $consumer . '/vendor/autoload.php';
if (is_link($consumer . '/vendor/anyvali/anyvali')) {
    throw new RuntimeException('Package must be mirrored, not symlinked');
}
if (Composer\InstalledVersions::getPrettyVersion('anyvali/anyvali') !== $version) {
    throw new RuntimeException('Installed version differs from the requested source version');
}
$schema = AnyVali\AnyVali::string()->minLength(2);
if ($schema->parse('consumer') !== 'consumer' || $schema->safeParse(42)->success) {
    throw new RuntimeException('Installed SDK failed validation smoke test');
}
if (!AnyVali\AnyVali::import($schema->export()->toJson())->safeParse('consumer')->success) {
    throw new RuntimeException('Installed SDK failed interchange smoke test');
}
echo "PHP clean-consumer install passed ($version)\n";
PHP
