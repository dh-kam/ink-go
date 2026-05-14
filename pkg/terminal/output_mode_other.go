//go:build !linux

package terminal

func captureOutputMode(_ int) (func() error, error) {
	return func() error {
		return nil
	}, nil
}
