package cliex

import (
	"errors"
	"net/http"
	"time"
)

// ServerErrorResponse is the error response from server (try to guess what it is)
type ServerErrorResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Details string `json:"details,omitempty"`
	Text    string `json:"text,omitempty"`
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Err     string `json:"err,omitempty"`
}

// RequestOpts is the options for resty client request.
type RequestOpts struct {
	// Method is the HTTP method to use.
	Method string

	// Headers is the headers of the request.
	Headers map[string]string

	// Query is the query string of the request.
	Query map[string]string

	// PathParams is the path parameters of the request, e.g. /v1/users/{userId} and userId is a path parameter
	// {"userId": "sample@sample.com"}
	PathParams map[string]string

	// Cookies is the cookies of the request.
	Cookies []*http.Cookie

	// FormData is the form data of the request.
	FormData map[string]string

	// Files is the files of the request, where key is fila name and value is file path.
	Files map[string]string

	// AuthToken is the token for authentication
	AuthToken string

	// BasicAuthUser is the user for basic authentication
	BasicAuthUser string

	// BasicAuthPass is the password for basic authentication
	BasicAuthPass string

	// ForceContentType tell Resty to parse a custom response (e.g. JSON if application/json) into your struct.
	ForceContentType string

	// Body is the body of the request
	Body any

	// Result is the variable where the response body will be stored
	Result any

	// OutputPath is the path to the output file where will be saved the response.
	OutputPath string

	// RequestName is the name of the request for logging retries.
	RequestName string

	// RetryCount is the number of times to retry the request.
	RetryCount int

	// RetryWaitTime is the starting wait time between retries.
	// Default is 100 milliseconds.
	RetryWaitTime time.Duration

	// RetryMaxWaitTime is the maximum wait time between retries.
	// Default is 2 seconds.
	RetryMaxWaitTime time.Duration

	// InfiniteRetry is whether to retry the request infinitely
	InfiniteRetry bool

	// RetryOnlyServerErrors is whether to retry only 5xx errors.
	RetryOnlyServerErrors bool

	// NoLogRetryError is whether to log the retry error
	NoLogRetryError bool

	// EnableTrace is whether to enable trace and return it in resp.Request.TraceInfo().
	EnableTrace bool
}

var (
	// ErrBadRequest is when the server cannot or will not process the request due to a client error
	// (e.g., malformed request syntax, size too large, invalid request message framing, or deceptive request routing).
	ErrBadRequest = errors.New("code 400, bad request")

	// ErrUnauthorized is used when authentication is required and has failed or has not yet been provided.
	// The response must include a WWW-Authenticate header field containing a challenge applicable to the requested resource.
	ErrUnauthorized = errors.New("code 401, unauthorized")

	// ErrPaymentRequired is reserved for future use. This code might indicate a digital cash or micropayment requirement.
	ErrPaymentRequired = errors.New("code 402, payment required")

	// ErrForbidden indicates that the server refuses to authorize the request, even though the server understands it.
	// This could be because the user does not have permissions for a resource, or because of another form of prohibition.
	ErrForbidden = errors.New("code 403, forbidden")

	// ErrNotFound indicates that the server can't find the requested resource. Further requests are allowable.
	ErrNotFound = errors.New("code 404, not found")

	// ErrMethodNotAllowed is when a request method is not supported for the requested resource.
	// For example, a GET request on a form designed only to be POSTed.
	ErrMethodNotAllowed = errors.New("code 405, method not allowed")

	// ErrNotAcceptable indicates that the resource is only capable of generating content not acceptable by the Accept headers.
	ErrNotAcceptable = errors.New("code 406, not acceptable")

	// ErrProxyAuthRequired suggests that the client must first authenticate itself with the proxy.
	ErrProxyAuthRequired = errors.New("code 407, proxy authentication required")

	// ErrRequestTimeout means the server timed out waiting for the request.
	// "The client did not produce a request within the time that the server was prepared to wait. The client MAY repeat the request without modifications at any later time."
	ErrRequestTimeout = errors.New("code 408, request timeout")

	// ErrConflict indicates the request could not be processed due to a conflict in the current state of the resource, such as a simultaneous edit conflict.
	ErrConflict = errors.New("code 409, conflict")

	// ErrGone signals that the resource requested is no longer available and will not be available again.
	// The client should not request the resource in the future.
	ErrGone = errors.New("code 410, gone")

	// ErrLengthRequired indicates that the request did not specify the length of its content, which is required.
	ErrLengthRequired = errors.New("code 411, length required")

	// ErrPreconditionFailed demonstrates the server does not meet one of the preconditions given in the request headers.
	ErrPreconditionFailed = errors.New("code 412, precondition failed")

	// ErrPayloadTooLarge denotes that the request is larger than the server is willing or able to process.
	ErrPayloadTooLarge = errors.New("code 413, payload too large")

	// ErrURITooLong means the URI provided was too long for the server to process.
	ErrURITooLong = errors.New("code 414, URI too long")

	// ErrUnsupportedMediaType is when the request entity has a media type which the server or resource does not support.
	ErrUnsupportedMediaType = errors.New("code 415, unsupported media type")

	// ErrRangeNotSatisfiable indicates the client has asked for a portion of the file, but the server cannot supply that portion.
	ErrRangeNotSatisfiable = errors.New("code 416, range not satisfiable")

	// ErrExpectationFailed implies the server cannot meet the requirements of the Expect request-header field.
	ErrExpectationFailed = errors.New("code 417, expectation failed")

	// ErrTeapot is an Easter egg response code indicating the server is a teapot, not capable of brewing coffee.
	ErrTeapot = errors.New("code 418, I'm a teapot")

	// ErrMisdirectedRequest indicates the request was directed at a server that is not able to produce a response.
	ErrMisdirectedRequest = errors.New("code 421, misdirected request")

	// ErrUnprocessableEntity means the request was well-formed but could not be processed by the server.
	ErrUnprocessableEntity = errors.New("code 422, unprocessable entity")

	// ErrLocked suggests the resource that is being accessed is locked.
	ErrLocked = errors.New("code 423, locked")

	// ErrFailedDependency indicates the request failed because it depended on another request and that request failed.
	ErrFailedDependency = errors.New("code 424, failed dependency")

	// ErrTooEarly indicates the server is unwilling to process a request that might be replayed.
	ErrTooEarly = errors.New("code 425, too early")

	// ErrUpgradeRequired indicates the client should switch to a different protocol, as suggested in the Upgrade header.
	ErrUpgradeRequired = errors.New("code 426, upgrade required")

	// ErrPreconditionRequired is used when the server requires that the request is conditional.
	// This is to prevent 'lost update' problems when a client GETs a resource's state, modifies it, and PUTs it back to the server.
	ErrPreconditionRequired = errors.New("code 428, precondition required")

	// ErrTooManyRequests indicates the user has sent too many requests in a given time frame, usually indicative of a rate-limiting policy.
	ErrTooManyRequests = errors.New("code 429, too many requests")

	// ErrHeaderFieldsTooLarge suggests that either an individual header field, or all the header fields collectively, are too large.
	ErrHeaderFieldsTooLarge = errors.New("code 431, request header fields too large")

	// ErrUnavailableForLegalReasons indicates the requested resource is unavailable due to legal reasons.
	ErrUnavailableForLegalReasons = errors.New("code 451, unavailable for legal reasons")
)

var (
	// ErrInternalServerError represents the HTTP 500 error.
	// It is a generic error message, given when an unexpected condition was encountered
	// and no more specific message is suitable.
	ErrInternalServerError = errors.New("code 500, internal server error")

	// ErrNotImplemented represents the HTTP 501 error.
	// The server either does not recognize the request method, or it lacks the ability to fulfill the request.
	// Usually, this implies future availability, such as a new feature of a web-service API.
	ErrNotImplemented = errors.New("code 501, not implemented")

	// ErrBadGateway represents the HTTP 502 error.
	// The server was acting as a gateway or proxy and received an invalid response from the upstream server.
	ErrBadGateway = errors.New("code 502, bad gateway")

	// ErrServiceUnavailable represents the HTTP 503 error.
	// The server is currently unable to handle the request due to a temporary overloading or maintenance of the server.
	ErrServiceUnavailable = errors.New("code 503, service unavailable")

	// ErrGatewayTimeout represents the HTTP 504 error.
	// The server was acting as a gateway or proxy and did not receive a timely response from the upstream server.
	ErrGatewayTimeout = errors.New("code 504, gateway timeout")

	// ErrHTTPVersionNotSupported represents the HTTP 505 error.
	// The server does not support, or refuses to support, the HTTP protocol version that was used in the request message.
	ErrHTTPVersionNotSupported = errors.New("code 505, HTTP version not supported")

	// ErrVariantAlsoNegotiates represents the HTTP 506 error.
	// Transparent content negotiation for the request results in a circular reference.
	ErrVariantAlsoNegotiates = errors.New("code 506, variant also negotiates")

	// ErrInsufficientStorage represents the HTTP 507 error.
	// The server is unable to store the representation needed to complete the request.
	ErrInsufficientStorage = errors.New("code 507, insufficient storage")

	// ErrLoopDetected represents the HTTP 508 error.
	// The server detected an infinite loop while processing the request.
	ErrLoopDetected = errors.New("code 508, loop detected")

	// ErrNotExtended represents the HTTP 510 error.
	// Further extensions to the request are required for the server to fulfill it.
	ErrNotExtended = errors.New("code 510, not extended")

	// ErrNetworkAuthenticationRequired represents the HTTP 511 error.
	// The client needs to authenticate to gain network access, often used by intercepting
	// proxies used to control access to the network, e.g., for "captive portal" purposes.
	ErrNetworkAuthenticationRequired = errors.New("code 511, network authentication required")
)

var (
	// 4xx: Client Errors
	// The request contains bad syntax or cannot be fulfilled.

	// ErrPageExpired is used by the Laravel Framework when a CSRF Token is missing or expired.
	ErrPageExpired = errors.New("code 419, page expired")

	// ErrMethodFailure is used by the Spring Framework when a method has failed.
	ErrMethodFailure = errors.New("code 420, method failure")

	// ErrEnhanceYourCalm used by Twitter's API to signal that the client is being rate-limited.
	ErrEnhanceYourCalm = errors.New("code 420, enhance your calm")

	// ErrRequestHeaderFieldsTooLarge was used by Shopify when too many URLs were requested in a timeframe.
	ErrRequestHeaderFieldsTooLarge = errors.New("code 430, request headers fields too large")

	// ErrShopifySecurityRejection is returned by Shopify if the request was deemed malicious.
	ErrShopifySecurityRejection = errors.New("code 430, security rejection")

	// ErrBlockedByWindowsParentalControls signifies being blocked by Microsoft parental controls.
	ErrBlockedByWindowsParentalControls = errors.New("code 450, blocked by Windows Parental Controls")

	// ErrInvalidToken indicates an expired or otherwise invalid token in ArcGIS.
	ErrInvalidToken = errors.New("code 498, invalid token")

	// ErrTokenRequired indicates that a token is required but was not submitted in ArcGIS.
	ErrTokenRequired = errors.New("code 499, token required")

	// 5xx: Server Errors
	// The server failed to fulfill a valid request.

	// ErrBandwidthLimitExceeded indicates that the server's bandwidth limit has been exceeded.
	ErrBandwidthLimitExceeded = errors.New("code 509, bandwidth limit exceeded")

	// ErrSiteIsOverloaded is used by Qualys when the site cannot process the request due to overload.
	ErrSiteIsOverloaded = errors.New("code 529, site is overloaded")

	// ErrSiteIsFrozen indicates a site that has been frozen due to inactivity.
	ErrSiteIsFrozen = errors.New("code 530, site is frozen")

	// ErrOriginDNSError is used by Shopify to indicate that Cloudflare can't resolve the requested DNS record.
	ErrOriginDNSError = errors.New("code 530, origin DNS error")

	// ErrTemporarilyDisabled is returned by Shopify when an endpoint has been temporarily disabled.
	ErrTemporarilyDisabled = errors.New("code 540, temporarily disabled")

	// ErrNetworkReadTimeoutError is an informal convention used by proxies to indicate a read timeout.
	ErrNetworkReadTimeoutError = errors.New("code 598, network read timeout error")

	// ErrNetworkConnectTimeoutError indicates a network connect timeout.
	ErrNetworkConnectTimeoutError = errors.New("code 599, network connect timeout error")

	// ErrUnexpectedToken is used by Shopify to indicate a JSON syntax error.
	ErrUnexpectedToken = errors.New("code 783, unexpected token")

	// ErrNonStandard is a placeholder for LinkedIn's restriction message.
	ErrNonStandard = errors.New("code 999, non-standard or restricted access error")

	// ErrLoginTimeout is when the client's session has expired and must log in again.
	ErrLoginTimeout = errors.New("code 440, login timeout")

	// ErrRetryWith may be returned to ask for additional information from the client.
	ErrRetryWith = errors.New("code 449, retry with")

	// ErrRedirect in Exchange ActiveSync contexts for mailbox access or server availability.
	ErrRedirect = errors.New("code 451, redirect for efficient server")

	// ErrNoResponse indicates no information returned, with the connection closed immediately.
	ErrNoResponse = errors.New("code 444, no response")

	// ErrRequestHeaderTooLarge signals an excessively large request or header line.
	ErrRequestHeaderTooLarge = errors.New("code 494, request header too large")

	// ErrSSLCertificateError indicates a problem with an SSL certificate provided by the client.
	ErrSSLCertificateError = errors.New("code 495, SSL Certificate Error")

	// ErrSSLCertificateRequired indicates a missing client SSL certificate where required.
	ErrSSLCertificateRequired = errors.New("code 496, SSL Certificate Required")

	// ErrHTTPtoHTTPS denotes a bad request on a wrong port.
	ErrHTTPtoHTTPS = errors.New("code 497, HTTP request sent to HTTPS port")

	// ErrClientClosedRequest arises when a client closes a request before a server responds.
	ErrClientClosedRequest = errors.New("code 499, client closed request")

	// Cloudflare 5xx Errors
	// These represent various issues with origin servers, often pertaining to Cloudflare users.

	// ErrWebServerReturnedUnknownError is an empty or unexpected response error by Cloudflare.
	ErrWebServerReturnedUnknownError = errors.New("code 520, web server returned an unknown error")

	// ErrWebServerIsDown is when a server refuses connections from Cloudflare.
	ErrWebServerIsDown = errors.New("code 521, web server is down")

	// ErrConnectionTimedOut as per contact issues between Cloudflare and the origin server.
	ErrConnectionTimedOut = errors.New("code 522, connection timed out")

	// ErrOriginIsUnreachable indicates DNS or network errors preventing access to the origin server.
	ErrOriginIsUnreachable = errors.New("code 523, origin is unreachable")

	// ErrTimeoutOccurred reflects a failure to receive timely HTTP response despite a TCP connection.
	ErrTimeoutOccurred = errors.New("code 524, a timeout occurred")

	// ErrSSLHandshakeFailed is a failed SSL/TLS handshake negotiation.
	ErrSSLHandshakeFailed = errors.New("code 525, SSL handshake failed")

	// ErrInvalidSSLCertificate indicates issues with validating the SSL certificate on the origin server.
	ErrInvalidSSLCertificate = errors.New("code 526, invalid SSL certificate")

	// Deprecated and other specific cases

	// ErrRailgunError was related to interrupted Railgun connections, now deprecated.
	ErrRailgunError = errors.New("deprecated code 527, Railgun error")

	// Elastic Load Balancer specific errors

	// ErrClientClosedConnection signifies early termination by a client.
	ErrClientClosedConnection = errors.New("code 460, client closed the connection")

	// ErrTooManyXForwardedFor is when too many IPs are listed in the X-Forwarded-For header.
	ErrTooManyXForwardedFor = errors.New("code 463, too many X-Forwarded-For headers")

	// ErrIncompatibleProtocolVersions flags a protocol version mismatch.
	ErrIncompatibleProtocolVersions = errors.New("code 464, incompatible protocol versions")

	// ErrUnauthorized relates to authentication errors in load-balanced services.
	ErrUnauthorizedElastic = errors.New("code 561, unauthorized access")
)

// Mapping of HTTP status codes to their corresponding errors.
var ErrorMapping = map[int]error{
	400: ErrBadRequest,
	401: ErrUnauthorized,
	402: ErrPaymentRequired,
	403: ErrForbidden,
	404: ErrNotFound,
	405: ErrMethodNotAllowed,
	406: ErrNotAcceptable,
	407: ErrProxyAuthRequired,
	408: ErrRequestTimeout,
	409: ErrConflict,
	410: ErrGone,
	411: ErrLengthRequired,
	412: ErrPreconditionFailed,
	413: ErrPayloadTooLarge,
	414: ErrURITooLong,
	415: ErrUnsupportedMediaType,
	416: ErrRangeNotSatisfiable,
	417: ErrExpectationFailed,
	418: ErrTeapot,
	421: ErrMisdirectedRequest,
	422: ErrUnprocessableEntity,
	423: ErrLocked,
	424: ErrFailedDependency,
	425: ErrTooEarly,
	426: ErrUpgradeRequired,
	428: ErrPreconditionRequired,
	429: ErrTooManyRequests,
	431: ErrHeaderFieldsTooLarge,
	451: ErrUnavailableForLegalReasons,

	500: ErrInternalServerError,
	501: ErrNotImplemented,
	502: ErrBadGateway,
	503: ErrServiceUnavailable,
	504: ErrGatewayTimeout,
	505: ErrHTTPVersionNotSupported,
	506: ErrVariantAlsoNegotiates,
	507: ErrInsufficientStorage,
	508: ErrLoopDetected,
	510: ErrNotExtended,
	511: ErrNetworkAuthenticationRequired,

	419: ErrPageExpired,
	420: ErrMethodFailure, // Spring Framework
	430: ErrRequestHeaderFieldsTooLarge,
	450: ErrBlockedByWindowsParentalControls,
	498: ErrInvalidToken,
	499: ErrTokenRequired,
	509: ErrBandwidthLimitExceeded,
	529: ErrSiteIsOverloaded,
	530: ErrSiteIsFrozen,
	540: ErrTemporarilyDisabled,
	598: ErrNetworkReadTimeoutError,
	599: ErrNetworkConnectTimeoutError,
	783: ErrUnexpectedToken,
	999: ErrNonStandard, // LinkedIn, etc.
	440: ErrLoginTimeout,
	449: ErrRetryWith,
	444: ErrNoResponse,
	494: ErrRequestHeaderTooLarge,
	495: ErrSSLCertificateError,
	496: ErrSSLCertificateRequired,
	497: ErrHTTPtoHTTPS,
	520: ErrWebServerReturnedUnknownError,
	521: ErrWebServerIsDown,
	522: ErrConnectionTimedOut,
	523: ErrOriginIsUnreachable,
	524: ErrTimeoutOccurred,
	525: ErrSSLHandshakeFailed,
	526: ErrInvalidSSLCertificate,
	527: ErrRailgunError,
	460: ErrClientClosedConnection,
	463: ErrTooManyXForwardedFor,
	464: ErrIncompatibleProtocolVersions,
	561: ErrUnauthorizedElastic,
}

const (
	// AAC audio
	MIMETypeAAC = "audio/aac"

	// AbiWord document
	MIMETypeABW = "application/x-abiword"

	// Animated Portable Network Graphics (APNG)
	MIMETypeAPNG = "image/apng"

	// Archive document (multiple files embedded)
	MIMETypeARC = "application/x-freearc"

	// AVIF image
	MIMETypeAVIF = "image/avif"

	// AVI: Audio Video Interleave
	MIMETypeAVI = "video/x-msvideo"

	// Amazon Kindle eBook format
	MIMETypeAZW = "application/vnd.amazon.ebook"

	// Any kind of binary data
	MIMETypeBIN = "application/octet-stream"

	// Windows OS/2 Bitmap Graphics
	MIMETypeBMP = "image/bmp"

	// BZip archive
	MIMETypeBZ = "application/x-bzip"

	// BZip2 archive
	MIMETypeBZ2 = "application/x-bzip2"

	// CD audio
	MIMETypeCDA = "application/x-cdf"

	// C-Shell script
	MIMETypeCSH = "application/x-csh"

	// Cascading Style Sheets (CSS)
	MIMETypeCSS = "text/css"

	// Comma-separated values (CSV)
	MIMETypeCSV = "text/csv"

	// Microsoft Word
	MIMETypeDOC = "application/msword"

	// Microsoft Word (OpenXML)
	MIMETypeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

	// MS Embedded OpenType fonts
	MIMETypeEOT = "application/vnd.ms-fontobject"

	// Electronic publication (EPUB)
	MIMETypeEPUB = "application/epub+zip"

	// GZip Compressed Archive
	MIMETypeGZ = "application/gzip"

	// Graphics Interchange Format (GIF)
	MIMETypeGIF = "image/gif"

	// HyperText Markup Language (HTML)
	MIMETypeHTML = "text/html"

	// Icon format
	MIMETypeICO = "image/vnd.microsoft.icon"

	// iCalendar format
	MIMETypeICS = "text/calendar"

	// Java Archive (JAR)
	MIMETypeJAR = "application/java-archive"

	// JPEG images
	MIMETypeJPEG = "image/jpeg"

	// JavaScript
	MIMETypeJS = "text/javascript"

	// JSON format
	MIMETypeJSON = "application/json"

	// JSON-LD format
	MIMETypeJSONLD = "application/ld+json"

	// Musical Instrument Digital Interface (MIDI)
	MIMETypeMIDI = "audio/midi"

	// JavaScript module
	MIMETypeMJS = "text/javascript"

	// MP3 audio
	MIMETypeMP3 = "audio/mpeg"

	// MP4 video
	MIMETypeMP4 = "video/mp4"

	// MPEG Video
	MIMETypeMPEG = "video/mpeg"

	// Apple Installer Package
	MIMETypeMPKG = "application/vnd.apple.installer+xml"

	// OpenDocument presentation document
	MIMETypeODP = "application/vnd.oasis.opendocument.presentation"

	// OpenDocument spreadsheet document
	MIMETypeODS = "application/vnd.oasis.opendocument.spreadsheet"

	// OpenDocument text document
	MIMETypeODT = "application/vnd.oasis.opendocument.text"

	// Ogg audio
	MIMETypeOGA = "audio/ogg"

	// Ogg video
	MIMETypeOGV = "video/ogg"

	// Ogg
	MIMETypeOGX = "application/ogg"

	// Opus audio in Ogg container
	MIMETypeOPUS = "audio/ogg"

	// OpenType font
	MIMETypeOTF = "font/otf"

	// Portable Network Graphics
	MIMETypePNG = "image/png"

	// Adobe Portable Document Format (PDF)
	MIMETypePDF = "application/pdf"

	// Hypertext Preprocessor (Personal Home Page)
	MIMETypePHP = "application/x-httpd-php"

	// Microsoft PowerPoint
	MIMETypePPT = "application/vnd.ms-powerpoint"

	// Microsoft PowerPoint (OpenXML)
	MIMETypePPTX = "application/vnd.openxmlformats-officedocument.presentationml.presentation"

	// RAR archive
	MIMETypeRAR = "application/vnd.rar"

	// Rich Text Format (RTF)
	MIMETypeRTF = "application/rtf"

	// Bourne shell script
	MIMETypeSH = "application/x-sh"

	// Scalable Vector Graphics (SVG)
	MIMETypeSVG = "image/svg+xml"

	// Tape Archive (TAR)
	MIMETypeTAR = "application/x-tar"

	// Tagged Image File Format (TIFF)
	MIMETypeTIFF = "image/tiff"

	// MPEG transport stream
	MIMETypeTS = "video/mp2t"

	// TrueType Font
	MIMETypeTTF = "font/ttf"

	// Text, (generally ASCII or ISO 8859-n)
	MIMETypeTXT = "text/plain"

	// Microsoft Visio
	MIMETypeVSD = "application/vnd.visio"

	// Waveform Audio Format
	MIMETypeWAV = "audio/wav"

	// WEBM audio
	MIMETypeWEBA = "audio/webm"

	// WEBM video
	MIMETypeWEBM = "video/webm"

	// WEBP image
	MIMETypeWEBP = "image/webp"

	// Web Open Font Format (WOFF)
	MIMETypeWOFF = "font/woff"

	// Web Open Font Format (WOFF2)
	MIMETypeWOFF2 = "font/woff2"

	// XHTML
	MIMETypeXHTML = "application/xhtml+xml"

	// Microsoft Excel
	MIMETypeXLS = "application/vnd.ms-excel"

	// Microsoft Excel (OpenXML)
	MIMETypeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	// XML
	MIMETypeXML = "application/xml"

	// XUL
	MIMETypeXUL = "application/vnd.mozilla.xul+xml"

	// ZIP archive
	MIMETypeZIP = "application/zip"

	// 3GPP audio/video container
	MIMEType3GP = "video/3gpp"

	// 3GPP2 audio/video container
	MIMEType3G2 = "video/3gpp2"

	// 7-zip archive
	MIMEType7Z = "application/x-7z-compressed"
)
