# Img2ASCII

Convert images and text banners to ASCII art via a modern web interface.

## Features
- Upload images (PNG, JPEG, GIF) and convert them to ASCII art.
- **Multiple aspect ratio modes:**
  - **Aspect Ratio Scaling** - Maintains image proportions within 65x54 character limits
  - **1:1 Pixel Map** - Direct pixel-to-character mapping with safety limits (300x200 max)
  - **Fixed Output Size** - Custom width/height dimensions (10-200 width, 10-100 height)
- Generate ASCII art banners from custom text using included fonts.
- Download or view ASCII output directly in the browser.
- Modern, responsive web UI with intuitive controls.
- Fast, concurrent image processing in Go.
- Rate limiting and file validation for security.

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

   ```sh
   go build -tags embed
   ./img2ascii
   ```

   Or run directly:

   ```sh
   go run -tags embed main.go
   ```

## Usage

1. **Start the server:**

   ```sh
   ./img2ascii
   # or
   go run -tags embed main.go
   ```

2. **Open your browser:**
   Navigate to [http://localhost:8080](http://localhost:8080)

3. **Convert an image:**
   - Use the upload form to select an image file.
   - Choose your preferred aspect ratio mode:
     - **Aspect Ratio Scaling (default)** - Maintains proportions, fits in 65x54 characters
     - **1:1 Pixel Map** - Direct pixel mapping with reasonable size limits
     - **Fixed Output Size** - Specify exact width/height dimensions
   - For fixed size mode, adjust the width (10-200) and height (10-100) values as needed.
   - Submit to see the ASCII art displayed in the output area.

4. **Generate a banner (beta):**
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
main.go                # Main entrypoint (with embedded assets)
main_test.go           # Basic tests
source/
  banners/             # Banner rendering logic and fonts
  handlers/            # HTTP handlers with aspect ratio support
  img2ascii/           # Image-to-ASCII conversion logic
  middleware/          # Rate limiting and other middleware
  www/                 # Static web assets (HTML, CSS, JS)
```

## Development

- Static assets are embedded using Go's `embed` package for easy deployment.
- Banner fonts are stored in `source/banners/fonts/` and are included by default.
- The application supports three aspect ratio modes for flexible ASCII output.
- Rate limiting and file validation are implemented for security.
- Concurrent image processing provides fast conversion times.

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
