// Package adminui serves the local administration client from the Go binary.
package adminui

import (
 "embed"
 "net/http"
 "github.com/gin-gonic/gin"
)

//go:embed dist/*
var assets embed.FS

func RegisterRoutes(r *gin.Engine) {
 r.GET("/admin", func(c *gin.Context) { c.Redirect(http.StatusTemporaryRedirect, "/admin/") })
 // Only explicit public assets are exposed. Data remains protected by the API's
 // JWT and DB-backed admin middleware, including when this page is bypassed.
 for name, contentType := range map[string]string{"": "text/html; charset=utf-8", "app.js": "text/javascript; charset=utf-8", "style.css": "text/css; charset=utf-8", "api.json": "application/json; charset=utf-8"} {
  file := name
  if file == "" { file = "index.html" }
  body, err := assets.ReadFile("dist/" + file)
  if err != nil { panic(err) }
  r.GET("/admin/"+name, func(c *gin.Context) {
   c.Header("Cache-Control", "no-store")
   c.Header("X-Content-Type-Options", "nosniff")
   c.Header("Referrer-Policy", "no-referrer")
   c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' https: http:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
   c.Data(http.StatusOK, contentType, body)
  })
 }
}
