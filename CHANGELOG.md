## 0.2.0

BREAKING CHANGES:

* Target REST API v0.7 (previously v0.5)
* Remove `ImageInfo` and `VideoMeta` fields from `file.BasicFileInfo` — use `ContentInfo.Image` and `ContentInfo.Video`
* Remove `RecognitionInfo` field from `file.Info` — use `AppData`
* Add `Metadata` and `AppData` fields to `file.Info`
* Remove `group.Store()` method (endpoint removed in v0.7)
* Remove `file.Copy()` method and `file.CopyParams` type — use `LocalCopy()` and `RemoteCopy()`
* Remove `file.OrderBySizeAsc` and `file.OrderBySizeDesc` constants (not supported in v0.7)
* Remove `APIv05` and `APIv06` constants
* Minimum Go version is now 1.25

IMPROVEMENTS:

* Add `UserAgent` field to `ucare.Config` for custom agent identification
* Replace `http.NewRequest` + `WithContext` with `http.NewRequestWithContext`
* Throttle retry loops now respect context cancellation
* Replace `ioutil` usage with `io` equivalents
* Replace `go-env` dependency with `os.Getenv`
* Update `stretchr/testify` to v1.10.0
* Update CI: Go 1.25, modern GitHub Actions versions, remove deprecated golint
* Integration tests skip gracefully when credentials are not set
* Fix errors in package documentation examples

## 1.2.1 (September 1, 2020)

IMPROVEMENTS:

* Update delete method endpoint
* Remove useless code

## 1.2.0 (August 19, 2020)

FEATURES:

* Webhooks
* Project

BUG FIXES:

* Fix empty response handling

## 1.1.10 (June 6, 2020)

BUG FIXES:

* Fix throttling request empty body issue

## 1.1.9 (May 3, 2020)

BUG FIXES:

* Set default upload ToStore form param value to "auto" 
* Change "UPLOADCARE_STORE" upload.FromURL param to "store" according to specs

## 1.1.8 (Apr 22, 2020)

IMPROVEMENTS:

* Use HMAC-SHA256 signature for signed uploads
* Set upload TTL to 60 seconds

## 1.1.7 (Apr 14, 2020)

BUG FIXES:

* Change ImageInfo.Orientation type to interface{} 

## 1.1.6 (Apr 14, 2020)

BUG FIXES:

* Change ImageInfo.Orientation type to \*string

## 1.1.5 (Mar 26, 2020)

BUG FIXES:

* Change ImageInfo.DateTimeOrignal type to \*time.Time

## 1.1.4 (Mar 20, 2020)

BUG FIXES:

* Change ImageInfo.DPI field value type to []float64

## 1.1.3 (Mar 20, 2020)

BUG FIXES:

* Change Location field value types to float64

## 1.1.2 (Feb 20, 2020)

BUG FIXES:

* Change file.VideoStreamMeta.FrameRate type (uint64 to float64)

## 1.1.1 (Feb 18, 2020)

BUG FIXES:

* Change file.AudioStreamMeta.Channels type (uint64 to string)

## 1.1.0 (Nov 8, 2019)

FEATURES:

* Support for the APIv05 file Copy method

IMPROVEMENTS:

* Use caching during CI builds
* Run integration test on push

BUG FIXES:

* Some broken tests
* Broken conversion api request body construction

## 1.0.0 (Oct 17, 2019)

Initial version
