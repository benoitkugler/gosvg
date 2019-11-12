package gosvg

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Utils dealing with URLs.

var (
	userAgent = "gosvg " + VERSION
	URL       = regexp.MustCompile(`url\((.+)\)`)
)

// Normalize ``rawurl`` for underlying NT/Unix operating systems.
// The input ``rawurl`` may look like the following:
//     - C:\\Directory\\zzz.svg
//     - file://C:\\Directory\\zzz.svg
//     - zzz.svg
// The output ``rawurl`` on NT systems would look like below:
//     - file:///C:/Directory/zzz.svg
func normalizeUrl(rawurl string) string {
	if rawurl != "" && runtime.GOOS == "windows" && !strings.HasPrefix(rawurl, "data:") {
		// Match input ``rawurl`` like the following:
		//   - C:\\Directory\\zzz.svg
		//   - Blah.svg
		if filepath.IsAbs(rawurl) {
			rawurl = filepath.ToSlash(rawurl)
		}
	}
	return rawurl
}

// Fetch the content of ``url``.
//  ``resourceType`` is the mimetype of the resource (currently one of
//  image/*, image/svg+xml, text/css).
func fetch(urlTarget, _ string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, urlTarget, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return ioutil.ReadAll(response.Body)
}

// Parse an URL.
// The URL can be surrounded by a ``url()`` string. If ``base`` is not `None`,
// the "folder" part of it is prepended to the URL.
// func parseUrl(rawurl, base string) *url.URL {
// if rawurl != "" {
//     match = URL.Find(rawurl)
//     if len(match) > 0 {
//         rawurl = match[1]
// 	}
// 	if base {
//         parsedBase = urlparse(base)
//         parsedUrl = urlparse(rawurl)
//         if parsedBase.scheme := range ("", "file") {
//             if parsedUrl.scheme := range ("", "file") {
//                 parsedBasePath = ntCompatiblePath(parsedBase.path)
//                 parsedUrlPath = ntCompatiblePath(parsedUrl.path)
//                 // We are sure that `rawurl` && `base` are both file-like URLs
//                 if os.path.isfile(parsedBasePath) {
//                     if parsedUrlPath {
//                         // Take the "folder" part of `base`, as
//                         // `os.path.join` doesn"t strip the file name
//                         rawurl = os.path.join(
//                             os.path.dirname(parsedBasePath),
//                             parsedUrlPath)
//                     } else {
//                         rawurl = parsedBasePath
//                     }
//                 } else if os.path.isdir(parsedBasePath) {
//                     if parsedUrlPath {
//                         rawurl = os.path.join(
//                             parsedBasePath, parsedUrlPath)
//                     } else {
//                         rawurl = ""
//                     }
//                 } else {
//                     rawurl = ""
//                 } if parsedUrl.fragment {
//                     rawurl = "{}#{}".format(rawurl, parsedUrl.fragment)
//                 }
//             }
//         } else if parsedUrl.scheme := range ("", parsedBase.scheme) {
//             // `urljoin` automatically uses the "folder" part of `base`
//             rawurl = urljoin(base, rawurl)
//         }
// 	}
// 	rawurl = normalizeUrl(rawurl.strip(`\""`))
// }
// return urlparse(rawurl || "")
// }

// Get bytes :in a parsed ``url`` using ``urlFetcher``.
// If ``urlFetcher`` == nil  a default (no limitations) URLFetcher is used.
func readUrl(u *url.URL, urlFetcher UrlFectcher, resourceType string) ([]byte, error) {
	var target string
	if u.Scheme != "" {
		target = u.String()
	} else {
		p, err := filepath.Abs(u.String())
		if err != nil {
			return nil, err
		}
		target = "file://" + p
		target = normalizeUrl(target)
	}
	return urlFetcher(target, resourceType)
}
