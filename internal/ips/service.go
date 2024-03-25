/*
 * Copyright (c) 2023 shenjunzheng@gmail.com
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ips

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Service initializes and runs the main web service for the application.
// It sets up middlewares, routes and starts the HTTP server.
func (m *Manager) Service() {
	router := gin.New()

	// Handle error from SetTrustedProxies
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Error("Failed to set trusted proxies:", err)
	}

	// Middleware
	router.Use(
		gin.Recovery(),
		gin.Logger(),
	)

	m.router = router

	m.InitRouter()

	// Handle error from Run
	if err := m.router.Run(m.Conf.Addr); err != nil {
		log.Error("Failed to run server:", err)
	}
}

// EFS holds embedded file system data for static assets.
//
//go:embed static
var EFS embed.FS

// InitRouter sets up routes and static file servers for the web service.
// It defines endpoints for API as well as serving static content.
func (m *Manager) InitRouter() {
	staticDir, _ := fs.Sub(EFS, "static")
	// m.router.StaticFS("/static", http.FS(staticDir))
	m.router.StaticFileFS("/favicon.ico", "./favicon.ico", http.FS(staticDir))
	// m.router.StaticFileFS("/", "./index.htm", http.FS(staticDir))

	// // API Router
	// api := m.router.Group("/api")
	// {
	// 	api.GET("/v1/ip", m.GetIP)
	// 	api.GET("/v1/query", m.GetQuery)
	// }
	m.router.GET("/", m.GetIP)
	m.router.GET("/:ip", m.GetIP)

	m.router.NoRoute(m.NoRoute)
}

// NoRoute handles 404 Not Found errors. If the request URL starts with "/api"
// or "/static", it responds with a JSON error. Otherwise, it redirects to the root path.
func (m *Manager) NoRoute(c *gin.Context) {
	path := c.Request.URL.Path
	switch {
	case strings.HasPrefix(path, "/api"), strings.HasPrefix(path, "/static"):
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	default:
		c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate, value")
		c.Redirect(http.StatusFound, "/")
	}
}

func (m *Manager) GetIP(c *gin.Context) {
	ip := c.Param("ip")

	if ip == "" {
		ip = getClientIP(c)
	}

	info, err := m.parseIP(ip)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info.Output(m.Conf.UseDBFields))
}

func getClientIP(c *gin.Context) string {
	// 首先尝试从X-Forwarded-For头信息获取客户端IP
	xForwardedFor := c.Request.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For可能包含多个IP，第一个通常是客户端的原始IP
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 && ips[0] != "" {
			return strings.TrimSpace(ips[0])
		}
	}

	// 如果X-Forwarded-For头信息不存在或不包含IP，则尝试从X-Real-IP获取
	xRealIP := c.Request.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// 如果以上方法都未能获取到IP，则回退到Gin的默认方法
	return c.ClientIP()
}
