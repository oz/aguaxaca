package collector

// Collector is a basic interface for types that can download images to
// be imported.
type Collector interface {
	DownloadImages() ([]string, error)
}
