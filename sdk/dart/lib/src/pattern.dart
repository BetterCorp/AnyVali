/// Bounded regular-expression matching for portable schema patterns.
/// Thompson states avoid backtracking; captures and match ordering are irrelevant
/// because schema validation asks only whether a match exists.
library;

/// Raised when a valid pattern exceeds the per-match work or input limit.
class PatternWorkLimit implements Exception {}

/// A compiled pattern with a bounded number of Thompson states.
class BoundedPattern {
  final List<_State> _states;
  final int _start;
  final int _accept;
  BoundedPattern._(this._states, this._start, this._accept);

  /// Search using deduplicated states at each UTF-16 input position.
  bool hasMatch(String input) {
    if (input.length > 100000) throw PatternWorkLimit();
    return _match(input, _start, _accept, 0, _Work(), search: true);
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

class _Work {
  int remaining = 2000000;

  /// Count nested assertion work against a single per-match allowance.
  void step() {
    if (--remaining < 0) throw PatternWorkLimit();
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
  final parser = _Parser(source);
  final tree = parser.expression(0);
  if (parser.position != source.length)
    throw FormatException('Unmatched pattern group');
  final compiler = _Compiler();
  final fragment = compiler.compile(tree);
  final accept = compiler.add(_State('accept'));
  compiler.patch(fragment.$2, accept);
  final result = BoundedPattern._(compiler.states, fragment.$1, accept);
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
  _State(this.kind,
      {this.atom,
      this.out = -1,
      this.other = -1,
      this.innerStart,
      this.innerAccept,
      this.negative = false,
      this.behind = false,
      this.lookMinimum = 0,
      this.lookMaximum});
}

class _Node {
  final String kind;
  final String? atom;
  final List<_Node> children;
  final int minimum;
  final int? maximum;
  _Node(this.kind,
      {this.atom, this.children = const [], this.minimum = 0, this.maximum});
}

/// Bound consumed widths at the maximum supported input length plus one.
/// Assertions consume nothing even when their inner expressions inspect text.
(int, int?) _width(_Node node) {
  int bound(int value) => value > 100000 ? 100001 : value;
  if (node.kind == 'atom') return (1, 1);
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

class _Parser {
  final String source;
  int position = 0;
  _Parser(this.source);
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
      if (peek == '?') {
        if (source.startsWith('?:', position)) {
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
          if (assertion == null)
            throw FormatException('Named groups are not supported');
        }
      }
      final inner = expression(depth + 1);
      if (peek != ')') throw FormatException('Unclosed pattern group');
      position++;
      return assertion == null ? inner : _Node(assertion, children: [inner]);
    }
    if (character == '[') {
      while (peek.isNotEmpty && peek != ']') {
        if (peek == '\\') {
          position++;
          if (peek.isEmpty) throw FormatException('Unclosed character class');
        }
        position++;
      }
      if (peek != ']') throw FormatException('Unclosed character class');
      position++;
    } else if (character == '\\') {
      if (peek.isEmpty) throw FormatException('Trailing pattern escape');
      final escaped = source[position++];
      if ('0123456789kpP'.contains(escaped)) {
        throw FormatException(
            'Backreferences, octal and property escapes are not supported');
      }
      if (escaped == 'x' || escaped == 'u' || escaped == 'c') {
        final count = escaped == 'x'
            ? 2
            : escaped == 'u'
                ? 4
                : 1;
        if (position + count > source.length)
          throw FormatException('Incomplete pattern escape');
        final digits = source.substring(position, position + count);
        final valid = escaped == 'c'
            ? RegExp(r'^[a-zA-Z]$').hasMatch(digits)
            : RegExp(r'^[0-9a-fA-F]+$').hasMatch(digits);
        if (!valid) throw FormatException('Invalid pattern escape');
        position += count;
      }
      if (escaped == 'b' || escaped == 'B') {
        return _Node('assert', atom: source.substring(start, position));
      }
    } else if ('*+?{}'.contains(character)) {
      throw FormatException('Unexpected pattern quantifier');
    }
    final text = source.substring(start, position);
    // Compiling a single atom also validates class ranges and escape syntax.
    RegExp(text);
    return _Node(character == '^' || character == r'$' ? 'assert' : 'atom',
        atom: text);
  }

  /// Parse bounded repetitions, treating lazy order identically for existence.
  _Node repeated(_Node inner) {
    if (peek.isEmpty || !'*+?{'.contains(peek)) return inner;
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
    if (peek == '?') position++;
    return _Node('repeat',
        children: [inner], minimum: minimum, maximum: maximum);
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

typedef _Fragment = (int, List<(int, bool)>);

class _Compiler {
  final states = <_State>[];

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

  /// Concatenate two fragments without introducing backtracking choices.
  _Fragment join(_Fragment left, _Fragment right) {
    patch(left.$2, right.$1);
    return (left.$1, right.$2);
  }

  /// Compile syntax into Thompson states with bounded repetition expansion.
  _Fragment compile(_Node node) {
    if (node.kind.endsWith('Lookahead') || node.kind.endsWith('Lookbehind')) {
      final inner = compile(node.children.single);
      final accept = add(_State('accept'));
      patch(inner.$2, accept);
      final width = _width(node.children.single);
      final index = add(_State('look',
          innerStart: inner.$1,
          innerAccept: accept,
          negative: node.kind.startsWith('negative'),
          behind: node.kind.endsWith('Lookbehind'),
          lookMinimum: width.$1,
          lookMaximum: width.$2));
      return (index, [(index, false)]);
    }
    if (node.kind == 'atom' || node.kind == 'assert') {
      final index = add(_State(node.kind, atom: RegExp(node.atom!)));
      return (index, [(index, false)]);
    }
    if (node.kind == 'alternatives') {
      var result = compile(node.children.first);
      for (final child in node.children.skip(1)) {
        final right = compile(child);
        final split = add(_State('split', out: result.$1, other: right.$1));
        result = (split, [...result.$2, ...right.$2]);
      }
      return result;
    }
    final empty = add(_State('epsilon'));
    _Fragment result = (empty, [(empty, false)]);
    if (node.kind == 'sequence') {
      for (final child in node.children) {
        result = join(result, compile(child));
      }
      return result;
    }
    for (var index = 0; index < node.minimum; index++) {
      result = join(result, compile(node.children.single));
    }
    if (node.maximum == null) {
      final body = compile(node.children.single);
      final split = add(_State('split', out: body.$1));
      patch(body.$2, split);
      return join(result, (split, [(split, true)]));
    }
    for (var index = node.minimum; index < node.maximum!; index++) {
      final body = compile(node.children.single);
      final split = add(_State('split', out: body.$1));
      result = join(result, (split, [...body.$2, (split, true)]));
    }
    return result;
  }
}
