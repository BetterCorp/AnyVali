cmake_minimum_required(VERSION 3.14)
project(anyvali-cpp VERSION 0.0.1 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Fetch nlohmann/json
include(FetchContent)
FetchContent_Declare(
    json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG        v3.11.3
    GIT_SHALLOW    TRUE
)
FetchContent_MakeAvailable(json)

# Header-only library
add_library(anyvali INTERFACE)
target_include_directories(anyvali INTERFACE
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>
)
target_link_libraries(anyvali INTERFACE nlohmann_json::nlohmann_json)

# Optional: compile a translation unit for any non-inline implementations
add_library(anyvali_impl STATIC src/anyvali.cpp)
target_link_libraries(anyvali_impl PUBLIC anyvali)

# Tests
option(ANYVALI_BUILD_TESTS "Build tests" ON)
if(ANYVALI_BUILD_TESTS)
    enable_testing()

    add_executable(anyvali_tests
        tests/main.cpp
        tests/test_schemas.cpp
        tests/test_coercion.cpp
        tests/test_format.cpp
        tests/test_interchange.cpp
        tests/test_conformance.cpp
    )
    target_link_libraries(anyvali_tests PRIVATE anyvali)

    # Pass the path to the spec corpus for conformance tests
    target_compile_definitions(anyvali_tests PRIVATE
        ANYVALI_CORPUS_DIR="${CMAKE_CURRENT_SOURCE_DIR}/../../spec/corpus"
    )

    add_test(NAME anyvali_tests COMMAND anyvali_tests)
endif()
