package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/iceber/iouring-go"
)

var totalLinks int = 256

func BenchmarkRegular(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dir := b.TempDir()
		sourceFile := filepath.Join(dir, "source")
		if err := os.WriteFile(sourceFile, []byte("i am the file contents"), 0666); err != nil {
			b.Fatal(err)
		}
		for i := 0; i < totalLinks; i++ {
			if err := os.Symlink(sourceFile, filepath.Join(dir, fmt.Sprint(i))); err != nil {
				b.Fatal(err)
			}
		}
	}
}
func BenchmarkIOUring(b *testing.B) {
	for i := 0; i < b.N; i++ {
		iour, err := iouring.New(uint(totalLinks))
		if err != nil {
			b.Fatal(err)
		}

		dir := b.TempDir()
		sourceFile := filepath.Join(dir, "source")
		if err := os.WriteFile(sourceFile, []byte("i am the file contents"), 0666); err != nil {
			b.Fatal(err)
		}
		f, err := os.Open(dir)
		if err != nil {
			b.Fatal(err)
		}
		ch := make(chan iouring.Result, totalLinks)
		count := totalLinks
		for i := 0; i < totalLinks; i++ {
			req, err := iouring.Symlinkat(sourceFile, int(f.Fd()), filepath.Join(dir, fmt.Sprint(i)))
			if err != nil {
				b.Fatal(err)
			}
			if _, err := iour.SubmitRequest(req, ch); err != nil {
				b.Fatal(err)
			}
			select {
			case result := <-ch:
				count--
				if err := result.Err(); err != nil {
					b.Fatal(err)
				}
			default:
			}
		}
		for i := 0; i < count; i++ {
			result := <-ch
			if err := result.Err(); err != nil {
				b.Fatal(err)
			}
		}
		if err := iour.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
