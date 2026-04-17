package conversion

import (
	"fmt"
	"strings"
)

type DocumentPathOptions struct {
	UUID string
	// Format is the target document format. Defaults to "pdf" when empty.
	Format string
	// Page extracts a single page when greater than zero.
	Page int
}

func BuildDocumentPath(opts DocumentPathOptions) string {
	format := opts.Format
	if format == "" {
		format = "pdf"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s/document/-/format/%s/", opts.UUID, format)

	if opts.Page > 0 {
		fmt.Fprintf(&b, "-/page/%d/", opts.Page)
	}

	return b.String()
}

type VideoPathOptions struct {
	UUID   string
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
	var b strings.Builder
	fmt.Fprintf(&b, "%s/video/-/format/%s/", opts.UUID, opts.Format)

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
