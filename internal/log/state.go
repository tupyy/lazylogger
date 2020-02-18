package log

const (
	HEALTHY = iota

	// DEGRADED means the fetcher is still running but there is a problem with the file
	DEGRADED = iota

	// FAILED means that the fetcher is stopped
	FAILED = iota

	// STOPPED means the logger has been stopped from a healthy state.
	STOPPED = iota
)

// State represents the state of the logger.
type State struct {
	ID     int
	Health int
	Err    error
}

func (s *State) String() string {
	switch s.Health {
	default:
		return "unknown"
	case HEALTHY:
		return "healthy"
	case DEGRADED:
		return "degraded"
	case FAILED:
		return "failed"
	case STOPPED:
		return "stopped"
	}
}

func NewState(id int) *State {
	return &State{id, HEALTHY, nil}
}

// HandleStateChange change the state. If stderr is not nil it means that logger has a problem reading the file and
// the logger will have a DEGRADED health. If err is not nil it means the connection is down and logger is supposed
// to be FAILED.
// HandleStateChange returns true if state changed.
func (state *State) HandleStateChange(stderr, err error) bool {
	oldHealth := state.Health
	if stderr == nil && err == nil {
		state.Health = HEALTHY
	} else if stderr != nil && err == nil {
		state.Health = DEGRADED
		state.Err = stderr
	} else {
		state.Health = FAILED
		state.Err = err
	}

	return oldHealth != state.Health
}
