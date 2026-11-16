// ตัวอย่าง //go:embed: string, []byte, embed.FS และเสิร์ฟ static ด้วย http.FileServerFS
package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"text/template"
)

// ฝังไฟล์เดียวเป็น string
//
//go:embed static/hello.txt
var hello string

// ฝังไฟล์เดียวเป็น []byte
//
//go:embed static/index.html
var indexHTML []byte

// ฝังทั้งโฟลเดอร์เป็น embed.FS (read-only file system)
//
//go:embed static
var staticFS embed.FS

// ฝังด้วย pattern
//
//go:embed templates/*.tmpl
var tmplFS embed.FS

func main() {
	fmt.Print("string: ", hello)
	fmt.Printf("[]byte: %d bytes\n", len(indexHTML))

	// เดินดูไฟล์ทั้งหมดใน embed.FS
	fs.WalkDir(staticFS, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			fmt.Println("file:", path)
		}
		return nil
	})

	// ใช้ template ที่ฝังไว้
	t := template.Must(template.ParseFS(tmplFS, "templates/*.tmpl"))
	t.ExecuteTemplate(os.Stdout, "greet.tmpl", "Gopher")

	// fs.Sub ตัด prefix "static/" ออก แล้วเสิร์ฟด้วย http.FileServerFS
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(sub)))

	// ทดสอบด้วย httptest แทนการเปิด port จริง
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/static/hello.txt")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("HTTP", resp.StatusCode, resp.Header.Get("Content-Type"))
	fmt.Println("body:", strings.TrimSpace(string(body)))
}
