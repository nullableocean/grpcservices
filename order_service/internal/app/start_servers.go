package app

import (
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func (app *App) startGrpcServer(errChan chan<- error) error {
	lis, err := net.Listen("tcp", app.config.App.Address+":"+app.config.App.Port)
	if err != nil {
		return fmt.Errorf("create listen tcp error: %w", err)
	}

	go func() {
		app.logger.Info("order grpc server started", zap.String("address", app.config.App.Address+":"+app.config.App.Port))

		err = app.grpc.server.Serve(lis)
		if err != nil {
			errChan <- fmt.Errorf("start serve grpc error: %w", err)
		}
	}()

	return nil
}

func (app *App) startHttpServer(errChan chan<- error) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(app.prometheus.reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	app.http.server = &http.Server{
		Addr:    ":" + app.config.Metrics.Port,
		Handler: mux,
	}

	go func() {
		app.logger.Info("start listen metrics http", zap.String("address", app.config.App.Address+":"+app.config.Metrics.Port))

		err := app.http.server.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()
}
