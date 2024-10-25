package esox

import "net/http"

type URL struct {
	Name    string
	Handler http.Handler
	Path    string
}

type URLs []URL

// AddURL returns a new URLs with the given URL added.
func (urls URLs) AddURL(url URL) URLs {
	out := make(URLs, len(urls)+1)
	copy(out, urls)
	out[len(urls)] = url
	return out
}

// AddURLs returns a new URLs with the given URLSet added.
func (urls URLs) AddURLs(urlSet URLs) URLs {
	out := make(URLs, len(urls), len(urls)+len(urlSet))
	copy(out, urls)
	return append(out, urlSet...)
}

// WithPrefix returns a new URLs with the given prefix added to each URL's path.
func (urls URLs) WithPrefix(prefix string) URLs {
	out := make(URLs, 0, len(urls))
	for _, url := range urls {
		out = append(out, URL{
			Name:    url.Name,
			Handler: url.Handler,
			Path:    prefix + url.Path,
		})
	}

	return out
}
