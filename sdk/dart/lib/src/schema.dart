import 'dart:convert';

/// An omitted value, distinct from an explicit null.
const absent = _Absent();

class _Absent {
  const _Absent();
}

typedef JsonMap = Map<String, dynamic>;
typedef _UnionTrace = Map<JsonMap, Map<String, int>>;

typedef SensitiveTransform = Object? Function(List<Object> path, Object? value);

/// A validation failure with an ordered object-key and array-index path.
class ValidationIssue {
  final String code;
  final String message;
  final List<Object> path;
  final String? expected;
  final String? received;
  ValidationIssue(this.code, this.message, List<Object> path,
      {this.expected, this.received})
      : path = List.unmodifiable(path);

  /// Encode portable issue fields, omitting unavailable diagnostics.
  JsonMap toJson() => {
        'code': code,
        'message': message,
        'path': path,
        if (expected != null) 'expected': expected,
        if (received != null) 'received': received
      };
}

/// Thrown by parse and transform helpers when validation fails.
class ValidationError implements Exception {
  final List<ValidationIssue> issues;
  ValidationError(List<ValidationIssue> issues)
      : issues = List.unmodifiable(issues);
  @override
  String toString() =>
      'ValidationError: ${issues.map((i) => '${i.path}: ${i.message}').join('; ')}';
}

/// A non-throwing validation result; inspect success before reading data.
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

  /// Snapshot node context and retain the typed output decoder.
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

  /// Require a numeric value greater than or equal to value.
  Schema<T> min(num value) => _set('min', value);

  /// Require a numeric value less than or equal to value.
  Schema<T> max(num value) => _set('max', value);

  /// Require a numeric value strictly greater than value.
  Schema<T> exclusiveMin(num value) => _set('exclusiveMin', value);

  /// Require a numeric value strictly less than value.
  Schema<T> exclusiveMax(num value) => _set('exclusiveMax', value);

  /// Require a numeric multiple of a positive finite divisor.
  Schema<T> multipleOf(num value) {
    if (!value.isFinite || value <= 0)
      throw ArgumentError.value(value, 'value');
    return _set('multipleOf', value);
  }

  /// Set the minimum string length in Unicode code points.
  Schema<T> minLength(int value) => _set('minLength', value);

  /// Set the maximum string length in Unicode code points.
  Schema<T> maxLength(int value) => _set('maxLength', value);

  /// Set the minimum array length.
  Schema<T> minItems(int value) => _set('minItems', value);

  /// Set the maximum array length.
  Schema<T> maxItems(int value) => _set('maxItems', value);

  /// Require a string to match the supplied regular expression.
  Schema<T> pattern(String value) {
    RegExp(value);
    return _set('pattern', value);
  }

  /// Require a string prefix.
  Schema<T> startsWith(String value) => _set('startsWith', value);

  /// Require a string suffix.
  Schema<T> endsWith(String value) => _set('endsWith', value);

  /// Require a string substring.
  Schema<T> includes(String value) => _set('includes', value);

  /// Require one of the seven portable string formats.
  Schema<T> format(String value) => _set('format', value);

  /// Configure strip, allow, or reject behavior for object extras.
  Schema<T> unknownKeys(String mode) {
    if (!['strip', 'allow', 'reject'].contains(mode))
      throw ArgumentError.value(mode);
    return _set('unknownKeys', mode);
  }

  /// Materialize and validate a cloned default only when input is absent.
  Schema<T> defaultValue(T value) => _set('default', value);

  /// Enable type coercion and optional trim, lower, or upper string transforms.
  Schema<T> coerce(
          {bool trim = false, bool lower = false, bool upper = false}) =>
      _set('coerce', {
        if (trim) 'trim': true,
        if (lower) 'lower': true,
        if (upper) 'upper': true
      });

  /// Merge a description and typed reserved metadata options into this schema.
  Schema<T> describe(String description,
          {String? title,
          bool? deprecated,
          String? deprecatedMessage,
          bool? notStable,
          String? since,
          bool? sensitive,
          bool? readonly,
          bool? writeonly,
          List<Object?>? examples}) =>
      _set('metadata', {
        ...?_node['metadata'] as Map?,
        'description': description,
        if (title != null) 'title': title,
        if (deprecated != null) 'deprecated': deprecated,
        if (deprecatedMessage != null) 'deprecatedMessage': deprecatedMessage,
        if (notStable != null) 'notStable': notStable,
        if (since != null) 'since': since,
        if (sensitive != null) 'sensitive': sensitive,
        if (readonly != null) 'readonly': readonly,
        if (writeonly != null) 'writeonly': writeonly,
        if (examples != null) 'examples': examples,
      });

  /// Merge custom metadata; reserved names must be configured with describe.
  /// With replace, remove previous custom keys while retaining reserved fields.
  Schema<T> metadata(JsonMap value, {bool replace = false}) {
    for (final key in value.keys) {
      if (_reservedMetadata.contains(key))
        throw ArgumentError(
            'Reserved metadata key $key: use describe() instead');
    }
    final previous = _node['metadata'] as Map?;
    return _set('metadata', {
      if (previous != null)
        for (final entry in previous.entries)
          if (!replace || _reservedMetadata.contains(entry.key))
            entry.key: entry.value,
      ...value,
    });
  }

  /// Accept absent input while retaining validation for present values.
  Schema<T?> optional() => Schema._({'kind': 'optional', 'schema': _node},
      definitions: _definitions,
      extensions: _extensions,
      decode: (v) => v == null || identical(v, absent) ? null : _decode(v));

  /// Accept an explicit null as well as the inner schema values.
  Schema<T?> nullable() => Schema._({'kind': 'nullable', 'schema': _node},
      definitions: _definitions,
      extensions: _extensions,
      decode: (v) => v == null ? null : _decode(v));

  /// Return typed parsed data or throw ValidationError with all issues.
  T parse([Object? input = absent]) {
    final result = safeParse(input);
    if (!result.success) throw ValidationError(result.issues);
    return result.data as T;
  }

  /// Validate input without throwing for ordinary validation failures.
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
      'root': mode == 'portable' ? _portableNode(_node) : _node,
      'definitions': mode == 'portable'
          ? _definitions
              .map((key, value) => MapEntry(key, _portableNode(value)))
          : _definitions,
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

/// Build a string schema.
Schema<String> string() => Schema._({'kind': 'string'});

/// Build a boolean schema.
Schema<bool> bool_() => Schema._({'kind': 'bool'});

/// Build a schema accepting only null.
Schema<Null> null_() => Schema._({'kind': 'null'});

/// Build a schema accepting any value, including absence.
Schema<Object?> any() => Schema._({'kind': 'any'});

/// Build a schema accepting unknown values without constraints.
Schema<Object?> unknown() => Schema._({'kind': 'unknown'});

/// Build a schema rejecting every value.
Schema<Never> never() => Schema._({'kind': 'never'});

/// Build a finite numeric schema using the float64 alias.
Schema<num> number() => Schema._({'kind': 'number'});

/// Build a finite numeric schema restricted to the IEEE 754 binary32 range.
Schema<num> float32() => Schema._({'kind': 'float32'});

/// Build a finite numeric schema for the float64 kind.
Schema<num> float64() => Schema._({'kind': 'float64'});

/// Build a signed 64-bit integer schema.
Schema<int> int_() => Schema._({'kind': 'int'});

/// Build a signed 8-bit integer schema.
Schema<int> int8() => Schema._({'kind': 'int8'});

/// Build a signed 16-bit integer schema.
Schema<int> int16() => Schema._({'kind': 'int16'});

/// Build a signed 32-bit integer schema.
Schema<int> int32() => Schema._({'kind': 'int32'});

/// Build a signed 64-bit integer schema.
Schema<int> int64() => Schema._({'kind': 'int64'});

/// Build a unsigned 8-bit integer schema.
Schema<int> uint8() => Schema._({'kind': 'uint8'});

/// Build a unsigned 16-bit integer schema.
Schema<int> uint16() => Schema._({'kind': 'uint16'});

/// Build a unsigned 32-bit integer schema.
Schema<int> uint32() => Schema._({'kind': 'uint32'});

/// Accepts int or BigInt, allowing the full unsigned 64-bit range.
Schema<Object> uint64() => Schema._({'kind': 'uint64'});

/// Build a schema accepting exactly the given JSON primitive.
Schema<T> literal<T>(T value) {
  if (!_primitive(value))
    throw ArgumentError('Literals must be JSON primitives');
  return Schema._({'kind': 'literal', 'value': value});
}

/// Build a schema accepting any of the supplied JSON primitives.
Schema<T> enum_<T>(List<T> values) {
  if (values.isEmpty || !values.every(_primitive))
    throw ArgumentError('An enum needs JSON primitive values');
  return Schema._({'kind': 'enum', 'values': values});
}

/// Combine child schemas while retaining imported definitions and extensions.
/// Conflicting imported names fail instead of changing reference semantics.
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

/// Build a homogeneous array preserving the item output type.
Schema<List<T>> array<T>(Schema<T> items) =>
    _composite({'kind': 'array', 'items': items._node}, [items],
        decode: (v) => (v as List).map(items._decode).toList());

/// Build a fixed-length array with a schema for each position.
Schema<List<Object?>> tuple(List<Schema> elements) => _composite({
      'kind': 'tuple',
      'elements': elements.map((s) => s._node).toList()
    }, elements, decode: (v) => List<Object?>.from(v));

/// Build an object, stripping unknown keys by default.
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

/// Build a string-keyed record with typed homogeneous values.
Schema<Map<String, T>> record<T>(Schema<T> values) =>
    _composite({'kind': 'record', 'values': values._node}, [values],
        decode: (v) => (v as Map)
            .map((k, value) => MapEntry(k as String, values._decode(value))));

/// Try alternatives in order and return the first successful parsed value.
Schema<Object?> union(List<Schema> variants) {
  if (variants.isEmpty)
    throw ArgumentError('A union needs at least one variant');
  return _composite(
      {'kind': 'union', 'variants': variants.map((s) => s._node).toList()},
      variants);
}

/// Require every member and merge their parsed results.
Schema<Object?> intersection(List<Schema> schemas) => _composite(
    {'kind': 'intersection', 'allOf': schemas.map((s) => s._node).toList()},
    schemas);

/// Validate a portable document and retain its definitions and metadata.
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
      !_stringMap(document['definitions']) ||
      !_stringMap(document['extensions'])) {
    throw FormatException('Invalid AnyVali document');
  }
  final defs = Map<String, dynamic>.from(document['definitions']);
  final extensions = Map<String, dynamic>.from(document['extensions']);
  for (final name in defs.keys) {
    if (!RegExp(r'^[A-Za-z_][A-Za-z0-9_-]*$').hasMatch(name))
      throw FormatException('Invalid definition name');
  }
  _checkExtensions(extensions);
  _checkNode(document['root'], defs, 0);
  for (final node in defs.values) {
    _checkNode(node, defs, 0);
  }
  return Schema._(Map<String, dynamic>.from(document['root']),
      definitions: defs, extensions: extensions);
}

/// Accept JSON objects by their string keys rather than Dart generic arguments.
/// Empty map literals commonly have dynamic key types before normalization.
bool _stringMap(dynamic value) =>
    value is Map && value.keys.every((key) => key is String);

/// Validate criticality and reject custom validators or unknown semantic rules.
void _checkExtensions(dynamic extensions) {
  if (!_stringMap(extensions)) throw FormatException('Invalid extensions');
  for (final value in extensions.values) {
    if (!_stringMap(value))
      throw FormatException('Invalid extension namespace');
    if (value['customValidator'] != null || value['customValidators'] != null) {
      _importError('custom_validation_not_portable');
    }
    if (value.containsKey('_criticality') &&
        !['informational', 'semantic'].contains(value['_criticality']))
      throw FormatException('Invalid extension criticality');
    if (value['_criticality'] == 'semantic')
      _importError('unsupported_extension');
  }
}

/// Raise a schema-import failure using a portable validation issue code.
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
const _reservedMetadata = {
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
};
const _formats = {'email', 'url', 'uuid', 'ipv4', 'ipv6', 'date', 'date-time'};
const _numericProperties = {
  'min',
  'max',
  'exclusiveMin',
  'exclusiveMax',
  'multipleOf'
};
const _nodeProperties = {
  'string': {
    'minLength',
    'maxLength',
    'pattern',
    'startsWith',
    'endsWith',
    'includes',
    'format'
  },
  'literal': {'value'},
  'enum': {'values'},
  'array': {'items', 'minItems', 'maxItems'},
  'tuple': {'elements', 'items'},
  'object': {'properties', 'required', 'unknownKeys'},
  'record': {'values', 'valueSchema'},
  'union': {'variants', 'schemas'},
  'intersection': {'allOf', 'schemas'},
  'optional': {'schema', 'inner'},
  'nullable': {'schema', 'inner'},
  'ref': {'ref'},
};

/// Check a node against its kind-specific shape before constructing a schema.
/// Recurse through structural children while refs resolve through definitions.
void _checkNode(dynamic node, JsonMap definitions, int depth) {
  if (depth > 200) throw FormatException('Schema nesting too deep');
  if (!_stringMap(node) || !_kinds.contains(node['kind']))
    _importError('unsupported_schema_kind');
  node = Map<String, dynamic>.from(node);
  if (node.containsKey('customValidator') ||
      node.containsKey('customValidators'))
    _importError('custom_validation_not_portable');
  final kind = node['kind'];
  final allowed = {
    'kind',
    'default',
    'coerce',
    'metadata',
    'extensions',
    ...?_nodeProperties[kind],
    if (_integerKinds.contains(kind) || _floatKinds.contains(kind))
      ..._numericProperties
  };
  if ((node as Map).keys.any((key) => !allowed.contains(key)))
    throw FormatException('Unsupported property for $kind');
  for (final pair in const {
    'schema': 'inner',
    'values': 'valueSchema',
    'elements': 'items',
    'variants': 'schemas',
    'allOf': 'schemas'
  }.entries) {
    if (allowed.contains(pair.value) &&
        node.containsKey(pair.key) &&
        node.containsKey(pair.value) &&
        !_equal(node[pair.key], node[pair.value]))
      throw FormatException('Conflicting schema aliases');
  }
  node = _aliases(node as JsonMap);
  if (node.containsKey('format') && !_formats.contains(node['format']))
    throw FormatException('Unsupported portable format');
  if (node.containsKey('extensions')) _checkExtensions(node['extensions']);
  if (node.containsKey('customValidator') ||
      node.containsKey('customValidators'))
    _importError('custom_validation_not_portable');
  if (node.containsKey('metadata') && !_stringMap(node['metadata']))
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
    if (metadata.containsKey('deprecatedMessage') &&
        metadata['deprecated'] != true)
      throw FormatException('deprecatedMessage requires deprecated: true');
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
  if (identical(1, 1.0) && _integerKinds.contains(kind)) {
    for (final key in _numericProperties) {
      if (node.containsKey(key) &&
          (node[key] as num).abs() > 9007199254740991) {
        throw FormatException(
            'Integer constraints beyond the JavaScript safe range cannot be imported exactly');
      }
    }
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
      if (!_stringMap(node['properties']) ||
          node['required'] is! List ||
          !(node['required'] as List)
              .every((k) => k is String && node['properties'].containsKey(k)) ||
          !['strip', 'allow', 'reject'].contains(
              node.containsKey('unknownKeys') ? node['unknownKeys'] : 'strip'))
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

/// Expose canonical child keys for supported legacy interchange spellings.
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

/// Remove informational extensions from schema nodes in portable mode.
/// Traverse schema edges only; defaults and metadata remain unchanged as data.
JsonMap _portableNode(JsonMap node, [int depth = 0]) {
  if (depth > 200) throw FormatException('Schema nesting too deep');
  final output = {...node}..remove('extensions');
  switch (node['kind']) {
    case 'object':
      output['properties'] = (node['properties'] as Map)
          .map((key, value) => MapEntry(key, _portableNode(value, depth + 1)));
    case 'array' || 'record' || 'optional' || 'nullable':
      final keys = switch (node['kind']) {
        'array' => ['items'],
        'record' => ['values', 'valueSchema'],
        _ => ['schema', 'inner'],
      };
      for (final key in keys) {
        if (node.containsKey(key))
          output[key] = _portableNode(node[key], depth + 1);
      }
    case 'tuple' || 'union' || 'intersection':
      final keys = switch (node['kind']) {
        'tuple' => ['elements', 'items'],
        'union' => ['variants', 'schemas'],
        _ => ['allOf', 'schemas'],
      };
      for (final key in keys) {
        if (node.containsKey(key))
          output[key] = [
            for (final child in node[key]) _portableNode(child, depth + 1)
          ];
      }
  }
  return output;
}

/// Recognize finite JSON primitive values allowed in literals and enums.
bool _primitive(dynamic v) =>
    v == null || v is String || v is bool || (v is num && v.isFinite);

/// Recursively copy map and list containers used by schemas and defaults.
dynamic _clone(dynamic value) {
  if (value is Map)
    return <String, dynamic>{
      for (final entry in value.entries)
        entry.key as String: _clone(entry.value)
    };
  if (value is List) return value.map(_clone).toList();
  return value;
}

/// Compare JSON structures with bounded recursion and shared-node shortcuts.
bool _equal(dynamic a, dynamic b, [int depth = 0]) {
  if (identical(a, b)) return true;
  if (depth > 200) return false;
  if (a is Map && b is Map)
    return a.length == b.length &&
        a.keys.every((k) => b.containsKey(k) && _equal(a[k], b[k], depth + 1));
  if (a is List && b is List)
    return a.length == b.length &&
        List.generate(a.length, (i) => i)
            .every((i) => _equal(a[i], b[i], depth + 1));
  return a == b;
}

/// Preserve values parsed by one intersection member when another member
/// passes the original key through unchanged.
dynamic _mergeParsed(dynamic original, Map left, Map right, [int depth = 0]) {
  if (identical(left, right) || depth > 200) return left;
  return <String, dynamic>{
    ...left,
    for (final entry in right.entries)
      entry.key as String: left.containsKey(entry.key) &&
              original is Map &&
              original.containsKey(entry.key)
          ? left[entry.key] is Map && entry.value is Map
              ? _mergeParsed(
                  original[entry.key], left[entry.key], entry.value, depth + 1)
              : _equal(entry.value, original[entry.key])
                  ? left[entry.key]
                  : entry.value
          : entry.value,
  };
}

/// Describe a value using portable type names, including the absence sentinel.
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

/// Format scalar diagnostic values while preserving native integer precision.
String _text(dynamic v) => v is int
    ? v.toString()
    : v is num && v.isFinite && v == v.truncateToDouble()
        ? v.toStringAsFixed(0)
        : '$v';

/// Normalize string, list, or legacy flag coercions into ordered portable tokens.
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

/// Find a default through transparent wrappers and references without looping.
(bool, dynamic) _defaultFor(JsonMap node, JsonMap definitions,
    [int depth = 0]) {
  if (depth > 200) return (false, null);
  if (node.containsKey('default')) return (true, node['default']);
  node = _aliases(node);
  if (node['kind'] == 'optional' || node['kind'] == 'nullable') {
    return _defaultFor(node['schema'], definitions, depth + 1);
  }
  if (node['kind'] == 'ref') {
    final target = definitions[(node['ref'] as String).substring(14)];
    if (target != null) return _defaultFor(target, definitions, depth + 1);
  }
  return (false, null);
}

/// Run presence, coercion, defaults, and validation for one schema node.
/// Preserve inherited key policies and successful union selections, limit depth,
/// and skip root coercion when a transparent wrapper materializes a default.
dynamic _run(JsonMap node, dynamic input, List<Object> path,
    List<ValidationIssue> issues, JsonMap definitions, int depth,
    {bool encrypted = false,
    _UnionTrace? unions,
    String? inheritedUnknownKeys,
    bool skipCoercion = false}) {
  /// Append an issue at this node without stopping sibling validation.
  void issue(String code, dynamic expected, dynamic received) =>
      issues.add(ValidationIssue(
          code, '$code: expected $expected, received $received', path,
          expected: _text(expected), received: _text(received)));
  if (depth > 200) {
    issue('too_deep', 'bounded nesting', 'too deep');
    return null;
  }
  final sourceNode = node;
  node = _aliases(node);
  final kind = node['kind'] as String;
  final unknownKeys = kind == 'object'
      ? inheritedUnknownKeys ?? node['unknownKeys'] ?? 'strip'
      : inheritedUnknownKeys;
  var value = input;
  final encryptedSensitive =
      encrypted && (node['metadata'] as Map?)?['sensitive'] == true;
  if (!skipCoercion &&
      !identical(value, absent) &&
      !(encryptedSensitive && value != null) &&
      node.containsKey('coerce')) {
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
              RegExp(r'^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$')
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
  final defaultInfo =
      identical(value, absent) ? _defaultFor(node, definitions) : (false, null);
  final usedDefault = identical(value, absent) && defaultInfo.$1;
  if (usedDefault) value = _clone(defaultInfo.$2);
  final start = issues.length;

  /// Validate a child path, optionally using isolated union-branch state.
  dynamic child(JsonMap childNode, dynamic v, List<Object> p,
          [List<ValidationIssue>? into, _UnionTrace? selected]) =>
      _run(childNode, v, p, into ?? issues, definitions, depth + 1,
          encrypted: encrypted,
          unions: selected ?? unions,
          inheritedUnknownKeys: unknownKeys == 'allow' ? null : unknownKeys);

  /// Delegate a transparent wrapper while preserving default/coercion order.
  dynamic delegate(JsonMap inner) =>
      _run(inner, value, path, issues, definitions, depth + 1,
          encrypted: encrypted,
          unions: unions,
          inheritedUnknownKeys: unknownKeys == 'allow' ? null : unknownKeys,
          skipCoercion: usedDefault || skipCoercion);

  /// Validate the current value after presence, coercion, and default handling.
  dynamic validate() {
    if (encryptedSensitive && !identical(value, absent) && value != null) {
      if (value is! String) {
        issue('invalid_type', 'encrypted:<value>', _type(value));
      } else if (!value.startsWith('encrypted:') || value.length == 10) {
        issue('invalid_string', 'encrypted:<value>', value);
      }
      return value;
    }
    if (kind == 'ref')
      return delegate(definitions[(node['ref'] as String).substring(14)]);
    if (kind == 'optional' && identical(value, absent)) return absent;
    if (kind == 'nullable' && value == null) return null;
    if (kind == 'optional' || kind == 'nullable')
      return delegate(node['schema']);
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
      final variants = node['variants'] as List;
      for (var index = 0; index < variants.length; index++) {
        final branchIssues = <ValidationIssue>[];
        final branchUnions =
            unions == null ? null : <JsonMap, Map<String, int>>{};
        final result =
            child(variants[index], value, path, branchIssues, branchUnions);
        if (branchIssues.isEmpty) {
          if (unions != null) {
            for (final entry in branchUnions!.entries) {
              unions.putIfAbsent(entry.key, () => {}).addAll(entry.value);
            }
            unions.putIfAbsent(sourceNode, () => {})[jsonEncode(path)] = index;
          }
          return result;
        }
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
        const maxFloat32 = 3.4028234663852886e38;
        if ((value as num).abs() > maxFloat32) {
          issue('invalid_number', 'float32', value);
          return null;
        }
      }

      /// Compare finite bounds with ordinary numeric equality for signed zero.
      int compare(num limit) {
        if (value is! BigInt) {
          final numeric = value as num;
          return numeric == limit
              ? 0
              : numeric < limit
                  ? -1
                  : 1;
        }
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
        if (!_multipleOf(value, divisor,
            allowTolerance: !_integerKinds.contains(kind))) {
          issue('invalid_number', divisor, value);
        }
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
        final present =
            map.containsKey(entry.key) && !identical(map[entry.key], absent);
        if (!present && !_defaultFor(field, definitions).$1) {
          if (required.contains(entry.key)) {
            issues.add(ValidationIssue('required',
                'Required field ${entry.key} is missing', [...path, entry.key],
                expected: field['kind'], received: 'undefined'));
            continue;
          }
          if (field['kind'] != 'optional') continue;
        }
        final parsed = child(
            field, present ? map[entry.key] : absent, [...path, entry.key]);
        if (!identical(parsed, absent)) result[entry.key] = parsed;
      }
      for (final key in map.keys.where((k) => !props.containsKey(k))) {
        if (unknownKeys == 'allow') result[key] = map[key];
        if (unknownKeys == 'reject')
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

/// Convert a finite numeric value's shortest decimal spelling to an exact ratio.
/// This avoids binary remainder error growing with the magnitude of the input.
(BigInt, BigInt) _decimalRatio(dynamic value) {
  if (value is BigInt) return (value, BigInt.one);
  final parts = (value as num).toString().toLowerCase().split('e');
  final decimal = parts.first.split('.');
  final coefficient = BigInt.parse(decimal.join());
  final exponent = (parts.length == 2 ? int.parse(parts[1]) : 0) -
      (decimal.length == 2 ? decimal[1].length : 0);
  return exponent >= 0
      ? (coefficient * BigInt.from(10).pow(exponent), BigInt.one)
      : (coefficient, BigInt.from(10).pow(-exponent));
}

/// Check decimal multiples exactly, with a bounded relative tolerance for floats.
/// Tolerance never grows with the value and never exceeds the spec's 1e-10 cap.
bool _multipleOf(dynamic value, num divisor, {required bool allowTolerance}) {
  final (numerator, denominator) = _decimalRatio(value);
  final (step, scale) = _decimalRatio(divisor);
  final modulus = step * denominator;
  final remainder = (numerator * scale) % modulus;
  if (remainder == BigInt.zero) return true;
  if (!allowTolerance) return false;
  final tolerance = divisor < 1 ? divisor * 1e-10 : 1e-10;
  final (epsilon, epsilonScale) = _decimalRatio(tolerance);
  final distance =
      remainder < modulus - remainder ? remainder : modulus - remainder;
  return distance * epsilonScale <= epsilon * denominator * scale;
}

/// Validate a portable string format without changing the input.
bool _format(String format, String value) {
  switch (format) {
    case 'email':
      final at = value.indexOf('@');
      if (at <= 0 ||
          at != value.lastIndexOf('@') ||
          RegExp(r'\s').hasMatch(value)) return false;
      final domain = value.substring(at + 1);
      return domain.length >= 3 &&
          domain.substring(1, domain.length - 1).contains('.');
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
        Uri.parseIPv6Address(value);
        return value.contains(':');
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
      return false;
  }
}

/// Validate encrypted storage without decrypting sensitive values.
ParseResult<Object?> safeParseEncrypted(Schema schema,
    [Object? input = absent]) {
  final issues = <ValidationIssue>[];
  final value = _run(schema._node, input, [], issues, schema._definitions, 0,
      encrypted: true);
  return ParseResult._(issues.isEmpty, issues.isEmpty ? value : null, issues);
}

/// Validate before and after walking sensitive nodes with caller-owned transforms.
/// Retain original union selections and cache callback results by data path.
Object? _transform(Schema schema, Object? input, SensitiveTransform transform,
    bool decrypting) {
  final unions = <JsonMap, Map<String, int>>{};
  final issues = <ValidationIssue>[];
  final checked = _run(schema._node, input, [], issues, schema._definitions, 0,
      encrypted: decrypting, unions: unions);
  if (issues.isNotEmpty) throw ValidationError(issues);
  // Reuse the branches selected from the original input. Parsed values may
  // match an earlier, non-sensitive variant after coercion or defaults.
  /// Walk the validated shape using retained union selections and opaque secrets.
  dynamic walk(JsonMap node, dynamic value, List<Object> path, int depth) {
    final sourceNode = node;
    node = _aliases(node);
    if (depth > 200)
      throw ValidationError(
          [ValidationIssue('too_deep', 'Maximum depth exceeded', path)]);
    if (value == null || identical(value, absent)) return value;
    if ((node['metadata'] as Map?)?['sensitive'] == true) {
      final transformed = transform(List.unmodifiable(path), value);
      if (!decrypting &&
          (transformed is! String ||
              !transformed.startsWith('encrypted:') ||
              transformed.length <= 10)) {
        throw ValidationError([
          ValidationIssue(
              transformed is String ? 'invalid_string' : 'invalid_type',
              'Sensitive encryption must return encrypted:<payload>',
              path,
              expected: 'encrypted:<value>',
              received: _type(transformed))
        ]);
      }
      return transformed;
    }
    final kind = node['kind'];
    if (kind == 'ref')
      return walk(schema._definitions[(node['ref'] as String).substring(14)],
          value, path, depth + 1);
    if (kind == 'optional' || kind == 'nullable')
      return walk(node['schema'], value, path, depth + 1);
    if (kind == 'union') {
      final selected = unions[sourceNode]?[jsonEncode(path)];
      if (selected == null) throw StateError('Missing validated union branch');
      return walk(node['variants'][selected], value, path, depth + 1);
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
  final output = walk(schema._node, checked, [], 0);
  final result = decrypting
      ? schema.safeParse(output)
      : safeParseEncrypted(schema, output);
  if (!result.success) throw ValidationError(result.issues);
  return result.data;
}

/// Parse plaintext, transform sensitive nodes, and validate encrypted storage.
Object? encrypt(Schema schema, Object? input, SensitiveTransform transform) =>
    _transform(schema, input, transform, false);

/// Validate encrypted storage, transform sensitive nodes, and parse plaintext.
Object? decrypt(Schema schema, Object? input, SensitiveTransform transform) =>
    _transform(schema, input, transform, true);
