package parser

// DocsURLBase is the root URL for the version of the docs associated with this
// binary.
var DocsURLBase = "https://www.cockroachlabs.com/docs/" + "v20.1.11"

// DocsURL generates the URL to pageName in the version of the docs associated
// with this binary.
func DocsURL(pageName string) string { return DocsURLBase + "/" + pageName }

