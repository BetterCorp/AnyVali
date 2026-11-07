import 'dart:convert';
import 'package:anyvali/anyvali.dart' as av;
import 'package:test/test.dart';

void main() {
  test('parent unknown-key restrictions apply through arrays and composition',
      () {
    final child = av.object({'name': av.string()}).unknownKeys('allow');
    final schema = av.object({
      'users': av.array(av.union([child, av.null_()]))
    });
    final input = {
      'users': [
        {'name': 'Alice', 'admin': true}
      ]
    };
    expect(schema.parse(input), {
      'users': [
        {'name': 'Alice'}
      ]
    });
    final rejected = schema.unknownKeys('reject').safeParse(input);
    expect(rejected.success, isFalse);
    expect(rejected.issues.single.code, 'invalid_union');
    final direct = av
        .object({'users': av.array(child)})
        .unknownKeys('reject')
        .safeParse(input);
    expect(direct.issues.single.path, ['users', 0, 'admin']);
    final allowed = schema.unknownKeys('allow').parse(input);
    expect(allowed, input);
  });
  test('deprecation messages require the deprecated flag', () {
    for (final metadata in [
      {'deprecatedMessage': 'old'},
      {'deprecated': false, 'deprecatedMessage': 'old'}
    ]) {
      expect(() => av.string().metadata(metadata), throwsFormatException);
      final document = av.string().exportSchema();
      expect(
          () => av.importSchema({
                ...document,
                'root': {'kind': 'string', 'metadata': metadata}
              }),
          throwsFormatException);
    }
    expect(
        av.string().metadata(
            {'deprecated': true, 'deprecatedMessage': 'use next'}).parse('ok'),
        'ok');
  });

  test('sensitive union retains the branch chosen before coercion', () {
    final schema = av.union(
        [av.int_(), av.int_().coerce().describe('PIN', sensitive: true)]);
    final paths = <List<Object>>[];
    final encrypted = av.encrypt(schema, '1234', (path, value) {
      paths.add(path);
      return 'encrypted:$value';
    });
    expect(encrypted, 'encrypted:1234');
    expect(paths, [[]]);
    expect(
        av.decrypt(schema, encrypted,
            (p, v) => int.parse((v as String).substring(10))),
        1234);
    final nested = av.object({'items': av.array(schema)});
    paths.clear();
    expect(
        av.encrypt(nested, {
          'items': ['1234', '5678']
        }, (p, v) {
          paths.add(p);
          return 'encrypted:$v';
        }),
        {
          'items': ['encrypted:1234', 'encrypted:5678']
        });
    expect(paths, [
      ['items', 0],
      ['items', 1]
    ]);
  });
  test('IPv6 validator rejects URL fragments and zone suffixes', () {
    final schema = av.string().format('ipv6');
    for (final input in [
      '::1]/x',
      '::1%25eth0',
      'fe80::1%eth0',
      '[::1]',
      '1.2.3.4'
    ]) {
      expect(schema.safeParse(input).success, isFalse, reason: input);
    }
    for (final input in ['::1', '2001:db8::1', '::ffff:192.0.2.128']) {
      expect(schema.safeParse(input).success, isTrue, reason: input);
    }
  });

  test('wide integer bounds cannot silently round on JavaScript targets', () {
    final document = av.uint64().exportSchema();
    final root = jsonDecode('{"kind":"uint64","min":9007199254740993}');
    if (identical(1, 1.0)) {
      expect(() => av.importSchema({...document, 'root': root}),
          throwsFormatException);
    } else {
      final schema = av.importSchema({...document, 'root': root});
      expect(
          schema.safeParse(BigInt.parse('9007199254740992')).success, isFalse);
      expect(
          schema.safeParse(BigInt.parse('9007199254740993')).success, isTrue);
    }
  });

  test(
      'portable imports reject ignored properties, unknown formats and criticality typos',
      () {
    final base = av.string().exportSchema();
    for (final root in [
      {'kind': 'string', 'min': 10},
      {
        'kind': 'object',
        'properties': <String, dynamic>{},
        'required': <String>[],
        'unknownKeys': null
      },
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
    expect(av.number().multipleOf(1e100).safeParse(1e101).success, isTrue);
    expect(av.number().multipleOf(1e-300).safeParse(2e-300).success, isTrue);
    expect(av.number().multipleOf(1e-12).safeParse(1126.0000000000011).success,
        isFalse);
    expect(av.number().multipleOf(0.1).safeParse(12345678.9).success, isTrue);
    expect(av.number().multipleOf(0.1).safeParse(-12345678.9).success, isTrue);
    expect(av.number().multipleOf(0.1).safeParse(1000000.1).success, isTrue);
    expect(av.number().multipleOf(0.1).safeParse(1000000.15).success, isFalse);
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
