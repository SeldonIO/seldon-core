package main

import (
	"flag"
	"context"
	"github.com/prometheus/common/log"
	"net/http"
	"os"
	"fmt"
     "github.com/gorilla/mux"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"github.com/seldonio/seldon-core/executor/api/client"
	"os/signal"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"syscall"
	"time"
)

var (
	configPath   = flag.String("config", "", "Path to kubconfig")
	sdepName   = flag.String("sdep", "", "Seldon deployment name")
	namespace   = flag.String("namespace", "", "Namespace")
	port        = flag.Int("port", 8080, "Executor port")
	wait        = flag.Duration( "graceful-timeout", time.Second * 15, "Graceful shutdown secs")
)

func ArticlesCategoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Category: %v\n", vars["category"])
}

func main() {
	flag.Parse()

	if *sdepName == "" {
		log.Error("Seldon deployment name must be provided")
		os.Exit(-1)
	}

	if *namespace == "" {
		log.Error("Namespace must be provied")
		os.Exit(-1)
	}

	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("entrypoint")

	seldonDeploymentClient := client.NewSeldonDeploymentClient(configPath)
	sdep, err := seldonDeploymentClient.Get(*sdepName,*namespace)
	if err != nil {
		log.Error(err,"Failed to find seldon deployment","name",sdepName)
		panic(err)
	} else {
		log.Info("Found seldon deployment","name",sdep.Name)
	}


	seldonRest := rest.NewSeldonRestApi(seldonDeploymentClient)
	seldonRest.Initialise()

	//http.Handle("/", router)

	srv := &http.Server{
		Handler: seldonRest.Router,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error(err, "Server error")
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), *wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Info("shutting down")
	os.Exit(0)
}
