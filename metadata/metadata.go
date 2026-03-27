// Package metadata holds primitives and logic for the file metadata API.
//
// File metadata is a key-value store associated with each file.
// Keys are strings matching ^[-_.:A-Za-z0-9]{1,64}$ and values are
// strings of up to 512 characters.
package metadata
