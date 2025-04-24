package errors

import (
	"runtime"

	"github.com/syseleven/terraform-provider-sys11iam/internal/logging"
)

func Trace(err error) error {
	if err != nil {
		pc, filename, line, _ := runtime.Caller(1)
		logging.Errorf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), filename, line, err)
	}
	return err
}
