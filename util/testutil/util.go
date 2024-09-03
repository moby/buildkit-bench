package testutil

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func runBuildkitd(conf *BackendConfig, args []string, logs map[string]*bytes.Buffer, extraEnv []string) (address string, cl func() error, err error) {
	deferF := &MultiCloser{}
	cl = deferF.F()

	defer func() {
		if err != nil {
			deferF.F()()
			cl = nil
		}
	}()

	tmpdir, err := os.MkdirTemp("", "bkbench_buildkitd")
	if err != nil {
		return "", nil, err
	}

	if err := os.MkdirAll(filepath.Join(tmpdir, "tmp"), 0711); err != nil {
		return "", nil, err
	}

	deferF.Append(func() error { return os.RemoveAll(tmpdir) })

	cfgfile, err := writeConfig(append(conf.DaemonConfig))
	if err != nil {
		return "", nil, err
	}
	deferF.Append(func() error {
		return os.RemoveAll(filepath.Dir(cfgfile))
	})

	args = append(args, "--config="+cfgfile)
	address = getBuildkitdAddr(tmpdir)

	args = append(args, "--root", tmpdir, "--addr", address, "--debug")
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec // test utility
	cmd.Env = append(
		os.Environ(),
		"BUILDKIT_DEBUG_EXEC_OUTPUT=1",
		"BUILDKIT_DEBUG_PANIC_ON_ERROR=1",
		"TMPDIR="+filepath.Join(tmpdir, "tmp"))
	if v := os.Getenv("GO_TEST_COVERPROFILE"); v != "" {
		coverDir := filepath.Join(filepath.Dir(v), "helpers")
		cmd.Env = append(cmd.Env, "GOCOVERDIR="+coverDir)
	}
	cmd.Env = append(cmd.Env, extraEnv...)
	cmd.SysProcAttr = getSysProcAttr()

	stop, err := startCmd(cmd, logs)
	if err != nil {
		return "", nil, err
	}
	deferF.Append(stop)

	if err := waitSocket(address, 15*time.Second, cmd); err != nil {
		return "", nil, err
	}

	// separated out since it's not required in windows
	deferF.Append(func() error {
		return mountInfo(tmpdir)
	})

	return address, cl, err
}

func startCmd(cmd *exec.Cmd, logs map[string]*bytes.Buffer) (func() error, error) {
	if logs != nil {
		setCmdLogs(cmd, logs)
	}

	fmt.Fprintf(cmd.Stderr, "> StartCmd %v %+v\n", time.Now(), cmd.String())

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	eg, ctx := errgroup.WithContext(context.TODO())

	stopped := make(chan struct{})
	stop := make(chan struct{})
	eg.Go(func() error {
		err := cmd.Wait()
		fmt.Fprintf(cmd.Stderr, "> stopped %v %+v %v\n", time.Now(), cmd.ProcessState, cmd.ProcessState.ExitCode())
		close(stopped)
		select {
		case <-stop:
			return nil
		default:
			return err
		}
	})

	eg.Go(func() error {
		select {
		case <-ctx.Done():
		case <-stopped:
		case <-stop:
			signal := syscall.SIGTERM
			signalStr := "SIGTERM"
			fmt.Fprintf(cmd.Stderr, "> sending sigterm %v\n", time.Now())
			fmt.Fprintf(cmd.Stderr, "> sending %s %v\n", signalStr, time.Now())
			cmd.Process.Signal(signal)
			go func() {
				select {
				case <-stopped:
				case <-time.After(20 * time.Second):
					cmd.Process.Kill()
				}
			}()
		}
		return nil
	})

	return func() error {
		close(stop)
		return eg.Wait()
	}, nil
}

// WaitSocket will dial a socket opened by a command passed in as cmd.
func waitSocket(address string, d time.Duration, cmd *exec.Cmd) error {
	address = strings.TrimPrefix(address, "unix://")
	step := 50 * time.Millisecond
	i := 0
	for {
		if cmd != nil && cmd.ProcessState != nil {
			return errors.Errorf("process exited: %s", cmd.String())
		}
		if conn, err := dialPipe(address); err == nil {
			conn.Close()
			break
		}
		i++
		if time.Duration(i)*step > d {
			return errors.Errorf("failed dialing: %s", address)
		}
		time.Sleep(step)
	}
	return nil
}

func getSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true, // stretch sudo needs this for sigterm
	}
}

func getBuildkitdAddr(tmpdir string) string {
	return "unix://" + filepath.Join(tmpdir, "buildkitd.sock")
}

func mountInfo(tmpdir string) error {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return errors.Wrap(err, "failed to open mountinfo")
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if strings.Contains(s.Text(), tmpdir) {
			return errors.Errorf("leaked mountpoint for %s", tmpdir)
		}
	}
	return s.Err()
}

// abstracted function to handle pipe dialing on unix.
// some simplification has been made to discard
// laddr for unix -- left as nil.
func dialPipe(address string) (net.Conn, error) {
	addr, err := net.ResolveUnixAddr("unix", address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed resolving unix addr: %s", address)
	}
	return net.DialUnix("unix", nil, addr)
}

type MultiCloser struct {
	fns []func() error
}

func (mc *MultiCloser) F() func() error {
	return func() error {
		var err error
		for i := range mc.fns {
			if err1 := mc.fns[len(mc.fns)-1-i](); err == nil {
				err = err1
			}
		}
		mc.fns = nil
		return err
	}
}

func (mc *MultiCloser) Append(f func() error) {
	mc.fns = append(mc.fns, f)
}

func setCmdLogs(cmd *exec.Cmd, logs map[string]*bytes.Buffer) {
	b := new(bytes.Buffer)
	logs["stdout: "+cmd.String()] = b
	cmd.Stdout = &lockingWriter{Writer: b}
	b = new(bytes.Buffer)
	logs["stderr: "+cmd.String()] = b
	cmd.Stderr = &lockingWriter{Writer: b}
}

type lockingWriter struct {
	mu sync.Mutex
	io.Writer
}

func (w *lockingWriter) Write(dt []byte) (int, error) {
	w.mu.Lock()
	n, err := w.Writer.Write(dt)
	w.mu.Unlock()
	return n, err
}
