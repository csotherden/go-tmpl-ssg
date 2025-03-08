# Go Template Static Site Generator

This project is a simple static site generator written in Go, designed to generate HTML files from Go templates. It supports component-based templating to make managing common site elements like headers and footers easier. It also supports layout-based templating to apply consistent structure across multiple pages. The generator also supports injecting dynamic data into templates via JSON files and provides built-in pagination support. The generated static site can be deployed easily anywhere you can serve static files, such as an S3 bucket behind a CloudFront distribution for efficient hosting.

## Features

- Uses Go templates for reusable components (`components/` directory)
- Supports layout-based page rendering for consistent structure (`layouts/` directory)
- Processes `.tmpl` files in the `pages/` directory and outputs HTML
- Allows passing dynamic data to templates via `.tmpl.json` files
- Supports automatic pagination for pages with large datasets
- Copies non-template files (e.g., `robots.txt`) to the output directory
- Simple CLI interface to specify template and output directories
- Optional sitemap generation (`sitemap.xml`)
- Supports defining a base URL for fully qualified sitemap entries
- Built-in development server with live reload support

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

### Development Server

The built-in development server allows for live reloading of templates when changes are detected.

#### Running the Dev Server
```sh
go run ./cmd/go-tmpl-dev-server --port 8080 --root output --source templates
```

#### Features
- Serves files from the output directory.
- Watches the template directory for changes and automatically recompiles templates.
- Notifies connected browsers via WebSocket to refresh the page after changes.

A WebSocket client is automatically injected into the `<head>` section of HTML pages when the dev server is enabled, ensuring a seamless live preview experience.

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

## Built-in Template Functions

The generator provides several built-in functions that can be used in your templates:

| Function | Description | Example |
|----------|-------------|---------|
| `html` | Outputs a string as HTML without escaping | `{{ html "<strong>Bold</strong>" }}` |
| `now` | Returns the current time as a string | `{{ now }}` |
| `nowFormat` | Returns the current time formatted according to the Go time format layout | `{{ nowFormat "2006-01-02" }}` |
| `formatDate` | Formats a date string (RFC3339 format) using the specified layout | `{{ formatDate "2023-04-01T15:04:05Z" "January 2, 2006" }}` |
| `upper` | Converts string to uppercase | `{{ upper "text" }}` |
| `lower` | Converts string to lowercase | `{{ lower "TEXT" }}` |
| `title` | Converts string to title case | `{{ title "hello world" }}` |
| `add` | Adds two integers | `{{ add 1 2 }}` |
| `sub` | Subtracts second integer from first | `{{ sub 5 2 }}` |
| `seq` | Generates a sequence of integers from start to end | `{{ range seq 1 5 }}{{ . }}{{ end }}` |
| `join` | Joins array elements with a separator | `{{ join .Items ", " }}` |

### Examples

#### Formatting current date:
```html
<p>Generated on {{ nowFormat "January 2, 2006" }}</p>
```

#### Creating pagination links:
```html
<nav>
  {{ range seq 1 .Iterations.TotalPages }}
    <a href="{{ $.Iterations.PageRoot }}/{{ . }}/">{{ . }}</a>
  {{ end }}
</nav>
```

#### Using HTML content without escaping:
```html
{{ html .RichTextContent }}
```

## Running Tests

To run tests, use:

```sh
go test ./pkg/generator
```

## License

MIT License

## Author

Chris Sotherden

