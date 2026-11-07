import 'dart:math';
import 'package:anyvali/anyvali.dart' as av;
import 'package:test/test.dart';

/// Verify pattern semantics and compilation/matching resource boundaries.
void main() {
  test('property-looking escapes use flags-free JavaScript semantics', () {
    for (final pattern in [
      r'^\p{L}$',
      r'^[\p{L}]$',
      r'^\P{Nd}$',
      r'^[\P{Nd}]$',
      r'^\p{2}$',
      r'^\p{L}+$',
      r'^a{x}$',
      r'^a{2x}$',
      r'^a{$',
      r'^{L}$',
    ]) {
      final schema = av.string().pattern(pattern);
      final imported = av.importSchema(schema.exportSchema());
      final reference = RegExp(pattern);
      for (final input in [
        'p{L}',
        'P{Nd}',
        'p{L}}',
        'pp',
        'p',
        'L',
        '{',
        '}',
        'P',
        'N',
        'd',
        'é',
        'a{x}',
        'a{2x}',
        'a{',
        '{L}'
      ]) {
        expect(schema.safeParse(input).success, reference.hasMatch(input),
            reason: '$pattern on $input');
        expect(imported.safeParse(input).success, reference.hasMatch(input));
      }
    }
  });
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
      r'^[\\p]+$',
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
      r'^(a)\1$',
      r'^(a*)*\1$',
      r'^(a?)+\1$',
      r'^(a?)*\1$',
      r'^(a*)+\1$',
      r'^(a?){2,4}\1$',
      r'^(a)?b\1$',
      r'^\1(a)$',
      r'^(a|(b))*\2$',
      r'^(?:(a)|b)+\1$',
      r'^(?=(a+))a*b\1$',
      r'^(?=(a+?))a*b\1$',
      r'(?<=([ab]+)([bc]+))\1$',
      r'^(?<letter>[ab])\k<letter>$',
      r'^(?<\u0061>a)\k<a>$',
      r'(?<=(?<x>a))\k<x>$',
      r'^(😀)\1$',
      r'^(?=.*[A-Z]).+$',
      r'a(?=b)',
      r'a(?!b)',
      r'(?<=a)b',
      r'(?<!a)b',
      r'(?<=a+)b',
      r'(?<=(?:ab|a))b',
      r'(?=(a+)+$)a+',
      r'(?=(?=a)a)a',
      r'(?!a*)b',
      r'(?=)',
      r'(?!)',
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
      'aa',
      'bb',
      'aba',
      'abb',
      'abba',
      'aaba',
      'abca',
      'abcab',
      '😀😀',
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
      'p',
      r'\',
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
      switch (random.nextInt(7)) {
        case 0:
          return expression(0);
        case 1:
          return '(${expression(depth - 1)}|${expression(depth - 1)})';
        case 2:
          return '${expression(depth - 1)}${expression(depth - 1)}';
        case 3:
          return '(?=${expression(depth - 1)})${expression(depth - 1)}';
        case 4:
          return '(?!${expression(depth - 1)})${expression(depth - 1)}';
        case 5:
          return '(?<=${expression(depth - 1)})${expression(depth - 1)}';
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

  test('portable lookarounds share compilation and execution limits', () {
    final document = av.string().exportSchema();
    final password = av.importSchema({
      ...document,
      'root': {'kind': 'string', 'pattern': r'^(?=.*[A-Z])(?=.*[0-9]).{8,}$'}
    });
    expect(password.parse('Password7'), 'Password7');
    expect(av.string().pattern(r'(?<!a)b').parse('${'c' * 10000}b'),
        '${'c' * 10000}b');
    expect(av.string().pattern('${'(?=' * 64}a${')' * 64}').parse('a'), 'a');
    expect(password.safeParse('password7').success, isFalse);
    expect(password.safeParse('Password').success, isFalse);
    final behind =
        av.string().pattern(r'(?<=a+)b').safeParse('${'a' * 10000}!');
    expect(behind.success, isFalse);
    expect(behind.issues.single.message, contains('work limit'));
    expect(
        av
            .string()
            .pattern(r'(?=(a+)+$)a+')
            .safeParse('${'a' * 10000}!')
            .success,
        isFalse);
    expect(() => av.string().pattern(r'(?=a{1000})(?=b{1000})(?=c{1000})'),
        throwsFormatException);
  });

  test('capture-aware backreferences are portable and remain bounded', () {
    var inputs = [''];
    var layer = [''];
    for (var length = 1; length <= 4; length++) {
      layer = [
        for (final prefix in layer)
          for (final letter in ['a', 'b']) '$prefix$letter'
      ];
      inputs.addAll(layer);
    }
    for (final body in [
      'a',
      'b',
      '[ab]',
      'a?',
      'a+',
      'a*',
      'a|b',
      '(?:ab|a)'
    ]) {
      for (final tail in [
        r'\1',
        r'(?:\1)?',
        r'(?=\1)\1',
        r'(?!\1)b',
        r'(?<=\1)'
      ]) {
        final pattern = '^($body)$tail\$';
        final schema = av.string().pattern(pattern);
        final reference = RegExp(pattern);
        for (final input in inputs) {
          expect(schema.safeParse(input).success, reference.hasMatch(input),
              reason: '$pattern on $input');
        }
      }
    }
    final document = av.string().exportSchema();
    final repeated = av.importSchema({
      ...document,
      'root': {'kind': 'string', 'pattern': r'^(a)\1$'}
    });
    expect(repeated.parse('aa'), 'aa');
    expect(repeated.safeParse('ab').success, isFalse);
    expect(
        av.string().pattern(r'(?<=([ab]+)([bc]+))\1$').parse('abca'), 'abca');
    expect(av.string().pattern(r'^(?=(a+))a*b\1$').safeParse('aaba').success,
        isFalse);
    expect(av.string().pattern(r'^(?=(a+?))a*b\1$').parse('aaba'), 'aaba');
    final attack =
        av.string().pattern(r'^(a+)+\1$').safeParse('${'a' * 10000}!');
    expect(attack.success, isFalse);
    expect(attack.issues.single.message, contains('limit'));
    expect(
        () => av.string().pattern('${'(a)' * 65}\\1'), throwsFormatException);
    expect(() => av.string().pattern(r'(a)\2'), throwsFormatException);
    expect(() => av.string().pattern(r'(?<x>a)(?<x>b)'), throwsFormatException);
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
