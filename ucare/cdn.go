package ucare

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// DefaultCDNDomain is the per-project CDN domain used for file delivery.
const DefaultCDNDomain = "ucarecd.net"

// CnamePrefix computes the per-project CDN subdomain prefix from a public key.
// It SHA-256 hashes the key and base-36 encodes the result, returning the
// first 10 characters.
func CnamePrefix(pubkey string) string {
	hash := sha256.Sum256([]byte(pubkey))
	hexStr := hex.EncodeToString(hash[:])
	n := new(big.Int)
	n.SetString(hexStr, 16)
	return n.Text(36)[:10]
}

// CDNBaseURL returns the full CDN base URL for a given public key.
func CDNBaseURL(pubkey string) string {
	return fmt.Sprintf("https://%s.%s", CnamePrefix(pubkey), DefaultCDNDomain)
}
