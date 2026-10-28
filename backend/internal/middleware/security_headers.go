package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets defense-in-depth response headers the Go server
// otherwise never sends. A reverse proxy in front of this app may already
// set some of these, but the app should not depend on that — this is the
// one process that always serves both the API and the built frontend.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "same-origin")
		// unsafe-inline/eval only for script/style because the built Vue
		// bundle still needs inline styles from component <style> blocks
		// (Vite emits some at runtime) — connect-src 'self' is what
		// actually matters here, keeping XHR/fetch confined to this origin.
		// style-src/font-src also allow the two CDNs main.css loads at
		// build time (Google Fonts for JetBrains Mono, jsDelivr for the
		// Tabler Icons webfont) — without them the browser silently blocks
		// both @import rules and every icon in the UI renders blank.
		h.Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net; "+
				"img-src 'self' data:; "+
				"font-src 'self' data: https://fonts.gstatic.com https://cdn.jsdelivr.net; "+
				"connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		c.Next()
	}
}
