package modes

import (
	"fmt"
	"os/exec"

	"github.com/kolesnikovae/go-winjob"
)

func runInJobObject(cmd *exec.Cmd) error {
	job, err := winjob.Start(cmd,
		winjob.WithKillOnJobClose(),
		winjob.WithActiveProcessLimit(2),
		winjob.WithProcessMemoryLimit(256<<20),
	)
	if err != nil {
		return fmt.Errorf("winjob.Start: %w", err)
	}
	defer job.Close()

	return cmd.Wait()
}
