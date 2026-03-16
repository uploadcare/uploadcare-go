package conversion

import (
	"fmt"
	"strings"
)

// DocumentPathOptions holds options for building a document conversion path.
type DocumentPathOptions struct {
	// UUID is the source file UUID.
	UUID string
	// Format is the target document format. Defaults to "pdf" when empty.
	Format string
	// Page extracts a single page when greater than zero.
	Page int
}

// BuildDocumentPath constructs a document conversion path.
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

// VideoPathOptions holds options for building a video conversion path.
type VideoPathOptions struct {
	// UUID is the source file UUID.
	UUID string
	// Format is the target video format.
	Format string
	// Size is the output dimensions, for example "640x480".
	Size string
	// ResizeMode controls how the video is resized.
	ResizeMode string
	// Quality controls output quality.
	Quality string
	// CutStart is the start time for cutting.
	CutStart string
	// CutLength is the duration to cut.
	CutLength string
	// Thumbs is the number of thumbnails to generate.
	Thumbs int
}

// BuildVideoPath constructs a video conversion path.
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
