# Sample AGENTS.md file

## Dev environment tips
- Use `make build` to build the Flagger binary for your platform
- Use `make test` to run all unit tests
- Use `make test-coverage` to run tests with coverage report
- Use `make fmt` to format Go code
- Use `make vet` for static analysis
- Use `make build-image` to build Docker image
- Use `make loadtester-build` to build the load tester image
- For development, run Flagger locally with `go run cmd/flagger/main.go` after setting up kubeconfig

## Testing instructions
- Find the CI plan in the .github/workflows folder
- Run `make test` to execute all tests
- Run `go test ./pkg/...` to run tests for all packages
- To focus on one package, run `go test ./pkg/controller/` (for example)
- Use `go test -v` for verbose output
- Fix any test or type errors until the whole suite is green
- After making changes, run `make fmt test-codegen` to ensure code is properly formatted
- Add or update tests for the code you change, even if nobody asked

## PR instructions
- Title format: [<component>] <Title> (e.g. [controller] Add manual step feature)
- Always run `make fmt test-codegen` and `make test` before committing
- Update documentation in README.md or /docs if needed
- Include unit tests for new functionality
- Follow existing code style and conventions
- Ensure all CI checks pass before merging