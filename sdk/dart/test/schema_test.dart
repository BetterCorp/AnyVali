import 'dart:convert';
import 'package:anyvali/anyvali.dart' as av;
import 'package:test/test.dart';

void main() {
  test('numeric precision, decimal multiples and web boundaries', () {
    expect(av.number().multipleOf(0.1).safeParse(0.3).success, isTrue);
    expect(av.number().multipleOf(0.1).safeParse(0.35).success, isFalse);
    expect(
        av
            .uint64()
            .multipleOf(0.5)
            .safeParse(BigInt.parse('18446744073709551615'))
            .success,
        isTrue);
    expect(av.int_().coerce().parse('9007199254740991'), 9007199254740991);
    final wide = av.int_().coerce().safeParse('9007199254740992');
    expect(wide.success, !identical(1, 1.0));
    expect(av.uint64().coerce().parse('18446744073709551615'),
        BigInt.parse('18446744073709551615'));
    expect(av.float32().safeParse(1e40).issues.single.code, 'invalid_number');
    expect(() => av.uint64().defaultValue(BigInt.one).exportSchema(),
        throwsFormatException);
  });
  test('invalid native constraints and metadata fail at authoring time', () {
    expect(() => av.string().minLength(-1), throwsFormatException);
    expect(() => av.number().min(double.nan), throwsFormatException);
    expect(() => av.number().multipleOf(0), throwsArgumentError);
    expect(() => av.literal({'a': 1}), throwsArgumentError);
    expect(() => av.enum_([double.nan]), throwsArgumentError);
    expect(() => av.string().metadata({'sensitive': 'yes'}),
        throwsFormatException);
  });
  test('typed output, immutable builders and nested arrays', () {
    final int literal = av.literal(1).parse(1.0);
    final int choice = av.enum_([1, 2]).parse(2.0);
    expect([literal, choice], [1, 2]);
    final base = av.string();
    final constrained = base.minLength(2);
    final String text = base.parse('a');
    expect(text, 'a');
    expect(constrained.safeParse('a').success, isFalse);
    final List<List<int>> numbers = av.array(av.array(av.int_())).parse([
      [1, 2]
    ]);
    expect(numbers, [
      [1, 2]
    ]);
    final Map<String, List<int>> mapping =
        av.record(av.array(av.int_())).parse({
      'a': [1]
    });
    expect(mapping, {
      'a': [1]
    });
    final int? optional = av.int_().optional().parse();
    expect(optional, isNull);
  });
  test('defaults are cloned, explicit null remains present', () {
    final schema = av.array(av.string()).defaultValue(['first']);
    schema.parse().add('changed');
    expect(schema.parse(), ['first']);
    expect(schema.safeParse(null).success, isFalse);
    expect(av.int_().min(10).defaultValue(5).safeParse().issues.single.code,
        'default_invalid');
    expect(av.int_().coerce().defaultValue(7).parse(), 7);
    expect(av.int_().coerce().min(5).parse('6'), 6);
    expect(av.int_().coerce().safeParse('1.2').issues.single.code,
        'coercion_failed');
  });
  test('Flutter validator callback and aggregated field errors', () {
    final validator = av.string().minLength(1).format('email').validator();
    expect(validator('alice@example.com'), isNull);
    expect(validator('bad'), isNotNull);
    expect(validator(null), isNotNull);
    final schema = av.object({'name': av.string(), 'age': av.int_()});
    expect(
        schema.safeParse({'name': 1, 'age': 'x'}).issues.map((i) => i.path), [
      ['name'],
      ['age']
    ]);
    expect(() => schema.parse({}), throwsA(isA<av.ValidationError>()));
  });
  test('Unicode lengths and strict date formats', () {
    expect(av.string().maxLength(1).safeParse('😀').success, isTrue);
    expect(av.string().format('date').safeParse('2025-02-29').success, isFalse);
    expect(
        av
            .string()
            .format('date-time')
            .safeParse('2024-02-29T12:30:00Z')
            .success,
        isTrue);
    expect(av.string().format('ipv6').safeParse(':::').success, isFalse);
  });
  test('exact uint64 using BigInt and bounded signed coercion', () {
    final max = BigInt.parse('18446744073709551615');
    expect(av.uint64().parse(max), max);
    expect(av.uint64().safeParse(max + BigInt.one).issues.single.code,
        'too_large');
    expect(
        av.int_().coerce().safeParse('9223372036854775808').issues.single.code,
        'too_large');
    expect(av.int_().safeParse(true).success, isFalse);
    expect(
        av.number().safeParse(double.nan).issues.single.code, 'invalid_number');
  });
  test('rejects malformed imports and semantic extensions', () {
    final doc = av.string().exportSchema();
    expect(
        () => av.importSchema({...doc, 'extra': true}), throwsFormatException);
    expect(
        () => av.importSchema({
              ...doc,
              'root': {'kind': 'wat'}
            }),
        throwsA(isA<av.ValidationError>()));
    expect(
        () => av.importSchema({
              ...doc,
              'root': {'kind': 'ref', 'ref': '#/definitions/Missing'}
            }),
        throwsFormatException);
    expect(
        () => av.importSchema({
              ...doc,
              'root': {'kind': 'array'}
            }),
        throwsA(isA<av.ValidationError>()));
    expect(
        () => av.importSchema({
              ...doc,
              'extensions': {
                'dart': {'_criticality': 'semantic', 'foo': true}
              }
            }),
        throwsA(isA<av.ValidationError>()));
    expect(
        () => av.importSchema({
              ...doc,
              'root': {'kind': 'string', 'coerce': 'wat'}
            }),
        throwsFormatException);
  });
  test('recursive references retain definitions inside native parents', () {
    final doc = av.string().exportSchema();
    final imported = av.importSchema({
      ...doc,
      'root': {'kind': 'ref', 'ref': '#/definitions/Node'},
      'definitions': {
        'Node': {
          'kind': 'object',
          'properties': {
            'next': {
              'kind': 'optional',
              'schema': {'kind': 'ref', 'ref': '#/definitions/Node'}
            }
          },
          'required': []
        }
      }
    });
    final parent = av.object({'payload': imported});
    expect(
        av.importSchema(jsonDecode(jsonEncode(parent.exportSchema()))).parse({
          'payload': {'next': {}}
        }),
        {
          'payload': {'next': {}}
        });
    final loop = <String, dynamic>{};
    loop['next'] = loop;
    expect(imported.safeParse(loop).issues.single.code, 'too_deep');
  });
  test('intersection preserves coercions and encrypts overlapping paths once',
      () {
    final left = av.object({
      'a': av.int_().coerce().describe('A', sensitive: true)
    }).unknownKeys('allow');
    final right = av.object(
        {'b': av.string().describe('B', sensitive: true)}).unknownKeys('allow');
    final schema = av.intersection([left, right, left]);
    final paths = <List<Object>>[];
    final stored = av.encrypt(schema, {'a': '42', 'b': 'hello'}, (path, value) {
      paths.add(path);
      return 'encrypted:${jsonEncode(value)}';
    });
    expect(paths, [
      ['a'],
      ['b']
    ]);
    expect(
        av.decrypt(
            schema, stored, (p, v) => jsonDecode((v as String).substring(10))),
        {'a': 42, 'b': 'hello'});
  });
  test('sensitive composites are opaque and callbacks run once', () {
    final schema = av.object({
      'secret':
          av.object({'value': av.string()}).describe('Secret', sensitive: true)
    });
    final paths = <List<Object>>[];
    final encrypted = av.encrypt(schema, {
      'secret': {'value': 'hello'}
    }, (path, value) {
      paths.add(path);
      return 'encrypted:${jsonEncode(value)}';
    });
    expect(paths, [
      ['secret']
    ]);
    expect(av.safeParseEncrypted(schema, encrypted).success, isTrue);
    expect(
        av.decrypt(schema, encrypted,
            (path, value) => jsonDecode((value as String).substring(10))),
        {
          'secret': {'value': 'hello'}
        });
    expect(
        () => av.encrypt(
            schema,
            {
              'secret': {'value': 'hello'}
            },
            (p, v) => 'bad'),
        throwsA(isA<av.ValidationError>()));
  });
}
