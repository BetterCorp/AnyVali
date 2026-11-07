import 'dart:convert';
import 'package:anyvali/anyvali.dart' as av;

/// Demonstrate native schema parsing and a portable document round trip.
void main() {
  final user = av.object({
    'name': av.string().coerce(trim: true).minLength(1),
    'email': av.string().format('email'),
    'age': av.int_().coerce().min(0).optional(),
  });
  final result = user.safeParse({
    'name': ' Alice ',
    'email': 'alice@example.com',
    'age': '30',
  });
  print(result.data);
  final imported = av.importSchema(jsonDecode(jsonEncode(user.exportSchema())));
  print(imported.parse({'name': 'Alice', 'email': 'alice@example.com'}));
}
