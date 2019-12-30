package clock

import "time"

type Interface interface {
	Now() time.Time
}
