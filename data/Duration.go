package data

import (
	"errors"
	"math/rand"
	"time"
)

type Duration struct {
	time.Duration
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v interface{}
	if err := unmarshal(&v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		d.Duration = tmp
		return nil
	default:
		return errors.New("invalid duration")
	}
}

// RandomDuration returns a random duration between min and max, but not contained within not.
func RandomDuration(min, max Duration, not []Duration) Duration {
	if min == max {
		return max
	}
	var r Duration
	if len(not) > 0 {
		for done := false; !done; {
			r.Duration = time.Duration(rand.Intn(int(max.Duration)-int(min.Duration))) + min.Duration
			match := 0
			for _, i := range not {
				if r == i {
					match++
				}
			}
			if match == 0 {
				done = true
			}
		}
		return r
	}
	r.Duration = time.Duration(rand.Intn(int(max.Duration)-int(min.Duration))) + min.Duration
	return r
}
