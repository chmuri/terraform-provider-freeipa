# Contributing

Thank you for contributing to `terraform-provider-freeipa`!

## Development Setup

### Prerequisites

- **Go** 1.26+
- **Docker** and **Docker Compose** (for integration tests)
- **Terraform CLI** 1.x
- **`tfplugindocs`** (for documentation generation)
- **`goreleaser`** (for releases)

```bash
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
```

### Clone & Build

```bash
git clone git@github.com:chmuri/terraform-provider-freeipa.git
cd terraform-provider-freeipa
go mod download
make build
```

### Local Development

Use `test_local/` to test changes against a local FreeIPA server:

```bash
cd test_local
TF_CLI_CONFIG_FILE=terraformrc terraform plan
```

## Testing

### Unit Tests

```bash
make test-unit
```

All 28 unit tests validate resource and data source schemas. They run in <1 second.

### Acceptance Tests

Requires Docker and a running FreeIPA container:

```bash
make test-acc                              # all 31 tests
make test-acc TESTARGS='-run TestAcc_User' # specific test
```

The test suite:
- Starts a clean FreeIPA container from a Docker volume snapshot
- Runs CRUD tests, option variants, membership scenarios
- Tests all 19 resources and 6 data sources
- Cleans up via sweepers

### Adding Tests

1. **Schema test**: Add to `provider/provider_test.go`
2. **Acceptance test**: Add to `provider/resource_acc_test.go`

Follow the existing patterns:

```go
func TestAcc_MyResource_CRUD(t *testing.T) {
    skipIfNotAcc(t)
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {Config: accProviderConfig() + `...`, Check: ...},
            {Config: accProviderConfig() + `...`, Check: ...},
            {ResourceName: "...", ImportState: true, ImportStateVerify: true},
        },
    })
}
```

## Code Style

- Run `gofmt -w .` before committing
- Run `go vet ./...` to check for issues
- Follow existing code patterns in the provider package
- JSON tags must match FreeIPA API response fields

## Documentation

Documentation is generated with `tfplugindocs`:

```bash
tfplugindocs generate --provider-name freeipa
```

Schema descriptions in Go code automatically generate `docs/` markdown files.

## Release Process

1. Bump version in `main.go`
2. Update `README.md` and `test_local/main.tf`
3. Update `CHANGELOG.md`
4. Run full test suite: `make test-all`
5. Commit, push, and tag: `git tag v1.0.x && git push origin v1.0.x`
6. GitHub Actions builds and publishes the release via GoReleaser

## Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Add tests for your changes
4. Run `make test-unit` and `make test-acc`
5. Run `gofmt -w . && go vet ./...`
6. Submit a PR with a clear description

## Communication

- Issues: [GitHub Issues](https://github.com/chmuri/terraform-provider-freeipa/issues)
- Discussions: [GitHub Discussions](https://github.com/chmuri/terraform-provider-freeipa/discussions)
