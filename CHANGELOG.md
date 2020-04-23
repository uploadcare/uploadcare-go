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
