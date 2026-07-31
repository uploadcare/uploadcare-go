## 2.1.0 (August 10, 2026)

FEATURES:

* Add `file.Service.Search()` for the file search endpoint (`POST /files/search/`), with full-text, exact, metadata, tag, size, upload-time and image filters, sorting, and `include=appdata` support; results come as a paginated `file.SearchResult` iterator yielding `file.SearchMatch` values (file fields plus match highlights; CDN base rewrite for `original_file_url`). Custom implementations and mocks of the `file.Service` interface must add the new method
* Export `ucare.ErrEndOfResults`, returned by result iterators when no results are left to read

## 2.0.0 (May 28, 2026)

BREAKING CHANGES:

* Target REST API v0.7 (previously v0.5); remove the `APIv05` and `APIv06` constants
* Minimum Go version is now 1.25
* Throttled requests no longer retry by default — automatic retries are now opt-in via `ucare.Config.Retry`
* `file.Service.Info()` now takes `*file.InfoParams` to pass optional `include` query parameters
* `webhook.Service.Delete()` now deletes by webhook ID instead of by target URL
* Remove `RecognitionInfo` field from `file.Info` — use `AppData` instead
* Remove `ImageInfo` and `VideoMeta` fields from `file.BasicFileInfo` — use `ContentInfo.Image` and `ContentInfo.Video`
* Remove `file.Copy()` method and `file.CopyParams` type — use `LocalCopy()` and `RemoteCopy()`
* Remove `group.Store()` method (endpoint removed in v0.7)
* Remove `file.OrderBySizeAsc` and `file.OrderBySizeDesc` constants (not supported in v0.7)

FEATURES:

* Add `projectapi` package for the Project API, with bearer token authentication via `ucare.NewBearerConfig()` and `ucare.NewBearerClient()` — manage projects, project features, secret keys, usage metrics, moderation thresholds, and meta reference lists for MIME types and moderation categories
* Add `addon` package for Addons API execution and status polling, with typed params for Remove.bg and ClamAV requests
* Add `metadata` package with file metadata CRUD operations
* Add `upload.Service.Upload()` for automatic direct-vs-multipart upload selection, with metadata support across direct, multipart, from-URL, and unified uploads
* Resolve a per-project CDN base URL automatically when `ucare.Config.CDNBase` is empty (overridable with an explicit absolute URL), and rewrite the scheme/host of API-returned URLs — `file.Info.OriginalFileURL`, `group.Info.CDNLink`, and `upload.GroupInfo.CDNLink` — to point at it while preserving the full path; exposed via the `ucare.ClientCDNBase(Client)` and `ucare.RewriteCDNURL(originalURL, cdnBase)` helpers
* Export structured error types for inspecting HTTP status and detail: `APIError`, `AuthError`, `ThrottleError`, `ValidationError`, `ForbiddenError`, and Project API equivalents `ProjectAPIError`, `ProjectAuthError`, `ProjectForbiddenError`
* Add `ucare.Config.Retry` and `RetryConfig` for configurable throttling retries
* Add `Metadata` and `AppData` fields to `file.Info`, plus `file.InfoParams.Include` and `file.ListParams.Include` for requesting `include=appdata`
* Add `group.Delete()` for deleting group metadata without deleting files
* Add `conversion.Params.SaveInGroup` to persist multi-page document conversion output as a file group, and `conversion.BuildDocumentPath()` / `conversion.BuildVideoPath()` helpers for constructing conversion paths
* Add webhook event constants for `file.stored`, `file.deleted`, `file.info_updated`, and deprecated `file.infected`
* Add typed Project API usage metric constants for `traffic`, `storage`, and `operations`

IMPROVEMENTS:

* Add `UserAgent` field to `ucare.Config` for custom agent identification
* Throttle retries now use the server `Retry-After` when present, falling back to exponential backoff (capped at 30s), respect context cancellation, and cap the effective wait via `MaxWaitSeconds`
* Extend form/query encoding to support Upload API metadata fields in `metadata[key]=value` bracket notation
* Replace `http.NewRequest` + `WithContext` with `http.NewRequestWithContext`
* Replace `ioutil` usage with `io` equivalents
* Replace `go-env` dependency with `os.Getenv`
* Update `stretchr/testify` to v1.10.0
* Update CI: Go 1.25, modern GitHub Actions versions, remove deprecated golint
* Fix errors in package documentation examples and update public examples for the new `file.Info()` signature

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
