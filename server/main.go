package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/gin-gonic/gin"
	"golang.org/x/mod/semver"
)

const binaryRoot = "/dist"

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.GET("/ping", Ping)
	router.GET("/versions", Versions)
	router.GET("/latest", Latest)
	router.GET("/download/:filename", Download)

	log.Print("Starting Server")
	router.Run(":8080")
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"response": "pong"})
}

func Versions(c *gin.Context) {
	entries, err := os.ReadDir(binaryRoot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var versions []string
	for _, entry := range entries {
		versions = append(versions, entry.Name())
	}
	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

func Latest(c *gin.Context) {
	entries, err := os.ReadDir(binaryRoot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	slices.SortFunc(entries, func(a, b os.DirEntry) int {
		return semver.Compare(b.Name(), a.Name())
	})

	c.JSON(http.StatusOK, gin.H{"latest": entries[0].Name()})
}

func Download(c *gin.Context) {
	filename := c.Param("filename")

	filePath := filepath.Join(binaryRoot, filename)
	stat, err := os.Stat(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", string(stat.Size()))
	c.File(filePath)
}
