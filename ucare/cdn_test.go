package ucare

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestCnamePrefix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		pubkey string
		want   string
	}{
		{"demopublickey", "1s4oyld5dc"},
		{"anotherkey", "4073mye3t0"},
	}

	for _, c := range cases {
		assert.Equal(t, c.want, CnamePrefix(c.pubkey), "pubkey: %s", c.pubkey)
	}
}

func TestCDNBaseURL(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		"https://1s4oyld5dc.ucarecd.net",
		CDNBaseURL("demopublickey"),
	)
}
