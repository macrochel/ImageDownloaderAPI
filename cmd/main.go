package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// Define a route to handle image download
	app.Post("/download", func(c *fiber.Ctx) error {
		// Get the URL from the request body
		url := c.FormValue("url")
		if url == "" {
			return c.Status(http.StatusBadRequest).SendString("Missing 'url' parameter")
		}

		// Download and save the image
		fileName := downloadImage(url)

		return c.SendString(":2050/static/" + fileName)
	})

	// Serve static files from the 'statics' folder
	app.Static("/static", "./statics")

	// Start the server
	log.Fatal(app.Listen(":2050"))
}

func downloadImage(url string) string {
	response, err := http.Get(url)
	print(err)
	defer response.Body.Close()

	hash := md5.New()
	hash.Write([]byte(url))
	hashInBytes := hash.Sum(nil)
	fileName := hex.EncodeToString(hashInBytes) + ".jpg"

	file, err := os.Create(filepath.Join("./statics", fileName))
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	return fileName
}
