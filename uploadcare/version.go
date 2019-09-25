package uploadcare

type RESTAPIVersion string

const (
	APIv05 RESTAPIVersion = "v0.5"
	APIV06 RESTAPIVersion = "v0.6"
)

var DefaultAPIVersion = APIv05
