package bbhash_test

import (
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relab/iago"
	fs "github.com/relab/wrfs"
)

func TestIago(t *testing.T) {
	// dir, _ := os.Getwd()

	hosts := []string{"bbchain1"}
	g, err := iago.NewSSHGroup(hosts, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		g.Close()
	})
	execPath, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	exe, err := filepath.Abs(execPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Executable path: %s", exe)

	g.Run("Create temporary directory on remote hosts",
		func(ctx context.Context, host iago.Host) (err error) {
			testDir := tempDirPath(host, "bbhash."+randString(8))
			dataDir := filepath.Join(testDir, "data")
			host.SetVar("test-dir", testDir)
			host.SetVar("data-dir", dataDir)
			return fs.MkdirAll(host.GetFS(), dataDir, 0o755)
		})

	g.Run(
		"Upload benchmark binary to remote hosts",
		func(ctx context.Context, host iago.Host) (err error) {
			dest, err := iago.NewPath("/", iago.GetStringVar(host, "test-dir")+"/bbhash")
			if err != nil {
				return err
			}
			host.SetVar("exe", dest.String())
			src, err := iago.NewPathFromAbs(exe)
			if err != nil {
				return err
			}
			return iago.Upload{
				Src:  src,
				Dest: dest,
				Perm: iago.NewPerm(0o755),
			}.Apply(ctx, host)
		})

	g.Run("Start benchmark binary",
		func(ctx context.Context, host iago.Host) (err error) {
			cmd, err := host.NewCommand()
			if err != nil {
				return err
			}
			// stdin, err := cmd.StdinPipe()
			// if err != nil {
			// 	return err
			// }

			// stdout, err := cmd.StdoutPipe()
			// if err != nil {
			// 	return err
			// }

			// stderr, err := cmd.StderrPipe()
			// if err != nil {
			// 	return err
			// }

			var sb strings.Builder
			sb.WriteString(iago.GetStringVar(host, "exe"))
			// sb.WriteString(" ")

			if err = cmd.Start(iago.Expand(host, sb.String())); err != nil {
				return err
			}

			return nil
		})

	g.Run("Download files", func(ctx context.Context, host iago.Host) error {
		return nil
	})
}

func tempDirPath(host iago.Host, dirName string) string {
	tmp := host.GetEnv("TMPDIR")
	if tmp == "" {
		tmp = "/tmp"
	}
	return filepath.Join(tmp, dirName)
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func randString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rnd.Intn(len(letters))]
	}
	return string(s)
}
