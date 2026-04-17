package conversion

import (
	"fmt"
	"strings"
)

type DocumentPathOptions struct {
	UUID string
	// Target document format. When empty, defaults to "png" if Page > 0
	// (the backend only accepts /page/ for image formats) and "pdf" otherwise.
	Format string
	// Page extracts a single page when greater than zero. Requires Format
	// to be an image format (jpg, png, tiff, webp, enhanced.jpg).
	Page int
}

func BuildDocumentPath(opts DocumentPathOptions) string {
	format := opts.Format
	if format == "" {
		if opts.Page > 0 {
			format = "png"
		} else {
			format = "pdf"
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s/document/-/format/%s/", opts.UUID, format)

	if opts.Page > 0 {
		fmt.Fprintf(&b, "-/page/%d/", opts.Page)
	}

	return b.String()
}

type VideoPathOptions struct {
	UUID string
	// Target video format. Defaults to "mp4".
	Format string
	// Output dimensions, for example "640x480".
	Size       string
	ResizeMode string
	Quality    string
	CutStart   string
	CutLength  string
	// Number of thumbnails to generate.
	Thumbs int
}

func BuildVideoPath(opts VideoPathOptions) string {
	format := opts.Format
	if format == "" {
		format = "mp4"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s/video/-/format/%s/", opts.UUID, format)

	if opts.Size != "" {
		fmt.Fprintf(&b, "-/size/%s/", opts.Size)
		if opts.ResizeMode != "" {
			fmt.Fprintf(&b, "%s/", opts.ResizeMode)
		}
	}

	if opts.Quality != "" {
		fmt.Fprintf(&b, "-/quality/%s/", opts.Quality)
	}

	if opts.CutStart != "" && opts.CutLength != "" {
		fmt.Fprintf(&b, "-/cut/%s/%s/", opts.CutStart, opts.CutLength)
	}

	if opts.Thumbs > 0 {
		fmt.Fprintf(&b, "-/thumbs~%d/", opts.Thumbs)
	}

	return b.String()
}
