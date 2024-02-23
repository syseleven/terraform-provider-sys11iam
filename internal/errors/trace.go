package errors

import (
	"gitlab.syseleven.de/ncs/terraform-provider-ncs/internal/logging"
	"runtime"
)

func Trace(err error) error {
	if err != nil {
		pc, filename, line, _ := runtime.Caller(1)
		logging.Errorf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), filename, line, err)
	}
	return err
}
