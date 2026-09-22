package main

// Disposable in-process coupled-rehearsal server. It initializes the VM with
// an in-memory database, serves the exact VM HTTP surface on an explicit
// loopback address, and accepts pending transitions through the real
// BuildBlock/Verify/Accept path. It exists so the FIELD runtime can drive the
// coupled lifecycle against the real VM without any network, chain creation,
// or persisted state. It must never bind a non-loopback address.

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ava-labs/avalanchego/database/memdb"
	"github.com/ava-labs/avalanchego/snow"
)

func serveRehearsal(genesisPath, address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("rehearsal address must be host:port: %w", err)
	}
	if host != "127.0.0.1" && host != "::1" {
		return errors.New("rehearsal server only binds explicit loopback addresses")
	}
	genesisBytes, err := os.ReadFile(genesisPath)
	if err != nil {
		return fmt.Errorf("read genesis: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	vm := new(VM)
	if err := vm.Initialize(ctx, nil, memdb.New(), genesisBytes, nil, nil, nil, nil); err != nil {
		return fmt.Errorf("initialize VM: %w", err)
	}
	if err := vm.SetState(ctx, snow.NormalOp); err != nil {
		return fmt.Errorf("enter normal operation: %w", err)
	}

	go func() {
		for {
			if _, err := vm.WaitForEvent(ctx); err != nil {
				return
			}
			block, err := vm.BuildBlock(ctx)
			if err != nil {
				continue
			}
			if err := block.Verify(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "rehearsal block verify: %v\n", err)
				continue
			}
			if err := block.Accept(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "rehearsal block accept: %v\n", err)
				continue
			}
		}
	}()

	server := &http.Server{
		Addr:              address,
		Handler:           vm.apiHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	fmt.Printf("rehearsal VM serving on http://%s (loopback only, in-memory, disposable)\n", listener.Addr().String())
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	if err := vm.Shutdown(context.Background()); err != nil && !strings.Contains(err.Error(), "closed") {
		return err
	}
	return nil
}
