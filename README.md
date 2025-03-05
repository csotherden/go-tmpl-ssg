# Go Template Static Site Generator

This project is a simple static site generator written in Go, designed to generate HTML files from Go templates. It supports component-based templating to make managing common site elements like headers and footers easier. It also supports layout-based templating to apply consistent structure across multiple pages. The generator also supports injecting dynamic data into templates via JSON files. The generated static site can be deployed easily anywhere you can serve static files. For example, to an S3 bucket behind a CloudFront distribution for efficient hosting.

## Features

- Uses Go templates for reusable components (`components/` directory)
- Supports layout-based page rendering for consistent structure (`layouts/` directory)
- Processes `.tmpl` files in the `pages/` directory and outputs HTML
- Allows passing dynamic data to templates via `.tmpl.json` files
- Copies non-template files (e.g., `robots.txt`) to the output directory
- Simple CLI interface to specify template and output directories
- **Optional sitemap generation (`sitemap.xml`)**
- **Supports defining a base URL for fully qualified sitemap entries**

## Installation

Ensure you have Go installed, then clone the repository:

```sh
git clone https://github.com/csotherden/go-tmpl-ssg.git
cd go-tmpl-ssg
```

Build the binary:

```sh
go build ./cmd/go-tmpl-generate
```

## Usage

Run the generator with:

```sh
./go-tmpl-generate --template path/to/templates --output path/to/output
```

### Additional Flags

| Flag           | Description                                         | Example                                  |
|---------------|-----------------------------------------------------|------------------------------------------|
| `--sitemap`   | Generates a `sitemap.xml` in the output directory  | `--sitemap`                              |
| `--base-url`  | Specifies the base URL for sitemap entries         | `--base-url "https://example.com"`       |

### Template Directory Structure

```
templates/
├── components/
│   ├── header.tmpl
│   ├── footer.tmpl
├── layouts/
│   ├── base.tmpl
├── pages/
│   ├── index.tmpl
│   ├── about.tmpl
│   ├── robots.txt
```

- `components/`: Contains common reusable templates like `header.tmpl` and `footer.tmpl`
- `layouts/`: Contains layout templates that define a page structure (e.g., `base.tmpl`)
- `pages/`: Contains individual page templates
- `.tmpl.json` files: Provide dynamic data to templates
- Non-template files (e.g., `robots.txt`) are copied as-is

### Using Layout Templates

Pages can specify a parent layout by including a comment at the beginning of the file:

#### Example Layout Template (`layouts/base.tmpl`):
```html
{{template "header.tmpl" .}}
<main>
    {{template %CONTENT% .}}
</main>
{{template "footer.tmpl" .}}
```

#### Example Page Template (`pages/index.tmpl`):
```html
{{- /* layout:base.tmpl */ -}}
<p>Welcome to My Site</p>
```

This results in the following rendered HTML:
```html
<header>Header</header>
<main>
    <p>Welcome to My Site</p>
</main>
<footer>Footer</footer>
```

### Using JSON Data with Templates

If a `page.tmpl.json` file exists alongside a `page.tmpl`, it will be parsed and made available as template data.

#### Example JSON File (`pages/index.tmpl.json`):
```json
{
  "Title": "Home Page"
}
```

#### Example Template (`pages/index.tmpl`):
```html
{{- /* layout:base.tmpl */ -}}
<h1>{{.Title}}</h1>
<p>Welcome to My Site</p>
```

This allows dynamic data (like page titles) to be injected without modifying the templates directly.

## Running Tests

To run tests, use:

```sh
go test ./pkg/generator
```

## License

MIT License

## Author

Chris Sotherden

