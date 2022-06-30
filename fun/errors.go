package fun

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
)

// RecoverToErrorVar recovers and places the recovered error into the given variable
func RecoverToErrorVar(name string, err *error) {
	err2 := recover()
	if err2 != nil {
		log.Printf("RecoverToErrorVar(%s) (err=%+v), (err2: %+v)\n", name, *err, err2)
		switch err2 := err2.(type) {
		case error:
			err4 := errors.Wrapf(err2, "%s: Recover from panic", name)
			*err = err4
		case string:
			err4 := errors.New(name + ": Recover from string-panic: " + err2)
			*err = err4
		default:
			err4 := fmt.Errorf("%s: Recover from unknown-panic: %+v", name, err2)
			*err = err4
		}
	}
}

// RecoverToLog in case of error just logs it.
func RecoverToLog(name string) {
	err2 := recover()
	if err2 != nil {
		log.Printf("RecoverToLog(%s) (err2: %+v\n", name, err2)
	}
}
