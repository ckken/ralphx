package validate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ckken/ralphx/internal/execx"
)

func Run(ctx context.Context, workdir, command, logPath string) error {
	if command == "" {
		return nil
	}
	res, err := execx.Run(ctx, "bash", []string{"-lc", command}, nil, workdir)
	if logPath != "" {
		_ = os.MkdirAll(filepath.Dir(logPath), 0o755)
		_ = os.WriteFile(logPath, res.Output, 0o644)
	}
	if err != nil {
		return fmt.Errorf("validation command failed: %w", err)
	}
	return nil
}
