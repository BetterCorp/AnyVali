#pragma once

#include <vector>
#include <string>
#include <map>
#include <memory>
#include "types.hpp"
#include "validation_issue.hpp"

namespace anyvali {

// Forward declaration
class Schema;

struct ValidationContext {
    Path path;
    std::vector<ValidationIssue> issues;
    const std::map<std::string, std::shared_ptr<Schema>>* definitions = nullptr;

    void push_path(const std::string& key) {
        path.push_back(key);
    }

    void push_path(int index) {
        path.push_back(index);
    }

    void pop_path() {
        if (!path.empty()) path.pop_back();
    }

    void add_issue(const std::string& code, const std::string& expected,
                   const std::string& received, const std::string& message = "") {
        issues.emplace_back(code, path, expected, received, message);
    }

    bool has_issues() const {
        return !issues.empty();
    }
};

} // namespace anyvali
