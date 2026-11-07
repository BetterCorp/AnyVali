import 'dart:convert';
import 'dart:typed_data';

/// An omitted value, distinct from an explicit null.
const absent = _Absent();

class _Absent {
  const _Absent();
}

typedef JsonMap = Map<String, dynamic>;
typedef SensitiveTransform = Object? Function(List<Object> path, Object? value);

class ValidationIssue {
  final String code;
  final String message;
  final List<Object> path;
  final String? expected;
  final String? received;
  ValidationIssue(this.code, this.message, List<Object> path,
      {this.expected, this.received})
      : path = List.unmodifiable(path);
  JsonMap toJson() => {
        'code': code,
        'message': message,
        'path': path,
        if (expected != null) 'expected': expected,
        if (received != null) 'received': received
      };
}

class ValidationError implements Exception {
  final List<ValidationIssue> issues;
  ValidationError(List<ValidationIssue> issues)
      : issues = List.unmodifiable(issues);
  @override
  String toString() =>
      'ValidationError: ${issues.map((i) => '${i.path}: ${i.message}').join('; ')}';
}

class ParseResult<T> {
  final bool success;
  final T? data;
  final List<ValidationIssue> issues;
  ParseResult._(this.success, this.data, List<ValidationIssue> issues)
      : issues = List.unmodifiable(issues);
}

/// Immutable schema. Modifier methods return a new schema.
class Schema<T> {
  final JsonMap _node;
  final JsonMap _definitions;
  final JsonMap _extensions;
  final T Function(dynamic) _decode;
  Schema._(JsonMap node,
      {JsonMap definitions = const {},
      JsonMap extensions = const {},
      T Function(dynamic)? decode})
      : _node = _clone(node),
        _definitions = _clone(definitions),
        _extensions = _clone(extensions),
        _decode = decode ?? ((v) => v as T);

  Schema<T> _set(String key, dynamic value) {
    final node = {..._node, key: value};
    _checkNode(node, _definitions, 0);
    return Schema._(node,
        definitions: _definitions, extensions: _extensions, decode: _decode);
  }

  Schema<T> min(num value) => _set('min', value);
  Schema<T> max(num value) => _set('max', value);
  Schema<T> exclusiveMin(num value) => _set('exclusiveMin', value);
  Schema<T> exclusiveMax(num value) => _set('exclusiveMax', value);
  Schema<T> multipleOf(num value) {
    if (!value.isFinite || value <= 0)
      throw ArgumentError.value(value, 'value');
    return _set('multipleOf', value);
  }

  Schema<T> minLength(int value) => _set('minLength', value);
  Schema<T> maxLength(int value) => _set('maxLength', value);
  Schema<T> minItems(int value) => _set('minItems', value);
  Schema<T> maxItems(int value) => _set('maxItems', value);
  Schema<T> pattern(String value) {
    RegExp(value);
    return _set('pattern', value);
  }

  Schema<T> startsWith(String value) => _set('startsWith', value);
  Schema<T> endsWith(String value) => _set('endsWith', value);
  Schema<T> includes(String value) => _set('includes', value);
  Schema<T> format(String value) => _set('format', value);
  Schema<T> unknownKeys(String mode) {
    if (!['strip', 'allow', 'reject'].contains(mode))
      throw ArgumentError.value(mode);
    return _set('unknownKeys', mode);
  }

  Schema<T> defaultValue(T value) => _set('default', value);
  Schema<T> coerce(
          {bool trim = false, bool lower = false, bool upper = false}) =>
      _set('coerce', {
        if (trim) 'trim': true,
        if (lower) 'lower': true,
        if (upper) 'upper': true
      });
  Schema<T> describe(String description, {bool? sensitive, String? title}) =>
      _set('metadata', {
        ...?_node['metadata'] as Map?,
        'description': description,
        if (sensitive != null) 'sensitive': sensitive,
        if (title != null) 'title': title
      });
  Schema<T> metadata(JsonMap value) =>
      _set('metadata', {...?_node['metadata'] as Map?, ...value});
  Schema<T?> optional() => Schema._({'kind': 'optional', 'schema': _node},
      definitions: _definitions,
      extensions: _extensions,
      decode: (v) => v == null || identical(v, absent) ? null : _decode(v));
  Schema<T?> nullable() => Schema._({'kind': 'nullable', 'schema': _node},
      definitions: _definitions,
      extensions: _extensions,
      decode: (v) => v == null ? null : _decode(v));

  T parse([Object? input = absent]) {
    final result = safeParse(input);
    if (!result.success) throw ValidationError(result.issues);
    return result.data as T;
  }

  ParseResult<T> safeParse([Object? input = absent]) {
    final issues = <ValidationIssue>[];
    final value = _run(_node, input, [], issues, _definitions, 0);
    return issues.isEmpty
        ? ParseResult._(
            true, _decode(identical(value, absent) ? null : value), [])
        : ParseResult._(false, null, issues);
  }

  /// Returns a callback suitable for Flutter's TextFormField.validator.
  String? Function(String?) validator() => (value) {
        final result = safeParse(value);
        return result.success ? null : result.issues.first.message;
      };

  /// `export` is a Dart keyword, so the interchange operation is exportSchema.
  JsonMap exportSchema({String mode = 'portable'}) {
    if (!['portable', 'extended'].contains(mode))
      throw ArgumentError.value(mode);
    final document = _clone({
      'anyvaliVersion': '1.0',
      'schemaVersion': '1.1',
      'root': _node,
      'definitions': _definitions,
      'extensions': mode == 'extended' ? _extensions : <String, dynamic>{}
    });
    try {
      jsonEncode(document);
    } on JsonUnsupportedObjectError {
      throw FormatException('Schema values must be JSON-compatible');
    }
    return document;
  }
}

Schema<String> string() => Schema._({'kind': 'string'});
Schema<bool> bool_() => Schema._({'kind': 'bool'});
Schema<Null> null_() => Schema._({'kind': 'null'});
Schema<Object?> any() => Schema._({'kind': 'any'});
Schema<Object?> unknown() => Schema._({'kind': 'unknown'});
Schema<Never> never() => Schema._({'kind': 'never'});
Schema<num> number() => Schema._({'kind': 'number'});
Schema<num> float32() => Schema._({'kind': 'float32'});
Schema<num> float64() => Schema._({'kind': 'float64'});
Schema<int> int_() => Schema._({'kind': 'int'});
Schema<int> int8() => Schema._({'kind': 'int8'});
Schema<int> int16() => Schema._({'kind': 'int16'});
Schema<int> int32() => Schema._({'kind': 'int32'});
Schema<int> int64() => Schema._({'kind': 'int64'});
Schema<int> uint8() => Schema._({'kind': 'uint8'});
Schema<int> uint16() => Schema._({'kind': 'uint16'});
Schema<int> uint32() => Schema._({'kind': 'uint32'});

/// Accepts int or BigInt, allowing the full unsigned 64-bit range.
Schema<Object> uint64() => Schema._({'kind': 'uint64'});
Schema<T> literal<T>(T value) {
  if (!_primitive(value))
    throw ArgumentError('Literals must be JSON primitives');
  return Schema._({'kind': 'literal', 'value': value});
}

Schema<T> enum_<T>(List<T> values) {
  if (values.isEmpty || !values.every(_primitive))
    throw ArgumentError('An enum needs JSON primitive values');
  return Schema._({'kind': 'enum', 'values': values});
}

Schema<T> _composite<T>(JsonMap node, Iterable<Schema> children,
    {T Function(dynamic)? decode}) {
  final definitions = <String, dynamic>{};
  final extensions = <String, dynamic>{};
  for (final child in children) {
    for (final entry in child._definitions.entries) {
      if (definitions.containsKey(entry.key) &&
          !_equal(definitions[entry.key], entry.value)) {
        throw ArgumentError('Conflicting definition: ${entry.key}');
      }
      definitions[entry.key] = entry.value;
    }
    for (final entry in child._extensions.entries) {
      if (extensions.containsKey(entry.key) &&
          !_equal(extensions[entry.key], entry.value)) {
        throw ArgumentError('Conflicting extension: ${entry.key}');
      }
      extensions[entry.key] = entry.value;
    }
  }
  return Schema._(node,
      definitions: definitions, extensions: extensions, decode: decode);
}

Schema<List<T>> array<T>(Schema<T> items) =>
    _composite({'kind': 'array', 'items': items._node}, [items],
        decode: (v) => (v as List).map(items._decode).toList());
Schema<List<Object?>> tuple(List<Schema> elements) => _composite({
      'kind': 'tuple',
      'elements': elements.map((s) => s._node).toList()
    }, elements, decode: (v) => List<Object?>.from(v));
Schema<Map<String, Object?>> object(Map<String, Schema> properties) =>
    _composite({
      'kind': 'object',
      'properties': properties.map((k, v) => MapEntry(k, v._node)),
      'required': properties.entries
          .where((e) =>
              e.value._node['kind'] != 'optional' &&
              !e.value._node.containsKey('default'))
          .map((e) => e.key)
          .toList(),
      'unknownKeys': 'strip'
    }, properties.values, decode: (v) => Map<String, Object?>.from(v));
Schema<Map<String, T>> record<T>(Schema<T> values) =>
    _composite({'kind': 'record', 'values': values._node}, [values],
        decode: (v) => (v as Map)
            .map((k, value) => MapEntry(k as String, values._decode(value))));
Schema<Object?> union(List<Schema> variants) {
  if (variants.isEmpty)
    throw ArgumentError('A union needs at least one variant');
  return _composite(
      {'kind': 'union', 'variants': variants.map((s) => s._node).toList()},
      variants);
}

Schema<Object?> intersection(List<Schema> schemas) => _composite(
    {'kind': 'intersection', 'allOf': schemas.map((s) => s._node).toList()},
    schemas);

Schema<Object?> importSchema(JsonMap document) {
  const keys = {
    'anyvaliVersion',
    'schemaVersion',
    'root',
    'definitions',
    'extensions'
  };
  if (document.length != keys.length ||
      !keys.every(document.containsKey) ||
      document['anyvaliVersion'] != '1.0' ||
      !['1', '1.0', '1.1'].contains(document['schemaVersion']) ||
      document['definitions'] is! Map<String, dynamic> ||
      document['extensions'] is! Map<String, dynamic>) {
    throw FormatException('Invalid AnyVali document');
  }
  final defs = document['definitions'] as JsonMap;
  for (final name in defs.keys) {
    if (!RegExp(r'^[A-Za-z_][A-Za-z0-9_-]*$').hasMatch(name))
      throw FormatException('Invalid definition name');
  }
  _checkExtensions(document['extensions']);
  _checkNode(document['root'], defs, 0);
  for (final node in defs.values) {
    _checkNode(node, defs, 0);
  }
  return Schema._(document['root'],
      definitions: defs, extensions: document['extensions']);
}

void _checkExtensions(dynamic extensions) {
  if (extensions is! Map) throw FormatException('Invalid extensions');
  for (final value in extensions.values) {
    if (value is! Map) throw FormatException('Invalid extension namespace');
    if (value['customValidator'] != null || value['customValidators'] != null) {
      _importError('custom_validation_not_portable');
    }
    if (value['_criticality'] == 'semantic')
      _importError('unsupported_extension');
  }
}

Never _importError(String code) =>
    throw ValidationError([ValidationIssue(code, code, [])]);

const _integerKinds = {
  'int',
  'int8',
  'int16',
  'int32',
  'int64',
  'uint8',
  'uint16',
  'uint32',
  'uint64'
};
const _floatKinds = {'number', 'float32', 'float64'};
const _kinds = {
  'any',
  'unknown',
  'never',
  'null',
  'bool',
  'string',
  'literal',
  'enum',
  'array',
  'tuple',
  'object',
  'record',
  'union',
  'intersection',
  'optional',
  'nullable',
  'ref',
  ..._integerKinds,
  ..._floatKinds
};
void _checkNode(dynamic node, JsonMap definitions, int depth) {
  if (depth > 200) throw FormatException('Schema nesting too deep');
  if (node is! JsonMap || !_kinds.contains(node['kind']))
    _importError('unsupported_schema_kind');
  node = _aliases(node);
  if (node.containsKey('format') &&
      node['format'] is String &&
      !RegExp(r'^[A-Za-z][A-Za-z0-9-]*$').hasMatch(node['format']))
    throw FormatException('Invalid format name');
  if (node.containsKey('extensions')) _checkExtensions(node['extensions']);
  if (node.containsKey('customValidator') ||
      node.containsKey('customValidators'))
    _importError('custom_validation_not_portable');
  if (node.containsKey('metadata') && node['metadata'] is! Map)
    throw FormatException('Invalid metadata');
  final metadata = node['metadata'] as Map?;
  if (metadata != null) {
    for (final key in ['title', 'description', 'deprecatedMessage', 'since']) {
      if (metadata.containsKey(key) && metadata[key] is! String)
        throw FormatException('Invalid metadata $key');
    }
    for (final key in [
      'deprecated',
      'notStable',
      'sensitive',
      'readonly',
      'writeonly'
    ]) {
      if (metadata.containsKey(key) && metadata[key] is! bool)
        throw FormatException('Invalid metadata $key');
    }
    if (metadata['readonly'] == true && metadata['writeonly'] == true)
      throw FormatException('Conflicting metadata access flags');
    if (metadata.containsKey('examples') && metadata['examples'] is! List)
      throw FormatException('Invalid metadata examples');
  }
  for (final key in [
    'min',
    'max',
    'exclusiveMin',
    'exclusiveMax',
    'multipleOf'
  ]) {
    if (node.containsKey(key) &&
        (node[key] is! num ||
            !(node[key] as num).isFinite ||
            (key == 'multipleOf' && node[key] <= 0)))
      throw FormatException('Invalid $key');
  }
  for (final key in ['minLength', 'maxLength', 'minItems', 'maxItems']) {
    if (node.containsKey(key) && (node[key] is! int || node[key] < 0))
      throw FormatException('Invalid $key');
  }
  for (final key in [
    'pattern',
    'startsWith',
    'endsWith',
    'includes',
    'format'
  ]) {
    if (node.containsKey(key) && node[key] is! String)
      throw FormatException('Invalid $key');
  }
  if (node.containsKey('pattern')) RegExp(node['pattern']);
  if (node.containsKey('coerce')) _coercions(node['coerce'], node['kind']);
  void child(dynamic value) => _checkNode(value, definitions, depth + 1);
  switch (node['kind']) {
    case 'ref':
      final ref = node['ref'];
      if (ref is! String ||
          !ref.startsWith('#/definitions/') ||
          !definitions.containsKey(ref.substring(14)))
        throw FormatException('Unresolved ref: $ref');
    case 'object':
      if (node['properties'] is! JsonMap ||
          node['required'] is! List ||
          !(node['required'] as List)
              .every((k) => k is String && node['properties'].containsKey(k)) ||
          !['strip', 'allow', 'reject']
              .contains(node['unknownKeys'] ?? 'strip'))
        throw FormatException('Invalid object schema');
      for (final value in (node['properties'] as Map).values) {
        child(value);
      }
    case 'array':
      child(node['items']);
    case 'record':
      child(node['values']);
    case 'optional' || 'nullable':
      child(node['schema']);
    case 'tuple' || 'union' || 'intersection':
      final key = {
        'tuple': 'elements',
        'union': 'variants',
        'intersection': 'allOf'
      }[node['kind']]!;
      if (node[key] is! List ||
          (node['kind'] == 'union' && (node[key] as List).isEmpty))
        throw FormatException('Invalid $key');
      for (final value in node[key]) {
        child(value);
      }
    case 'literal':
      if (!node.containsKey('value') || !_primitive(node['value']))
        throw FormatException('Invalid literal');
    case 'enum':
      if (node['values'] is! List ||
          (node['values'] as List).isEmpty ||
          !(node['values'] as List).every(_primitive))
        throw FormatException('Invalid enum');
  }
}

JsonMap _aliases(JsonMap node) => {
      ...node,
      if (!node.containsKey('schema') && node.containsKey('inner'))
        'schema': node['inner'],
      if (node['kind'] == 'record' &&
          !node.containsKey('values') &&
          node.containsKey('valueSchema'))
        'values': node['valueSchema'],
      if (node['kind'] == 'tuple' &&
          !node.containsKey('elements') &&
          node.containsKey('items'))
        'elements': node['items'],
      if (node['kind'] == 'union' &&
          !node.containsKey('variants') &&
          node.containsKey('schemas'))
        'variants': node['schemas'],
      if (node['kind'] == 'intersection' &&
          !node.containsKey('allOf') &&
          node.containsKey('schemas'))
        'allOf': node['schemas'],
    };

bool _primitive(dynamic v) =>
    v == null || v is String || v is bool || (v is num && v.isFinite);

dynamic _clone(dynamic value) {
  if (value is Map)
    return <String, dynamic>{
      for (final entry in value.entries)
        entry.key as String: _clone(entry.value)
    };
  if (value is List) return value.map(_clone).toList();
  return value;
}

bool _equal(dynamic a, dynamic b) {
  if (a is Map && b is Map)
    return a.length == b.length &&
        a.keys.every((k) => b.containsKey(k) && _equal(a[k], b[k]));
  if (a is List && b is List)
    return a.length == b.length &&
        List.generate(a.length, (i) => i).every((i) => _equal(a[i], b[i]));
  return a == b;
}

// Preserve values parsed by one intersection member when another member
// passes the original key through unchanged.
dynamic _mergeParsed(dynamic original, Map left, Map right) =>
    <String, dynamic>{
      ...left,
      for (final entry in right.entries)
        entry.key as String: left.containsKey(entry.key) &&
                original is Map &&
                original.containsKey(entry.key)
            ? left[entry.key] is Map && entry.value is Map
                ? _mergeParsed(
                    original[entry.key], left[entry.key], entry.value)
                : _equal(entry.value, original[entry.key])
                    ? left[entry.key]
                    : entry.value
            : entry.value,
    };

String _type(dynamic v) => identical(v, absent)
    ? 'undefined'
    : v == null
        ? 'null'
        : v is bool
            ? 'boolean'
            : v is String
                ? 'string'
                : v is num || v is BigInt
                    ? 'number'
                    : v is List
                        ? 'array'
                        : 'object';
String _text(dynamic v) => v is int
    ? v.toString()
    : v is num && v.isFinite && v == v.truncateToDouble()
        ? v.toStringAsFixed(0)
        : '$v';

List<String> _coercions(dynamic config, String kind) {
  List<String> tokens;
  if (config is String) {
    tokens = [config];
  } else if (config is List && config.every((v) => v is String)) {
    tokens = config.cast<String>();
  } else if (config is Map &&
      config.keys.every((k) => [
            'toInt',
            'toNumber',
            'toBool',
            'trim',
            'lower',
            'upper',
            'from'
          ].contains(k)) &&
      config.entries.every(
          (e) => e.key == 'from' ? e.value == 'string' : e.value is bool)) {
    tokens = [
      if (config['toInt'] == true) 'string->int',
      if (config['toNumber'] == true) 'string->number',
      if (config['toBool'] == true) 'string->bool'
    ];
    if (tokens.isEmpty) {
      if (_integerKinds.contains(kind)) tokens.add('string->int');
      if (_floatKinds.contains(kind)) tokens.add('string->number');
      if (kind == 'bool') tokens.add('string->bool');
    }
    tokens.addAll([
      for (final key in ['trim', 'lower', 'upper'])
        if (config[key] == true) key
    ]);
  } else {
    throw FormatException('Invalid coercion');
  }
  if (!tokens.every((v) => [
        'string->int',
        'string->number',
        'string->bool',
        'trim',
        'lower',
        'upper'
      ].contains(v))) throw FormatException('Unknown coercion');
  return tokens;
}

dynamic _run(JsonMap node, dynamic input, List<Object> path,
    List<ValidationIssue> issues, JsonMap definitions, int depth,
    {bool encrypted = false}) {
  void issue(String code, dynamic expected, dynamic received) =>
      issues.add(ValidationIssue(
          code, '$code: expected $expected, received $received', path,
          expected: _text(expected), received: _text(received)));
  if (depth > 200) {
    issue('too_deep', 'bounded nesting', 'too deep');
    return null;
  }
  node = _aliases(node);
  final kind = node['kind'] as String;
  var value = input;
  if (encrypted &&
      !identical(value, absent) &&
      value != null &&
      (node['metadata'] as Map?)?['sensitive'] == true) {
    if (value is! String) {
      issue('invalid_type', 'encrypted:<value>', _type(value));
    } else if (!value.startsWith('encrypted:') || value.length == 10) {
      issue('invalid_string', 'encrypted:<value>', value);
    }
    return value;
  }
  if (!identical(value, absent) && node.containsKey('coerce')) {
    for (final token in _coercions(node['coerce'], kind)) {
      if (value is! String) continue;
      final original = value;
      final trimmed = value.trim();
      switch (token) {
        case 'trim':
          value = trimmed;
        case 'lower':
          value = value.toLowerCase();
        case 'upper':
          value = value.toUpperCase();
        case 'string->int':
          final parsed = RegExp(r'^-?[0-9]+$').hasMatch(trimmed)
              ? BigInt.tryParse(trimmed)
              : null;
          if (parsed == null) {
            issue('coercion_failed', kind, original);
            return null;
          }
          // Avoid wrapping on native platforms and rounding on the web.
          final candidate = int.tryParse(trimmed);
          value = candidate != null &&
                  BigInt.from(candidate) == parsed &&
                  (!identical(1, 1.0) || candidate.abs() <= 9007199254740991)
              ? candidate
              : parsed;
        case 'string->number':
          final parsed =
              RegExp(r'^[+-]?(?:[0-9]+\.?[0-9]*|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$')
                      .hasMatch(trimmed)
                  ? double.tryParse(trimmed)
                  : null;
          if (parsed == null || !parsed.isFinite) {
            issue('coercion_failed', kind, original);
            return null;
          }
          value = parsed;
        case 'string->bool':
          if (['true', '1'].contains(trimmed.toLowerCase())) {
            value = true;
          } else if (['false', '0'].contains(trimmed.toLowerCase())) {
            value = false;
          } else {
            issue('coercion_failed', kind, original);
            return null;
          }
      }
    }
  }
  final usedDefault = identical(value, absent) && node.containsKey('default');
  if (usedDefault) value = _clone(node['default']);
  final start = issues.length;
  dynamic child(JsonMap childNode, dynamic v, List<Object> p,
          [List<ValidationIssue>? into]) =>
      _run(childNode, v, p, into ?? issues, definitions, depth + 1,
          encrypted: encrypted);
  dynamic validate() {
    if (kind == 'ref')
      return child(
          definitions[(node['ref'] as String).substring(14)], value, path);
    if (kind == 'optional' && identical(value, absent)) return absent;
    if (kind == 'nullable' && value == null) return null;
    if (kind == 'optional' || kind == 'nullable')
      return child(node['schema'], value, path);
    if (kind == 'any' || kind == 'unknown') return value;
    if (kind == 'never') {
      issue('invalid_type', 'never', _type(value));
      return null;
    }
    if (kind == 'literal' || kind == 'enum') {
      final choices =
          kind == 'literal' ? [node['value']] : node['values'] as List;
      if (!choices.any((v) => _equal(v, value))) {
        if (kind == 'enum') {
          issue('invalid_type', 'enum(${choices.map(_text).join(',')})', value);
        } else {
          issue('invalid_literal', node['value'], value);
        }
      }
      return choices.firstWhere((v) => _equal(v, value), orElse: () => value);
    }
    if (kind == 'union') {
      for (final variant in node['variants']) {
        final branchIssues = <ValidationIssue>[];
        final result = child(variant, value, path, branchIssues);
        if (branchIssues.isEmpty) return result;
      }
      issue(
          'invalid_union',
          (node['variants'] as List).map((v) => v['kind']).join(' | '),
          _type(value));
      return null;
    }
    if (kind == 'intersection') {
      dynamic result = value;
      for (final member in node['allOf']) {
        final parsed = child(member, value, path);
        if (result is Map && parsed is Map) {
          result = _mergeParsed(value, result, parsed);
        } else {
          result = parsed;
        }
      }
      return result;
    }
    final correct = switch (kind) {
      'null' => value == null,
      'bool' => value is bool,
      'string' => value is String,
      'array' || 'tuple' => value is List,
      'object' ||
      'record' =>
        value is Map && value.keys.every((k) => k is String),
      _ => _integerKinds.contains(kind)
          ? value is BigInt ||
              value is int ||
              (value is double &&
                  value.isFinite &&
                  value == value.truncateToDouble())
          : value is num,
    };
    if (!correct) {
      issue('invalid_type', kind, _type(value));
      return null;
    }
    if (kind == 'string') {
      final s = value as String;
      if (node['minLength'] != null && s.runes.length < node['minLength'])
        issue('too_small', node['minLength'], s.runes.length);
      if (node['maxLength'] != null && s.runes.length > node['maxLength'])
        issue('too_large', node['maxLength'], s.runes.length);
      if (node['pattern'] != null && !RegExp(node['pattern']).hasMatch(s))
        issue('invalid_string', node['pattern'], s);
      if (node['startsWith'] != null && !s.startsWith(node['startsWith']))
        issue('invalid_string', node['startsWith'], s);
      if (node['endsWith'] != null && !s.endsWith(node['endsWith']))
        issue('invalid_string', node['endsWith'], s);
      if (node['includes'] != null && !s.contains(node['includes']))
        issue('invalid_string', node['includes'], s);
      if (node['format'] != null && !_format(node['format'], s))
        issue('invalid_string', node['format'], s);
    }
    if (_integerKinds.contains(kind) || _floatKinds.contains(kind)) {
      if (value is num && !value.isFinite) {
        issue('invalid_number', 'finite number', value);
        return null;
      }
      if (_integerKinds.contains(kind)) {
        if (identical(1, 1.0) &&
            value is num &&
            value.abs() > 9007199254740991) {
          issue('invalid_number', 'safe integer or BigInt', value);
          return null;
        }
        final integer = value is BigInt ? value as BigInt : BigInt.from(value);
        final unsigned = kind.startsWith('uint');
        final bits = int.parse(kind.replaceAll(RegExp(r'[^0-9]'), '').isEmpty
            ? '64'
            : kind.replaceAll(RegExp(r'[^0-9]'), ''));
        final min = unsigned ? BigInt.zero : -(BigInt.one << (bits - 1));
        final max = (BigInt.one << (unsigned ? bits : bits - 1)) - BigInt.one;
        if (integer < min) {
          issue('too_small', kind, value);
          return null;
        }
        if (integer > max) {
          issue('too_large', kind, value);
          return null;
        }
        if (identical(1, 1.0) &&
            kind != 'uint64' &&
            integer.abs() > BigInt.from(9007199254740991)) {
          issue('invalid_number', 'safe integer', value);
          return null;
        }
        value = kind == 'uint64' &&
                (integer > BigInt.parse('9223372036854775807') ||
                    (identical(1, 1.0) &&
                        integer.abs() > BigInt.from(9007199254740991)))
            ? integer
            : integer.toInt();
      }
      if (kind == 'float32') {
        final rounded = (Float32List(1)..[0] = (value as num).toDouble())[0];
        if (!rounded.isFinite) {
          issue('invalid_number', 'float32', value);
          return null;
        }
        value = rounded;
      }
      int compare(num limit) {
        if (value is! BigInt) return (value as num).compareTo(limit);
        final truncated = BigInt.from(limit);
        final comparison = (value as BigInt).compareTo(truncated);
        if (comparison != 0 || limit == limit.truncateToDouble())
          return comparison;
        return limit > 0 ? -1 : 1;
      }

      for (final key in ['min', 'max', 'exclusiveMin', 'exclusiveMax']) {
        if (node[key] == null) continue;
        final c = compare(node[key]);
        if ((key == 'min' && c < 0) || (key == 'exclusiveMin' && c <= 0))
          issue('too_small', node[key], value);
        if ((key == 'max' && c > 0) || (key == 'exclusiveMax' && c >= 0))
          issue('too_large', node[key], value);
      }
      if (node['multipleOf'] != null) {
        final divisor = node['multipleOf'] as num;
        bool matches;
        if (value is BigInt || value is int) {
          final integer =
              value is BigInt ? value as BigInt : BigInt.from(value);
          final parts = divisor.toString().toLowerCase().split('e');
          final decimal = parts.first.split('.');
          final coefficient = BigInt.parse(decimal.join());
          final exponent = (parts.length == 2 ? int.parse(parts[1]) : 0) -
              (decimal.length == 2 ? decimal[1].length : 0);
          final numerator = exponent >= 0
              ? coefficient * BigInt.from(10).pow(exponent)
              : coefficient;
          final denominator =
              exponent < 0 ? BigInt.from(10).pow(-exponent) : BigInt.one;
          matches = (integer * denominator) % numerator == BigInt.zero;
        } else {
          final remainder = (value as num) % divisor;
          matches =
              remainder.abs() <= 1e-10 || (divisor - remainder).abs() <= 1e-10;
        }
        if (!matches) issue('invalid_number', divisor, value);
      }
    }
    if (kind == 'array' || kind == 'tuple') {
      final list = value as List;
      if (kind == 'tuple' && list.length != (node['elements'] as List).length) {
        issue(
            list.length < (node['elements'] as List).length
                ? 'too_small'
                : 'too_large',
            (node['elements'] as List).length,
            list.length);
        return null;
      }
      if (node['minItems'] != null && list.length < node['minItems'])
        issue('too_small', node['minItems'], list.length);
      if (node['maxItems'] != null && list.length > node['maxItems'])
        issue('too_large', node['maxItems'], list.length);
      return [
        for (var i = 0; i < list.length; i++)
          child(kind == 'array' ? node['items'] : node['elements'][i], list[i],
              [...path, i])
      ];
    }
    if (kind == 'object' || kind == 'record') {
      final map = value as Map;
      final result = <String, dynamic>{};
      if (kind == 'record') {
        for (final entry in map.entries) {
          result[entry.key] =
              child(node['values'], entry.value, [...path, entry.key]);
        }
        return result;
      }
      final props = node['properties'] as Map;
      final required = node['required'] as List;
      for (final entry in props.entries) {
        final field = entry.value as JsonMap;
        final present = map.containsKey(entry.key);
        if (!present &&
            !field.containsKey('default') &&
            field['kind'] != 'optional') {
          if (required.contains(entry.key))
            issues.add(ValidationIssue('required',
                'Required field ${entry.key} is missing', [...path, entry.key],
                expected: field['kind'], received: 'undefined'));
          continue;
        }
        final parsed = child(
            field, present ? map[entry.key] : absent, [...path, entry.key]);
        if (!identical(parsed, absent)) result[entry.key] = parsed;
      }
      for (final key in map.keys.where((k) => !props.containsKey(k))) {
        if (node['unknownKeys'] == 'allow') result[key] = map[key];
        if (node['unknownKeys'] == 'reject')
          issues.add(ValidationIssue(
              'unknown_key', 'Unknown key $key', [...path, key],
              expected: 'undefined', received: key));
      }
      return result;
    }
    return value;
  }

  final result = validate();
  if (usedDefault && issues.length > start) {
    final expected = issues[start].expected;
    issues.removeRange(start, issues.length);
    issue('default_invalid', expected, value);
    return null;
  }
  return result;
}

bool _format(String format, String value) {
  switch (format) {
    case 'email':
      return RegExp(r'^[^\s@]+@[^\s@]+\.[^\s@]+$').hasMatch(value);
    case 'url':
      return RegExp(r'^https?://.+$').hasMatch(value);
    case 'uuid':
      return RegExp(
              r'^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$')
          .hasMatch(value);
    case 'ipv4':
      final parts = value.split('.');
      return parts.length == 4 &&
          parts.every((p) =>
              RegExp(r'^(0|[1-9][0-9]{0,2})$').hasMatch(p) &&
              int.parse(p) <= 255);
    case 'ipv6':
      try {
        return Uri.parse('http://[$value]/').host.isNotEmpty &&
            value.contains(':');
      } on FormatException {
        return false;
      }
    case 'date':
      if (!RegExp(r'^[0-9]{4}-[0-9]{2}-[0-9]{2}$').hasMatch(value))
        return false;
      final date = DateTime.tryParse(value);
      return date != null && date.toIso8601String().substring(0, 10) == value;
    case 'date-time':
      if (!RegExp(
              r'^[0-9]{4}-[0-9]{2}-[0-9]{2}T([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]+)?(Z|[+-]([01][0-9]|2[0-3]):[0-5][0-9])$')
          .hasMatch(value)) return false;
      return _format('date', value.substring(0, 10)) &&
          DateTime.tryParse(value) != null;
    default:
      return RegExp(r'^[A-Za-z][A-Za-z0-9-]*$').hasMatch(format);
  }
}

ParseResult<Object?> safeParseEncrypted(Schema schema,
    [Object? input = absent]) {
  final issues = <ValidationIssue>[];
  final value = _run(schema._node, input, [], issues, schema._definitions, 0,
      encrypted: true);
  return ParseResult._(issues.isEmpty, issues.isEmpty ? value : null, issues);
}

Object? _transform(Schema schema, Object? input, SensitiveTransform transform,
    bool decrypting) {
  final checked =
      decrypting ? safeParseEncrypted(schema, input) : schema.safeParse(input);
  if (!checked.success) throw ValidationError(checked.issues);
  // Resolve unions against validated data before invoking user callbacks.
  dynamic walk(JsonMap node, dynamic value, List<Object> path, int depth) {
    node = _aliases(node);
    if (depth > 200)
      throw ValidationError(
          [ValidationIssue('too_deep', 'Maximum depth exceeded', path)]);
    if (value == null || identical(value, absent)) return value;
    if ((node['metadata'] as Map?)?['sensitive'] == true)
      return transform(List.unmodifiable(path), value);
    final kind = node['kind'];
    if (kind == 'ref')
      return walk(schema._definitions[(node['ref'] as String).substring(14)],
          value, path, depth + 1);
    if (kind == 'optional' || kind == 'nullable')
      return walk(node['schema'], value, path, depth + 1);
    if (kind == 'union') {
      for (final branch in node['variants']) {
        final issues = <ValidationIssue>[];
        _run(branch, value, path, issues, schema._definitions, depth + 1,
            encrypted: decrypting);
        if (issues.isEmpty) return walk(branch, value, path, depth + 1);
      }
    }
    if (kind == 'intersection') {
      // Cache callback results by path so overlapping members transform once.
      dynamic result = value;
      for (final member in node['allOf']) {
        final next = walk(member, value, path, depth + 1);
        result = result is Map && next is Map
            ? _mergeParsed(value, result, next)
            : next;
      }
      return result;
    }
    if (value is Map && (kind == 'object' || kind == 'record')) {
      return value.map((k, v) {
        final child = kind == 'record' ? node['values'] : node['properties'][k];
        return MapEntry(
            k, child == null ? v : walk(child, v, [...path, k], depth + 1));
      });
    }
    if (value is List && (kind == 'array' || kind == 'tuple')) {
      return [
        for (var i = 0; i < value.length; i++)
          walk(kind == 'array' ? node['items'] : node['elements'][i], value[i],
              [...path, i], depth + 1)
      ];
    }
    return value;
  }

  final cache = <String, Object?>{};
  final originalTransform = transform;
  transform = (path, value) =>
      cache.putIfAbsent(jsonEncode(path), () => originalTransform(path, value));
  final output = walk(schema._node, checked.data, [], 0);
  final result = decrypting
      ? schema.safeParse(output)
      : safeParseEncrypted(schema, output);
  if (!result.success) throw ValidationError(result.issues);
  return result.data;
}

Object? encrypt(Schema schema, Object? input, SensitiveTransform transform) =>
    _transform(schema, input, transform, false);
Object? decrypt(Schema schema, Object? input, SensitiveTransform transform) =>
    _transform(schema, input, transform, true);
