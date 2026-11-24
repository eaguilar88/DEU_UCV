package backups

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/eaguilar88/deu/internal/backups/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var fixedNow = time.Date(2026, 11, 20, 7, 0, 0, 0, time.UTC)

func newTestService(t *testing.T, retention time.Duration) (*Service, *mocks.MockDumper, *mocks.MockStorage) {
	dumper := mocks.NewMockDumper(t)
	storage := mocks.NewMockStorage(t)
	svc := NewService(dumper, storage, Config{
		DatabaseName: "deu",
		Prefix:       "backups/postgres",
		Retention:    retention,
	}, zap.NewNop())
	svc.now = func() time.Time { return fixedNow }
	svc.tempDir = t.TempDir()
	return svc, dumper, storage
}

func dumpWrites(content string) func(context.Context, io.Writer) error {
	return func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, content)
		return err
	}
}

func TestRun_UploadsDumpAndDeletesExpired(t *testing.T) {
	svc, dumper, storage := newTestService(t, 30*24*time.Hour)
	const key = "backups/postgres/deu_20261120_070000.dump"

	dumper.EXPECT().Dump(mock.Anything, mock.Anything).RunAndReturn(dumpWrites("PGDMP-data"))
	storage.EXPECT().PutObject(mock.Anything, key, dumpContentType, mock.Anything).
		RunAndReturn(func(_ context.Context, _, _ string, body io.Reader) error {
			data, err := io.ReadAll(body)
			require.NoError(t, err)
			assert.Equal(t, "PGDMP-data", string(data), "the uploaded body is the whole dump")
			return nil
		})
	storage.EXPECT().ListObjects(mock.Anything, "backups/postgres/").Return([]entities.StoredObject{
		{Key: "backups/postgres/deu_20261001_070000.dump", LastModified: fixedNow.AddDate(0, 0, -50)}, // expired
		{Key: "backups/postgres/deu_20261019_070000.dump", LastModified: fixedNow.AddDate(0, 0, -32)}, // expired
		{Key: "backups/postgres/deu_20261101_070000.dump", LastModified: fixedNow.AddDate(0, 0, -19)}, // kept
		{Key: "backups/postgres/notes.txt", LastModified: fixedNow.AddDate(-1, 0, 0)},                 // not a dump
		{Key: key, LastModified: fixedNow},
	}, nil)
	storage.EXPECT().DeleteFile(mock.Anything, "backups/postgres/deu_20261001_070000.dump").Return(nil)
	storage.EXPECT().DeleteFile(mock.Anything, "backups/postgres/deu_20261019_070000.dump").Return(nil)

	result, err := svc.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, key, result.Key)
	assert.Equal(t, int64(len("PGDMP-data")), result.Size)
	assert.Equal(t, []string{
		"backups/postgres/deu_20261001_070000.dump",
		"backups/postgres/deu_20261019_070000.dump",
	}, result.Deleted)
	assertNoTempFiles(t, svc.tempDir)
}

func TestRun_DumpFailsNothingUploaded(t *testing.T) {
	svc, dumper, _ := newTestService(t, 30*24*time.Hour)
	dumper.EXPECT().Dump(mock.Anything, mock.Anything).Return(errors.New("connection refused"))

	_, err := svc.Run(context.Background())

	require.ErrorContains(t, err, "connection refused")
	assertNoTempFiles(t, svc.tempDir)
}

func TestRun_EmptyDumpNotUploaded(t *testing.T) {
	svc, dumper, _ := newTestService(t, 30*24*time.Hour)
	dumper.EXPECT().Dump(mock.Anything, mock.Anything).RunAndReturn(dumpWrites(""))

	_, err := svc.Run(context.Background())

	require.ErrorContains(t, err, "empty file")
}

func TestRun_UploadFailsNoCleanup(t *testing.T) {
	svc, dumper, storage := newTestService(t, 30*24*time.Hour)
	dumper.EXPECT().Dump(mock.Anything, mock.Anything).RunAndReturn(dumpWrites("PGDMP"))
	storage.EXPECT().PutObject(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("b2 down"))

	_, err := svc.Run(context.Background())

	require.ErrorContains(t, err, "b2 down")
}

func TestRun_CleanupErrorKeepsBackupResult(t *testing.T) {
	svc, dumper, storage := newTestService(t, 7*24*time.Hour)
	dumper.EXPECT().Dump(mock.Anything, mock.Anything).RunAndReturn(dumpWrites("PGDMP"))
	storage.EXPECT().PutObject(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	storage.EXPECT().ListObjects(mock.Anything, mock.Anything).Return([]entities.StoredObject{
		{Key: "backups/postgres/deu_a.dump", LastModified: fixedNow.AddDate(0, 0, -10)},
		{Key: "backups/postgres/deu_b.dump", LastModified: fixedNow.AddDate(0, 0, -9)},
	}, nil)
	storage.EXPECT().DeleteFile(mock.Anything, "backups/postgres/deu_a.dump").Return(errors.New("forbidden"))
	storage.EXPECT().DeleteFile(mock.Anything, "backups/postgres/deu_b.dump").Return(nil)

	result, err := svc.Run(context.Background())

	require.ErrorContains(t, err, "forbidden")
	assert.Equal(t, "backups/postgres/deu_20261120_070000.dump", result.Key, "the new backup is still reported")
	assert.Equal(t, []string{"backups/postgres/deu_b.dump"}, result.Deleted)
}

func TestRun_NoRetentionKeepsEverything(t *testing.T) {
	svc, dumper, storage := newTestService(t, 0)
	dumper.EXPECT().Dump(mock.Anything, mock.Anything).RunAndReturn(dumpWrites("PGDMP"))
	storage.EXPECT().PutObject(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// ListObjects and DeleteFile must not be called: the mock fails the test if they are.

	result, err := svc.Run(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result.Deleted)
}

func TestRun_NeverDeletesTheNewBackup(t *testing.T) {
	svc, dumper, storage := newTestService(t, time.Hour)
	const key = "backups/postgres/deu_20261120_070000.dump"
	dumper.EXPECT().Dump(mock.Anything, mock.Anything).RunAndReturn(dumpWrites("PGDMP"))
	storage.EXPECT().PutObject(mock.Anything, key, mock.Anything, mock.Anything).Return(nil)
	// B2's clock reports the new object as older than the retention: it must still be kept.
	storage.EXPECT().ListObjects(mock.Anything, mock.Anything).Return([]entities.StoredObject{
		{Key: key, LastModified: fixedNow.Add(-2 * time.Hour)},
	}, nil)

	result, err := svc.Run(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result.Deleted)
}

func TestPgDumpArgs(t *testing.T) {
	p := PgDump{Host: "db", Port: 5432, User: "deu", Password: "secret", Database: "deu"}

	args := p.args()

	assert.Equal(t, []string{
		"--format=custom", "--no-owner", "--no-privileges",
		"--host", "db", "--port", "5432", "--username", "deu", "--dbname", "deu",
	}, args)
	assert.NotContains(t, args, "secret", "the password is passed through PGPASSWORD, not argv")
}

func TestPgDump_ReportsStderr(t *testing.T) {
	p := PgDump{Binary: "sh", Database: "deu"}
	// "sh --format=custom ..." fails; the error must carry the command's stderr.
	err := p.Dump(context.Background(), io.Discard)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "pg_dump")
}

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries, "the temporary dump file is removed")
}
