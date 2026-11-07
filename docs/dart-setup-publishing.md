# Dart and Flutter setup and publishing

AnyVali's Dart package lives in `sdk/dart` and works without a Flutter dependency.
Install Flutter when building a Flutter app; install standalone Dart when working
only on the validation package. The package supports Dart 3.4 and later within
Dart 3, and CI exercises 3.4.4 and stable on native and JavaScript targets.

The package is currently distributed from source. Its initial pub.dev upload
requires an authorized Google account; subsequent releases can use GitHub OIDC.
The instructions below prepare publishing without enabling an upload automatically.

## Install the development tools

For Flutter, follow the [official installation instructions](https://docs.flutter.dev/install)
for your OS, add Flutter's `bin` directory to PATH, then check the installation:

```bash
flutter --version
dart --version
flutter doctor -v
```

Flutter includes the full Dart SDK. Use that bundled version for the Flutter app
rather than adding a second Dart installation. Complete the platform toolchain
steps that `flutter doctor` reports for your intended targets. Android requires
Android tooling; iOS development requires macOS and Xcode.

For standalone Dart, use the [official Dart installer](https://dart.dev/get-dart)
or a stable [SDK archive](https://dart.dev/get-dart/archive), add `dart-sdk/bin`
to PATH, and verify `dart --version`. Install Node.js 24 to run the JavaScript
target tests. The minimum Dart version is a compatibility target; use current
stable for daily development.

From the AnyVali repository root:

```bash
bash runner.sh ci dart
cd sdk/dart
dart test --platform node test/schema_test.dart test/robustness_test.dart test/pattern_test.dart
dart run example/main.dart
dart pub publish --dry-run
```

The VM suite reads both corpora under `spec/`; run it from a complete repository
checkout. That repository-only test is excluded from the package archive; the
standalone schema and robustness tests remain included. The package dry run
lists the actual upload contents and performs
package validation. It does not establish ownership of the package name.

## Add AnyVali to a Flutter app

Before pub.dev publication, use a local path dependency in your app's
`pubspec.yaml`:

```yaml
dependencies:
  flutter:
    sdk: flutter
  anyvali:
    path: ../AnyVali/sdk/dart
```

Adjust the path relative to the app's `pubspec.yaml`, then run `flutter pub get`.
For a repository dependency after the SDK commit is available remotely, pin an
actual reviewed commit rather than a moving branch:

```yaml
dependencies:
  anyvali:
    git:
      url: https://github.com/BetterCorp/AnyVali.git
      ref: REPLACE_WITH_REVIEWED_COMMIT
      path: sdk/dart
```

Use a field schema as a Flutter validator:

```dart
import 'package:anyvali/anyvali.dart' as av;
import 'package:flutter/material.dart';

final email = av.string().minLength(1).format('email');
final emailField = TextFormField(validator: email.validator());
```

See the [Dart SDK guide](sdk-dart.md) for complete schema usage, numeric limits,
and encrypted-storage behavior. This pub.dev process distributes a library;
releasing your Flutter application to app stores is a separate process.

## Prepare the first pub.dev release

1. Sign in to [pub.dev](https://pub.dev) with the Google account that will own
   the initial package. Confirm that the name `anyvali` is available, or that
   the account has uploader access if the package already exists. GitHub
   repository access does not confer pub.dev publishing rights.
2. Review `sdk/dart/pubspec.yaml`, the package README, CHANGELOG, and LICENSE.
   Select the release version and commit it with the reviewed source. The
   current source version is `0.0.1`; align it with the intended unified
   release version before adopting the tag workflow below.
3. Run the native and JavaScript checks above and inspect the dry-run file
   list. The SDK needs no runtime secrets or generated Flutter artifacts.
4. From the reviewed checkout, run the interactive publishing command:

   ```bash
   cd sdk/dart
   dart pub publish
   ```

   Follow the browser authentication link using the authorized Google account
   and confirm the upload contents. This command publishes a public version.
5. Verify the version and package contents on pub.dev, then test consumption
   from a separate Dart or Flutter app with `dart pub add anyvali` or
   `flutter pub add anyvali`. Update installation docs only after publication.

Pub.dev requires a first publication before automated publishing can be enabled.
New packages are initially uploaded through a Google account. For organization
ownership, create or use a verified publisher for a domain BetterCorp controls
and transfer the package to it through the package's Admin tab. Creating the
publisher requires domain ownership verification through Google Search Console.
See [publishing](https://dart.dev/tools/pub/publishing) and
[verified publishers](https://dart.dev/tools/pub/verified-publishers).

Published versions cannot be overwritten. Fixes require a new version.

## Configure GitHub OIDC after the first publication

The current `build-release.yml` does not publish Dart. To add Dart to future
unified releases, configure the existing package's Admin tab on pub.dev with:

| Setting | Value |
| --- | --- |
| Automated publishing | GitHub Actions |
| Repository | `BetterCorp/AnyVali` |
| Tag pattern | `v{{version}}` |
| Optional deployment environment | Use the same environment name in pub.dev and the workflow |

Configure these settings as a package uploader or publisher admin. No pub.dev
API token or Google service-account key is needed for GitHub OIDC. Pub.dev
accepts this flow only from matching tag-push events, not a branch push or
manual workflow dispatch. See the [official automation guide](https://dart.dev/tools/pub/automated-publishing).

After completing that configuration, the following can be saved as
`.github/workflows/publish-dart.yml`. It uses the repository's unified `v*`
release tags, including prereleases such as `v1.2.3-rc.1` and pins the Dart team's reusable publishing workflow. The initial
PR adds this example to documentation only.

```yaml
name: Publish Dart
on:
  push:
    tags:
      - 'v*'
permissions:
  contents: read
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7
      - uses: dart-lang/setup-dart@6afc89df92d6eb3834022f73cd65adc8cdfcb92d # v1
      - uses: actions/setup-node@949feb2413d6458794dcd2491c4babbbce0c15c1 # v7
        with:
          node-version: 24
      - name: Check committed package version matches tag
        shell: bash
        run: |
          package_version=$(sed -n 's/^version: *//p' sdk/dart/pubspec.yaml)
          test "$package_version" = "${GITHUB_REF_NAME#v}"
      - run: bash runner.sh ci dart
      - run: cd sdk/dart && dart test --platform node test/schema_test.dart test/robustness_test.dart test/pattern_test.dart
      - run: cd sdk/dart && dart pub publish --dry-run
  publish:
    needs: validate
    permissions:
      contents: read
      id-token: write
    uses: dart-lang/setup-dart/.github/workflows/publish.yml@6afc89df92d6eb3834022f73cd65adc8cdfcb92d # v1
    with:
      working-directory: sdk/dart
      # environment: pub.dev
```

If pub.dev requires a GitHub environment, create that environment in GitHub and
pass its exact name in the reusable workflow's `environment` input. Existing
repository environment rules still apply.

For each release, commit the matching `pubspec.yaml` version and changelog,
merge through the normal review process, and follow the [release process](release-process.md)
to create a signed unified tag on the verified commit. The package version must
already match the tag; do not reuse an existing tag or bump only in a temporary
publishing workspace. Verify the publishing run and the new pub.dev version
before announcing availability. If `anyvali` belongs to another publisher,
resolve the package ownership or choose a new package name first.
