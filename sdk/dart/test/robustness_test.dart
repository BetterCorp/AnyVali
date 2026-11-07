import 'package:anyvali/anyvali.dart' as av;
import 'package:test/test.dart';

void main() {
  test(
      'portable imports reject ignored properties, unknown formats and criticality typos',
      () {
    final base = av.string().exportSchema();
    for (final root in [
      {'kind': 'string', 'min': 10},
      {'kind': 'string', 'minLenght': 10},
      {'kind': 'string', 'format': 'hostname'},
      {
        'kind': 'string',
        'extensions': {
          'x': {'_criticality': 'sematic'}
        }
      },
      {
        'kind': 'optional',
        'schema': {'kind': 'string'},
        'inner': {'kind': 'int'}
      },
    ]) {
      expect(() => av.importSchema({...base, 'root': root}),
          throwsFormatException);
    }
    for (final criticality in ['sematic', null, true]) {
      expect(
          () => av.importSchema({
                ...base,
                'extensions': {
                  'x': {'_criticality': criticality}
                }
              }),
          throwsFormatException);
    }
    expect(() => av.string().format('hostname'), throwsFormatException);
    expect(() => av.string().min(10), throwsFormatException);
    expect(() => av.array(av.string()).minLength(1), throwsFormatException);
    expect(
        av.importSchema({
          ...base,
          'extensions': {
            'x': {'_criticality': 'informational', 'note': 'ok'}
          }
        }).parse('ok'),
        'ok');
  });
  test('tiny floating divisors do not admit fractional multiples', () {
    expect(av.number().multipleOf(1e-12).safeParse(1.5e-12).success, isFalse);
    expect(av.number().multipleOf(1e-12).safeParse(2e-12).success, isTrue);
  });
  test('numeric coercion and email reject long invalid input promptly', () {
    expect(av.number().coerce().safeParse('${'1' * 20000}x').success, isFalse);
    expect(av.string().format('email').safeParse('a@${'a.' * 10000} ').success,
        isFalse);
  }, timeout: Timeout(Duration(seconds: 5)));
  test('safeParse bounds cyclic intersections', () {
    final loop = <String, dynamic>{};
    loop['self'] = loop;
    final s = av.intersection([av.any(), av.any()]);
    expect(() => s.safeParse(loop), returnsNormally);
  });
  test('encrypted storage validates materialized sensitive defaults', () {
    final schema = av
        .string()
        .describe('Secret', sensitive: true)
        .defaultValue('plaintext');
    final checked = av.safeParseEncrypted(schema);
    expect(checked.success, isFalse);
    expect(checked.issues.single.code, 'default_invalid');
    final encrypted = av.encrypt(schema, av.absent, (p, v) => 'encrypted:$v');
    expect(encrypted, 'encrypted:plaintext');
  });
  test('optional array and record outputs remain typed', () {
    final List<int?> a = av.array(av.int_().optional()).parse([av.absent, 1]);
    final Map<String, int?> b =
        av.record(av.int_().optional()).parse({'a': av.absent});
    expect(a, [null, 1]);
    expect(b, {'a': null});
  });
}
