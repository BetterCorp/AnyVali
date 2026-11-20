@TestOn('vm')
library;

import 'dart:convert';
import 'dart:io';
import 'package:anyvali/anyvali.dart' as av;
import 'package:test/test.dart';

/// Exercise both repository conformance corpora, round trips, and sensitive paths.
void main() {
  test(
      'JavaScript portable fixture retains ref coercions, schema edges and defaults',
      () {
    final fixture = jsonDecode(
        File('../go/testdata/js-ref-coercions.json').readAsStringSync()) as Map;
    final schema =
        av.importSchema(Map<String, dynamic>.from(fixture['schema']));
    expect(schema.parse(fixture['input']), fixture['output']);
    expect(av.importSchema(schema.exportSchema()).parse(fixture['input']),
        fixture['output']);
  });
  final roots = [
    Directory('../../spec/corpus'),
    Directory('../../spec/interchange-corpus')
  ];
  final files = roots
      .expand((d) => d.listSync(recursive: true).whereType<File>())
      .where((f) => f.path.endsWith('.json'))
      .toList()
    ..sort((a, b) => a.path.compareTo(b.path));
  if (files.isEmpty) throw StateError('Conformance corpus is missing');
  for (final file in files) {
    final suite = jsonDecode(file.readAsStringSync()) as Map;
    for (final c in suite['cases']) {
      test('${file.path}: ${c['description']}', () {
        var schema = av.importSchema(Map<String, dynamic>.from(c['schema']));
        final input = c['input'] == '__ABSENT__' ? av.absent : c['input'];
        if (c['nativeParent'] == true) schema = av.object({'payload': schema});

        /// Compare portable results and required diagnostics for this corpus case.
        void verify(av.Schema s) {
          final result = s.safeParse(input);
          expect(result.success, c['valid'],
              reason: '${result.issues.map((i) => i.toJson()).toList()}');
          if (c['valid']) {
            expect(result.data, c['output']);
          } else {
            for (final expected in c['issues']) {
              expect(
                  result.issues.any((i) =>
                      i.code == expected['code'] &&
                      equals(expected['path']).matches(i.path, {}) &&
                      (!expected.containsKey('expected') ||
                          i.expected == expected['expected']) &&
                      (!expected.containsKey('received') ||
                          i.received == expected['received'])),
                  isTrue,
                  reason:
                      'Expected $expected; actual ${result.issues.map((i) => i.toJson()).toList()}');
            }
          }
        }

        verify(schema);
        final document = schema.describe('Imported contract').exportSchema();
        // Extended mode preserves imported spellings; portable mode canonicalizes them.
        expect(schema.exportSchema(mode: 'extended')['definitions'],
            c['schema']['definitions']);
        final roundtrip = av.importSchema(jsonDecode(jsonEncode(document)));
        verify(roundtrip);
        if (c.containsKey('sensitivePaths')) {
          for (final s in [schema, roundtrip]) {
            expect(av.safeParseEncrypted(s, input).success, isFalse);
            final paths = <List<Object>>[];
            final encrypted = av.encrypt(s, input, (path, value) {
              paths.add(path);
              return 'encrypted:${jsonEncode(value)}';
            });
            expect(paths, c['sensitivePaths']);
            expect(av.safeParseEncrypted(s, encrypted).success, isTrue);
            paths.clear();
            final clear = av.decrypt(s, encrypted, (path, value) {
              paths.add(path);
              return jsonDecode((value as String).substring(10));
            });
            expect(clear, c['output']);
            expect(paths, c['sensitivePaths']);
          }
        }
      });
    }
  }
}
