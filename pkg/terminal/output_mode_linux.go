//go:build linux

package terminal

import "golang.org/x/sys/unix"

func captureOutputMode(fd int) (func() error, error) {
	original, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}

	return func() error {
		current, err := unix.IoctlGetTermios(fd, unix.TCGETS)
		if err != nil {
			return err
		}

		current.Oflag = original.Oflag
		return unix.IoctlSetTermios(fd, unix.TCSETS, current)
	}, nil
}
