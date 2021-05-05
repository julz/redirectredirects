package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gernest/front"
)

func main() {
	base := flag.String("path", "docs", "")
	flag.Parse()

	m := front.NewMatter()
	m.Handle("---", front.YAMLHandler)

	redirects := make(map[string]string)
	if err := filepath.WalkDir(*base, func(path string, d fs.DirEntry, err error) error {
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %v: %w", path, err)
		}

		defer f.Close()

		yaml, _, err := m.Parse(f)
		if err != nil && err != front.ErrUnknownDelim {
			return fmt.Errorf("failed to parse %v: %w", path, err)
		}

		if a, ok := yaml["aliases"].([]interface{}); ok {
			for _, v := range a {
				from := v.(string)
				from = strings.TrimPrefix(from, "/docs")
				if !strings.HasPrefix(from, "/") {
					from, err = filepath.Rel(*base, from)
					if err != nil {
						return fmt.Errorf("making alias %q relative at %v: %w", from, path, err)
					}
				}

				from = strings.TrimPrefix(from, "/")

				if strings.HasPrefix(from, "..") {
					continue
				}

				redirects[from], err = filepath.Rel(*base, path)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		log.Fatal("FAILED:", err)
	}

	for k, v := range redirects {
		if !strings.HasSuffix(k, ".md") {
			k = strings.TrimSuffix(k, "/")
			k = k + "/index.md"
		}
		fmt.Printf("          %v: %v\n", k, v)
	}
}
