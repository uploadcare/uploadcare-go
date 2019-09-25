package uploadcare

type RESTAPIVersion string

const (
	APIv05 RESTAPIVersion = "v0.5"
	APIV06 RESTAPIVersion = "v0.6"
)

var (
	supportedVersions = map[RESTAPIVersion]bool{
		APIv05: true,
		APIV06: true,
	}

	DefaultAPIVersion = APIv05
)
