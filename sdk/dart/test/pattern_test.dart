import 'dart:math';
import 'package:anyvali/anyvali.dart' as av;
import 'package:test/test.dart';

/// Verify pattern semantics and compilation/matching resource boundaries.
void main() {
  test('bounded patterns agree with Dart for the supported regular subset', () {
    final patterns = <String>[
      '',
      'a',
      '^a',
      r'a$',
      r'^$',
      r'^[a-z]+$',
      r'[^a-z]',
      r'[^]',
      r'[]',
      r'^.$',
      r'.*',
      r'a.b',
      r'a|b',
      r'|a',
      r'a|',
      r'(a|)b',
      r'(ab|a)+',
      r'(?:ab)?c',
      r'a*?b',
      r'a+?b',
      r'a??b',
      r'a{0}',
      r'a{2}',
      r'a{1,3}',
      r'a{2,}',
      r'(a?)*b',
      r'\d+',
      r'\D+',
      r'\w+',
      r'\W',
      r'\s+',
      r'\S+',
      r'\bword\b',
      r'\Boo\B',
      r'\x61\u0062',
      r'\cA',
      r'\[a\]',
      r'\^\$',
      r'[\]a]+',
      r'[a-zA-Z0-9_-]{1,3}',
      r'^(a+)+$',
      r'^(a|a)+$',
      r'a*a*a*b',
    ];
    final inputs = <String>[
      '',
      'a',
      'b',
      'c',
      'ab',
      'abc',
      'aaab',
      'aaaa!',
      'bbbc',
      'a\nb',
      '\n',
      '\r\n',
      '1',
      '123',
      'A9_',
      ' word ',
      'wood',
      'word!',
      '[a]',
      '^\$',
      '\x01',
      '😀',
      'hello',
      'abc\n',
      'aa a',
    ];
    for (final pattern in patterns) {
      final schema = av.string().pattern(pattern);
      final expected = RegExp(pattern);
      for (final input in inputs) {
        expect(schema.safeParse(input).success, expected.hasMatch(input),
            reason: 'pattern=$pattern input=${input.codeUnits}');
      }
    }
  });

  test('nested alternatives and repetitions agree with Dart', () {
    final random = Random(1827);

    /// Generate bounded regular syntax independently of the matcher parser.
    String expression(int depth) {
      const atoms = ['a', 'b', '.', '[ab]', '[^a]', r'\d'];
      if (depth == 0) return atoms[random.nextInt(atoms.length)];
      switch (random.nextInt(4)) {
        case 0:
          return expression(0);
        case 1:
          return '(${expression(depth - 1)}|${expression(depth - 1)})';
        case 2:
          return '${expression(depth - 1)}${expression(depth - 1)}';
        default:
          const quantifiers = [
            '*',
            '+',
            '?',
            '{0,2}',
            '{1,3}',
            '*?',
            '+?',
            '??'
          ];
          return '(?:${expression(depth - 1)})${quantifiers[random.nextInt(quantifiers.length)]}';
      }
    }

    var inputs = [''];
    var layer = [''];
    for (var length = 1; length <= 3; length++) {
      layer = [
        for (final prefix in layer)
          for (final letter in ['a', 'b', '1']) '$prefix$letter'
      ];
      inputs.addAll(layer);
    }
    for (var index = 0; index < 200; index++) {
      final pattern =
          '${random.nextBool() ? '^' : ''}${expression(3)}${random.nextBool() ? r'$' : ''}';
      final schema = av.string().pattern(pattern);
      final reference = RegExp(pattern);
      for (final input in inputs) {
        expect(schema.safeParse(input).success, reference.hasMatch(input),
            reason: '$pattern on $input');
      }
    }
  });

  test('nested and overlapping repetitions do not backtrack', () {
    final input = '${'a' * 10000}!';
    for (final pattern in [
      r'^(a+)+$',
      r'^(a|a)+$',
      r'^([a-zA-Z0-9]+)*$',
      r'^a*a*a*b$'
    ]) {
      final schema = av.string().pattern(pattern);
      final result = schema.safeParse(input);
      expect(result.success, isFalse);
      expect(result.issues.single.code, 'invalid_string');
    }
    expect(av.string().pattern(r'^(a+)+$').parse('a' * 10000), 'a' * 10000);
  });

  test(
      'unsupported or excessively large patterns fail during construction and import',
      () {
    final base = av.any().exportSchema();
    for (final pattern in [
      r'^(a)\1$',
      r'(?=a)',
      r'(?<=a)b',
      r'(?<name>a)',
      r'\p{L}',
      r'a{1001}',
      'a' * 1025,
      '${'(' * 65}a${')' * 65}',
      r'a{1000}a{1000}a{1000}',
      r'[z-a]',
      r'[',
      r'a{2,1}',
      r'a**',
      '(',
      ')',
      '\\',
    ]) {
      expect(() => av.string().pattern(pattern), throwsFormatException);
      expect(
          () => av.importSchema({
                ...base,
                'root': {'kind': 'string', 'pattern': pattern}
              }),
          throwsFormatException);
    }
  });

  test('input and work limits fail validation with an explicit diagnostic', () {
    final tooLong = av.string().pattern('a').safeParse('a' * 100001);
    expect(tooLong.success, isFalse);
    expect(tooLong.issues.single.code, 'invalid_string');
    expect(tooLong.issues.single.message, contains('limit'));
    final expensive = av.string().pattern(r'a{1000}b').safeParse('a' * 3000);
    expect(expensive.success, isFalse);
    expect(expensive.issues.single.message, contains('work limit'));
  });

  test('patterns survive interchange and bounded cache eviction', () {
    final schema = av.string().pattern(r'^(ab|c){1,3}$');
    final imported = av.importSchema(schema.exportSchema());
    for (var i = 0; i < 32; i++) {
      expect(av.string().pattern('^value$i\$').parse('value$i'), 'value$i');
    }
    expect(imported.parse('abc'), 'abc');
    expect(schema.parse('ababab'), 'ababab');
    expect(imported.safeParse('abcd').success, isFalse);
  });
}
