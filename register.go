package eventsourcing

import (
	"reflect"
)

type registerFunc = func() interface{}
type RegisterFunc = func(events ...interface{})

type register struct {
	r map[string]registerFunc
}

func newRegister() *register {
	return &register{
		r: make(map[string]registerFunc),
	}
}

// Type return the func to generate the correct event data type
func (r *register) Type(typ, reason string) (registerFunc, bool) {
	d, ok := r.r[typ+"_"+reason]
	return d, ok
}

func (r *register) Register(a aggregate) {
	typ := reflect.TypeOf(a).Elem().Name()

	// fe is a helper function to make the event type registration simpler
	fe := func(events ...interface{}) []registerFunc {
		res := []registerFunc{}
		for _, e := range events {
			res = append(res, eventToFunc(e))
		}
		return res
	}

	fu := func(events ...interface{}) {
		eventsF := fe(events...)
		for _, f := range eventsF {
			event := f()
			reason := reflect.TypeOf(event).Elem().Name()
			r.r[typ+"_"+reason] = f
		}
	}
	a.Register(fu)
}

func eventToFunc(event interface{}) registerFunc {
	return func() interface{} { return event }
}
