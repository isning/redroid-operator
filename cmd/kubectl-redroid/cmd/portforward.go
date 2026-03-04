package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// portForwardOptions holds everything needed to set up a single port-forward session.
type portForwardOptions struct {
	REST        *rest.Config
	Namespace   string
	ServiceName string // Service name to forward to (operator always creates one per instance)
	LocalPort   int
	PodPort     int
	Out         io.Writer
	ErrOut      io.Writer
	// StopCh is closed by the caller to terminate forwarding. If nil a
	// SIGINT/SIGTERM handler is installed automatically.
	StopCh <-chan struct{}
	// ReadyCh is closed when the tunnel is ready. If nil an internal channel is used.
	ReadyCh chan struct{}
}

// portForwardResult is returned when the ready channel fires or tunnelling ends.
type portForwardResult struct {
	LocalPort int
	StopCh    chan struct{} // caller closes to tear down
}

// startPortForward starts a port-forward and returns once the tunnel is ready.
// The caller must close result.StopCh to terminate the tunnel.
// If opts.StopCh is nil (interactive mode) the function blocks until the user sends SIGINT/SIGTERM.
func startPortForward(opts portForwardOptions) (*portForwardResult, error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(opts.REST)
	if err != nil {
		return nil, fmt.Errorf("build SPDY round-tripper: %w", err)
	}

	// Forward to the Service rather than a Pod directly. The operator guarantees
	// a ClusterIP Service named "redroid-instance-<name>" exists per instance,
	// so this survives pod restarts without the CLI tracking pod names.
	path := fmt.Sprintf("/api/v1/namespaces/%s/services/%s/portforward", opts.Namespace, opts.ServiceName)
	serverURL, err := url.Parse(opts.REST.Host)
	if err != nil {
		return nil, err
	}
	serverURL.Path = path

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, serverURL)

	stopCh := opts.StopCh
	interactive := stopCh == nil
	internalStop := make(chan struct{})
	if interactive {
		stopCh = internalStop
	}

	readyCh := opts.ReadyCh
	if readyCh == nil {
		readyCh = make(chan struct{})
	}

	out := opts.Out
	if out == nil {
		out = io.Discard
	}
	errOut := opts.ErrOut
	if errOut == nil {
		errOut = os.Stderr
	}

	ports := []string{fmt.Sprintf("%d:%d", opts.LocalPort, opts.PodPort)}
	fw, err := portforward.New(dialer, ports, stopCh, readyCh, out, errOut)
	if err != nil {
		return nil, fmt.Errorf("create port-forwarder: %w", err)
	}

	result := &portForwardResult{
		LocalPort: opts.LocalPort,
		StopCh:    internalStop,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- fw.ForwardPorts() }()

	if interactive {
		// Wait for ready then hand control to the user (block until Ctrl-C).
		select {
		case <-readyCh:
		case err := <-errCh:
			return nil, err
		}

		_, _ = fmt.Fprintf(opts.Out, "Forwarding %s → localhost:%d\nPress Ctrl-C to stop.\n",
			opts.ServiceName, opts.LocalPort)

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sig:
		case err := <-errCh:
			return nil, err
		}
		close(internalStop)
		return nil, nil
	}

	// Non-interactive: wait until ready, then hand back to caller.
	select {
	case <-readyCh:
	case err := <-errCh:
		return nil, err
	}

	// Collect tunnelling errors in background.
	go func() {
		if err := <-errCh; err != nil {
			_, _ = fmt.Fprintln(errOut, "port-forward error:", err)
		}
	}()

	return result, nil
}
