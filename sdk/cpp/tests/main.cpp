#include "test_framework.hpp"

int main(int argc, char** argv) {
    std::string filter = "";
    if (argc > 1) filter = argv[1];

    int total = 0;
    int passed = 0;
    int failed = 0;
    int skipped = 0;

    for (auto& tc : test_registry()) {
        if (!filter.empty() && tc.name.find(filter) == std::string::npos) {
            skipped++;
            continue;
        }
        total++;
        current_test() = tc.name;
        int prev_fails = fail_count();
        try {
            tc.func();
            if (fail_count() == prev_fails) {
                passed++;
                std::cout << "  PASS: " << tc.name << std::endl;
            } else {
                failed++;
                std::cerr << "  FAIL: " << tc.name << std::endl;
            }
        } catch (const std::exception& e) {
            failed++;
            std::cerr << "  FAIL: " << tc.name << " (" << e.what() << ")" << std::endl;
        }
    }

    std::cout << "\n========================================" << std::endl;
    std::cout << "Results: " << passed << " passed, " << failed << " failed, "
              << skipped << " skipped, " << total << " total" << std::endl;

    return failed > 0 ? 1 : 0;
}
