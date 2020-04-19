package datasource

import "errors"

var (
	// ErrNofile means the read operation failed for some reason
	ErrRead = errors.New("read error")

	// ErrDatasource means the underlining datasource has crashed
	ErrDatasource = errors.New("datasource error")
)
