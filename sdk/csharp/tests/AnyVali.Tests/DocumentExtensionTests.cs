using AnyVali;
using System.Text.Json;
using Xunit;

namespace AnyVali.Tests;

public class DocumentExtensionTests
{
    private static AnyValiDocument Document(Dictionary<string, object?>? extensions, Schema? schema = null) => new()
    {
        AnyvaliVersion = "1.0", SchemaVersion = "1.1",
        Root = (schema ?? V.String()).Export().Root,
        Definitions = new(), Extensions = extensions!,
    };

    [Theory]
    [InlineData("vendor")]
    [InlineData("default")]
    [InlineData("csharp")]
    public void RejectsUnsupportedSemanticNamespaces(string name)
    {
        var doc = Document(new() { [name] = new Dictionary<string, object?> { ["_criticality"] = "semantic" } });
        foreach (var input in new[] { doc, AnyValiDocument.FromJson(doc.ToJson()) })
        {
            var issue = Assert.Single(Assert.Throws<ValidationError>(() => V.Import(input)).Issues);
            Assert.Equal(IssueCodes.UnsupportedExtension, issue.Code);
            Assert.Equal(new object[] { "extensions", name }, issue.Path);
        }
    }

    [Fact]
    public void RejectsMalformedExtensions()
    {
        Assert.Throws<InvalidOperationException>(() => V.Import(Document(null)));
        foreach (var payload in new object?[] { null, "invalid", new List<object?>(),
            new Dictionary<string, object?> { ["_criticality"] = null },
            new Dictionary<string, object?> { ["_criticality"] = 1 },
            new Dictionary<string, object?> { ["_criticality"] = "unknown" } })
        {
            Assert.Throws<InvalidOperationException>(() => V.Import(Document(new() { ["vendor"] = payload })));
        }
    }

    [Theory]
    [InlineData(false)]
    [InlineData(true)]
    public void RetainsInformationalExtensionsThroughCompositionAndRoundtrips(bool explicitCriticality)
    {
        var payload = new Dictionary<string, object?> { ["hint"] = "keep" };
        if (explicitCriticality) payload["_criticality"] = "informational";
        var extensions = new Dictionary<string, object?> { ["vendor"] = payload, ["default"] = payload };
        var child = V.Import(Document(extensions));
        var cases = new (Schema, object)[]
        {
            (child, "ok"),
            (V.Object(new() { ["payload"] = child }), new Dictionary<string, object?> { ["payload"] = "ok" }),
            (V.Array(child), new List<object?> { "ok" }),
            (V.Record(child), new Dictionary<string, object?> { ["key"] = "ok" }),
            (V.Tuple(child), new List<object?> { "ok" }),
            (V.Optional(child), "ok"), (V.Nullable(child), "ok"),
            (V.Union(child, V.Bool()), "ok"), (V.Intersection(child, V.String()), "ok"),
        };
        foreach (var (schema, input) in cases)
        {
            var exported = V.Export(schema, ExportMode.Extended);
            Assert.Equal(JsonSerializer.Serialize(extensions), JsonSerializer.Serialize(exported.Extensions));
            var roundtripped = V.Import(AnyValiDocument.FromJson(exported.ToJson()));
            Assert.Equal(exported.ToJson(), V.Export(roundtripped, ExportMode.Extended).ToJson());
            Assert.Equal(JsonSerializer.Serialize(schema.Parse(input)), JsonSerializer.Serialize(roundtripped.Parse(input)));
            Assert.Empty(V.Export(schema, ExportMode.Portable).Extensions);
        }
    }

    [Fact]
    public void MergesNamespacesAndRejectsConflictsOnlyInExtendedMode()
    {
        var child = V.Import(Document(new() { ["vendor"] = new Dictionary<string, object?>
            { ["hint"] = "keep", ["_criticality"] = "informational" } }));
        var reordered = V.Import(Document(new() { ["vendor"] = new Dictionary<string, object?>
            { ["_criticality"] = "informational", ["hint"] = "keep" } }));
        var other = V.Import(Document(new() { ["other"] = new Dictionary<string, object?>() }));
        var parent = V.Object(new() { ["first"] = child, ["second"] = child, ["third"] = reordered, ["fourth"] = other });
        Assert.Equal(2, V.Export(parent, ExportMode.Extended).Extensions.Count);

        var conflict = V.Import(Document(new() { ["vendor"] = new Dictionary<string, object?> { ["hint"] = "different" } }));
        var mixed = V.Object(new() { ["first"] = child, ["second"] = conflict });
        Assert.Equal("Conflicting extension namespace: vendor",
            Assert.Throws<InvalidOperationException>(() => V.Export(mixed, ExportMode.Extended)).Message);
        Assert.Empty(V.Export(mixed, ExportMode.Portable).Extensions);
    }

    [Fact]
    public void ImportAndExportIsolateNestedPayloads()
    {
        var values = new List<object?> { "keep" };
        var payload = new Dictionary<string, object?> { ["nested"] = new Dictionary<string, object?> { ["values"] = values } };
        var doc = Document(new() { ["vendor"] = payload });
        var expected = JsonSerializer.Serialize(doc.Extensions);
        var schema = V.Import(doc);
        values[0] = "changed";
        doc.Extensions.Clear();
        var exported = V.Export(schema, ExportMode.Extended);
        Assert.Equal(expected, JsonSerializer.Serialize(exported.Extensions));
        var exportedPayload = (Dictionary<string, object?>)exported.Extensions["vendor"]!;
        var nested = (Dictionary<string, object?>)exportedPayload["nested"]!;
        ((List<object?>)nested["values"]!)[0] = "changed again";
        Assert.Equal(expected, JsonSerializer.Serialize(V.Export(schema, ExportMode.Extended).Extensions));
    }

    [Fact]
    public void FluentClonesRetainExtensionsForEverySchemaKind()
    {
        Schema[] schemas = [V.String(), V.Number(), V.Float32(), V.Float64(), V.Int(), V.Int8(), V.Int16(),
            V.Int32(), V.Int64(), V.Uint8(), V.Uint16(), V.Uint32(), V.Uint64(), V.Bool(), V.Null(),
            V.Any(), V.Unknown(), V.Never(), V.Literal("ok"), V.Enum("ok"), V.Array(V.String()),
            V.Tuple(V.String()), V.Object(new()), V.Record(V.String()), V.Union(V.String(), V.Bool()),
            V.Intersection(V.String(), V.String()), V.Optional(V.String()), V.Nullable(V.String())];
        var extensions = new Dictionary<string, object?> { ["vendor"] = new Dictionary<string, object?> { ["hint"] = "keep" } };
        var imported = schemas.Select(schema => V.Import(Document(extensions, schema))).ToList();
        imported.Add(V.Import(AnyValiDocument.FromJson("""
            {"anyvaliVersion":"1.0","schemaVersion":"1.1","root":{"kind":"ref","ref":"#/definitions/Value"},
             "definitions":{"Value":{"kind":"string"}},"extensions":{"vendor":{"hint":"keep"}}}
            """)));
        foreach (var schema in imported)
        {
            var clone = schema.Default("ok").Coerce().Describe("description").Metadata(new() { ["custom"] = true });
            Assert.Equal(JsonSerializer.Serialize(extensions), JsonSerializer.Serialize(V.Export(clone, ExportMode.Extended).Extensions));
        }
    }
}
