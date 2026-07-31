// Package tag provides access to the Uploadcare per-file tags API.
package tag

// UpdateParams adds and removes tags atomically. Deletions are applied before
// additions by the API.
type UpdateParams struct {
	Add    []string `json:"add,omitempty"`
	Delete []string `json:"delete,omitempty"`
}

// Result describes the tag state after Replace or Update.
type Result struct {
	Tags    []string `json:"tags"`
	Added   []string `json:"added"`
	Deleted []string `json:"deleted"`
}
