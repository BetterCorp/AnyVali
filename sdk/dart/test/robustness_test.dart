import 'dart:convert';
import 'package:anyvali/anyvali.dart' as av;
import 'package:test/test.dart';

/// Cover portability, numeric boundaries, defaults, and sensitive-data regressions.
void main() {
  test('overlapping sensitive members encrypt the final parsed value once', () {
    final secret = av.string().describe('Secret', sensitive: true);
    final schemas = <(av.Schema, Object?)>[
      (
        av.intersection([secret.defaultValue('a'), secret.defaultValue('b')]),
        av.absent
      ),
      (
        av.intersection(
            [secret.coerce(lower: true), secret.coerce(upper: true)]),
        'Mixed'
      ),
      (
        av.intersection([
          av.object({'value': secret.defaultValue('a')}),
          av.object({'value': secret.defaultValue('b')})
        ]),
        <String, Object?>{}
      ),
      (
        av.intersection([
          av.object({'value': secret.coerce(lower: true)}),
          av.object({'value': secret.coerce(upper: true)})
        ]),
        {'value': 'Mixed'}
      ),
    ];
    for (final (schema, input) in schemas) {
      final parsed = schema.parse(input);
      var calls = 0;
      final stored = av.encrypt(schema, input, (path, value) {
        calls++;
        expect(value, path.isEmpty ? parsed : (parsed as Map)['value']);
        return 'encrypted:$value';
      });
      expect(calls, 1);
      expect(av.safeParseEncrypted(schema, stored).success, isTrue);
      final restored = av.decrypt(
          schema, stored, (path, value) => (value as String).substring(10));
      // Decryption applies ordinary parsing, including any further coercion.
      expect(restored, schema.parse(parsed));
    }
  });
  test('intersection transforms use member results before merging defaults',
      () {
    final choice = av.object({
      'value': av.union([av.string(), av.int_()]).optional()
    });
    final plain = av.intersection([
      choice,
      av.object({'value': av.string().defaultValue('default')})
    ]);
    Object? unexpected(List<Object> path, Object? value) =>
        throw StateError('No sensitive node');
    expect(av.safeParseEncrypted(plain, {}).data, {'value': 'default'});
    expect(av.encrypt(plain, {}, unexpected), {'value': 'default'});
    expect(av.decrypt(plain, {}, unexpected), {'value': 'default'});
    final secret = av.string().describe('Secret', sensitive: true);
    final sensitive = av.intersection([
      choice,
      av.object({'value': secret.defaultValue('default')})
    ]);
    var calls = 0;
    final stored = av.encrypt(sensitive, {}, (path, value) {
      calls++;
      return 'encrypted:$value';
    });
    expect(stored, {'value': 'encrypted:default'});
    expect(calls, 1);
    expect(av.safeParseEncrypted(sensitive, stored).data, stored);
    expect(
        av.decrypt(sensitive, stored,
            (path, value) => (value as String).substring(10)),
        {'value': 'default'});
    final tuples = av.intersection([
      av.tuple([secret, av.string()]),
      av.tuple([av.string(), secret])
    ]);
    final encoded =
        av.encrypt(tuples, ['a', 'b'], (path, value) => 'encrypted:$value');
    expect(encoded, ['encrypted:a', 'encrypted:b']);
    expect(
        av.decrypt(
            tuples, encoded, (path, value) => (value as String).substring(10)),
        ['a', 'b']);
  });
  test('informational extension conflicts defer until extended export', () {
    final base = av.string().exportSchema();
    av.Schema<Object?> imported(String note) => av.importSchema({
          ...base,
          'extensions': {
            'app': {'_criticality': 'informational', 'note': note}
          }
        });
    final left = imported('left');
    final right = imported('right');
    final schemas = <(av.Schema, Object?)>[
      (av.object({'a': left, 'b': right}), {'a': 'x', 'b': 'y'}),
      (av.tuple([left, right]), ['x', 'y']),
      (av.union([left, right]), 'x'),
      (av.intersection([left, right]), 'x'),
    ];
    for (final (schema, input) in schemas) {
      expect(schema.safeParse(input).success, isTrue);
      expect(schema.exportSchema()['extensions'], isEmpty);
      expect(av.importSchema(schema.exportSchema()).safeParse(input).success,
          isTrue);
      for (final wrapped in [
        schema,
        schema.optional(),
        schema.nullable(),
        schema.describe('Example'),
        av.array(schema),
        av.record(schema)
      ]) {
        expect(wrapped.exportSchema()['extensions'], isEmpty);
        expect(() => wrapped.exportSchema(mode: 'extended'), throwsStateError);
      }
    }
    final same = av.tuple([left, imported('left')]);
    expect(same.exportSchema(mode: 'extended')['extensions']['app']['note'],
        'left');
  });
  test(
      'sensitive intersections preserve transformed values in either member order',
      () {
    final secret = av.string().describe('Secret', sensitive: true);
    for (final members in [
      [secret, av.string()],
      [av.string(), secret],
      [secret, secret]
    ]) {
      final schema = av.intersection(members);
      var calls = 0;
      final stored = av.encrypt(schema, 'value', (path, value) {
        calls++;
        return 'encrypted:$value';
      });
      expect(stored, 'encrypted:value');
      expect(calls, 1);
      calls = 0;
      expect(
          av.decrypt(schema, stored, (path, value) {
            calls++;
            return (value as String).substring(10);
          }),
          'value');
      expect(calls, 1);
    }
    final arrays = av.intersection([av.array(secret), av.array(av.string())]);
    final stored =
        av.encrypt(arrays, ['value'], (path, value) => 'encrypted:$value');
    expect(stored, ['encrypted:value']);
    expect(
        av.decrypt(
            arrays, stored, (path, value) => (value as String).substring(10)),
        ['value']);
    final coercing = av.intersection([secret, av.string().coerce(lower: true)]);
    expect(
        () =>
            av.encrypt(coercing, 'value', (path, value) => 'encrypted:PAYLOAD'),
        throwsA(isA<av.ValidationError>()));
    expect(
        av.safeParseEncrypted(coercing, 'encrypted:PAYLOAD').success, isFalse);
    var called = false;
    expect(
        () => av.decrypt(coercing, 'encrypted:PAYLOAD', (path, value) {
              called = true;
              return 'value';
            }),
        throwsA(isA<av.ValidationError>()));
    expect(called, isFalse);
    final opaque = coercing.describe('Whole secret', sensitive: true);
    final opaqueStored =
        av.encrypt(opaque, 'VALUE', (path, value) => 'encrypted:PAYLOAD');
    expect(opaqueStored, 'encrypted:PAYLOAD');
    expect(av.safeParseEncrypted(opaque, opaqueStored).data, opaqueStored);
    expect(av.decrypt(opaque, opaqueStored, (path, value) => 'VALUE'), 'value');
  });
  test('native unions decode the original selected branch at each data path',
      () {
    final plain =
        av.object({'tag': av.literal('x'), 'items': av.array(av.any())});
    final typed = av.object({
      'tag': av.string().coerce(trim: true),
      'items': av.array(av.int_().optional())
    });
    final variants = <av.Schema>[plain, typed];
    final schema = av.union(variants);
    variants.clear();
    final decoded = schema.parse({
      'tag': ' x ',
      'items': [av.absent]
    }) as Map;
    expect(decoded['tag'], 'x');
    expect(decoded['items'], <int?>[null]);
    expect(decoded['items'], isA<List<int?>>());
    final item = av.union([av.int_().optional(), av.string()]);
    expect(av.array(item).parse([av.absent, 'x']), [null, 'x']);
    expect(av.record(item).parse({'a': av.absent, 'b': 'x'}),
        {'a': null, 'b': 'x'});
    expect(
        av.tuple([item.nullable(), item.optional()]).parse([null, av.absent]),
        [null, null]);
  });
  test('native intersections decode retained member results before merging',
      () {
    final members = <av.Schema>[
      av.object({'items': av.array(av.int_().optional())}),
      av.object({'name': av.string()})
    ];
    final schema = av.intersection(members);
    members.clear();
    final decoded = schema.parse({
      'items': [av.absent],
      'name': 'Alice'
    }) as Map;
    expect(decoded['items'], <int?>[null]);
    expect(decoded['items'], isA<List<int?>>());
    expect(decoded['name'], 'Alice');
    expect(
        av.intersection([av.any(), av.array(av.int_().optional())]).parse(
            [av.absent]),
        <int?>[null]);
  });
  test('native objects decode children and snapshot the property schemas', () {
    final fields = <String, av.Schema>{
      'items': av.array(av.int_().optional()),
      'nested': av.tuple([av.bool_().optional()])
    };
    final schema = av.object(fields).unknownKeys('allow');
    fields.clear();
    final parsed = schema.parse({
      'items': [av.absent],
      'nested': [av.absent],
      'extra': 'keep'
    });
    expect(parsed['items'], <int?>[null]);
    expect(parsed['items'], isA<List<int?>>());
    expect(parsed['nested'], [null]);
    expect(parsed['extra'], 'keep');
  });
  test('native tuples decode child values and snapshot their element list', () {
    final slots = <av.Schema>[
      av.any(),
      av.int_().optional(),
      av.array(av.string().optional()),
      av.record(av.bool_().optional())
    ];
    final schema = av.tuple(slots);
    slots[0] = av.string();
    slots.add(av.int_());
    final parsed = schema.parse([
      av.absent,
      av.absent,
      [av.absent],
      {'missing': av.absent}
    ]);
    expect(identical(parsed[0], av.absent), isTrue);
    expect(parsed[1], isNull);
    expect(parsed[2], <String?>[null]);
    expect(parsed[2], isA<List<String?>>());
    expect(parsed[3], <String, bool?>{'missing': null});
    expect(parsed[3], isA<Map<String, bool?>>());
    expect(av.tuple([av.int_().coerce()]).parse(['7']), [7]);
  });
  test('permissive schemas preserve absence while native optionals decode null',
      () {
    for (final schema in [
      av.any(),
      av.unknown(),
      av.any().nullable(),
      av.unknown().nullable()
    ]) {
      for (final current in [schema, av.importSchema(schema.exportSchema())]) {
        expect(identical(current.parse(), av.absent), isTrue);
        expect(identical(current.parse(av.absent), av.absent), isTrue);
        expect(identical(current.safeParse().data, av.absent), isTrue);
        expect(current.parse(null), isNull);
        expect(current.safeParse(null).data, isNull);
      }
    }
    expect(av.string().optional().parse(), isNull);
    expect(av.array(av.string()).optional().parse(), isNull);
    expect(av.any().optional().parse(), isNull);
    expect(av.any().defaultValue(null).parse(), isNull);
  });
  test('empty intersections fail native construction and interchange import',
      () {
    expect(() => av.intersection([]), throwsArgumentError);
    final base = av.any().exportSchema();
    for (final key in ['allOf', 'schemas']) {
      expect(
          () => av.importSchema({
                ...base,
                'root': {'kind': 'intersection', key: []}
              }),
          throwsFormatException);
    }
    expect(av.intersection([av.int_()]).parse(7), 7);
  });
  test(
      'portable coercions use ordered token lists across nodes and definitions',
      () {
    for (final entry in [
      (av.int_().coerce(), ['string->int']),
      (av.uint64().coerce(), ['string->int']),
      (av.number().coerce(), ['string->number']),
      (av.bool_().coerce(), ['string->bool']),
      (av.string().coerce(trim: true, lower: true), ['trim', 'lower']),
    ]) {
      expect(entry.$1.exportSchema()['root']['coerce'], entry.$2);
      expect(
          entry.$1.exportSchema(mode: 'extended')['root']['coerce'], entry.$2);
    }
    final base = av.any().exportSchema();
    for (final entry in [
      ('int', '7', 7),
      ('number', '2.5', 2.5),
      ('bool', 'true', true)
    ]) {
      final imported = av.importSchema({
        ...base,
        'root': {
          'kind': entry.$1,
          'coerce': [entry.$1]
        }
      });
      expect(imported.parse(entry.$2), entry.$3);
      expect(
          imported.exportSchema()['root']['coerce'], ['string->${entry.$1}']);
    }
    final legacy = av.importSchema({
      ...base,
      'root': {
        'kind': 'object',
        'required': ['count', 'text'],
        'properties': {
          'count': {
            'kind': 'ref',
            'ref': '#/definitions/Count',
            'coerce': {'toInt': true}
          },
          'text': {
            'kind': 'optional',
            'schema': {
              'kind': 'string',
              'coerce': ['trim', 'lower']
            }
          }
        }
      },
      'definitions': {
        'Count': {'kind': 'int', 'coerce': {}}
      }
    });
    final portable = legacy.exportSchema();
    expect(portable['root']['properties']['count']['coerce'], ['string->int']);
    expect(portable['definitions']['Count']['coerce'], ['string->int']);
    expect(portable['root']['properties']['text']['schema']['coerce'],
        ['trim', 'lower']);
    expect(
        legacy.exportSchema(mode: 'extended')['definitions']['Count']['coerce'],
        {});
    expect(av.importSchema(portable).parse({'count': '7', 'text': ' A '}),
        {'count': 7, 'text': 'a'});
    final single = av.importSchema({
      ...base,
      'root': {'kind': 'string', 'coerce': 'trim'}
    });
    expect(single.exportSchema()['root']['coerce'], ['trim']);
  });
  test('schema data must contain actual JSON values', () {
    final base = av.any().exportSchema();
    final cycle = <String, dynamic>{};
    cycle['self'] = cycle;
    for (final value in [
      DateTime.utc(2026),
      BigInt.one,
      double.nan,
      double.infinity,
      av.absent,
      () => 1,
      {1, 2},
      {1: 'bad key'},
      cycle,
      _EncodableValue()
    ]) {
      for (final document in [
        {
          ...base,
          'root': {'kind': 'any', 'default': value}
        },
        {
          ...base,
          'root': {
            'kind': 'any',
            'metadata': {
              'custom': {'value': value}
            }
          }
        },
        {
          ...base,
          'extensions': {
            'dart': {'_criticality': 'informational', 'value': value}
          }
        }
      ]) {
        expect(() => av.importSchema(document), throwsFormatException);
      }
      expect(() => av.any().defaultValue(value), throwsFormatException);
      expect(() => av.any().metadata({'custom': value}), throwsFormatException);
    }
    final shared = <String, dynamic>{
      'value': [null, false, 0, 'text']
    };
    final schema = av.importSchema({
      ...base,
      'root': {
        'kind': 'any',
        'default': [shared, shared]
      }
    });
    expect(schema.parse(), [shared, shared]);
    expect(jsonDecode(jsonEncode(schema.exportSchema()))['root']['default'],
        [shared, shared]);
  });
  test('malformed coercion targets report schema import errors', () {
    final base = av.any().exportSchema();
    for (final target in [
      {},
      {'kind': 1},
      {1: 'int'},
      null,
      'int'
    ]) {
      expect(
          () => av.importSchema({
                ...base,
                'root': {
                  'kind': 'ref',
                  'ref': '#/definitions/Value',
                  'coerce': 'string->int'
                },
                'definitions': {'Value': target}
              }),
          throwsFormatException);
      expect(
          () => av.importSchema({
                ...base,
                'root': {
                  'kind': 'optional',
                  'schema': target,
                  'coerce': 'string->int'
                }
              }),
          throwsFormatException);
    }
  });
  test('nested unknown-key policies match inherited JS and Go behavior', () {
    final input = {
      'child': {'name': 'Alice', 'admin': true}
    };
    for (final parent in ['allow', 'strip', 'reject']) {
      for (final child in ['allow', 'strip', 'reject']) {
        final schema = av.object({
          'child': av.object({'name': av.string()}).unknownKeys(child)
        }).unknownKeys(parent);
        final effective = parent == 'allow' ? child : parent;
        for (final current in [
          schema,
          av.importSchema(schema.exportSchema())
        ]) {
          final result = current.safeParse(input);
          if (effective == 'reject') {
            expect(result.success, isFalse);
            expect(result.issues.single.code, 'unknown_key');
            expect(result.issues.single.path, ['child', 'admin']);
          } else {
            expect(result.success, isTrue);
            expect(((result.data as Map)['child'] as Map).containsKey('admin'),
                effective == 'allow');
          }
        }
      }
    }
  });
  test('reference defaults and coercions follow each node pipeline', () {
    final base = av.any().exportSchema();
    final schema = av.importSchema({
      ...base,
      'root': {
        'kind': 'object',
        'properties': {
          'count': {
            'kind': 'ref',
            'ref': '#/definitions/Count',
            'default': '42'
          },
          'wrapped': {
            'kind': 'ref',
            'ref': '#/definitions/Count',
            'coerce': 'string->int'
          }
        },
        'required': ['count', 'wrapped']
      },
      'definitions': {
        'Count': {'kind': 'int', 'coerce': 'string->int'}
      }
    });
    expect(schema.parse({'wrapped': '7'}), {'count': 42, 'wrapped': 7});
  });
  test('portable exports canonicalize every legacy child key', () {
    final base = av.any().exportSchema();
    for (final entry in {
      'optional': ('inner', 'schema', {'kind': 'int'}),
      'nullable': ('inner', 'schema', {'kind': 'int'}),
      'record': ('valueSchema', 'values', {'kind': 'int'}),
      'tuple': (
        'items',
        'elements',
        [
          {'kind': 'int'}
        ]
      ),
      'union': (
        'schemas',
        'variants',
        [
          {'kind': 'int'}
        ]
      ),
      'intersection': (
        'schemas',
        'allOf',
        [
          {'kind': 'int'}
        ]
      )
    }.entries) {
      final root = av.importSchema({
        ...base,
        'root': {'kind': entry.key, entry.value.$1: entry.value.$3}
      }).exportSchema()['root'] as Map;
      expect(root.containsKey(entry.value.$1), isFalse);
      expect(root.containsKey(entry.value.$2), isTrue);
    }
  });
  test('encrypted union storage is checked after normalization', () {
    final schema = av.union([
      av.object({'type': av.literal('x'), 'secret': av.string()}),
      av.object({
        'type': av.string().coerce(trim: true).defaultValue(' x '),
        'secret': av.string().describe('Secret', sensitive: true)
      })
    ]);
    expect(
        () => av.encrypt(schema, {'secret': 'value'}, (p, v) => 'encrypted:$v'),
        throwsA(isA<av.ValidationError>()));
  });
  test('coercions must match their primitive schema target', () {
    final base = av.string().exportSchema();
    for (final root in [
      for (final transform in ['trim', 'lower', 'upper'])
        for (final kind in ['int', 'number', 'bool', 'any'])
          {'kind': kind, 'coerce': transform},
      {'kind': 'bool', 'coerce': 'string->int'},
      {'kind': 'int', 'coerce': 'string->number'},
      {'kind': 'string', 'coerce': 'string->bool'},
      {
        'kind': 'int',
        'coerce': ['string->int', 'string->number']
      },
      {
        'kind': 'bool',
        'coerce': {'toInt': true, 'toBool': true}
      },
    ]) {
      expect(() => av.importSchema({...base, 'root': root}),
          throwsFormatException);
    }
    expect(
        av.importSchema({
          ...base,
          'root': {'kind': 'uint8', 'coerce': 'string->int'}
        }).parse('1'),
        1);
    expect(
        av.importSchema({
          ...base,
          'root': {'kind': 'bool', 'coerce': 'string->bool'}
        }).parse('1'),
        true);
  });
  test(
      'encryption rejects ambiguous union storage while preserving opaque nodes',
      () {
    final both = av.union([
      av.string().describe('Text', sensitive: true),
      av.int_().describe('PIN', sensitive: true)
    ]);
    final bothStored = av.encrypt(both, 1234, (p, v) => 'encrypted:$v');
    expect(
        av.decrypt(
            both, bothStored, (p, v) => int.parse((v as String).substring(10))),
        1234);
    final markerLikePlaintext =
        av.union([av.int_().describe('PIN', sensitive: true), av.string()]);
    expect(
        () => av.encrypt(markerLikePlaintext, 'encrypted:1234',
            (p, v) => throw StateError('Plain branch')),
        throwsA(isA<av.ValidationError>()));

    final plain = av.union([av.int_(), av.int_().coerce()]);
    expect(
        av.encrypt(
            plain, '42', (p, v) => throw StateError('No sensitive data')),
        42);
    final sibling = av.object({
      'choice': plain,
      'secret': av.string().describe('Secret', sensitive: true)
    });
    expect(
        av.encrypt(
            sibling, {'choice': '42', 'secret': 'x'}, (p, v) => 'encrypted:$v'),
        {'choice': 42, 'secret': 'encrypted:x'});
    final parallel = av.intersection([
      av.union([av.string(), av.int_()]),
      av.int_().describe('Secret', sensitive: true)
    ]);
    final parallelStored = av.encrypt(parallel, 1234, (p, v) => 'encrypted:$v');
    expect(
        av.decrypt(parallel, parallelStored,
            (p, v) => int.parse((v as String).substring(10))),
        1234);

    final ambiguous =
        av.union([av.string(), av.int_().describe('PIN', sensitive: true)]);
    expect(
        () => av.encrypt(ambiguous, 1234, (p, v) => 'encrypted:$v'),
        throwsA(predicate((e) =>
            e is av.ValidationError &&
            e.issues.single.code == 'invalid_union')));
    final opaque = av.object({
      'choice': av.union([av.string(), av.int_()])
    }).describe('Opaque', sensitive: true);
    final schema = av.object({'secret': opaque});
    final input = {
      'secret': {'choice': 1234}
    };
    final stored =
        av.encrypt(schema, input, (p, v) => 'encrypted:${jsonEncode(v)}');
    expect(
        av.decrypt(
            schema, stored, (p, v) => jsonDecode((v as String).substring(10))),
        input);
    final whole = ambiguous.describe('Whole union', sensitive: true);
    final encrypted = av.encrypt(whole, 1234, (p, v) => 'encrypted:$v');
    expect(
        av.decrypt(
            whole, encrypted, (p, v) => int.parse((v as String).substring(10))),
        1234);
  });

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
        (portable['root']['properties']['value']['schema'] as Map)
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
      expect(av.float32().safeParse(value).issues.single.code, 'too_large');
    }
    expect(av.float32().max(0.3).parse(0.3), 0.3);
    expect(av.float32().multipleOf(0.1).parse(0.3), 0.3);
    expect(av.float32().parse(1 / 3), 1 / 3);
    expect(av.float32().safeParse(1e40).issues.single.code, 'too_large');
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

/// Represent an encodable Dart object that is not itself a JSON value.
class _EncodableValue {
  /// Exercise rejection without calling user serialization hooks.
  Object toJson() => throw StateError('Must not be invoked');
}
