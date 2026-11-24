package backups

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// PgDump runs the pg_dump binary (postgresql-client in the backup image) against the database.
// It produces a custom-format archive (-Fc): compressed, and restorable table by table with
// pg_restore.
type PgDump struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	// Binary is the pg_dump executable; empty means "pg_dump" from PATH.
	Binary string
}

// Dump streams the archive to w. The password goes through PGPASSWORD so it never appears in the
// process list.
func (p PgDump) Dump(ctx context.Context, w io.Writer) error {
	binary := p.Binary
	if binary == "" {
		binary = "pg_dump"
	}

	cmd := exec.CommandContext(ctx, binary, p.args()...)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+p.Password)
	cmd.Stdout = w
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func (p PgDump) args() []string {
	return []string{
		"--format=custom",
		"--no-owner",
		"--no-privileges",
		"--host", p.Host,
		"--port", strconv.Itoa(p.Port),
		"--username", p.User,
		"--dbname", p.Database,
	}
}
