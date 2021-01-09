//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package data

import (
	"errors"
	"fmt"
	"mime"
	"strings"
)

// Format specifies input data format.
type Format int

// Known input data formats.
const (
	FormatUnknown Format = iota
	FormatCSV
	FormatHTML
	FormatJSON
)

var mediatypes = map[string]Format{
	"text/csv":         FormatCSV,
	"text/html":        FormatHTML,
	"application/json": FormatJSON,
}

var suffixes = map[string]Format{
	".csv":  FormatCSV,
	".html": FormatHTML,
	".json": FormatJSON,
}

var formats = map[Format]NewSource{
	FormatCSV:  NewCSV,
	FormatHTML: NewHTML,
	FormatJSON: NewJSON,
}

var formatNames = map[Format]string{
	FormatUnknown: "unknown",
	FormatCSV:     "csv",
	FormatHTML:    "html",
	FormatJSON:    "json",
}

func (f Format) String() string {
	name, ok := formatNames[f]
	if ok {
		return name
	}
	return fmt.Sprintf("{Format %d}", f)
}

// Resolver resolves data format from input meta data.
type Resolver struct {
	format Format
	err    error
}

// Format returns the resolved input format.
func (r Resolver) Format() (Format, error) {
	if r.format == FormatUnknown {
		if r.err != nil {
			return r.format, r.err
		}
		return FormatUnknown, errors.New("could not resolve input format")
	}
	return r.format, nil
}

// ResolvePath resolves the input format from file path.
func (r *Resolver) ResolvePath(path string) {
	idx := strings.LastIndexByte(path, '.')
	if idx < 0 {
		r.err = errors.New("no file suffix")
		return
	}
	f, ok := suffixes[strings.ToLower(path[idx:])]
	if !ok {
		r.err = fmt.Errorf("unknown file suffix '%s'", path[idx:])
		return
	}
	r.format = f
}

// ResolveMediaType resolves the input format from content media type.
func (r *Resolver) ResolveMediaType(t string) {
	if len(t) == 0 {
		r.err = errors.New("no Content-Type")
		return
	}
	mediatype, _, err := mime.ParseMediaType(t)
	if err != nil {
		r.err = err
		return
	}
	var ok bool
	f, ok := mediatypes[mediatype]
	if !ok {
		r.err = fmt.Errorf("unknown Content-Type: %s", mediatype)
		return
	}
	r.format = f
}
