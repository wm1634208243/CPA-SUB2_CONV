# Contributing

Thanks for your interest in improving this project.

## Getting started

1. Fork the repository
2. Create a feature branch
3. Run the test suite
4. Open a pull request with a short summary of the change

## Local development

```bash
go test ./...
go build ./...
```

If you want to use a local build cache inside the project directory on Windows:

```powershell
$env:GOCACHE="$PWD\.gocache"
go test ./...
go build ./...
```

## Pull request notes

- Keep changes focused and easy to review
- Add or update tests when behavior changes
- Update `README.md` if the user-facing workflow changes
