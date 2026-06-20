package com.anyvali;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

/**
 * Internal context for collecting validation issues during parsing.
 */
public class ValidationContext {
    private final List<Object> path;
    private final List<ValidationIssue> issues;
    private final Map<String, Schema<?>> definitions;
    private final Set<String> activeRefs;
    private UnknownKeyMode inheritedUnknownKeys;

    public ValidationContext() {
        this(new ArrayList<>(), new ArrayList<>(), new HashMap<>(), new HashSet<>());
    }

    public ValidationContext(List<Object> path, List<ValidationIssue> issues,
                             Map<String, Schema<?>> definitions) {
        this(path, issues, definitions, new HashSet<>());
    }

    private ValidationContext(List<Object> path, List<ValidationIssue> issues,
                             Map<String, Schema<?>> definitions, Set<String> activeRefs) {
        this.path = path;
        this.issues = issues;
        this.definitions = definitions;
        this.activeRefs = activeRefs;
        this.inheritedUnknownKeys = null;
    }

    public void addIssue(String code, String message, Object expected, Object received) {
        addIssue(code, message, expected, received, null);
    }

    public void addIssue(String code, String message, Object expected, Object received,
                         Map<String, Object> meta) {
        issues.add(new ValidationIssue(code, message, List.copyOf(path),
                expected, received, meta));
    }

    public void addIssue(String code, String message) {
        addIssue(code, message, null, null);
    }

    /**
     * Create a child context with an additional path element.
     * Shares the same issues list.
     */
    public ValidationContext child(Object key) {
        var childPath = new ArrayList<>(this.path);
        childPath.add(key);
        var child = new ValidationContext(childPath, this.issues, this.definitions, this.activeRefs);
        child.setInheritedUnknownKeys(this.inheritedUnknownKeys);
        return child;
    }

    public List<Object> getPath() {
        return path;
    }

    public List<ValidationIssue> getIssues() {
        return issues;
    }

    public boolean hasIssues() {
        return !issues.isEmpty();
    }

    public int issueCount() {
        return issues.size();
    }

    public Map<String, Schema<?>> getDefinitions() {
        return definitions;
    }

    public boolean enterRef(String ref) {
        return activeRefs.add(ref);
    }

    public void exitRef(String ref) {
        activeRefs.remove(ref);
    }

    public UnknownKeyMode getInheritedUnknownKeys() {
        return inheritedUnknownKeys;
    }

    public void setInheritedUnknownKeys(UnknownKeyMode inheritedUnknownKeys) {
        this.inheritedUnknownKeys = inheritedUnknownKeys;
    }
}
