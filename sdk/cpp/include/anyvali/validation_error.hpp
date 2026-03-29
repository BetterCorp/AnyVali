#pragma once

#include <stdexcept>
#include <vector>
#include "validation_issue.hpp"

namespace anyvali {

class ValidationError : public std::runtime_error {
public:
    std::vector<ValidationIssue> issues;

    explicit ValidationError(std::vector<ValidationIssue> issues_)
        : std::runtime_error(build_message(issues_)), issues(std::move(issues_)) {}

private:
    static std::string build_message(const std::vector<ValidationIssue>& issues) {
        std::string msg = "Validation failed with " + std::to_string(issues.size()) + " issue(s)";
        for (const auto& issue : issues) {
            msg += "\n  - [" + issue.code + "] " + issue.message;
            if (!issue.expected.empty()) {
                msg += " (expected: " + issue.expected + ", received: " + issue.received + ")";
            }
        }
        return msg;
    }
};

} // namespace anyvali
