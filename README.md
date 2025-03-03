# Go Template Static Site Generator

This project is a simple static site generator written in Go, designed to generate HTML files from Go templates. It supports component-based templating to make managing common site elements like headers and footers easier. The generated static site can be deployed easily anywhere you can serve static files. For example, to an S3 bucket behind a CloudFront distribution for efficient hosting.

## Features
- Uses Go templates for reusable components (`components/` directory)
- Processes `.tmpl` files in the `pages/` directory and outputs HTML
- Copies non-template files (e.g., `robots.txt`) to the output directory
- Simple CLI interface to specify template and output directories

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

### Template Directory Structure
```
templates/
├── components/
│   ├── header.tmpl
│   ├── footer.tmpl
├── pages/
│   ├── index.tmpl
│   ├── about.tmpl
│   ├── robots.txt
```

- `components/`: Contains common reusable templates like `header.tmpl` and `footer.tmpl`
- `pages/`: Contains individual page templates
- Non-template files (e.g., `robots.txt`) are copied as-is

### Example Page Template
```html
{{template "header.tmpl"}}
<main>
    <h1>Welcome to My Site</h1>
</main>
{{template "footer.tmpl"}}
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

