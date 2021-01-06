//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/markkurossi/iql/types"
)

var (
	_ types.Source = &CSV{}
	_ types.Source = &HTML{}
)

// NewSource defines a constructor for data sources.
type NewSource func(in []io.ReadCloser, filter string,
	columns []types.ColumnSelector) (types.Source, error)

// New creates a new data source for the URL.
func New(url, filter string, columns []types.ColumnSelector) (
	types.Source, error) {

	input, format, err := openInput(url)
	if err != nil {
		return nil, err
	}

	n, ok := formats[format]
	if !ok {
		return nil, fmt.Errorf("unknown data format '%s'", format)
	}
	return n(input, filter, columns)
}

func openInput(input string) ([]io.ReadCloser, Format, error) {
	var resolver Resolver

	u, err := url.Parse(input)
	if err != nil {
		resolver.ResolvePath(input)
	} else {
		resolver.ResolvePath(u.Path)
	}
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		resp, err := http.Get(input)
		if err != nil {
			return nil, 0, err
		}
		if resp.StatusCode != http.StatusOK {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
			return nil, 0, fmt.Errorf("HTTP URL '%s' not found", input)
		}

		resolver.ResolveMediaType(resp.Header.Get("Content-Type"))

		format, err := resolver.Format()
		return []io.ReadCloser{resp.Body}, format, err
	}
	if err == nil && u.Scheme == "data" {
		idx := strings.IndexByte(input, ',')
		if idx < 0 {
			return nil, 0, fmt.Errorf("malformed data URI: %s", input)
		}
		data := input[idx+1:]
		contentType := input[5:idx]
		var encoding string

		idx = strings.IndexByte(contentType, ';')
		if idx >= 0 {
			encoding = contentType[idx+1:]
			contentType = contentType[:idx]
		}

		var decoded []byte

		// Decode data.
		switch encoding {
		case "base64":
			decoded, err = base64.StdEncoding.DecodeString(data)
			if err != nil {
				return nil, 0, err
			}
		case "":
			decoded = []byte(data)
		default:
			return nil, 0, fmt.Errorf("unknown data URI encoding: %s", encoding)
		}

		// Resolve format.
		resolver.ResolveMediaType(contentType)

		format, err := resolver.Format()

		return []io.ReadCloser{
			&memory{
				in: bytes.NewReader(decoded),
			},
		}, format, err
	}

	matches, err := filepath.Glob(input)
	if err != nil {
		return nil, 0, err
	}
	var result []io.ReadCloser
	for _, match := range matches {
		f, err := os.Open(match)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, f)
	}

	format, err := resolver.Format()

	return result, format, err
}

type memory struct {
	in io.Reader
}

func (m *memory) Read(p []byte) (n int, err error) {
	return m.in.Read(p)
}

func (m *memory) Close() error {
	return nil
}
