using AnyVali;
using AnyVali.Schemas;
using Xunit;

namespace AnyVali.Tests;

public class StringSchemaTests
{
    [Fact]
    public void AcceptsValidString()
    {
        var s = V.String();
        Assert.Equal("hello", s.Parse("hello"));
    }

    [Fact]
    public void AcceptsEmptyString()
    {
        var s = V.String();
        Assert.Equal("", s.Parse(""));
    }

    [Fact]
    public void RejectsNonString()
    {
        var s = V.String();
        var result = s.SafeParse(42L);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void RejectsNull()
    {
        var s = V.String();
        var result = s.SafeParse(null);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void MinLengthValidation()
    {
        var s = V.String().MinLength(3);
        Assert.Equal("abc", s.Parse("abc"));

        var result = s.SafeParse("ab");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooSmall, result.Issues[0].Code);
    }

    [Fact]
    public void MaxLengthValidation()
    {
        var s = V.String().MaxLength(3);
        Assert.Equal("abc", s.Parse("abc"));

        var result = s.SafeParse("abcd");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooLarge, result.Issues[0].Code);
    }

    [Fact]
    public void PatternValidation()
    {
        var s = V.String().Pattern(@"^\d+$");
        Assert.Equal("123", s.Parse("123"));

        var result = s.SafeParse("abc");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidString, result.Issues[0].Code);
    }

    [Fact]
    public void StartsWithValidation()
    {
        var s = V.String().StartsWith("hello");
        Assert.Equal("hello world", s.Parse("hello world"));

        var result = s.SafeParse("world hello");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidString, result.Issues[0].Code);
    }

    [Fact]
    public void EndsWithValidation()
    {
        var s = V.String().EndsWith("world");
        Assert.Equal("hello world", s.Parse("hello world"));

        var result = s.SafeParse("world hello");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidString, result.Issues[0].Code);
    }

    [Fact]
    public void IncludesValidation()
    {
        var s = V.String().Includes("lo wo");
        Assert.Equal("hello world", s.Parse("hello world"));

        var result = s.SafeParse("goodbye");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidString, result.Issues[0].Code);
    }

    [Fact]
    public void MethodChainingCreatesNewSchema()
    {
        var s1 = V.String();
        var s2 = s1.MinLength(3);
        // s1 should still accept short strings
        Assert.Equal("ab", s1.Parse("ab"));
        var result = s2.SafeParse("ab");
        Assert.False(result.Success);
    }

    [Fact]
    public void ParseThrowsValidationError()
    {
        var s = V.String();
        var ex = Assert.Throws<ValidationError>(() => s.Parse(42L));
        Assert.Single(ex.Issues);
        Assert.Equal(IssueCodes.InvalidType, ex.Issues[0].Code);
    }
}

public class NumberSchemaTests
{
    [Fact]
    public void AcceptsDouble()
    {
        var s = V.Number();
        Assert.Equal(3.14, s.Parse(3.14));
    }

    [Fact]
    public void AcceptsLong()
    {
        var s = V.Number();
        Assert.Equal(42.0, s.Parse(42L));
    }

    [Fact]
    public void RejectsString()
    {
        var s = V.Number();
        var result = s.SafeParse("hello");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void MinConstraint()
    {
        var s = V.Number().Min(10);
        Assert.Equal(10.0, s.Parse(10L));
        var result = s.SafeParse(9L);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooSmall, result.Issues[0].Code);
    }

    [Fact]
    public void MaxConstraint()
    {
        var s = V.Number().Max(10);
        Assert.Equal(10.0, s.Parse(10L));
        var result = s.SafeParse(11L);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooLarge, result.Issues[0].Code);
    }

    [Fact]
    public void ExclusiveMinConstraint()
    {
        var s = V.Number().ExclusiveMin(10);
        Assert.Equal(11.0, s.Parse(11L));
        var result = s.SafeParse(10L);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooSmall, result.Issues[0].Code);
    }

    [Fact]
    public void ExclusiveMaxConstraint()
    {
        var s = V.Number().ExclusiveMax(10);
        Assert.Equal(9.0, s.Parse(9L));
        var result = s.SafeParse(10L);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooLarge, result.Issues[0].Code);
    }

    [Fact]
    public void MultipleOfConstraint()
    {
        var s = V.Number().MultipleOf(3);
        Assert.Equal(9.0, s.Parse(9L));
        var result = s.SafeParse(10L);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidNumber, result.Issues[0].Code);
    }

    [Fact]
    public void Float32Schema()
    {
        var s = V.Float32();
        Assert.Equal(3.14, s.Parse(3.14));
    }

    [Fact]
    public void Float64Schema()
    {
        var s = V.Float64();
        Assert.Equal(3.14, s.Parse(3.14));
    }
}

public class IntSchemaTests
{
    [Fact]
    public void AcceptsInteger()
    {
        var s = V.Int();
        Assert.Equal(42L, s.Parse(42L));
    }

    [Fact]
    public void RejectsFloat()
    {
        var s = V.Int();
        var result = s.SafeParse(3.14);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void RejectsString()
    {
        var s = V.Int();
        var result = s.SafeParse("hello");
        Assert.False(result.Success);
    }

    [Fact]
    public void Int8Range()
    {
        var s = V.Int8();
        Assert.Equal(127L, s.Parse(127L));
        Assert.Equal(-128L, s.Parse(-128L));

        var result = s.SafeParse(128L);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooLarge, result.Issues[0].Code);

        result = s.SafeParse(-129L);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooSmall, result.Issues[0].Code);
    }

    [Fact]
    public void Int16Range()
    {
        var s = V.Int16();
        Assert.Equal(32767L, s.Parse(32767L));
        var result = s.SafeParse(32768L);
        Assert.False(result.Success);
    }

    [Fact]
    public void Int32Range()
    {
        var s = V.Int32();
        Assert.Equal(2147483647L, s.Parse(2147483647L));
        var result = s.SafeParse(2147483648L);
        Assert.False(result.Success);
    }

    [Fact]
    public void Uint8Range()
    {
        var s = V.Uint8();
        Assert.Equal(255L, s.Parse(255L));
        Assert.Equal(0L, s.Parse(0L));

        var result = s.SafeParse(256L);
        Assert.False(result.Success);

        result = s.SafeParse(-1L);
        Assert.False(result.Success);
    }

    [Fact]
    public void Uint16Range()
    {
        var s = V.Uint16();
        Assert.Equal(65535L, s.Parse(65535L));
        var result = s.SafeParse(65536L);
        Assert.False(result.Success);
    }

    [Fact]
    public void Uint32Range()
    {
        var s = V.Uint32();
        Assert.Equal(4294967295L, s.Parse(4294967295L));
        var result = s.SafeParse(4294967296L);
        Assert.False(result.Success);
    }

    [Fact]
    public void IntWithMinConstraint()
    {
        var s = V.Int().Min(10);
        Assert.Equal(10L, s.Parse(10L));
        var result = s.SafeParse(9L);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooSmall, result.Issues[0].Code);
    }
}

public class BoolSchemaTests
{
    [Fact]
    public void AcceptsTrue() => Assert.Equal(true, V.Bool().Parse(true));

    [Fact]
    public void AcceptsFalse() => Assert.Equal(false, V.Bool().Parse(false));

    [Fact]
    public void RejectsString()
    {
        var result = V.Bool().SafeParse("true");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void RejectsNull()
    {
        var result = V.Bool().SafeParse(null);
        Assert.False(result.Success);
    }
}

public class NullSchemaTests
{
    [Fact]
    public void AcceptsNull() => Assert.Null(V.Null().Parse(null));

    [Fact]
    public void RejectsString()
    {
        var result = V.Null().SafeParse("hello");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }
}

public class AnySchemaTests
{
    [Fact]
    public void AcceptsAnything()
    {
        var s = V.Any();
        Assert.Equal("hello", s.Parse("hello"));
        Assert.Equal(42L, s.Parse(42L));
        Assert.Equal(true, s.Parse(true));
        Assert.Null(s.Parse(null));
    }
}

public class UnknownSchemaTests
{
    [Fact]
    public void AcceptsAnything()
    {
        var s = V.Unknown();
        Assert.Equal("hello", s.Parse("hello"));
        Assert.Equal(42L, s.Parse(42L));
    }
}

public class NeverSchemaTests
{
    [Fact]
    public void RejectsEverything()
    {
        var s = V.Never();
        var result = s.SafeParse("hello");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }
}

public class LiteralSchemaTests
{
    [Fact]
    public void AcceptsMatchingString()
    {
        var s = V.Literal("hello");
        Assert.Equal("hello", s.Parse("hello"));
    }

    [Fact]
    public void AcceptsMatchingNumber()
    {
        var s = V.Literal(42L);
        Assert.Equal(42L, s.Parse(42L));
    }

    [Fact]
    public void AcceptsMatchingBool()
    {
        var s = V.Literal(true);
        Assert.Equal(true, s.Parse(true));
    }

    [Fact]
    public void AcceptsMatchingNull()
    {
        var s = V.Literal(null);
        Assert.Null(s.Parse(null));
    }

    [Fact]
    public void RejectsMismatch()
    {
        var s = V.Literal("hello");
        var result = s.SafeParse("world");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidLiteral, result.Issues[0].Code);
    }
}

public class EnumSchemaTests
{
    [Fact]
    public void AcceptsValidValue()
    {
        var s = V.Enum("a", "b", "c");
        Assert.Equal("b", s.Parse("b"));
    }

    [Fact]
    public void AcceptsNumericValue()
    {
        var s = V.Enum("a", 1L, 2L);
        Assert.Equal(1L, s.Parse(1L));
    }

    [Fact]
    public void RejectsInvalidValue()
    {
        var s = V.Enum("a", "b", "c");
        var result = s.SafeParse("d");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }
}

public class ArraySchemaTests
{
    [Fact]
    public void AcceptsValidArray()
    {
        var s = V.Array(V.String());
        var input = new List<object?> { "a", "b" };
        var result = s.Parse(input);
        var list = Assert.IsType<List<object?>>(result);
        Assert.Equal(2, list.Count);
        Assert.Equal("a", list[0]);
    }

    [Fact]
    public void RejectsNonArray()
    {
        var s = V.Array(V.String());
        var result = s.SafeParse("not array");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void ValidatesItems()
    {
        var s = V.Array(V.String());
        var input = new List<object?> { "a", 42L };
        var result = s.SafeParse(input);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
        Assert.Contains(1, result.Issues[0].Path);
    }

    [Fact]
    public void MinItemsConstraint()
    {
        var s = V.Array(V.String()).MinItems(2);
        var result = s.SafeParse(new List<object?> { "a" });
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooSmall, result.Issues[0].Code);
    }

    [Fact]
    public void MaxItemsConstraint()
    {
        var s = V.Array(V.String()).MaxItems(1);
        var result = s.SafeParse(new List<object?> { "a", "b" });
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooLarge, result.Issues[0].Code);
    }
}

public class TupleSchemaTests
{
    [Fact]
    public void AcceptsValidTuple()
    {
        var s = V.Tuple(V.String(), V.Int());
        var input = new List<object?> { "hello", 42L };
        var result = (List<object?>)s.Parse(input)!;
        Assert.Equal("hello", result[0]);
        Assert.Equal(42L, result[1]);
    }

    [Fact]
    public void RejectsTooFew()
    {
        var s = V.Tuple(V.String(), V.Int());
        var result = s.SafeParse(new List<object?> { "hello" });
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooSmall, result.Issues[0].Code);
    }

    [Fact]
    public void RejectsTooMany()
    {
        var s = V.Tuple(V.String(), V.Int());
        var result = s.SafeParse(new List<object?> { "hello", 42L, "extra" });
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooLarge, result.Issues[0].Code);
    }

    [Fact]
    public void RejectsNonArray()
    {
        var s = V.Tuple(V.String());
        var result = s.SafeParse("not tuple");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }
}

public class ObjectSchemaTests
{
    [Fact]
    public void AcceptsValidObject()
    {
        var s = V.Object(new Dictionary<string, Schema>
        {
            ["name"] = V.String(),
            ["age"] = V.Int(),
        });
        var input = new Dictionary<string, object?>
        {
            ["name"] = "Alice",
            ["age"] = 30L,
        };
        var result = (Dictionary<string, object?>)s.Parse(input)!;
        Assert.Equal("Alice", result["name"]);
        Assert.Equal(30L, result["age"]);
    }

    [Fact]
    public void RejectsNonObject()
    {
        var s = V.Object(new Dictionary<string, Schema> { ["name"] = V.String() });
        var result = s.SafeParse("not object");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void RequiredFieldMissing()
    {
        var s = V.Object(new Dictionary<string, Schema>
        {
            ["name"] = V.String(),
        });
        var result = s.SafeParse(new Dictionary<string, object?>());
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.Required, result.Issues[0].Code);
    }

    [Fact]
    public void UnknownKeysRejectedByDefault()
    {
        var s = V.Object(new Dictionary<string, Schema>
        {
            ["name"] = V.String(),
        });
        var input = new Dictionary<string, object?> { ["name"] = "Alice", ["extra"] = "val" };
        var result = s.SafeParse(input);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.UnknownKey, result.Issues[0].Code);
    }

    [Fact]
    public void UnknownKeysStripped()
    {
        var s = V.Object(
            new Dictionary<string, Schema> { ["name"] = V.String() },
            UnknownKeyMode.Strip);
        var input = new Dictionary<string, object?> { ["name"] = "Alice", ["extra"] = "val" };
        var result = (Dictionary<string, object?>)s.Parse(input)!;
        Assert.Single(result);
        Assert.Equal("Alice", result["name"]);
    }

    [Fact]
    public void UnknownKeysAllowed()
    {
        var s = V.Object(
            new Dictionary<string, Schema> { ["name"] = V.String() },
            UnknownKeyMode.Allow);
        var input = new Dictionary<string, object?> { ["name"] = "Alice", ["extra"] = "val" };
        var result = (Dictionary<string, object?>)s.Parse(input)!;
        Assert.Equal(2, result.Count);
        Assert.Equal("val", result["extra"]);
    }

    [Fact]
    public void OptionalFieldMissing()
    {
        var s = V.Object(new Dictionary<string, Schema>
        {
            ["name"] = V.String(),
            ["nick"] = V.Optional(V.String()),
        });
        var input = new Dictionary<string, object?> { ["name"] = "Alice" };
        var result = s.SafeParse(input);
        Assert.True(result.Success);
    }

    [Fact]
    public void ChangeUnknownKeyMode()
    {
        var s = V.Object(new Dictionary<string, Schema> { ["name"] = V.String() })
            .UnknownKeys(UnknownKeyMode.Allow);
        var input = new Dictionary<string, object?> { ["name"] = "Alice", ["extra"] = "val" };
        var result = s.SafeParse(input);
        Assert.True(result.Success);
    }
}

public class RecordSchemaTests
{
    [Fact]
    public void AcceptsValidRecord()
    {
        var s = V.Record(V.Int());
        var input = new Dictionary<string, object?> { ["a"] = 1L, ["b"] = 2L };
        var result = (Dictionary<string, object?>)s.Parse(input)!;
        Assert.Equal(1L, result["a"]);
    }

    [Fact]
    public void ValidatesValues()
    {
        var s = V.Record(V.String());
        var input = new Dictionary<string, object?> { ["a"] = "ok", ["b"] = 42L };
        var result = s.SafeParse(input);
        Assert.False(result.Success);
    }

    [Fact]
    public void RejectsNonObject()
    {
        var s = V.Record(V.String());
        var result = s.SafeParse("not object");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }
}

public class UnionSchemaTests
{
    [Fact]
    public void MatchesFirstVariant()
    {
        var s = V.Union(V.String(), V.Int());
        Assert.Equal("hello", s.Parse("hello"));
    }

    [Fact]
    public void MatchesSecondVariant()
    {
        var s = V.Union(V.String(), V.Int());
        Assert.Equal(42L, s.Parse(42L));
    }

    [Fact]
    public void RejectsNonMatching()
    {
        var s = V.Union(V.String(), V.Int());
        var result = s.SafeParse(true);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidUnion, result.Issues[0].Code);
    }
}

public class IntersectionSchemaTests
{
    [Fact]
    public void MergesObjects()
    {
        var s = V.Intersection(
            V.Object(new Dictionary<string, Schema> { ["a"] = V.String() }, UnknownKeyMode.Strip),
            V.Object(new Dictionary<string, Schema> { ["b"] = V.Int() }, UnknownKeyMode.Strip));

        var input = new Dictionary<string, object?> { ["a"] = "hello", ["b"] = 42L };
        var result = (Dictionary<string, object?>)s.Parse(input)!;
        Assert.Equal("hello", result["a"]);
        Assert.Equal(42L, result["b"]);
    }

    [Fact]
    public void FailsIfAnyFails()
    {
        var s = V.Intersection(V.String(), V.Int());
        var result = s.SafeParse("hello");
        Assert.False(result.Success);
    }
}

public class OptionalSchemaTests
{
    [Fact]
    public void AcceptsAbsent()
    {
        var s = V.Optional(V.String());
        var result = s.SafeParse(AnyVali.Parse.Absent.Value);
        Assert.True(result.Success);
    }

    [Fact]
    public void AcceptsPresent()
    {
        var s = V.Optional(V.String());
        Assert.Equal("hello", s.Parse("hello"));
    }

    [Fact]
    public void ValidatesPresentValue()
    {
        var s = V.Optional(V.String());
        var result = s.SafeParse(42L);
        Assert.False(result.Success);
    }
}

public class NullableSchemaTests
{
    [Fact]
    public void AcceptsNull()
    {
        var s = V.Nullable(V.String());
        Assert.Null(s.Parse(null));
    }

    [Fact]
    public void AcceptsValid()
    {
        var s = V.Nullable(V.String());
        Assert.Equal("hello", s.Parse("hello"));
    }

    [Fact]
    public void RejectsInvalid()
    {
        var s = V.Nullable(V.String());
        var result = s.SafeParse(42L);
        Assert.False(result.Success);
    }
}

public class RefSchemaTests
{
    [Fact]
    public void ResolvesRef()
    {
        var stringSchema = V.String();
        var s = new RefSchema("#/definitions/MyString", () => stringSchema);
        Assert.Equal("hello", s.Parse("hello"));
    }

    [Fact]
    public void UnresolvedRefFails()
    {
        var s = new RefSchema("#/definitions/Missing");
        var result = s.SafeParse("hello");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.UnsupportedSchemaKind, result.Issues[0].Code);
    }
}

public class DefaultSchemaTests
{
    [Fact]
    public void DefaultAppliedWhenAbsent()
    {
        var s = V.String().Default("fallback");
        var result = s.SafeParse(AnyVali.Parse.Absent.Value);
        Assert.True(result.Success);
        Assert.Equal("fallback", result.Data);
    }

    [Fact]
    public void DefaultNotAppliedWhenPresent()
    {
        var s = V.String().Default("fallback");
        Assert.Equal("hello", s.Parse("hello"));
    }

    [Fact]
    public void InvalidDefaultProducesDefaultInvalidIssue()
    {
        var s = V.Int().Min(10).Default(5L);
        var result = s.SafeParse(AnyVali.Parse.Absent.Value);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.DefaultInvalid, result.Issues[0].Code);
    }
}
