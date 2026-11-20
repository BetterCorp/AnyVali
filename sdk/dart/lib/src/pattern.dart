/// Bounded regular-expression matching for portable schema patterns.
/// Regular patterns use Thompson states. Backreferences use ordered capture
/// states with explicit operation and explored-state limits.
library;

/// Raised when a valid pattern exceeds the per-match work or input limit.
class PatternWorkLimit implements Exception {}

/// A compiled pattern with a bounded number of Thompson states.
class BoundedPattern {
  final List<_State> _states;
  final int _start;
  final int _accept;
  final int _captureCount;
  final bool _backrefs;
  final int _registerCount;
  BoundedPattern._(this._states, this._start, this._accept, this._captureCount,
      this._backrefs, this._registerCount);

  /// Search using deduplicated states at each UTF-16 input position.
  bool hasMatch(String input) {
    if (input.length > 100000) throw PatternWorkLimit();
    final work = _Work();
    if (_backrefs) return _searchVm(input, work);
    return _match(input, _start, _accept, 0, work, search: true);
  }

  /// Search capture-aware patterns without using the platform backtracking engine.
  bool _searchVm(String input, _Work work) {
    for (var position = 0; position <= input.length; position++) {
      work.spend(_captureCount * 2 + _registerCount + 1);
      final captures = List<int>.filled(_captureCount * 2 + _registerCount, -1);
      if (_runVm(input, _start, _accept, position, captures, work) != null)
        return true;
    }
    return false;
  }

  /// Execute ordered alternatives with bounded states and immutable captures.
  _Frame? _runVm(String input, int start, int accept, int position,
      List<int> captures, _Work work,
      {bool backwards = false}) {
    final pending = [_Frame(start, position, captures)];
    final seen = <(int, int, String)>{};
    while (pending.isNotEmpty) {
      work.step();
      final frame = pending.removeLast();
      work.spend(frame.captures.length);
      final key = (frame.pc, frame.position, frame.captures.join(','));
      if (!seen.add(key)) continue;
      work.remember();
      if (frame.pc == accept) return frame;
      final state = _states[frame.pc];
      final at = frame.position;
      final current = frame.captures;
      void next(int pc, int offset, List<int> values) {
        if (pending.length >= 16384) throw PatternWorkLimit();
        pending.add(_Frame(pc, offset, values));
      }

      switch (state.kind) {
        case 'epsilon':
          next(state.out, at, current);
        case 'split':
          next(state.other, at, current);
          next(state.out, at, current);
        case 'atom':
          final index = backwards ? at - 1 : at;
          if (index >= 0 &&
              index < input.length &&
              state.atom!.matchAsPrefix(input, index) != null) {
            next(state.out, backwards ? at - 1 : at + 1, current);
          }
        case 'assert':
          if (state.atom!.matchAsPrefix(input, at) != null)
            next(state.out, at, current);
        case 'captureStart' || 'captureEnd':
          final values = work.copy(current);
          final offset = (state.capture! - 1) * 2;
          if (state.kind == 'captureStart') {
            values[offset] = backwards ? -1 : at;
            values[offset + 1] = backwards ? at : -1;
          } else {
            values[offset + (backwards ? 0 : 1)] = at;
          }
          next(state.out, at, values);
        case 'reset':
          next(state.out, at, work.clear(current, state.resetIds));
        case 'progressStart':
          final values = work.copy(current);
          values[state.capture!] = at;
          next(state.out, at, values);
        case 'progressEnd':
          if (current[state.capture!] != at) {
            final values = work.copy(current);
            values[state.capture!] = -1;
            next(state.out, at, values);
          }
        case 'backref':
          final offset = (state.capture! - 1) * 2;
          final left = current[offset];
          final right = current[offset + 1];
          final length = left < 0 || right < 0 ? 0 : right - left;
          final from = backwards ? at - length : at;
          if (from >= 0 && from + length <= input.length) {
            work.spend(length);
            var matches = true;
            for (var index = 0; index < length; index++) {
              if (input.codeUnitAt(from + index) !=
                  input.codeUnitAt(left + index)) {
                matches = false;
                break;
              }
            }
            if (matches)
              next(state.out, backwards ? at - length : at + length, current);
          }
        case 'look':
          final initial = work.clear(current, state.resetIds);
          final matched = _runVm(
              input, state.innerStart!, state.innerAccept!, at, initial, work,
              backwards: state.behind);
          if ((matched != null) != state.negative) {
            next(state.out, at, state.negative ? initial : matched!.captures);
          }
        case 'accept':
          break;
        default:
          throw StateError('Unknown compiled pattern state');
      }
    }
    return null;
  }

  /// Evaluate assertions with the same budget as the enclosing search.
  bool _match(String input, int start, int accept, int from, _Work work,
      {required bool search, int? endAt}) {
    final end = endAt ?? input.length;
    var active = <int>{};
    var visited = <int>{};

    /// Expand epsilon edges once per state at the current position.
    void add(int start, int position, Set<int> output, Set<int> seen) {
      final pending = [start];
      while (pending.isNotEmpty) {
        work.step();
        final index = pending.removeLast();
        if (!seen.add(index)) continue;
        final state = _states[index];
        switch (state.kind) {
          case 'epsilon':
            pending.add(state.out);
          case 'split':
            pending.addAll([state.out, state.other]);
          case 'assert':
            if (state.atom!.matchAsPrefix(input, position) != null) {
              pending.add(state.out);
            }
          case 'look':
            var matched = false;
            if (!state.behind || position >= state.lookMinimum) {
              final first = state.behind
                  ? position - (state.lookMaximum ?? position)
                  : position;
              matched = _match(input, state.innerStart!, state.innerAccept!,
                  first < 0 ? 0 : first, work,
                  search: state.behind, endAt: state.behind ? position : null);
            }
            if (matched != state.negative) pending.add(state.out);
          default:
            output.add(index);
        }
      }
    }

    for (var position = from; position <= end; position++) {
      if (search || position == from) add(start, position, active, visited);
      if (active.contains(accept) && (endAt == null || position == endAt))
        return true;
      if (position == end || (!search && active.isEmpty)) break;
      final next = <int>{};
      final nextVisited = <int>{};
      for (final index in active) {
        work.step();
        final state = _states[index];
        if (state.kind == 'accept') continue;
        // Atoms are one character or a character class, never quantified.
        if (state.atom!.matchAsPrefix(input, position) != null) {
          add(state.out, position + 1, next, nextVisited);
        }
      }
      active = next;
      visited = nextVisited;
    }
    return false;
  }
}

class _Frame {
  final int pc;
  final int position;
  final List<int> captures;
  _Frame(this.pc, this.position, this.captures);
}

class _Work {
  int remaining = 2000000;
  int explored = 0;

  /// Count all nested work against the same allowance.
  void step() => spend(1);
  void spend(int amount) {
    remaining -= amount;
    if (remaining < 0) throw PatternWorkLimit();
  }

  /// Bound capture-context memoisation across nested assertions and start offsets.
  void remember() {
    if (++explored > 16384) throw PatternWorkLimit();
  }

  List<int> copy(List<int> captures) {
    spend(captures.length);
    return List<int>.of(captures);
  }

  List<int> clear(List<int> captures, List<int> ids) {
    if (ids.isEmpty) return captures;
    final result = copy(captures);
    spend(ids.length * 2);
    for (final id in ids) {
      result[(id - 1) * 2] = -1;
      result[(id - 1) * 2 + 1] = -1;
    }
    return result;
  }
}

final _cache = <String, BoundedPattern>{};

/// Compile the supported regular subset and retain at most sixteen patterns.
BoundedPattern compilePattern(String source) {
  final cached = _cache.remove(source);
  if (cached != null) {
    _cache[source] = cached;
    return cached;
  }
  if (source.length > 1024)
    throw FormatException('Pattern exceeds 1024 characters');
  final parser = _Parser(source, _countCaptures(source));
  final tree = parser.expression(0);
  if (parser.position != source.length)
    throw FormatException('Unmatched pattern group');
  if (parser.hasBackreferences && parser.captureCount > 64)
    throw FormatException('Backreference patterns exceed 64 captures');
  _validateReferences(tree, parser.captureCount, parser.names);
  final compiler = _Compiler(
      recording: parser.hasBackreferences,
      captureCount: parser.captureCount,
      names: parser.names);
  final fragment = compiler.compile(tree);
  final accept = compiler.add(_State('accept'));
  compiler.patch(fragment.$2, accept);
  // Validate contextual syntax only after the bounded compiler has succeeded.
  // Matching always uses the bounded states, never this whole-pattern RegExp.
  RegExp(source);
  final result = BoundedPattern._(compiler.states, fragment.$1, accept,
      parser.captureCount, parser.hasBackreferences, compiler.registerCount);
  if (_cache.length >= 16) _cache.remove(_cache.keys.first);
  _cache[source] = result;
  return result;
}

class _State {
  final String kind;
  final RegExp? atom;
  int out;
  int other;
  final int? innerStart;
  final int? innerAccept;
  final bool negative;
  final bool behind;
  final int lookMinimum;
  final int? lookMaximum;
  final int? capture;
  final List<int> resetIds;
  _State(this.kind,
      {this.atom,
      this.out = -1,
      this.other = -1,
      this.innerStart,
      this.innerAccept,
      this.negative = false,
      this.behind = false,
      this.lookMinimum = 0,
      this.lookMaximum,
      this.capture,
      this.resetIds = const []});
}

class _Node {
  final String kind;
  final String? atom;
  final List<_Node> children;
  final int minimum;
  final int? maximum;
  final int? capture;
  final bool lazy;
  _Node(this.kind,
      {this.atom,
      this.children = const [],
      this.minimum = 0,
      this.maximum,
      this.capture,
      this.lazy = false});
}

/// Bound consumed widths at the maximum supported input length plus one.
/// Assertions consume nothing even when their inner expressions inspect text.
(int, int?) _width(_Node node) {
  int bound(int value) => value > 100000 ? 100001 : value;
  if (node.kind == 'atom') return (1, 1);
  if (node.kind == 'backref') return (0, null);
  if (node.kind == 'assert' ||
      node.kind.endsWith('Lookahead') ||
      node.kind.endsWith('Lookbehind')) return (0, 0);
  if (node.kind == 'repeat') {
    final child = _width(node.children.single);
    final maximum = node.maximum == 0 || child.$2 == 0
        ? 0
        : node.maximum == null || child.$2 == null
            ? null
            : bound(node.maximum! * child.$2!);
    return (bound(node.minimum * child.$1), maximum);
  }
  final alternative = node.kind == 'alternatives';
  var minimum = alternative ? 100001 : 0;
  int? maximum = 0;
  for (final child in node.children) {
    final width = _width(child);
    minimum = alternative
        ? (minimum < width.$1 ? minimum : width.$1)
        : bound(minimum + width.$1);
    final childMax = width.$2;
    maximum = maximum == null || childMax == null
        ? null
        : alternative
            ? (maximum > childMax ? maximum : childMax)
            : bound(maximum + childMax);
  }
  return (minimum, maximum);
}

/// Count captures before interpreting legacy decimal escapes as references.
/// Escaped delimiters and character classes do not introduce captures.
(int, bool) _countCaptures(String source) {
  var captures = 0;
  var named = false;
  var inClass = false;
  for (var i = 0; i < source.length; i++) {
    final character = source[i];
    if (character == '\\') {
      i++;
    } else if (inClass) {
      if (character == ']') inClass = false;
    } else if (character == '[') {
      inClass = true;
    } else if (character == '(') {
      if (i + 1 >= source.length || source[i + 1] != '?') {
        captures++;
      } else if (source.startsWith('(?<', i) &&
          i + 3 < source.length &&
          source[i + 3] != '=' &&
          source[i + 3] != '!') {
        captures++;
        named = true;
      }
    }
  }
  return (captures, named);
}

class _Parser {
  final String source;
  final int totalCaptures;
  final bool hasNamedCaptures;
  int position = 0;
  int captureCount = 0;
  bool hasBackreferences = false;
  final names = <String, int>{};
  _Parser(this.source, (int, bool) captures)
      : totalCaptures = captures.$1,
        hasNamedCaptures = captures.$2;
  String get peek => position < source.length ? source[position] : '';

  /// Parse alternatives and concatenation while bounding nested groups.
  _Node expression(int depth) {
    if (depth > 64) throw FormatException('Pattern nesting exceeds 64 groups');
    final alternatives = <_Node>[];
    while (true) {
      final sequence = <_Node>[];
      while (peek.isNotEmpty && peek != ')' && peek != '|') {
        sequence.add(repeated(atom(depth)));
      }
      alternatives.add(_Node('sequence', children: sequence));
      if (peek != '|') break;
      position++;
    }
    return _Node('alternatives', children: alternatives);
  }

  /// Parse one literal, class, assertion, or ordinary/noncapturing group.
  _Node atom(int depth) {
    final start = position;
    final character = source[position++];
    if (character == '(') {
      String? assertion;
      int? capture;
      if (peek != '?') {
        capture = ++captureCount;
      } else if (source.startsWith('?:', position)) {
        position += 2;
      } else {
        for (final entry in const {
          '?=': 'positiveLookahead',
          '?!': 'negativeLookahead',
          '?<=': 'positiveLookbehind',
          '?<!': 'negativeLookbehind'
        }.entries) {
          if (source.startsWith(entry.key, position)) {
            assertion = entry.value;
            position += entry.key.length;
            break;
          }
        }
        if (assertion == null) {
          if (!source.startsWith('?<', position))
            throw FormatException('Invalid pattern group');
          position += 2;
          final name = readName();
          if (names.containsKey(name))
            throw FormatException('Duplicate capture name');
          capture = ++captureCount;
          names[name] = capture;
        }
      }
      final inner = expression(depth + 1);
      if (peek != ')') throw FormatException('Unclosed pattern group');
      position++;
      if (assertion != null) return _Node(assertion, children: [inner]);
      return capture == null
          ? inner
          : _Node('capture', capture: capture, children: [inner]);
    }
    if (character == '[') {
      while (peek.isNotEmpty && peek != ']') {
        if (peek == '\\') {
          position++;
          if (peek.isEmpty) throw FormatException('Unclosed character class');
          if (peek == 'k' && hasNamedCaptures) {
            throw FormatException(
                'Named references are not allowed in classes');
          }
        }
        position++;
      }
      if (peek != ']') throw FormatException('Unclosed character class');
      position++;
    } else if (character == '\\') {
      if (peek.isEmpty) throw FormatException('Trailing pattern escape');
      final escaped = source[position++];
      if ('0123456789'.contains(escaped)) {
        var end = position;
        while (end < source.length && '0123456789'.contains(source[end])) end++;
        final id = int.tryParse(source.substring(start + 1, end));
        if (escaped != '0' && id != null && id <= totalCaptures) {
          position = end;
          hasBackreferences = true;
          return _Node('backref', capture: id);
        }
        // Annex B numeric escapes consume at most three octal digits (0..3)
        // or two (4..7); remaining digits are separate, quantifiable atoms.
        if ('01234567'.contains(escaped)) {
          final maximum = '0123'.contains(escaped) ? 3 : 2;
          var digits = escaped;
          while (digits.length < maximum &&
              peek.isNotEmpty &&
              '01234567'.contains(peek)) {
            digits += source[position++];
          }
          return _Node('atom',
              atom: RegExp.escape(
                  String.fromCharCode(int.parse(digits, radix: 8))));
        }
        return _Node('atom', atom: escaped);
      }
      if (escaped == 'k' && hasNamedCaptures) {
        if (peek != '<') throw FormatException('Invalid named backreference');
        position++;
        final name = readName();
        hasBackreferences = true;
        return _Node('backref', atom: name);
      }
      if (escaped == 'x' || escaped == 'u' || escaped == 'c') {
        final count = escaped == 'x'
            ? 2
            : escaped == 'u'
                ? 4
                : 1;
        final digits = source.substring(
            position,
            position + count <= source.length
                ? position + count
                : source.length);
        final valid = escaped == 'c'
            ? RegExp(r'^[a-zA-Z]$').hasMatch(digits)
            : RegExp(r'^[0-9a-fA-F]+$').hasMatch(digits);
        if (!valid || digits.length != count) {
          if (escaped == 'c') {
            position--;
            return _Node('atom', atom: r'\\');
          }
          return _Node('atom', atom: escaped);
        }
        position += count;
      }
      if (escaped == 'b' || escaped == 'B') {
        return _Node('assert', atom: source.substring(start, position));
      }
    } else if ('*+?'.contains(character) ||
        (character == '{' && _quantifierAt(start))) {
      throw FormatException('Unexpected pattern quantifier');
    }
    final text = character == '{' || character == '}'
        ? '\\$character'
        : source.substring(start, position);
    // Compiling a single atom also validates class ranges and escape syntax.
    RegExp(text);
    return _Node(character == '^' || character == r'$' ? 'assert' : 'atom',
        atom: text);
  }

  /// Validate identifier syntax using an empty group, never an input match.
  String readName() {
    final start = position;
    while (peek.isNotEmpty && peek != '>') position++;
    if (peek != '>' || start == position)
      throw FormatException('Invalid capture name');
    final raw = source.substring(start, position++);
    RegExp('(?<$raw>)');
    return raw.replaceAllMapped(
        RegExp(r'\\u(?:\{([0-9a-fA-F]+)\}|([0-9a-fA-F]{4}))'),
        (match) =>
            String.fromCharCode(int.parse(match[1] ?? match[2]!, radix: 16)));
  }

  /// Non-Unicode ECMA-262 treats braces outside valid quantifiers literally.
  bool _quantifierAt(int offset) =>
      RegExp(r'\{[0-9]+(?:,[0-9]*)?\}').matchAsPrefix(source, offset) != null;

  /// Parse bounded repetitions, treating lazy order identically for existence.
  _Node repeated(_Node inner) {
    if (peek.isEmpty ||
        !'*+?{'.contains(peek) ||
        (peek == '{' && !_quantifierAt(position))) return inner;
    if (inner.kind == 'assert' || inner.kind.endsWith('Lookbehind'))
      throw FormatException('Assertions cannot be repeated');
    final quantifier = source[position++];
    int minimum;
    int? maximum;
    switch (quantifier) {
      case '*':
        minimum = 0;
      case '+':
        minimum = 1;
      case '?':
        minimum = 0;
        maximum = 1;
      default:
        minimum = count();
        maximum = minimum;
        if (peek == ',') {
          position++;
          maximum = peek == '}' ? null : count();
        }
        if (peek != '}' || (maximum != null && maximum < minimum)) {
          throw FormatException('Invalid pattern repetition');
        }
        position++;
    }
    final lazy = peek == '?';
    if (lazy) position++;
    return _Node('repeat',
        children: [inner], minimum: minimum, maximum: maximum, lazy: lazy);
  }

  /// Bound expansion before the compiler allocates repetition states.
  int count() {
    final start = position;
    while (peek.isNotEmpty && '0123456789'.contains(peek)) position++;
    if (start == position || position - start > 4)
      throw FormatException('Invalid repetition count');
    final value = int.parse(source.substring(start, position));
    if (value > 1000) throw FormatException('Pattern repetition exceeds 1000');
    return value;
  }
}

/// Resolve numeric and named references after all groups have been parsed.
int _reference(_Node node, int count, Map<String, int> names) {
  final id = node.capture ?? names[node.atom];
  if (id == null || id < 1 || id > count)
    throw FormatException('Unknown backreference');
  return id;
}

void _validateReferences(_Node node, int count, Map<String, int> names) {
  if (node.kind == 'backref') _reference(node, count, names);
  for (final child in node.children) {
    _validateReferences(child, count, names);
  }
}

/// Capture groups nested inside a repeated expression must be reset each time.
List<int> _captureIds(_Node node) {
  final result = <int>[];
  void visit(_Node current) {
    if (current.kind == 'capture') result.add(current.capture!);
    for (final child in current.children) {
      visit(child);
    }
  }

  visit(node);
  return result;
}

typedef _Fragment = (int, List<(int, bool)>);

class _Compiler {
  final states = <_State>[];
  final bool recording;
  final int captureCount;
  final Map<String, int> names;
  int registerCount = 0;
  _Compiler(
      {this.recording = false, this.captureCount = 0, this.names = const {}});

  /// Cap compilation memory and per-position matching work.
  int add(_State state) {
    if (states.length >= 2048)
      throw FormatException('Pattern exceeds 2048 states');
    states.add(state);
    return states.length - 1;
  }

  /// Connect fragment exits to their continuation.
  void patch(List<(int, bool)> exits, int target) {
    for (final exit in exits) {
      if (exit.$2) {
        states[exit.$1].other = target;
      } else {
        states[exit.$1].out = target;
      }
    }
  }

  /// Concatenate two fragments in execution order.
  _Fragment join(_Fragment left, _Fragment right) {
    patch(left.$2, right.$1);
    return (left.$1, right.$2);
  }

  /// Reset captures inside each repetition before a new iteration participates.
  _Fragment iteration(_Node node, bool backwards, {bool progress = false}) {
    var body = compile(node, backwards);
    if (progress && recording && _width(node).$1 == 0) {
      final register = captureCount * 2 + registerCount++;
      final start =
          add(_State('progressStart', capture: register, out: body.$1));
      final end = add(_State('progressEnd', capture: register));
      patch(body.$2, end);
      body = (start, [(end, false)]);
    }
    final ids = recording ? _captureIds(node) : <int>[];
    if (ids.isEmpty) return body;
    final reset = add(_State('reset', resetIds: ids, out: body.$1));
    return (reset, body.$2);
  }

  /// Compile regular or capture-aware states; lookbehind consumes in reverse.
  _Fragment compile(_Node node, [bool backwards = false]) {
    if (node.kind == 'capture') {
      final inner = compile(node.children.single, backwards);
      if (!recording) return inner;
      final start =
          add(_State('captureStart', capture: node.capture, out: inner.$1));
      final end = add(_State('captureEnd', capture: node.capture));
      patch(inner.$2, end);
      return (start, [(end, false)]);
    }
    if (node.kind == 'backref') {
      final id = _reference(node, captureCount, names);
      final index = add(_State('backref', capture: id));
      return (index, [(index, false)]);
    }
    if (node.kind.endsWith('Lookahead') || node.kind.endsWith('Lookbehind')) {
      final behind = node.kind.endsWith('Lookbehind');
      final inner = compile(node.children.single, recording && behind);
      final accept = add(_State('accept'));
      patch(inner.$2, accept);
      final width = _width(node.children.single);
      final index = add(_State('look',
          innerStart: inner.$1,
          innerAccept: accept,
          negative: node.kind.startsWith('negative'),
          behind: behind,
          lookMinimum: width.$1,
          lookMaximum: width.$2,
          resetIds: recording ? _captureIds(node.children.single) : const []));
      return (index, [(index, false)]);
    }
    if (node.kind == 'atom' || node.kind == 'assert') {
      final index = add(_State(node.kind, atom: RegExp(node.atom!)));
      return (index, [(index, false)]);
    }
    if (node.kind == 'alternatives') {
      var result = compile(node.children.first, backwards);
      for (final child in node.children.skip(1)) {
        final right = compile(child, backwards);
        final split = add(_State('split', out: result.$1, other: right.$1));
        result = (split, [...result.$2, ...right.$2]);
      }
      return result;
    }
    final empty = add(_State('epsilon'));
    _Fragment result = (empty, [(empty, false)]);
    if (node.kind == 'sequence') {
      for (final child in backwards ? node.children.reversed : node.children) {
        result = join(result, compile(child, backwards));
      }
      return result;
    }
    for (var index = 0; index < node.minimum; index++) {
      result = join(result, iteration(node.children.single, backwards));
    }
    if (node.maximum == null) {
      final body = iteration(node.children.single, backwards, progress: true);
      final split = add(_State('split',
          out: node.lazy ? -1 : body.$1, other: node.lazy ? body.$1 : -1));
      patch(body.$2, split);
      return join(result, (split, [(split, !node.lazy)]));
    }
    for (var index = node.minimum; index < node.maximum!; index++) {
      final body = iteration(node.children.single, backwards, progress: true);
      final split = add(_State('split',
          out: node.lazy ? -1 : body.$1, other: node.lazy ? body.$1 : -1));
      result = join(result, (split, [...body.$2, (split, !node.lazy)]));
    }
    return result;
  }
}
