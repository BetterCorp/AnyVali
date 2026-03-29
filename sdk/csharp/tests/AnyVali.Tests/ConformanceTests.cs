using System.Text.Json;
using AnyVali;
using AnyVali.Parse;
using Xunit;
using Xunit.Abstractions;

namespace AnyVali.Tests;

/// <summary>
/// Runs the shared conformance corpus from spec/corpus/.
/// Each JSON file contains a suite of test cases with schema, input, expected validity, expected output, and expected issues.
/// </summary>
public class ConformanceTests
{
    private readonly ITestOutputHelper _output;

    public ConformanceTests(ITestOutputHelper output)
    {
        _output = output;
    }

    private static string FindCorpusPath()
    {
        // Navigate from the test project up to the repo root
        var dir = AppDomain.CurrentDomain.BaseDirectory;
        // Walk up to find spec/corpus
        for (var i = 0; i < 10; i++)
        {
            var candidate = Path.Combine(dir, "spec", "corpus");
            if (Directory.Exists(candidate)) return candidate;
            dir = Path.GetDirectoryName(dir)!;
        }
        // Try absolute path
        var abs = Path.GetFullPath(Path.Combine(AppDomain.CurrentDomain.BaseDirectory,
            "..", "..", "..", "..", "..", "..", "spec", "corpus"));
        if (Directory.Exists(abs)) return abs;
        throw new DirectoryNotFoundException("Could not find spec/corpus directory");
    }

    public static IEnumerable<object[]> GetCorpusFiles()
    {
        string corpusPath;
        try
        {
            corpusPath = FindCorpusPath();
        }
        catch
        {
            yield break;
        }

        foreach (var file in Directory.GetFiles(corpusPath, "*.json", SearchOption.AllDirectories))
        {
            yield return new object[] { file };
        }
    }

    [Theory]
    [MemberData(nameof(GetCorpusFiles))]
    public void RunCorpusFile(string filePath)
    {
        var json = File.ReadAllText(filePath);
        var doc = JsonSerializer.Deserialize<JsonElement>(json);
        var suiteName = doc.GetProperty("suite").GetString()!;
        var cases = doc.GetProperty("cases");

        foreach (var testCase in cases.EnumerateArray())
        {
            var description = testCase.GetProperty("description").GetString()!;
            var schemaElement = testCase.GetProperty("schema");
            var inputElement = testCase.GetProperty("input");
            var expectedValid = testCase.GetProperty("valid").GetBoolean();
            var expectedIssues = testCase.GetProperty("issues");

            _output.WriteLine($"[{suiteName}] {description}");

            // Parse the schema document
            var schemaJson = schemaElement.GetRawText();
            var schemaDoc = JsonSerializer.Deserialize<AnyValiDocument>(schemaJson, JsonHelper.SerializerOptions)!;

            // Import the schema
            var schema = V.Import(schemaDoc);

            // Convert input
            var input = JsonHelper.ElementToObject(inputElement);

            // Run SafeParse
            var result = schema.SafeParse(input);

            // Validate result
            Assert.Equal(expectedValid, result.Success);

            if (!expectedValid)
            {
                // Check that we got the expected issue codes and paths
                var expectedIssueArray = expectedIssues.EnumerateArray().ToList();
                Assert.Equal(expectedIssueArray.Count, result.Issues.Count);

                for (var i = 0; i < expectedIssueArray.Count; i++)
                {
                    var expected = expectedIssueArray[i];
                    var actual = result.Issues[i];

                    var expectedCode = expected.GetProperty("code").GetString()!;
                    Assert.Equal(expectedCode, actual.Code);

                    // Check path
                    var expectedPath = expected.GetProperty("path").EnumerateArray()
                        .Select(e => JsonHelper.ElementToObject(e))
                        .ToList();

                    Assert.Equal(expectedPath.Count, actual.Path.Count);
                    for (var j = 0; j < expectedPath.Count; j++)
                    {
                        var ep = expectedPath[j];
                        var ap = actual.Path[j];
                        // Normalize: both could be string or long/int
                        if (ep is long el && ap is int ai)
                            Assert.Equal(el, (long)ai);
                        else if (ep is long el2 && ap is long al)
                            Assert.Equal(el2, al);
                        else
                            Assert.Equal(ep?.ToString(), ap?.ToString());
                    }
                }
            }
            else
            {
                // Check output matches expected
                var expectedOutput = testCase.GetProperty("output");
                AssertOutputMatches(expectedOutput, result.Data);
            }
        }
    }

    private void AssertOutputMatches(JsonElement expected, object? actual)
    {
        switch (expected.ValueKind)
        {
            case JsonValueKind.Null:
                Assert.Null(actual);
                break;
            case JsonValueKind.String:
                Assert.Equal(expected.GetString(), actual);
                break;
            case JsonValueKind.Number:
                if (expected.TryGetInt64(out var l))
                {
                    // actual could be long or double
                    if (actual is long al)
                        Assert.Equal(l, al);
                    else if (actual is double ad)
                        Assert.Equal((double)l, ad);
                    else
                        Assert.Fail($"Expected number {l}, got {actual?.GetType().Name}: {actual}");
                }
                else
                {
                    var d = expected.GetDouble();
                    Assert.Equal(d, Convert.ToDouble(actual), 10);
                }
                break;
            case JsonValueKind.True:
                Assert.Equal(true, actual);
                break;
            case JsonValueKind.False:
                Assert.Equal(false, actual);
                break;
            case JsonValueKind.Array:
                var actualList = Assert.IsType<List<object?>>(actual);
                var expectedArray = expected.EnumerateArray().ToList();
                Assert.Equal(expectedArray.Count, actualList.Count);
                for (var i = 0; i < expectedArray.Count; i++)
                    AssertOutputMatches(expectedArray[i], actualList[i]);
                break;
            case JsonValueKind.Object:
                var actualDict = Assert.IsType<Dictionary<string, object?>>(actual);
                foreach (var prop in expected.EnumerateObject())
                {
                    Assert.True(actualDict.ContainsKey(prop.Name), $"Missing key: {prop.Name}");
                    AssertOutputMatches(prop.Value, actualDict[prop.Name]);
                }
                // Check no extra keys
                Assert.Equal(expected.EnumerateObject().Count(), actualDict.Count);
                break;
        }
    }
}
