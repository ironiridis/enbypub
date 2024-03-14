package enbypub

import "strings"

type ContentTypesExtraT struct {
	FromExtension map[string]string
}

var ContentTypesExtra *ContentTypesExtraT

// ContentTypeFromExtension returns a (guessed) content-type suitable for an HTTP header based on
// a given file extension, or an empty string.
// If ContentTypesExtra has been set, it will be referenced first.
func ContentTypeFromExtension(ext string) string {
	// main ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types

	// strip leading dots
	for strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}
	// drop case
	ext = strings.ToLower(ext)

	// allow override if specified
	if ContentTypesExtra != nil {
		if ContentTypesExtra.FromExtension != nil {
			if t := ContentTypesExtra.FromExtension[ext]; t != "" {
				return t
			}
		}
	}

	// we build in some very common types here...
	switch ext {

	// the web, in general
	case "htm":
	case "html":
		return "text/html"
	case "css":
		return "text/css"
	case "js":
	case "mjs":
		return "text/javascript"
	case "json":
		return "application/json"
	case "jsonld":
		return "application/ld+json"
	case "xml":
		return "application/xml"

	// common document formats
	case "txt":
		return "text/plain"
	case "csv":
		return "text/csv"
	case "tab":
	case "tsv":
		return "text/tab-separated-values"
	case "rtf":
		return "application/rtf"
	case "epub":
		return "application/epub+zip"

	// ODF document formats
	case "odp":
	case "fodp":
		return "application/vnd.oasis.opendocument.presentation"
	case "ods":
	case "fods":
		return "application/vnd.oasis.opendocument.spreadsheet"
	case "odt":
	case "fodt":
		return "application/vnd.oasis.opendocument.text"
	case "odg":
	case "fodg":
		return "application/vnd.oasis.opendocument.graphics"

	// adobe document format (singular)
	case "pdf":
		return "application/pdf"

	// microsoft document formats
	case "ppt":
		return "application/vnd.ms-powerpoint"
	case "pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case "xls":
		return "application/vnd.ms-excel"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "doc":
		return "application/msword"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "vsd":
		return "application/vnd.visio"
	case "vsdx":
		return "application/vnd.visio2013"

	// font formats
	case "ttf":
		return "font/ttf"
	case "otf":
		return "font/otf"
	case "eot":
		return "application/vnd.ms-fontobject"
	case "woff":
		return "font/woff"
	case "woff2":
		return "font/woff2"

	// graphic formats
	case "jpg":
	case "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "png":
		return "image/png"
	case "svg":
		return "image/svg+xml"
	case "webp":
		return "image/webp"
	case "ico":
		return "image/vnd.microsoft.icon"

	// audio formats
	case "aac":
		return "audio/aac"
	case "m4a":
		return "audio/mp4"
	case "mp3":
		return "audio/mpeg"
	case "oga":
	case "ogg":
		return "audio/ogg"
	case "opus":
		return "audio/opus"
	case "wav":
		return "audio/wav"
	case "weba":
		return "audio/webm" // not a typo

	// media (video or video+audio) formats
	case "mp4":
	case "m4v":
		return "video/mp4"
	case "mkv":
		return "video/x-matroska"
	case "ogv":
		return "video/ogg"
	case "webm":
		return "video/webm"
	case "ogx":
		return "application/ogg"

	// archive/compressed formats
	case "gz":
		return "application/gzip"
	case "zip":
		return "application/zip"
	case "bz2":
		return "application/x-bzip2"
	case "7z":
		return "application/x-7z-compressed"
	case "xz":
		return "application/x-xz"
	case "tar":
		return "application/x-tar"
	case "jar":
		return "application/java-archive"
	}

	return ""
}
