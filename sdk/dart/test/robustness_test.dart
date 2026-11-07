import 'dart:convert';
import 'package:anyvali/anyvali.dart' as av;
import 'package:test/test.dart';

/// Cover portability, numeric boundaries, defaults, and sensitive-data regressions.
void main() {
  test(
      'reserved metadata uses describe and replacement retains reserved fields',
      () {
    for (final key in [
      'title',
      'description',
      'deprecated',
      'deprecatedMessage',
      'notStable',
      'since',
      'sensitive',
      'readonly',
      'writeonly',
      'examples'
    ]) {
      expect(
          () => av.string().metadata({key: true}),
          throwsA(predicate(
              (e) => e is ArgumentError && e.toString().contains('describe'))));
    }
    final described = av.string().describe('Original',
        title: 'Title',
        deprecated: true,
        deprecatedMessage: 'Use next',
        notStable: true,
        since: '1.0',
        sensitive: true,
        readonly: true,
        writeonly: false,
        examples: ['example']);
    final original = described.metadata({
      'custom': {'owner': 'one'},
      'old': 1
    });
    final replaced = original.metadata({
      'custom': {'owner': 'two'}
    }, replace: true).describe('Updated');
    final metadata = replaced.exportSchema()['root']['metadata'] as Map;
    expect(metadata['description'], 'Updated');
    expect(metadata['title'], 'Title');
    expect(metadata['deprecated'], true);
    expect(metadata['sensitive'], true);
    expect(metadata['readonly'], true);
    expect(metadata['writeonly'], false);
    expect(metadata['examples'], ['example']);
    expect(metadata.containsKey('old'), false);
    expect(metadata['custom'], {'owner': 'two'});
    expect(original.exportSchema()['root']['metadata']['old'], 1);
    expect(
        () => av.string().describe('Invalid', readonly: true, writeonly: true),
        throwsFormatException);
  });

  test('manual Dart documents accept empty and raw string-keyed maps', () {
    final document = <String, dynamic>{
      'anyvaliVersion': '1.0',
      'schemaVersion': '1.1',
      'root': <dynamic, dynamic>{
        'kind': 'object',
        'properties': <dynamic, dynamic>{},
        'required': []
      },
      'definitions': {},
      'extensions': {}
    };
    expect(av.importSchema(document).parse({}), {});
    expect(
        () => av.importSchema({
              ...document,
              'definitions': {
                1: {'kind': 'string'}
              }
            }),
        throwsFormatException);
  });

  test('present sensitive encryption cannot return null or absence', () {
    final schemas = [
      av.string().optional().describe('Secret', sensitive: true),
      av.string().nullable().describe('Secret', sensitive: true)
    ];
    for (final schema in schemas) {
      for (final lost in [null, av.absent]) {
        expect(() => av.encrypt(schema, 'secret', (p, v) => lost),
            throwsA(isA<av.ValidationError>()));
      }
      expect(av.encrypt(schema, 'secret', (p, v) => 'encrypted:$v'),
          'encrypted:secret');
    }
    var calls = 0;
    expect(
        av.encrypt(schemas[1], null, (p, v) {
          calls++;
          return null;
        }),
        null);
    expect(
        identical(
            av.encrypt(schemas[0], av.absent, (p, v) {
              calls++;
              return null;
            }),
            av.absent),
        isTrue);
    expect(calls, 0);
    final object = av.object({'secret': schemas[1]});
    expect(
        () => av.encrypt(object, {'secret': 'value'}, (p, v) => null),
        throwsA(predicate((e) =>
            e is av.ValidationError &&
            e.issues.single.path.single == 'secret')));
  });
  test('portable export strips schema extensions without changing default data',
      () {
    final base = av.any().exportSchema();
    final extensions = {
      'dart': {'_criticality': 'informational', 'hint': 'example'}
    };
    final document = {
      ...base,
      'root': {
        'kind': 'object',
        'properties': {
          'value': {
            'kind': 'optional',
            'inner': {
              'kind': 'ref',
              'ref': '#/definitions/Value',
              'extensions': extensions
            },
            'extensions': extensions
          },
          'data': {
            'kind': 'any',
            'default': {
              'extensions': {'app': 'keep'}
            },
            'extensions': extensions
          },
        },
        'required': <String>[],
        'extensions': extensions,
        'metadata': {
          'custom': {'extensions': 'keep'}
        }
      },
      'definitions': {
        'Value': {'kind': 'string', 'extensions': extensions},
      }
    };
    final schema = av.importSchema(document);
    final portable = schema.exportSchema();
    expect((portable['root'] as Map).containsKey('extensions'), isFalse);
    expect(
        (portable['root']['properties']['value'] as Map)
            .containsKey('extensions'),
        isFalse);
    expect(
        (portable['root']['properties']['value']['inner'] as Map)
            .containsKey('extensions'),
        isFalse);
    expect((portable['definitions']['Value'] as Map).containsKey('extensions'),
        isFalse);
    expect(portable['root']['metadata']['custom']['extensions'], 'keep');
    expect(av.importSchema(portable).parse({'value': 'ok'}), {
      'value': 'ok',
      'data': {
        'extensions': {'app': 'keep'}
      }
    });
    final extended = schema.exportSchema(mode: 'extended');
    expect(extended['root']['extensions'], extensions);
    expect(extended['definitions']['Value']['extensions'], extensions);
  });

  test('optional inner defaults survive wrappers and round trips', () {
    final schema = av.object({
      'count': av.int_().defaultValue(0).optional(),
      'text': av.string().coerce(trim: true).defaultValue(' keep ').optional(),
      'tags': av.array(av.string()).defaultValue(['a']).optional(),
      'nullable': av.string().nullable().defaultValue(null).optional(),
    });
    for (final s in [schema, av.importSchema(schema.exportSchema())]) {
      final result = s.parse({}) as Map;
      expect(result, {
        'count': 0,
        'text': ' keep ',
        'tags': ['a'],
        'nullable': null
      });
      (result['tags'] as List).add('changed');
      expect((s.parse({}) as Map)['tags'], ['a']);
    }
    expect(
        av
            .int_()
            .min(1)
            .defaultValue(0)
            .optional()
            .safeParse()
            .issues
            .single
            .code,
        'default_invalid');
    final secret = av
        .string()
        .defaultValue('plaintext')
        .optional()
        .describe('Secret', sensitive: true);
    expect(av.safeParseEncrypted(secret).success, isFalse);
    expect(av.encrypt(secret, av.absent, (p, v) => 'encrypted:$v'),
        'encrypted:plaintext');
  });
  test('referenced defaults materialize missing required and optional fields',
      () {
    final base = av.string().exportSchema();
    final document = {
      ...base,
      'root': {
        'kind': 'object',
        'properties': {
          'required': {'kind': 'ref', 'ref': '#/definitions/Value'},
          'other': {'kind': 'ref', 'ref': '#/definitions/Value'},
        },
        'required': ['required']
      },
      'definitions': {
        'Value': {'kind': 'ref', 'ref': '#/definitions/Text'},
        'Text': {'kind': 'string', 'coerce': 'trim', 'default': ' fallback '},
      }
    };
    final schema = av.importSchema(document);
    for (final s in [schema, av.importSchema(schema.exportSchema())]) {
      expect(s.parse({}), {'required': ' fallback ', 'other': ' fallback '});
      expect(s.parse({'required': ' value '}),
          {'required': 'value', 'other': ' fallback '});
    }
    final loop = {
      ...base,
      'root': {
        'kind': 'object',
        'properties': {
          'x': {'kind': 'ref', 'ref': '#/definitions/Loop'}
        },
        'required': ['x']
      },
      'definitions': {
        'Loop': {'kind': 'ref', 'ref': '#/definitions/Loop'}
      }
    };
    expect(av.importSchema(loop).safeParse({}).issues.single.code, 'required');
  });

  test('required keys take precedence over imported optional wrappers', () {
    final base = av.string().exportSchema();
    final document = {
      ...base,
      'root': {
        'kind': 'object',
        'properties': {
          'name': {
            'kind': 'optional',
            'schema': {'kind': 'string'}
          }
        },
        'required': ['name']
      }
    };
    final schema = av.importSchema(document);
    for (final input in [
      <String, dynamic>{},
      {'name': av.absent}
    ]) {
      expect(schema.safeParse(input).issues.single.code, 'required');
    }
    expect(schema.parse({'name': 'Alice'}), {'name': 'Alice'});
    document['root'] = {
      'kind': 'object',
      'properties': {
        'name': {
          'kind': 'optional',
          'schema': {'kind': 'string'},
          'default': 'fallback'
        }
      },
      'required': ['name']
    };
    expect(av.importSchema(document).parse({}), {'name': 'fallback'});
  });
  test('numeric bounds treat negative and positive zero as equal', () {
    for (final schema in [av.number(), av.float32(), av.float64()]) {
      expect(schema.min(0).max(0).safeParse(-0.0).success, isTrue);
      expect(schema.min(-0.0).max(-0.0).safeParse(0).success, isTrue);
      expect(schema.exclusiveMin(0).safeParse(-0.0).success, isFalse);
      expect(schema.exclusiveMax(0).safeParse(-0.0).success, isFalse);
      expect(schema.coerce().min(0).safeParse('-0').success, isTrue);
    }
  });

  test('float32 preserves the original value for constraints and output', () {
    const maximum = 3.4028234663852886e38;
    expect(av.float32().parse(maximum), maximum);
    expect(av.float32().parse(-maximum), -maximum);
    for (final value in [3.40282355e38, -3.40282355e38]) {
      expect(
          av.float32().safeParse(value).issues.single.code, 'invalid_number');
    }
    expect(av.float32().max(0.3).parse(0.3), 0.3);
    expect(av.float32().multipleOf(0.1).parse(0.3), 0.3);
    expect(av.float32().parse(1 / 3), 1 / 3);
    expect(av.float32().safeParse(1e40).issues.single.code, 'invalid_number');
  });

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
      expect(
          () => av.string().describe('Old',
              deprecated: metadata['deprecated'] as bool?,
              deprecatedMessage: 'old'),
          throwsFormatException);
      final document = av.string().exportSchema();
      expect(
          () => av.importSchema({
                ...document,
                'root': {'kind': 'string', 'metadata': metadata}
              }),
          throwsFormatException);
    }
    expect(
        av
            .string()
            .describe('Old', deprecated: true, deprecatedMessage: 'use next')
            .parse('ok'),
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
