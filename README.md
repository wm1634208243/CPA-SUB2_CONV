# CPA-SUB2_Conv

A lightweight Go web app for converting account JSON between CPA (`codex`) and Sub2API formats.

## Highlights

- Convert `CPA -> Sub2API`
- Convert `Sub2API -> CPA`
- Paste JSON directly in the browser
- Upload a single `.json` file and download the converted result
- Upload a `.zip` archive and batch-convert all `.json` files inside it
- Download batch output as `converted_<timestamp>.zip`
- Choose how converted JSON files are named:
  - Timestamp
  - Output format + original file name
  - Output format + JSON `name`
  - Custom prefix

## Preview

Add a screenshot here after your first GitHub push, for example:

```md
![App Screenshot](./docs/screenshot.png)
```

## Quick Start

Build:

```bash
go build -o converter_server .
```

Run:

```bash
./converter_server
```

Open [http://localhost:8080](http://localhost:8080).

### Windows PowerShell

If you want the Go build cache to stay inside the project directory:

```powershell
$env:GOCACHE="$PWD\.gocache"
go build ./...
go test ./...
```

## Features

### Web UI

- Browser-based JSON paste and convert flow
- `.json` and `.zip` upload support
- Download naming options for pasted JSON and uploaded JSON files
- Batch ZIP output download

### Format handling

- Auto-detect source format in `auto` mode
- Convert CPA account JSON to Sub2API export JSON
- Convert Sub2API export JSON to CPA account JSON

## API

### `POST /api/detect`

Request:

```json
{ "input": "<json string>" }
```

Response:

```json
{ "format": "cpa" }
```

Possible values:

- `cpa`
- `sub2`
- `unknown`

### `POST /api/convert`

Request:

```json
{
  "input": "<json string>",
  "target": "auto"
}
```

Possible `target` values:

- `cpa`
- `sub2`
- `auto`

### `POST /api/convert-file`

Multipart form fields:

- `file`: `.json` or `.zip`
- `target`: `cpa`, `sub2`, or `auto`

Behavior:

- If the file is `.json`, the response is a downloadable converted JSON file
- If the file is `.zip`, each `.json` entry is converted and returned as `converted_<timestamp>.zip`

## Example Files

Sample input files are available in [`examples/`](D:/project/cpa-sub2-json-converter/examples):

- [cpa.sample.json](D:/project/cpa-sub2-json-converter/examples/cpa.sample.json)
- [sub2.sample.json](D:/project/cpa-sub2-json-converter/examples/sub2.sample.json)

## Project Structure

```text
.
|-- main.go
|-- static/
|   |-- index.html
|   `-- favicon.svg
|-- internal/
|   |-- converter/
|   `-- handler/
|-- examples/
|-- .github/
`-- README.md
```

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
go build ./...
```

## Roadmap

- Partial-success ZIP conversion with a failure report
- More robust input validation
- Better per-file progress or status for batch conversion
- Release packaging and versioned binaries

## Contributing

See [CONTRIBUTING.md](D:/project/cpa-sub2-json-converter/CONTRIBUTING.md).

## License

Released under the [MIT License](D:/project/cpa-sub2-json-converter/LICENSE).
