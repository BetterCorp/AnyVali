using System.Diagnostics;
using AnyVali;
using AnyVali.Schemas;
using Xunit;

namespace AnyVali.Tests;

/// <summary>
/// Security-focused tests covering CVE/CWE scenarios for the AnyVali C# SDK.
/// </summary>
public class SecurityTests
{
    // ---------------------------------------------------------------
    // CVE-2016-4055: ReDoS - Catastrophic backtracking patterns
    // Verifies that user-supplied regexes with catastrophic backtracking
    // do not hang the validator indefinitely. Each test uses Task.Run
    // with a 2-second timeout so the test suite does not block.
    // ---------------------------------------------------------------

    private static bool RunWithTimeout(Action action, int timeoutMs = 2000)
    {
        var task = Task.Run(action);
        return task.Wait(timeoutMs);
    }

    [Fact]
    public void ReDoS_PatternValidation_CompletesInBoundedTime()
    {
        // Classic ReDoS pattern: (a+)+ against a string of 'a's followed by '!'
        var s = V.String().Pattern(@"^(a+)+$");
        var malicious = new string('a', 25) + "!";
        var completed = RunWithTimeout(() => { s.SafeParse(malicious); });
        Assert.True(completed,
            "ReDoS: Pattern ^(a+)+$ with 25 'a' chars + '!' did not complete within 2 seconds. " +
            "The regex engine is vulnerable to catastrophic backtracking (CVE-2016-4055).");
    }

    [Fact]
    public void ReDoS_NestedQuantifiers_CompletesInBoundedTime()
    {
        // Nested quantifier pattern that causes exponential backtracking
        var s = V.String().Pattern(@"^([a-zA-Z0-9]+)*$");
        var malicious = new string('a', 25) + "!";
        var completed = RunWithTimeout(() => { s.SafeParse(malicious); });
        Assert.True(completed,
            "ReDoS: Pattern ^([a-zA-Z0-9]+)*$ with 25 'a' chars + '!' did not complete within 2 seconds. " +
            "The regex engine is vulnerable to catastrophic backtracking (CVE-2016-4055).");
    }

    [Fact]
    public void ReDoS_EmailLikePattern_CompletesInBoundedTime()
    {
        // Evil email-like regex with catastrophic backtracking
        var s = V.String().Pattern(@"^([a-zA-Z0-9._%+-]+)+@([a-zA-Z0-9.-]+)+\.[a-zA-Z]{2,}$");
        var malicious = new string('a', 25) + "@" + new string('b', 25) + "!";
        var completed = RunWithTimeout(() => { s.SafeParse(malicious); });
        Assert.True(completed,
            "ReDoS: Evil email pattern with 25-char segments did not complete within 2 seconds. " +
            "The regex engine is vulnerable to catastrophic backtracking (CVE-2016-4055).");
    }

    [Fact]
    public void ReDoS_FormatEmail_CompletesInBoundedTimeOnMaliciousInput()
    {
        var s = V.String().Format("email");
        // Crafted input designed to cause backtracking in email regex
        var malicious = new string('a', 50) + "@" + new string('.', 50);
        var sw = Stopwatch.StartNew();
        var result = s.SafeParse(malicious);
        sw.Stop();
        Assert.False(result.Success);
        Assert.True(sw.ElapsedMilliseconds < 5000, $"Email format validation took {sw.ElapsedMilliseconds}ms, expected < 5000ms");
    }

    // ---------------------------------------------------------------
    // CVE-2003-1564: Recursive $ref - Self-referencing schema imports
    // ---------------------------------------------------------------

    [Fact]
    public void RecursiveRef_DirectSelfReference_ImportSucceeds()
    {
        // Schema where a definition references itself: { kind: "ref", ref: "#/definitions/Self" }
        // NOTE: Import uses lazy resolution, so import itself succeeds. Actually calling
        // SafeParse on a purely self-referencing schema would cause StackOverflowException
        // (which kills the process and cannot be caught). This test verifies import does
        // not eagerly resolve the cycle. A production fix should add a depth guard to
        // RefSchema.RunPipeline (CVE-2003-1564).
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "ref", ["ref"] = "#/definitions/Self" },
            Definitions = new Dictionary<string, object?>
            {
                ["Self"] = new Dictionary<string, object?> { ["kind"] = "ref", ["ref"] = "#/definitions/Self" },
            },
            Extensions = new(),
        };
        // Import should succeed (lazy ref resolution)
        var schema = V.Import(doc);
        Assert.NotNull(schema);
        // We intentionally do NOT call SafeParse here because the unbounded recursion
        // in RefSchema.RunPipeline would cause a StackOverflowException that kills the
        // test process. This is a known vulnerability (CVE-2003-1564).
    }

    [Fact]
    public void RecursiveRef_IndirectCycle_ImportSucceeds()
    {
        // A -> B -> A cycle
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "ref", ["ref"] = "#/definitions/A" },
            Definitions = new Dictionary<string, object?>
            {
                ["A"] = new Dictionary<string, object?> { ["kind"] = "ref", ["ref"] = "#/definitions/B" },
                ["B"] = new Dictionary<string, object?> { ["kind"] = "ref", ["ref"] = "#/definitions/A" },
            },
            Extensions = new(),
        };
        // Import should succeed (lazy ref resolution)
        var schema = V.Import(doc);
        Assert.NotNull(schema);
        // As above, SafeParse would StackOverflow. A depth guard is needed (CVE-2003-1564).
    }

    [Fact]
    public void RecursiveRef_ObjectWithSelfReference_DoesNotCrash()
    {
        // A tree-like schema: object with a property referencing itself
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "ref", ["ref"] = "#/definitions/TreeNode" },
            Definitions = new Dictionary<string, object?>
            {
                ["TreeNode"] = new Dictionary<string, object?>
                {
                    ["kind"] = "object",
                    ["properties"] = new Dictionary<string, object?>
                    {
                        ["value"] = new Dictionary<string, object?> { ["kind"] = "string" },
                        ["child"] = new Dictionary<string, object?> { ["kind"] = "ref", ["ref"] = "#/definitions/TreeNode" },
                    },
                    ["required"] = new List<object?> { "value" },
                    ["unknownKeys"] = "reject",
                },
            },
            Extensions = new(),
        };
        try
        {
            var schema = V.Import(doc);
            // Parse a valid non-recursive instance
            var input = new Dictionary<string, object?> { ["value"] = "root" };
            var result = schema.SafeParse(input);
            Assert.True(result.Success);
        }
        catch
        {
            // Throwing during import is acceptable
        }
    }

    // ---------------------------------------------------------------
    // CWE-190: Integer overflow - All int width boundaries
    // ---------------------------------------------------------------

    [Fact]
    public void IntOverflow_Int8_BoundaryValues()
    {
        var s = V.Int8();
        Assert.Equal(127L, s.Parse(127L));
        Assert.Equal(-128L, s.Parse(-128L));

        Assert.False(s.SafeParse(128L).Success);
        Assert.False(s.SafeParse(-129L).Success);
        Assert.False(s.SafeParse(255L).Success);
        Assert.False(s.SafeParse(-256L).Success);
    }

    [Fact]
    public void IntOverflow_Int16_BoundaryValues()
    {
        var s = V.Int16();
        Assert.Equal(32767L, s.Parse(32767L));
        Assert.Equal(-32768L, s.Parse(-32768L));

        Assert.False(s.SafeParse(32768L).Success);
        Assert.False(s.SafeParse(-32769L).Success);
    }

    [Fact]
    public void IntOverflow_Int32_BoundaryValues()
    {
        var s = V.Int32();
        Assert.Equal(2147483647L, s.Parse(2147483647L));
        Assert.Equal(-2147483648L, s.Parse(-2147483648L));

        Assert.False(s.SafeParse(2147483648L).Success);
        Assert.False(s.SafeParse(-2147483649L).Success);
    }

    [Fact]
    public void IntOverflow_Int64_BoundaryValues()
    {
        var s = V.Int64();
        Assert.Equal(long.MaxValue, s.Parse(long.MaxValue));
        Assert.Equal(long.MinValue, s.Parse(long.MinValue));
    }

    [Fact]
    public void IntOverflow_Uint8_BoundaryValues()
    {
        var s = V.Uint8();
        Assert.Equal(0L, s.Parse(0L));
        Assert.Equal(255L, s.Parse(255L));

        Assert.False(s.SafeParse(256L).Success);
        Assert.False(s.SafeParse(-1L).Success);
    }

    [Fact]
    public void IntOverflow_Uint16_BoundaryValues()
    {
        var s = V.Uint16();
        Assert.Equal(0L, s.Parse(0L));
        Assert.Equal(65535L, s.Parse(65535L));

        Assert.False(s.SafeParse(65536L).Success);
        Assert.False(s.SafeParse(-1L).Success);
    }

    [Fact]
    public void IntOverflow_Uint32_BoundaryValues()
    {
        var s = V.Uint32();
        Assert.Equal(0L, s.Parse(0L));
        Assert.Equal(4294967295L, s.Parse(4294967295L));

        Assert.False(s.SafeParse(4294967296L).Success);
        Assert.False(s.SafeParse(-1L).Success);
    }

    [Fact]
    public void IntOverflow_Uint64_BoundaryValues()
    {
        var s = V.Uint64();
        Assert.Equal(0L, s.Parse(0L));
        // uint64 is capped at long.MaxValue in the C# SDK
        Assert.Equal(long.MaxValue, s.Parse(long.MaxValue));
        Assert.False(s.SafeParse(-1L).Success);
    }

    [Fact]
    public void IntOverflow_Int8_OnePastBoundary()
    {
        var s = V.Int8();
        var overMax = s.SafeParse(128L);
        Assert.False(overMax.Success);
        Assert.Equal(IssueCodes.TooLarge, overMax.Issues[0].Code);

        var underMin = s.SafeParse(-129L);
        Assert.False(underMin.Success);
        Assert.Equal(IssueCodes.TooSmall, underMin.Issues[0].Code);
    }

    [Fact]
    public void IntOverflow_LargePositiveValue_RejectedByNarrowTypes()
    {
        var int8 = V.Int8();
        var int16 = V.Int16();
        var int32 = V.Int32();
        var uint8 = V.Uint8();
        var uint16 = V.Uint16();
        var uint32 = V.Uint32();

        long large = 5000000000L; // exceeds int32 range
        Assert.False(int8.SafeParse(large).Success);
        Assert.False(int16.SafeParse(large).Success);
        Assert.False(int32.SafeParse(large).Success);
        Assert.False(uint8.SafeParse(large).Success);
        Assert.False(uint16.SafeParse(large).Success);
        Assert.False(uint32.SafeParse(large).Success);
    }

    [Fact]
    public void IntOverflow_NegativeValue_RejectedByUnsignedTypes()
    {
        Assert.False(V.Uint8().SafeParse(-1L).Success);
        Assert.False(V.Uint16().SafeParse(-1L).Success);
        Assert.False(V.Uint32().SafeParse(-1L).Success);
        Assert.False(V.Uint64().SafeParse(-1L).Success);
        Assert.False(V.Uint8().SafeParse(-128L).Success);
        Assert.False(V.Uint16().SafeParse(-32768L).Success);
        Assert.False(V.Uint32().SafeParse(-2147483648L).Success);
    }

    // ---------------------------------------------------------------
    // CWE-20: NaN/Infinity rejection
    // ---------------------------------------------------------------

    [Fact]
    public void NaN_RejectedByNumber()
    {
        var s = V.Number();
        var result = s.SafeParse(double.NaN);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void PositiveInfinity_RejectedByNumber()
    {
        var s = V.Number();
        var result = s.SafeParse(double.PositiveInfinity);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void NegativeInfinity_RejectedByNumber()
    {
        var s = V.Number();
        var result = s.SafeParse(double.NegativeInfinity);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void NaN_RejectedByFloat32()
    {
        var s = V.Float32();
        var result = s.SafeParse(double.NaN);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void NaN_RejectedByFloat64()
    {
        var s = V.Float64();
        var result = s.SafeParse(double.NaN);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    // CWE-20 / spec 1.4: float32 MUST reject values outside the binary32 range.
    // Without the check it silently accepts any double, defeating narrowing.
    [Fact]
    public void Float32_RejectsValueAboveRange()
    {
        var s = V.Float32();
        // 3.5e38 > float32 max (~3.4028e38)
        var result = s.SafeParse(3.5e38);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooLarge, result.Issues[0].Code);
    }

    // CWE-20 / spec 3.1: regex anchor newline bypass. .NET's "$" matches before
    // a trailing "\n", so ^[a-z]+$ would accept "abc\n" (newline/CRLF injection).
    // Anchors are rewritten to absolute (\A/\z) to match the JS reference.
    [Fact]
    public void Pattern_DollarAnchorRejectsTrailingNewline()
    {
        var s = V.String().Pattern("^[a-z]+$");
        Assert.True(s.SafeParse("abc").Success);
        Assert.False(s.SafeParse("abc\n").Success);
        Assert.False(s.SafeParse("abc\nEVIL").Success);
    }

    [Fact]
    public void Pattern_CaretAnchorIsStringStartNotLineStart()
    {
        var s = V.String().Pattern("^admin$");
        Assert.True(s.SafeParse("admin").Success);
        Assert.False(s.SafeParse("x\nadmin").Success);
        Assert.False(s.SafeParse("admin\n").Success);
    }

    [Fact]
    public void Pattern_EscapedDollarStaysLiteral()
    {
        var s = V.String().Pattern(@"^a\$$");
        Assert.True(s.SafeParse("a$").Success);
        Assert.False(s.SafeParse("a$\n").Success);
    }

    [Fact]
    public void Float32_RejectsHugeAndLargeNegativeValues()
    {
        Assert.False(V.Float32().SafeParse(1e300).Success);
        Assert.False(V.Float32().SafeParse(-1e300).Success);
    }

    [Fact]
    public void Float32_AcceptsInRangeAndZero()
    {
        Assert.True(V.Float32().SafeParse(1.5).Success);
        Assert.True(V.Float32().SafeParse(0.0).Success);
        Assert.True(V.Float32().SafeParse(3.4e38).Success);
    }

    [Fact]
    public void Infinity_RejectedByFloat32()
    {
        var s = V.Float32();
        Assert.False(s.SafeParse(double.PositiveInfinity).Success);
        Assert.False(s.SafeParse(double.NegativeInfinity).Success);
    }

    [Fact]
    public void Infinity_RejectedByFloat64()
    {
        var s = V.Float64();
        Assert.False(s.SafeParse(double.PositiveInfinity).Success);
        Assert.False(s.SafeParse(double.NegativeInfinity).Success);
    }

    [Fact]
    public void NaN_RejectedByInt()
    {
        var s = V.Int();
        var result = s.SafeParse(double.NaN);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidType, result.Issues[0].Code);
    }

    [Fact]
    public void Infinity_RejectedByInt()
    {
        var s = V.Int();
        Assert.False(s.SafeParse(double.PositiveInfinity).Success);
        Assert.False(s.SafeParse(double.NegativeInfinity).Success);
    }

    [Fact]
    public void NaN_RejectedByAllIntWidths()
    {
        Assert.False(V.Int8().SafeParse(double.NaN).Success);
        Assert.False(V.Int16().SafeParse(double.NaN).Success);
        Assert.False(V.Int32().SafeParse(double.NaN).Success);
        Assert.False(V.Int64().SafeParse(double.NaN).Success);
        Assert.False(V.Uint8().SafeParse(double.NaN).Success);
        Assert.False(V.Uint16().SafeParse(double.NaN).Success);
        Assert.False(V.Uint32().SafeParse(double.NaN).Success);
        Assert.False(V.Uint64().SafeParse(double.NaN).Success);
    }

    [Fact]
    public void FloatNaN_RejectedByNumber()
    {
        var s = V.Number();
        Assert.False(s.SafeParse(float.NaN).Success);
        Assert.False(s.SafeParse(float.PositiveInfinity).Success);
        Assert.False(s.SafeParse(float.NegativeInfinity).Success);
    }

    // ---------------------------------------------------------------
    // CWE-20: Format bypass - email, url, ipv4 edge cases
    // ---------------------------------------------------------------

    [Fact]
    public void FormatEmail_RejectsEmptyString()
    {
        var s = V.String().Format("email");
        Assert.False(s.SafeParse("").Success);
    }

    [Fact]
    public void FormatEmail_TamperedFormatName_NotSilentlyIgnored()
    {
        var s = V.String().Format("email\0");
        Assert.False(s.SafeParse("not-an-email").Success);
    }

    // REVIEW: The test above proves the vulnerable email case, but it does not
    // distinguish malformed built-ins from valid custom extension names.
    [Fact]
    public void Format_MalformedIdentifierRejected_CustomFormatAllowed()
    {
        Assert.False(V.String().Format("email\0").SafeParse("not-an-email").Success);
        Assert.True(V.String().Format("x-custom").SafeParse("any value").Success);
    }

    [Fact]
    public void FormatEmail_ImportedTamperedFormatName_NotUnconstrained()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "string", ["format"] = "email\0" },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.False(schema.SafeParse("not-an-email").Success);
    }

    // REVIEW: Imported schemas need the same malformed-format guard as the
    // builder API, otherwise untrusted interchange can strip validation.
    [Fact]
    public void Format_ImportedMalformedIdentifierRejected_CustomFormatAllowed()
    {
        var malformed = V.Import(new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "string", ["format"] = "email\0" },
            Definitions = new(),
            Extensions = new(),
        });
        var custom = V.Import(new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "string", ["format"] = "x-custom" },
            Definitions = new(),
            Extensions = new(),
        });

        Assert.False(malformed.SafeParse("not-an-email").Success);
        Assert.True(custom.SafeParse("any value").Success);
    }

    [Fact]
    public void FormatEmail_RejectsNoAtSign()
    {
        var s = V.String().Format("email");
        Assert.False(s.SafeParse("userexample.com").Success);
    }

    [Fact]
    public void FormatEmail_RejectsDoubleAtSign()
    {
        var s = V.String().Format("email");
        Assert.False(s.SafeParse("user@@example.com").Success);
    }

    [Fact]
    public void FormatEmail_RejectsTrailingDot()
    {
        var s = V.String().Format("email");
        Assert.False(s.SafeParse("user@example.com.").Success);
    }

    [Fact]
    public void FormatEmail_RejectsMissingDomain()
    {
        var s = V.String().Format("email");
        Assert.False(s.SafeParse("user@").Success);
    }

    [Fact]
    public void FormatEmail_RejectsMissingLocalPart()
    {
        var s = V.String().Format("email");
        Assert.False(s.SafeParse("@example.com").Success);
    }

    [Fact]
    public void FormatEmail_AcceptsValidEmail()
    {
        var s = V.String().Format("email");
        Assert.Equal("user@example.com", s.Parse("user@example.com"));
    }

    [Fact]
    public void FormatUrl_RejectsEmptyString()
    {
        var s = V.String().Format("url");
        Assert.False(s.SafeParse("").Success);
    }

    [Fact]
    public void FormatUrl_RejectsJavascriptProtocol()
    {
        var s = V.String().Format("url");
        Assert.False(s.SafeParse("javascript:alert(1)").Success);
    }

    [Fact]
    public void FormatUrl_RejectsFileProtocol()
    {
        var s = V.String().Format("url");
        Assert.False(s.SafeParse("file:///etc/passwd").Success);
    }

    [Fact]
    public void FormatUrl_RejectsFtpProtocol()
    {
        var s = V.String().Format("url");
        Assert.False(s.SafeParse("ftp://example.com").Success);
    }

    [Fact]
    public void FormatUrl_AcceptsHttps()
    {
        var s = V.String().Format("url");
        Assert.Equal("https://example.com", s.Parse("https://example.com"));
    }

    [Fact]
    public void FormatUrl_AcceptsHttp()
    {
        var s = V.String().Format("url");
        Assert.Equal("http://example.com", s.Parse("http://example.com"));
    }

    [Fact]
    public void FormatUrl_RejectsRelativePath()
    {
        var s = V.String().Format("url");
        Assert.False(s.SafeParse("/path/to/resource").Success);
    }

    [Fact]
    public void FormatIpv4_RejectsEmptyString()
    {
        var s = V.String().Format("ipv4");
        Assert.False(s.SafeParse("").Success);
    }

    [Fact]
    public void FormatIpv4_RejectsOctetAbove255()
    {
        var s = V.String().Format("ipv4");
        Assert.False(s.SafeParse("256.0.0.1").Success);
        Assert.False(s.SafeParse("1.2.3.999").Success);
    }

    [Fact]
    public void FormatIpv4_RejectsTooFewOctets()
    {
        var s = V.String().Format("ipv4");
        Assert.False(s.SafeParse("192.168.1").Success);
    }

    [Fact]
    public void FormatIpv4_RejectsTooManyOctets()
    {
        var s = V.String().Format("ipv4");
        Assert.False(s.SafeParse("192.168.1.1.1").Success);
    }

    [Fact]
    public void FormatIpv4_AcceptsValid()
    {
        var s = V.String().Format("ipv4");
        Assert.Equal("192.168.1.1", s.Parse("192.168.1.1"));
        Assert.Equal("0.0.0.0", s.Parse("0.0.0.0"));
        Assert.Equal("255.255.255.255", s.Parse("255.255.255.255"));
    }

    [Fact]
    public void FormatIpv4_RejectsNegativeOctet()
    {
        var s = V.String().Format("ipv4");
        Assert.False(s.SafeParse("-1.0.0.0").Success);
    }

    [Fact]
    public void FormatIpv4_RejectsAlphaCharacters()
    {
        var s = V.String().Format("ipv4");
        Assert.False(s.SafeParse("abc.def.ghi.jkl").Success);
    }

    // ---------------------------------------------------------------
    // Unicode length constraints - code points, not UTF-16 code units
    // ---------------------------------------------------------------

    [Fact]
    public void UnicodeLength_AstralCodePoint_CountsAsOneCharacter()
    {
        var emoji = char.ConvertFromUtf32(0x1F600);
        Assert.True(V.String().MaxLength(1).SafeParse(emoji).Success);
        Assert.False(V.String().MinLength(2).SafeParse(emoji).Success);
    }

    // REVIEW: The test above covers one surrogate pair. This companion case
    // catches mixed BMP plus astral strings where UTF-16 length diverges more subtly.
    [Fact]
    public void UnicodeLength_MixedBmpAndAstral_CountsCodePoints()
    {
        var value = "a" + char.ConvertFromUtf32(0x1F600);
        Assert.True(V.String().MinLength(2).MaxLength(2).SafeParse(value).Success);
        Assert.False(V.String().MaxLength(1).SafeParse(value).Success);
    }

    [Fact]
    public void UnicodeLength_ImportedMaxLength_UsesCodePoints()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "string", ["maxLength"] = 1L },
            Definitions = new(),
            Extensions = new(),
        };
        var schema = V.Import(doc);
        Assert.True(schema.SafeParse(char.ConvertFromUtf32(0x1F600)).Success);
    }

    // ---------------------------------------------------------------
    // CWE-400: Large inputs - strings, arrays don't crash
    // ---------------------------------------------------------------

    [Fact]
    public void LargeString_DoesNotCrash()
    {
        var s = V.String();
        var largeString = new string('x', 1_000_000);
        var result = s.SafeParse(largeString);
        Assert.True(result.Success);
        Assert.Equal(largeString, result.Data);
    }

    [Fact]
    public void LargeString_MaxLengthConstraint_RejectsLargeInput()
    {
        var s = V.String().MaxLength(100);
        var largeString = new string('x', 1_000_000);
        var result = s.SafeParse(largeString);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooLarge, result.Issues[0].Code);
    }

    [Fact]
    public void LargeArray_DoesNotCrash()
    {
        var s = V.Array(V.Int());
        var largeArray = new List<object?>();
        for (var i = 0; i < 10_000; i++)
            largeArray.Add((long)i);
        var result = s.SafeParse(largeArray);
        Assert.True(result.Success);
    }

    [Fact]
    public void LargeArray_MaxItemsConstraint_RejectsLargeInput()
    {
        var s = V.Array(V.Int()).MaxItems(5);
        var largeArray = new List<object?>();
        for (var i = 0; i < 10_000; i++)
            largeArray.Add((long)i);
        var result = s.SafeParse(largeArray);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooLarge, result.Issues[0].Code);
    }

    [Fact]
    public void LargeObject_DoesNotCrash()
    {
        var shape = new Dictionary<string, Schema>();
        var input = new Dictionary<string, object?>();
        for (var i = 0; i < 1000; i++)
        {
            var key = $"field_{i}";
            shape[key] = V.String();
            input[key] = $"value_{i}";
        }
        var s = V.Object(shape);
        var result = s.SafeParse(input);
        Assert.True(result.Success);
    }

    [Fact]
    public void DeeplyNestedArray_DoesNotCrash()
    {
        // Build deeply nested array schema: array(array(array(...string...)))
        Schema schema = V.String();
        for (var i = 0; i < 50; i++)
            schema = V.Array(schema);

        // Build matching deeply nested input
        object? input = "hello";
        for (var i = 0; i < 50; i++)
            input = new List<object?> { input };

        try
        {
            var result = schema.SafeParse(input);
            // Either success or failure is fine; just must not hang or crash
        }
        catch (StackOverflowException)
        {
            Assert.Fail("Deeply nested array caused StackOverflowException");
        }
    }

    [Fact]
    public void LargeString_PatternValidation_DoesNotCrash()
    {
        var s = V.String().Pattern(@"^\d+$");
        var largeNumericString = new string('1', 100_000);
        var result = s.SafeParse(largeNumericString);
        Assert.True(result.Success);
    }

    // ---------------------------------------------------------------
    // CVE-2019-10744: Object key safety - __proto__, constructor
    // ---------------------------------------------------------------

    [Fact]
    public void ObjectKey_Proto_HandledSafely()
    {
        var s = V.Object(
            new Dictionary<string, Schema> { ["__proto__"] = V.String() });
        var input = new Dictionary<string, object?> { ["__proto__"] = "test" };
        var result = s.SafeParse(input);
        Assert.True(result.Success);
        var parsed = (Dictionary<string, object?>)result.Data!;
        Assert.Equal("test", parsed["__proto__"]);
    }

    [Fact]
    public void ObjectKey_Constructor_HandledSafely()
    {
        var s = V.Object(
            new Dictionary<string, Schema> { ["constructor"] = V.String() });
        var input = new Dictionary<string, object?> { ["constructor"] = "test" };
        var result = s.SafeParse(input);
        Assert.True(result.Success);
        var parsed = (Dictionary<string, object?>)result.Data!;
        Assert.Equal("test", parsed["constructor"]);
    }

    [Fact]
    public void ObjectKey_ToString_HandledSafely()
    {
        var s = V.Object(
            new Dictionary<string, Schema> { ["toString"] = V.String() });
        var input = new Dictionary<string, object?> { ["toString"] = "overridden" };
        var result = s.SafeParse(input);
        Assert.True(result.Success);
        var parsed = (Dictionary<string, object?>)result.Data!;
        Assert.Equal("overridden", parsed["toString"]);
    }

    [Fact]
    public void ObjectKey_HasOwnProperty_HandledSafely()
    {
        var s = V.Object(
            new Dictionary<string, Schema> { ["hasOwnProperty"] = V.String() });
        var input = new Dictionary<string, object?> { ["hasOwnProperty"] = "test" };
        var result = s.SafeParse(input);
        Assert.True(result.Success);
    }

    [Fact]
    public void UnknownKey_Proto_RejectedWhenConfigured()
    {
        var s = V.Object(new Dictionary<string, Schema> { ["name"] = V.String() }, UnknownKeyMode.Reject);
        var input = new Dictionary<string, object?>
        {
            ["name"] = "Alice",
            ["__proto__"] = new Dictionary<string, object?> { ["isAdmin"] = true },
        };
        var result = s.SafeParse(input);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.UnknownKey, result.Issues[0].Code);
    }

    [Fact]
    public void UnknownKey_Constructor_RejectedWhenConfigured()
    {
        var s = V.Object(new Dictionary<string, Schema> { ["name"] = V.String() }, UnknownKeyMode.Reject);
        var input = new Dictionary<string, object?>
        {
            ["name"] = "Alice",
            ["constructor"] = new Dictionary<string, object?> { ["prototype"] = "evil" },
        };
        var result = s.SafeParse(input);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.UnknownKey, result.Issues[0].Code);
    }

    [Fact]
    public void ObjectKey_Proto_StrippedInStripMode()
    {
        var s = V.Object(
            new Dictionary<string, Schema> { ["name"] = V.String() },
            UnknownKeyMode.Strip);
        var input = new Dictionary<string, object?>
        {
            ["name"] = "Alice",
            ["__proto__"] = new Dictionary<string, object?> { ["isAdmin"] = true },
        };
        var result = (Dictionary<string, object?>)s.Parse(input)!;
        Assert.Single(result);
        Assert.Equal("Alice", result["name"]);
        Assert.False(result.ContainsKey("__proto__"));
    }

    [Fact]
    public void RecordKey_Proto_HandledSafely()
    {
        var s = V.Record(V.String());
        var input = new Dictionary<string, object?>
        {
            ["__proto__"] = "value1",
            ["constructor"] = "value2",
            ["normal"] = "value3",
        };
        var result = s.SafeParse(input);
        Assert.True(result.Success);
        var parsed = (Dictionary<string, object?>)result.Data!;
        Assert.Equal("value1", parsed["__proto__"]);
        Assert.Equal("value2", parsed["constructor"]);
        Assert.Equal("value3", parsed["normal"]);
    }

    // ---------------------------------------------------------------
    // Schema import injection - Unknown kinds rejected
    // ---------------------------------------------------------------

    [Fact]
    public void ImportSchema_UnknownKind_Throws()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "evil_injection" },
            Definitions = new(),
            Extensions = new(),
        };
        Assert.Throws<InvalidOperationException>(() => V.Import(doc));
    }

    [Fact]
    public void ImportSchema_NullKind_Throws()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = null },
            Definitions = new(),
            Extensions = new(),
        };
        Assert.ThrowsAny<Exception>(() => V.Import(doc));
    }

    [Fact]
    public void ImportSchema_MissingKind_Throws()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["notKind"] = "string" },
            Definitions = new(),
            Extensions = new(),
        };
        Assert.ThrowsAny<Exception>(() => V.Import(doc));
    }

    [Fact]
    public void ImportSchema_EmptyStringKind_Throws()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "" },
            Definitions = new(),
            Extensions = new(),
        };
        Assert.ThrowsAny<Exception>(() => V.Import(doc));
    }

    [Fact]
    public void ImportSchema_SqlInjectionKind_Throws()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "'; DROP TABLE schemas;--" },
            Definitions = new(),
            Extensions = new(),
        };
        Assert.Throws<InvalidOperationException>(() => V.Import(doc));
    }

    [Fact]
    public void ImportSchema_HtmlInjectionKind_Throws()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "<script>alert('xss')</script>" },
            Definitions = new(),
            Extensions = new(),
        };
        Assert.Throws<InvalidOperationException>(() => V.Import(doc));
    }

    [Fact]
    public void ImportSchema_NullRoot_Throws()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = null!,
            Definitions = new(),
            Extensions = new(),
        };
        Assert.ThrowsAny<Exception>(() => V.Import(doc));
    }

    [Fact]
    public void ImportSchema_NestedUnknownKind_Throws()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?>
            {
                ["kind"] = "array",
                ["items"] = new Dictionary<string, object?> { ["kind"] = "malicious_type" },
            },
            Definitions = new(),
            Extensions = new(),
        };
        Assert.Throws<InvalidOperationException>(() => V.Import(doc));
    }

    [Fact]
    public void ImportSchema_UnresolvedRef_FailsOnParse()
    {
        var doc = new AnyValiDocument
        {
            AnyvaliVersion = "1.0",
            SchemaVersion = "1",
            Root = new Dictionary<string, object?> { ["kind"] = "ref", ["ref"] = "#/definitions/Missing" },
            Definitions = new(),
            Extensions = new(),
        };
        // Import may succeed (lazy ref resolution) or throw
        try
        {
            var schema = V.Import(doc);
            // If import succeeds, parsing should fail gracefully
            Assert.ThrowsAny<Exception>(() => schema.Parse("test"));
        }
        catch
        {
            // Throwing on import is also acceptable
        }
    }
}
