package www

import "embed"

//go:embed index.html styles.css main.js
var StaticFiles embed.FS
