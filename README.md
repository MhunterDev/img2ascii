# Img2ASCII

Convert images and text banners to ASCII art via a modern web interface.

## Features
- Upload images (PNG, JPEG, GIF) and convert them to ASCII art.
- Generate ASCII art banners from custom text using included fonts.
- Download or view ASCII output directly in the browser.
- Modern, responsive web UI.
- Fast, concurrent image processing in Go.

## Requirements
- Go 1.24+
- [Go modules](https://golang.org/doc/go1.11#modules)
- Fonts for banner generation (included in `source/banners/fonts/`)

## Installation

1. **Clone the repository:**
   ```sh
   git clone https://github.com/MhunterDev/img2ascii.git
   cd img2ascii
   ```
2. **Install dependencies:**
   ```sh
   go mod download
   ```
3. **Build and run:**
   - For local development (no embed):
     ```sh
     go run main.go
     ```
   - For production (with embedded assets):
     ```sh
     go run -tags=embed main_embed.go
     ```

## Usage

1. **Start the server:**
   ```sh
   go run main.go
   # or
   go run -tags=embed main_embed.go
   ```
2. **Open your browser:**
   Navigate to [http://localhost:8080](http://localhost:8080)
3. **Convert an image:**
   - Use the upload form to select and submit an image file.
   - The ASCII art will be displayed in the output area.
4. **Generate a banner(beta):** 
   - Enter your text in the banner form and submit.
   - The ASCII banner will be displayed in the output area.

## Configuration
You can override default directories and output files using environment variables:
- `IMG2ASCII_OUTPUT_DIR` — Output directory for ASCII files (default: `/tmp/img2ascii`)
- `IMG2ASCII_OUTPUT_FILE` — Default output file (default: `/tmp/img2ascii/output.txt`)
- `IMG2ASCII_WWW_DIR` — Directory for static web assets (default: `/tmp/img2ascii/www`)

> **Note:** Output files are temporary and cleaned up after use unless you change the output directory.

## Project Structure
```
go.mod, go.sum         # Go module files
main.go                # Main entrypoint (no embed)
main_embed.go          # Main entrypoint (with embed)
main_test.go           # Basic tests
source/
  banners/             # Banner rendering logic and fonts
  handlers/            # HTTP handlers
  img2ascii/           # Image-to-ASCII conversion logic
  www/                 # Static web assets (HTML, CSS, JS)
```

## Development
- Static assets are served from `source/www/` in development.
- For production, assets can be embedded using Go's `embed` package (see `main_embed.go`).
- Banner fonts are stored in `source/banners/fonts/` and are included by default.

## Testing
Run the included tests with:
```sh
go test ./...
```
> Only basic tests are included. Consider adding more tests for production use.

## License
MIT License. See [LICENSE](LICENSE) for details.

## Credits
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Fogleman GG](https://github.com/fogleman/gg) for banner rendering
- [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) for image processing

---

For questions or contributions, please open an issue or pull request on GitHub.
