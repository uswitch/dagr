package execute

import (
	"github.com/uswitch/dagr/program"
)

type Execution struct {
}

func Execute(baseDir string, program *program.Program) (Execution, error) {
	return Execution{}, nil
}
