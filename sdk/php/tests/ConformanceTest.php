<?php

declare(strict_types=1);

namespace AnyVali\Tests;

use AnyVali\AnyValiDocument;
use AnyVali\Interchange\Importer;
use AnyVali\ValidationContext;
use PHPUnit\Framework\TestCase;

/**
 * Runs all conformance test cases from the shared spec/corpus/ directory.
 *
 * Each JSON file in the corpus contains a suite of test cases with:
 * - schema: an AnyValiDocument
 * - input: the input value
 * - valid: expected success/failure
 * - output: expected parsed output (null if invalid)
 * - issues: expected issue codes and paths
 */
final class ConformanceTest extends TestCase
{
    private static function corpusDir(): string
    {
        // Navigate from sdk/php/tests/ up to repo root, then into spec/corpus/
        return realpath(__DIR__ . '/../../../spec/corpus');
    }

    /**
     * @return iterable<string, array{0: array<string, mixed>}>
     */
    public static function corpusProvider(): iterable
    {
        $corpusDir = self::corpusDir();
        if ($corpusDir === false || !is_dir($corpusDir)) {
            return;
        }

        $iterator = new \RecursiveIteratorIterator(
            new \RecursiveDirectoryIterator($corpusDir, \RecursiveDirectoryIterator::SKIP_DOTS)
        );

        foreach ($iterator as $file) {
            if ($file->getExtension() !== 'json') {
                continue;
            }

            $content = file_get_contents($file->getPathname());
            $suite = json_decode($content, true, 512, JSON_THROW_ON_ERROR);
            // Also decode preserving objects (stdClass) so we can distinguish {} from []
            $rawSuite = json_decode($content, false, 512, JSON_THROW_ON_ERROR);
            $suiteName = $suite['suite'] ?? $file->getBasename('.json');

            foreach ($suite['cases'] as $i => $testCase) {
                // Use the object-preserved decode for the input to distinguish {} from []
                $testCase['input'] = self::convertJsonValue($rawSuite->cases[$i]->input);
                $label = "{$suiteName}: {$testCase['description']}";
                yield $label => [$testCase];
            }
        }
    }

    /**
     * Convert a JSON-decoded value (with stdClass for objects) to PHP types
     * while preserving the object/array distinction.
     * Non-empty stdClass → associative array, empty stdClass stays as stdClass
     * so that getTypeName() can report "object" instead of "array".
     */
    private static function convertJsonValue(mixed $value): mixed
    {
        if ($value instanceof \stdClass) {
            $props = (array)$value;
            if (empty($props)) {
                // Keep empty stdClass so Schema::getTypeName returns "object"
                return $value;
            }
            $result = [];
            foreach ($props as $k => $v) {
                $result[$k] = self::convertJsonValue($v);
            }
            return $result;
        }
        if (is_array($value)) {
            return array_map([self::class, 'convertJsonValue'], $value);
        }
        return $value;
    }

    /**
     * @dataProvider corpusProvider
     * @param array<string, mixed> $testCase
     */
    public function testConformance(array $testCase): void
    {
        $schemaDoc = AnyValiDocument::fromArray($testCase['schema']);
        $schema = Importer::import($schemaDoc);

        // Create a validation context with definitions from the document
        $ctx = new ValidationContext(
            path: [],
            definitions: $schemaDoc->definitions,
        );

        $result = $schema->safeParse($testCase['input'], $ctx);

        $this->assertSame(
            $testCase['valid'],
            $result->success,
            sprintf(
                "Expected %s but got %s for input: %s\nIssues: %s",
                $testCase['valid'] ? 'valid' : 'invalid',
                $result->success ? 'valid' : 'invalid',
                json_encode($testCase['input']),
                json_encode(array_map(fn($i) => $i->toArray(), $result->issues)),
            ),
        );

        if ($testCase['valid']) {
            // Check output matches
            $this->assertEquals(
                $testCase['output'],
                $result->value,
                sprintf(
                    "Output mismatch for input: %s\nExpected: %s\nGot: %s",
                    json_encode($testCase['input']),
                    json_encode($testCase['output']),
                    json_encode($result->value),
                ),
            );
        }

        if (!$testCase['valid'] && !empty($testCase['issues'])) {
            // Check issue codes match
            $expectedCodes = array_map(fn($i) => $i['code'], $testCase['issues']);
            $actualCodes = array_map(fn($i) => $i->code, $result->issues);

            $this->assertSame(
                $expectedCodes,
                $actualCodes,
                sprintf(
                    "Issue codes mismatch.\nExpected: %s\nGot: %s",
                    json_encode($expectedCodes),
                    json_encode($actualCodes),
                ),
            );

            // Check issue paths match
            foreach ($testCase['issues'] as $idx => $expectedIssue) {
                if (isset($result->issues[$idx])) {
                    $this->assertEquals(
                        $expectedIssue['path'],
                        $result->issues[$idx]->path,
                        sprintf(
                            "Issue path mismatch at index %d.\nExpected: %s\nGot: %s",
                            $idx,
                            json_encode($expectedIssue['path']),
                            json_encode($result->issues[$idx]->path),
                        ),
                    );
                }
            }

            // Check expected/received where specified
            foreach ($testCase['issues'] as $idx => $expectedIssue) {
                if (isset($result->issues[$idx])) {
                    if (isset($expectedIssue['expected'])) {
                        $this->assertSame(
                            $expectedIssue['expected'],
                            $result->issues[$idx]->expected,
                            sprintf(
                                "Issue expected mismatch at index %d.\nExpected: %s\nGot: %s",
                                $idx,
                                $expectedIssue['expected'],
                                $result->issues[$idx]->expected ?? 'null',
                            ),
                        );
                    }
                    if (isset($expectedIssue['received'])) {
                        $this->assertSame(
                            $expectedIssue['received'],
                            $result->issues[$idx]->received,
                            sprintf(
                                "Issue received mismatch at index %d.\nExpected: %s\nGot: %s",
                                $idx,
                                $expectedIssue['received'],
                                $result->issues[$idx]->received ?? 'null',
                            ),
                        );
                    }
                }
            }
        }
    }
}
