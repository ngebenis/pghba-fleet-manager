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
		h.Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; font-src 'self' data:; connect-src 'self'; "+
				"frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		c.Next()
	}
}
