# Restrum

rest·rum /ˈrestrəm/ - A simple, lightweight Go HTTP framework.

### Feature Overview

- Lightweight and easy-to-use routing system
- CORS support via middleware configuration
- Simple JSON, HTML, and data responses
- Route parameter and query parameter handling
- Support for form values and cookies management
- Easy grouping of routes
- IP address extraction from requests

## Guide

### Installation

```bash
go get -u github.com/bagasdisini/restrum
```

### Example

```go
package main

import (
    "net/http"
    "github.com/bagasdisini/restrum"
)

func main() {
	// Define CORS configuration
	config := &restrum.Config{
		AllowOrigins:     []string{"http://example.com", "http://example2.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowCredentials: true,
	}

	// Create a new Restrum engine
	r := restrum.New(config)

	// Use CORS middleware globally
	r.Use(restrum.CORSMiddleware(config))
	
	// Create a new group for API routes
	api := r.Group("/api")

	// Define a simple GET route
	api.GET("/hello", func(ctx *restrum.Context) {
		ctx.JSON(http.StatusOK, map[string]string{"message": "Hello, World!"})
	})
	
	// Define a simple GET route with a parameter
	api.GET("/hello/:name", func(ctx *restrum.Context) {
		name := ctx.Param("name")
		ctx.JSON(http.StatusOK, map[string]string{"message": "Hello, " + name + "!"})
	})
	
	// Define a simple POST route with form values
	r.POST("/login", func(ctx *restrum.Context) {
		username := ctx.FormValue("username")
		password := ctx.FormValue("password")

		ctx.JSON(http.StatusOK, map[string]string{
			"username": username,
			"password": password,
		})
	})

	// Run the server on port 8080
	r.Run(":8080")
}
```