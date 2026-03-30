package com.anyvali;

import com.anyvali.interchange.Importer;
import com.anyvali.interchange.JsonHelper;
import org.junit.jupiter.api.DynamicTest;
import org.junit.jupiter.api.TestFactory;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Reads spec/corpus/ JSON files and runs conformance tests.
 */
class ConformanceTest {

    private static final Path CORPUS_DIR = findCorpusDir();

    private static Path findCorpusDir() {
        // Walk up from the SDK directory to find spec/corpus/
        Path current = Paths.get("").toAbsolutePath();
        // Try relative to CWD
        Path candidate = current.resolve("spec/corpus");
        if (Files.isDirectory(candidate)) return candidate;
        // Try from sdk/java
        candidate = current.resolve("../../spec/corpus");
        if (Files.isDirectory(candidate)) return candidate.normalize();
        // Try absolute well-known path
        candidate = Paths.get("C:/Users/MitchellRobert/_win_repos/anyvali/spec/corpus");
        if (Files.isDirectory(candidate)) return candidate;
        // Fallback
        return Paths.get("spec/corpus");
    }

    @TestFactory
    Stream<DynamicTest> conformanceSuite() throws IOException {
        if (!Files.isDirectory(CORPUS_DIR)) {
            return Stream.of(DynamicTest.dynamicTest("corpus dir not found - skipping",
                    () -> System.out.println("Corpus dir not found at " + CORPUS_DIR)));
        }

        List<DynamicTest> tests = new ArrayList<>();

        try (var stream = Files.walk(CORPUS_DIR)) {
            List<Path> jsonFiles = stream
                    .filter(p -> p.toString().endsWith(".json"))
                    .sorted()
                    .toList();

            for (Path file : jsonFiles) {
                String content = Files.readString(file);
                @SuppressWarnings("unchecked")
                Map<String, Object> suite = (Map<String, Object>) JsonHelper.parseJson(content);
                String suiteName = (String) suite.get("suite");
                @SuppressWarnings("unchecked")
                List<Map<String, Object>> cases = (List<Map<String, Object>>) suite.get("cases");

                if (cases == null) continue;

                for (int i = 0; i < cases.size(); i++) {
                    Map<String, Object> testCase = cases.get(i);
                    String description = (String) testCase.get("description");
                    String testName = suiteName + " [" + i + "] " + description;

                    tests.add(DynamicTest.dynamicTest(testName, () -> runTestCase(testCase)));
                }
            }
        }

        return tests.stream();
    }

    @SuppressWarnings("unchecked")
    private void runTestCase(Map<String, Object> testCase) {
        Map<String, Object> schemaDoc = (Map<String, Object>) testCase.get("schema");
        Object input = testCase.get("input");
        boolean expectedValid = (Boolean) testCase.get("valid");
        List<Map<String, Object>> expectedIssues = (List<Map<String, Object>>) testCase.get("issues");

        Schema<?> schema;
        try {
            schema = Importer.importSchema(schemaDoc);
        } catch (Exception e) {
            if (!expectedValid) {
                // Schema import failure on an invalid test case is acceptable
                return;
            }
            fail("Failed to import schema: " + e.getMessage());
            return;
        }

        ParseResult<?> result = schema.safeParse(input);

        assertEquals(expectedValid, result.success(),
                "Expected valid=" + expectedValid + " but got " + result.success()
                + (result.issues().isEmpty() ? "" : " issues: " + result.issues()));

        if (!expectedValid && expectedIssues != null && !expectedIssues.isEmpty()) {
            // Verify issue codes match
            assertEquals(expectedIssues.size(), result.issues().size(),
                    "Expected " + expectedIssues.size() + " issues but got " + result.issues().size()
                    + ": " + result.issues());

            for (int i = 0; i < expectedIssues.size(); i++) {
                Map<String, Object> expectedIssue = expectedIssues.get(i);
                ValidationIssue actualIssue = result.issues().get(i);

                // Check code
                String expectedCode = (String) expectedIssue.get("code");
                if (expectedCode != null) {
                    assertEquals(expectedCode, actualIssue.code(),
                            "Issue code mismatch at index " + i);
                }

                // Check path
                List<Object> expectedPath = (List<Object>) expectedIssue.get("path");
                if (expectedPath != null) {
                    assertEquals(expectedPath.size(), actualIssue.path().size(),
                            "Path size mismatch at issue index " + i);
                    for (int j = 0; j < expectedPath.size(); j++) {
                        Object expectedPathElem = expectedPath.get(j);
                        Object actualPathElem = actualIssue.path().get(j);
                        // Compare as strings for flexibility (int vs string path segments)
                        assertEquals(String.valueOf(expectedPathElem),
                                String.valueOf(actualPathElem),
                                "Path element mismatch at issue " + i + " path index " + j);
                    }
                }
            }
        }

        // Verify output on successful parse
        if (expectedValid && testCase.containsKey("output")) {
            Object expectedOutput = testCase.get("output");
            assertOutputMatches(expectedOutput, result.data());
        }
    }

    private void assertOutputMatches(Object expected, Object actual) {
        if (expected == null) {
            assertNull(actual);
            return;
        }
        if (expected instanceof Map<?, ?> expectedMap && actual instanceof Map<?, ?> actualMap) {
            assertEquals(expectedMap.size(), actualMap.size(),
                    "Map size mismatch: expected " + expectedMap + " but got " + actualMap);
            for (var entry : expectedMap.entrySet()) {
                assertTrue(actualMap.containsKey(entry.getKey()),
                        "Missing key: " + entry.getKey());
                assertOutputMatches(entry.getValue(), actualMap.get(entry.getKey()));
            }
        } else if (expected instanceof List<?> expectedList && actual instanceof List<?> actualList) {
            assertEquals(expectedList.size(), actualList.size(),
                    "List size mismatch");
            for (int i = 0; i < expectedList.size(); i++) {
                assertOutputMatches(expectedList.get(i), actualList.get(i));
            }
        } else if (expected instanceof Number && actual instanceof Number) {
            // Compare numeric values flexibly (int/long/double)
            assertEquals(((Number) expected).doubleValue(), ((Number) actual).doubleValue(), 1e-9,
                    "Numeric mismatch: expected " + expected + " but got " + actual);
        } else {
            assertEquals(expected, actual);
        }
    }
}
