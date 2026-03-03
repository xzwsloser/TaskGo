package event

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

const (
	EXIT = "exit"
	WAIT = "wait"
)

var (
	_events = make(map[string][]func(any), 2)
)

func OnEvent(name string, fs ...func(any)) error {
	evs, ok := _events[name]
	if !ok {
		evs = make([]func(any), 0, len(fs))
	}

	for _, f := range fs {
		if f == nil {
			continue
		}

		fp := reflect.ValueOf(f).Pointer()
		for i := 0; i < len(evs); i++ {
			if reflect.ValueOf(evs[i]).Pointer() == fp {
				return fmt.Errorf("func[%v] already exists in event[%s]", fp, name)
			}
		}
		evs = append(evs, f)
	}

	_events[name] = evs
	return nil
}

func EmitEvent(name string, arg any) {
	evs, ok := _events[name]
	if !ok {
		return
	}

	for _, f := range evs {
		f(arg)
	}
}

func EmitAllEvent(arg any) {
	for _, fs := range _events {
		for _, f := range fs {
			f(arg)
		}
	}
}

func OffEvent(name string, f func(any)) error {
	evs, ok := _events[name]
	if !ok || len(evs) == 0 {
		return fmt.Errorf("envet[%s] doesn't have any funcs", name)
	}

	fp := reflect.ValueOf(f).Pointer()
	for i := 0; i < len(evs); i++ {
		if reflect.ValueOf(evs[i]).Pointer() == fp {
			evs = append(evs[:i], evs[i+1:]...)
			_events[name] = evs
			return nil
		}
	}

	return fmt.Errorf("%v func dones't exist in event[%s]", fp, name)
}

func OffAllEvent(name string) error {
	_events[name] = nil
	return nil
}

// @Description: Wait For OS Signal, Grace Stop Server
func WaitEvent(sig ...os.Signal) os.Signal {
	c := make(chan os.Signal, 1)
	if len(sig) == 0 {
		signal.Notify(c, syscall.SIGHUP, 
						 syscall.SIGINT, 
			             syscall.SIGTERM, 
					     syscall.SIGQUIT)
	} else {
		signal.Notify(c, sig...)
	}
	return <-c
}
