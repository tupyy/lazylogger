package reader

// DataNotification wil notify all the registered clients about new data arrived in the cache.
// Size is the new size of the cache. It the cache is full (e.g. size = 10Mb) the new size
// is set to 10Mb and the previousSize = size - number of bytes put in cache by the logger.
type DataNotification struct {
	ID           int
	Size         int64
	PreviousSize int64
}
