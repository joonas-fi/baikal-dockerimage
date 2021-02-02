package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/net/http/httputils"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/yookoala/gofast"
)

func main() {
	rootLogger := logex.StandardLogger()

	osutil.ExitIfError(logic(
		osutil.CancelOnInterruptOrTerminate(rootLogger),
		rootLogger))
}

func logic(ctx context.Context, logger *log.Logger) error {
	cgiTcpAddr := "127.0.0.1:8081"

	// ensure DB dir exists (convenience util so user can mount empty directory from Docker host)
	if err := os.MkdirAll("Specific/db", osutil.FileMode(osutil.OwnerRWX, osutil.GroupRWX, osutil.OtherNone)); err != nil {
		return err
	}

	// gofast would also serve static files, but it doesn't respect MIME types
	staticFilesBaikal := http.FileServer(http.Dir("html/"))

	routes := http.NewServeMux()
	routes.Handle("/res/", staticFilesBaikal)

	routes.Handle("/.well-known/caldav", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dav.php", http.StatusPermanentRedirect)
	}))
	routes.Handle("/.well-known/carddav", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dav.php", http.StatusPermanentRedirect)
	}))

	routes.Handle("/", phpFastCgiHandler("html/", "tcp", cgiTcpAddr))

	srv := &http.Server{
		Addr:    ":80",
		Handler: routes,
	}

	tasks := taskrunner.New(ctx, logger)

	tasks.Start("listener "+srv.Addr, func(ctx context.Context) error {
		return httputils.CancelableServer(ctx, srv, func() error { return srv.ListenAndServe() })
	})

	tasks.Start("php-fastcgi", func(ctx context.Context) error {
		return exec.CommandContext(
			ctx,
			"php-cgi",
			"-b", cgiTcpAddr,
		).Run()
	})

	return tasks.Wait()
}

func phpFastCgiHandler(docroot string, network string, address string) http.Handler {
	return gofast.NewHandler(
		gofast.NewPHPFS(docroot)(func(client gofast.Client, req *gofast.Request) (*gofast.ResponsePipe, error) {
			// we need to lie we're using HTTPS, because the app for some reason makes absolute URLs
			// and thus is needs to detect the scheme, and behind a proxy it can't know for sure.
			// TODO: x-forwarded-proto or some other kind?
			req.Params["HTTPS"] = "on" // TODO: is this really "on" and not e.g. "true"?
			return client.Do(req)
		}),
		gofast.SimpleClientFactory(gofast.SimpleConnFactory(network, address), 0),
	)
}
