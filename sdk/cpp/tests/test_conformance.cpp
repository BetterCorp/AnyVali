#include "test_framework.hpp"
#include <anyvali/anyvali.hpp>
#include <fstream>
#include <filesystem>

using namespace anyvali;
using json = nlohmann::json;
namespace fs = std::filesystem;

static void run_corpus_file(const fs::path& filepath) {
    std::ifstream ifs(filepath);
    if (!ifs.is_open()) {
        std::cerr << "  WARNING: Cannot open corpus file: " << filepath << std::endl;
        return;
    }

    json suite;
    try {
        ifs >> suite;
    } catch (const std::exception& e) {
        std::cerr << "  WARNING: Cannot parse corpus file: " << filepath << ": " << e.what() << std::endl;
        return;
    }

    std::string suite_name = suite.value("suite", filepath.filename().string());

    for (const auto& tc : suite["cases"]) {
        std::string desc = tc["description"].get<std::string>();
        bool expected_valid = tc["valid"].get<bool>();

        try {
            auto doc = import_schema(tc["schema"]);
            auto result = parse_document(doc, tc["input"]);

            if (result.success != expected_valid) {
                std::cerr << "  CORPUS FAIL [" << suite_name << "]: " << desc << std::endl;
                std::cerr << "    Expected valid=" << expected_valid
                          << ", got valid=" << result.success << std::endl;
                if (!result.success) {
                    for (const auto& issue : result.issues) {
                        std::cerr << "      issue: " << issue.code
                                  << " expected=" << issue.expected
                                  << " received=" << issue.received << std::endl;
                    }
                }
                fail_count()++;
                continue;
            }

            // Check output for valid cases
            if (expected_valid && tc.contains("output") && !tc["output"].is_null()) {
                if (result.value != tc["output"]) {
                    std::cerr << "  CORPUS FAIL [" << suite_name << "]: " << desc << std::endl;
                    std::cerr << "    Output mismatch:" << std::endl;
                    std::cerr << "      expected: " << tc["output"].dump() << std::endl;
                    std::cerr << "      actual:   " << result.value.dump() << std::endl;
                    fail_count()++;
                    continue;
                }
            }

            // Check issues for invalid cases
            if (!expected_valid && tc.contains("issues") && tc["issues"].is_array()) {
                const auto& expected_issues = tc["issues"];
                if (result.issues.size() != expected_issues.size()) {
                    std::cerr << "  CORPUS FAIL [" << suite_name << "]: " << desc << std::endl;
                    std::cerr << "    Issue count mismatch: expected "
                              << expected_issues.size() << ", got " << result.issues.size() << std::endl;
                    for (const auto& issue : result.issues) {
                        json path_json = json::array();
                        for (const auto& seg : issue.path) {
                            if (std::holds_alternative<std::string>(seg))
                                path_json.push_back(std::get<std::string>(seg));
                            else
                                path_json.push_back(std::get<int>(seg));
                        }
                        std::cerr << "      got: code=" << issue.code
                                  << " path=" << path_json.dump()
                                  << " expected=" << issue.expected
                                  << " received=" << issue.received << std::endl;
                    }
                    fail_count()++;
                    continue;
                }

                // Check that each expected issue exists in actual issues (order-independent)
                bool issues_match = true;
                for (size_t i = 0; i < expected_issues.size(); i++) {
                    const auto& ei = expected_issues[i];
                    bool found = false;

                    for (const auto& ai : result.issues) {
                        // Check code
                        if (ei.contains("code") && ai.code != ei["code"].get<std::string>()) continue;

                        // Check path
                        if (ei.contains("path")) {
                            json actual_path = json::array();
                            for (const auto& seg : ai.path) {
                                if (std::holds_alternative<std::string>(seg))
                                    actual_path.push_back(std::get<std::string>(seg));
                                else
                                    actual_path.push_back(std::get<int>(seg));
                            }
                            if (actual_path != ei["path"]) continue;
                        }

                        // Check expected field
                        if (ei.contains("expected") && ai.expected != ei["expected"].get<std::string>()) continue;

                        // Check received field
                        if (ei.contains("received") && ai.received != ei["received"].get<std::string>()) continue;

                        found = true;
                        break;
                    }

                    if (!found) {
                        std::cerr << "  CORPUS FAIL [" << suite_name << "]: " << desc << std::endl;
                        std::cerr << "    Expected issue not found: " << ei.dump() << std::endl;
                        issues_match = false;
                        break;
                    }
                }

                if (!issues_match) {
                    fail_count()++;
                    continue;
                }
            }

        } catch (const std::exception& e) {
            std::cerr << "  CORPUS FAIL [" << suite_name << "]: " << desc
                      << " (exception: " << e.what() << ")" << std::endl;
            fail_count()++;
        }
    }
}

TEST("conformance: run spec corpus") {
    std::string corpus_dir = ANYVALI_CORPUS_DIR;
    fs::path corpus_path(corpus_dir);

    if (!fs::exists(corpus_path)) {
        std::cerr << "  WARNING: Corpus directory not found: " << corpus_dir << std::endl;
        return;
    }

    int file_count = 0;
    int prev_fails = fail_count();

    for (const auto& entry : fs::recursive_directory_iterator(corpus_path)) {
        if (entry.is_regular_file() && entry.path().extension() == ".json") {
            std::cout << "  Running corpus: " << entry.path().filename().string() << std::endl;
            run_corpus_file(entry.path());
            file_count++;
        }
    }

    int new_fails = fail_count() - prev_fails;
    std::cout << "  Corpus: " << file_count << " files processed, "
              << new_fails << " failures" << std::endl;

    // Fail the test if any corpus cases failed
    ASSERT(new_fails == 0);
}
