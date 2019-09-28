package ucare

// Public configuration constants
const (
	APIv05 = "v0.5"
	APIv06 = "v0.6"
)

var (
	supportedVersions = map[string]bool{
		APIv05: true,
		APIv06: true,
	}

	defaultAPIVersion = APIv05
)
