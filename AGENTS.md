# Agent Development Guide

This guide defines a multi-agent workflow system for autonomous software development. Customize the commands and conventions sections for your specific project.

## Commands

Configure these commands for your project's build system:

- **Build**: `[BUILD_COMMAND]` - Full project build with production configuration
- **Restore**: `[RESTORE_COMMAND]` - Install/restore project dependencies
- **Test**: `[TEST_COMMAND]` - Run all test suites with CI-compatible output
- **Test Single**: `[TEST_SINGLE_COMMAND]` - Run specific test case or test class
- **Run**: `[RUN_COMMAND]` - Start application in development mode
- **Publish**: `[PUBLISH_COMMAND]` - Package application for deployment

**Examples:**
- Node.js: `npm run build`, `npm test`, `npm start`
- Python: `python -m build`, `pytest`, `python -m myapp`
- Go: `go build ./...`, `go test ./...`, `go run cmd/server/main.go`
- Rust: `cargo build --release`, `cargo test`, `cargo run`
- .NET: `dotnet build`, `dotnet test`, `dotnet run`

## Code Style

Configure these conventions for your project's programming language:

- **Framework**: `[FRAMEWORK_NAME_AND_VERSION]` - e.g., "Express.js 4.x", "Django 5.0", "Spring Boot 3.2"
- **Naming**: `[NAMING_CONVENTIONS]` - e.g., "camelCase for variables, PascalCase for classes"
- **Formatting**: `[FORMATTING_RULES]` - e.g., "2-space indent, 80 char line limit, trailing commas"
- **Types**: `[TYPE_CONVENTIONS]` - e.g., "Use strict typing, avoid any/unknown, prefer interfaces"
- **Patterns**: `[ARCHITECTURE_PATTERNS]` - e.g., "Dependency injection, repository pattern, DTOs for API"
- **Async**: `[ASYNC_CONVENTIONS]` - e.g., "Use async/await, handle promise rejections, timeout long operations"
- **Error Handling**: `[ERROR_PATTERNS]` - e.g., "Custom error classes, structured logging, graceful degradation"

**Language-Specific Examples:**
- **JavaScript/TypeScript**: camelCase variables, PascalCase classes, 2-space indent, ESLint/Prettier, async/await
- **Python**: snake_case, 4-space indent, type hints (PEP 484), Black formatter, pytest conventions
- **Go**: camelCase (unexported), PascalCase (exported), gofmt, error wrapping, defer cleanup
- **Rust**: snake_case, 4-space indent, rustfmt, Result/Option types, ownership patterns
- **Java/C#**: PascalCase classes/methods, camelCase variables, 4-space indent, LINQ/Stream patterns

## Testing

Configure testing conventions for your project:

- **Framework**: `[TEST_FRAMEWORK]` - e.g., "Jest", "pytest", "JUnit", "xUnit", "Go testing"
- **Patterns**: `[TEST_PATTERNS]` - e.g., "describe/it blocks", "Given-When-Then", "Arrange-Act-Assert"
- **Naming**: `[TEST_NAMING]` - e.g., "test_method_scenario_expected", "should_do_x_when_y"
- **Single Test**: `[SINGLE_TEST_COMMAND]` - Command to run specific test case

**Framework-Specific Examples:**
- **Jest/Vitest**: `describe()` blocks, `test()` or `it()`, `npm test -- -t "test name"`
- **pytest**: `test_function_scenario()`, fixtures, `pytest -k test_name`
- **Go**: `TestFunctionScenario()`, table-driven tests, `go test -run TestName`
- **JUnit/xUnit**: `@Test` annotations, `MethodName_Scenario_Expected()`, test filters
- **RSpec**: `describe` blocks, `it "should do x"`, `rspec spec/path/to/spec.rb:42`
