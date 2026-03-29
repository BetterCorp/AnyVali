using AnyVali;
using AnyVali.Schemas;
using Xunit;

namespace AnyVali.Tests;

public class ExportTests
{
    [Fact]
    public void ExportStringSchema()
    {
        var s = V.String().MinLength(1).MaxLength(100);
        var doc = s.Export();
        Assert.Equal("1.0", doc.AnyvaliVersion);
        Assert.Equal("1", doc.SchemaVersion);
        Assert.Equal("string", doc.Root["kind"]);
        Assert.Equal(1L, doc.Root["minLength"]);
        Assert.Equal(100L, doc.Root["maxLength"]);
    }

    [Fact]
    public void ExportNumberSchema()
    {
        var s = V.Number().Min(0).Max(100);
        var doc = s.Export();
        Assert.Equal("number", doc.Root["kind"]);
        Assert.Equal(0.0, doc.Root["min"]);
        Assert.Equal(100.0, doc.Root["max"]);
    }

    [Fact]
    public void ExportIntSchema()
    {
        var s = V.Int().Min(0);
        var doc = s.Export();
        Assert.Equal("int", doc.Root["kind"]);
    }

    [Fact]
    public void ExportObjectSchema()
    {
        var s = V.Object(new Dictionary<string, Schema>
        {
            ["name"] = V.String(),
            ["age"] = V.Optional(V.Int()),
        });
        var doc = s.Export();
        Assert.Equal("object", doc.Root["kind"]);
        var required = (List<object?>)doc.Root["required"]!;
        Assert.Contains("name", required);
        Assert.DoesNotContain("age", required);
    }

    [Fact]
    public void ExportArraySchema()
    {
        var s = V.Array(V.String()).MinItems(1);
        var doc = s.Export();
        Assert.Equal("array", doc.Root["kind"]);
        Assert.Equal(1L, doc.Root["minItems"]);
    }

    [Fact]
    public void ExportUnionSchema()
    {
        var s = V.Union(V.String(), V.Int());
        var doc = s.Export();
        Assert.Equal("union", doc.Root["kind"]);
    }

    [Fact]
    public void ExportWithDefault()
    {
        var s = V.String().Default("hello");
        var doc = s.Export();
        Assert.Equal("hello", doc.Root["default"]);
    }

    [Fact]
    public void PortableExportFailsForNonPortable()
    {
        var s = V.String();
        s.IsPortable = false;
        Assert.Throws<InvalidOperationException>(() => s.Export(ExportMode.Portable));
    }

    [Fact]
    public void ExtendedExportAllowed()
    {
        var s = V.String();
        s.IsPortable = false;
        var doc = s.Export(ExportMode.Extended);
        Assert.NotNull(doc);
    }

    [Fact]
    public void ExportDocumentToJson()
    {
        var s = V.String();
        var doc = s.Export();
        var json = doc.ToJson();
        Assert.Contains("\"kind\"", json);
        Assert.Contains("\"string\"", json);
    }
}

public class ImportTests
{
    [Fact]
    public void ImportStringSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "string", ["minLength"] = 1L },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal("hello", schema.Parse("hello"));
        var result = schema.SafeParse("");
        Assert.False(result.Success);
    }

    [Fact]
    public void ImportNumberSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "number", ["min"] = 0.0 },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal(5.0, schema.Parse(5L));
        var result = schema.SafeParse(-1L);
        Assert.False(result.Success);
    }

    [Fact]
    public void ImportIntSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "int", ["min"] = 1L },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal(5L, schema.Parse(5L));
    }

    [Fact]
    public void ImportObjectSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?>
            {
                ["kind"] = "object",
                ["properties"] = new Dictionary<string, object?>
                {
                    ["name"] = new Dictionary<string, object?> { ["kind"] = "string" },
                    ["age"] = new Dictionary<string, object?> { ["kind"] = "int" },
                },
                ["required"] = new List<object?> { "name", "age" },
                ["unknownKeys"] = "reject",
            },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        var input = new Dictionary<string, object?> { ["name"] = "Alice", ["age"] = 30L };
        var result = (Dictionary<string, object?>)schema.Parse(input)!;
        Assert.Equal("Alice", result["name"]);
    }

    [Fact]
    public void ImportArraySchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?>
            {
                ["kind"] = "array",
                ["items"] = new Dictionary<string, object?> { ["kind"] = "string" },
            },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        var input = new List<object?> { "a", "b" };
        var result = (List<object?>)schema.Parse(input)!;
        Assert.Equal(2, result.Count);
    }

    [Fact]
    public void ImportUnionSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?>
            {
                ["kind"] = "union",
                ["variants"] = new List<object?>
                {
                    new Dictionary<string, object?> { ["kind"] = "string" },
                    new Dictionary<string, object?> { ["kind"] = "int" },
                },
            },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal("hello", schema.Parse("hello"));
        Assert.Equal(42L, schema.Parse(42L));
    }

    [Fact]
    public void ImportWithDefault()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "string", ["default"] = "fallback" },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        var result = schema.SafeParse(AnyVali.Parse.Absent.Value);
        Assert.True(result.Success);
        Assert.Equal("fallback", result.Data);
    }

    [Fact]
    public void ImportWithCoercion()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "int", ["coerce"] = "string->int" },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal(42L, schema.Parse("42"));
    }

    [Fact]
    public void ImportRefSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "ref", ["ref"] = "#/definitions/MyStr" },
            Definitions = new Dictionary<string, object?>
            {
                ["MyStr"] = new Dictionary<string, object?> { ["kind"] = "string" },
            },
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal("hello", schema.Parse("hello"));
    }

    [Fact]
    public void ImportNullableSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?>
            {
                ["kind"] = "nullable",
                ["schema"] = new Dictionary<string, object?> { ["kind"] = "string" },
            },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Null(schema.Parse(null));
        Assert.Equal("hello", schema.Parse("hello"));
    }

    [Fact]
    public void ImportOptionalSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?>
            {
                ["kind"] = "optional",
                ["schema"] = new Dictionary<string, object?> { ["kind"] = "string" },
            },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        var result = schema.SafeParse(AnyVali.Parse.Absent.Value);
        Assert.True(result.Success);
    }

    [Fact]
    public void ImportTupleSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?>
            {
                ["kind"] = "tuple",
                ["elements"] = new List<object?>
                {
                    new Dictionary<string, object?> { ["kind"] = "string" },
                    new Dictionary<string, object?> { ["kind"] = "int" },
                },
            },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        var input = new List<object?> { "hello", 42L };
        var result = (List<object?>)schema.Parse(input)!;
        Assert.Equal("hello", result[0]);
        Assert.Equal(42L, result[1]);
    }

    [Fact]
    public void ImportRecordSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?>
            {
                ["kind"] = "record",
                ["values"] = new Dictionary<string, object?> { ["kind"] = "int" },
            },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        var input = new Dictionary<string, object?> { ["a"] = 1L, ["b"] = 2L };
        var result = (Dictionary<string, object?>)schema.Parse(input)!;
        Assert.Equal(1L, result["a"]);
    }

    [Fact]
    public void ImportIntersectionSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?>
            {
                ["kind"] = "intersection",
                ["allOf"] = new List<object?>
                {
                    new Dictionary<string, object?>
                    {
                        ["kind"] = "object",
                        ["properties"] = new Dictionary<string, object?>
                        {
                            ["a"] = new Dictionary<string, object?> { ["kind"] = "string" },
                        },
                        ["required"] = new List<object?> { "a" },
                        ["unknownKeys"] = "strip",
                    },
                    new Dictionary<string, object?>
                    {
                        ["kind"] = "object",
                        ["properties"] = new Dictionary<string, object?>
                        {
                            ["b"] = new Dictionary<string, object?> { ["kind"] = "int" },
                        },
                        ["required"] = new List<object?> { "b" },
                        ["unknownKeys"] = "strip",
                    },
                },
            },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        var input = new Dictionary<string, object?> { ["a"] = "hello", ["b"] = 42L };
        var result = (Dictionary<string, object?>)schema.Parse(input)!;
        Assert.Equal("hello", result["a"]);
        Assert.Equal(42L, result["b"]);
    }

    [Fact]
    public void ImportLiteralSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "literal", ["value"] = "hello" },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal("hello", schema.Parse("hello"));
    }

    [Fact]
    public void ImportEnumSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?>
            {
                ["kind"] = "enum",
                ["values"] = new List<object?> { "a", "b", "c" },
            },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal("b", schema.Parse("b"));
    }

    [Fact]
    public void ImportBoolSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "bool" },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal(true, schema.Parse(true));
    }

    [Fact]
    public void ImportNullSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "null" },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Null(schema.Parse(null));
    }

    [Fact]
    public void ImportAnySchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "any" },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal("anything", schema.Parse("anything"));
    }

    [Fact]
    public void ImportNeverSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "never" },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        var result = schema.SafeParse("anything");
        Assert.False(result.Success);
    }

    [Fact]
    public void ImportUnsupportedKindThrows()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "not_a_real_kind" },
            Definitions = new(),
            Extensions = new(),
        };
        Assert.Throws<InvalidOperationException>(() => V.Import(doc));
    }

    [Fact]
    public void RoundTripExportImport()
    {
        var original = V.Object(new Dictionary<string, Schema>
        {
            ["name"] = V.String().MinLength(1),
            ["age"] = V.Optional(V.Int().Min(0)),
        });

        var doc = original.Export();
        var imported = V.Import(doc);

        var input = new Dictionary<string, object?> { ["name"] = "Alice", ["age"] = 30L };
        var result = imported.SafeParse(input);
        Assert.True(result.Success);
    }

    [Fact]
    public void ImportAllIntTypes()
    {
        foreach (var kind in new[] { "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64" })
        {
            var doc = new AnyValiDocument
            {
                AnyvaliVersion = "1.0",
                SchemaVersion = "1",
                Root = new Dictionary<string, object?> { ["kind"] = kind },
                Definitions = new(),
                Extensions = new(),
            };
            var schema = V.Import(doc);
            Assert.Equal(0L, schema.Parse(0L));
        }
    }

    [Fact]
    public void ImportFloat32Schema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "float32" },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal(1.5, schema.Parse(1.5));
    }

    [Fact]
    public void ImportFloat64Schema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "float64" },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal(1.5, schema.Parse(1.5));
    }

    [Fact]
    public void ImportUnknownSchema()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "unknown" },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.Equal("x", schema.Parse("x"));
    }
}
