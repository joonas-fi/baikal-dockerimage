package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/net/http/httputils"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/gorilla/mux"
	"github.com/yookoala/gofast"
)

func main() {
	rootLogger := logex.StandardLogger()

	osutil.ExitIfError(logic(
		osutil.CancelOnInterruptOrTerminate(rootLogger),
		rootLogger))
}

func logic(ctx context.Context, logger *log.Logger) error {
	cgiTcpAddr := "127.0.0.1:9000"

	// ensure DB dir exists (convenience util so user can mount empty directory from Docker host)
	if err := os.MkdirAll("Specific/db", osutil.FileMode(osutil.OwnerRWX, osutil.GroupRWX, osutil.OtherNone)); err != nil {
		return err
	}

	phpCgiBackend := gofast.SimpleClientFactory(gofast.SimpleConnFactory("tcp", cgiTcpAddr), 0)

	// gofast would also serve static files (or rather, I guess it hands them to the CGI server for serving),
	// but it doesn't respect MIME types
	staticFilesBaikal := http.FileServer(http.Dir("html/"))

	// need Mux because http.NewServeMux() cannot set route for exact "/" (uses prefix => a huge catch-all)
	routes := mux.NewRouter()

	routes.PathPrefix("/res/").Handler(staticFilesBaikal)

	routes.Handle("/.well-known/caldav", permanentRedirectTo("/dav.php"))
	routes.Handle("/.well-known/carddav", permanentRedirectTo("/dav.php"))

	// declaring each PHP file separately to skip gofast.NewPHPFS() which has a serious security vulnerability
	phpEndpoint(routes, "/cal.php", phpCgiBackend)
	phpEndpoint(routes, "/dav.php", phpCgiBackend)
	phpEndpoint(routes, "/card.php", phpCgiBackend)
	phpEndpoint(routes, "/admin/index.php", phpCgiBackend)
	phpEndpoint(routes, "/admin/install/index.php", phpCgiBackend)
	phpEndpoint(routes, "/index.php", phpCgiBackend)

	srv := &http.Server{
		Addr:    ":80",
		Handler: routes,
	}

	tasks := taskrunner.New(ctx, logger)

	tasks.Start("listener "+srv.Addr, func(ctx context.Context) error {
		return httputils.CancelableServer(ctx, srv, func() error { return srv.ListenAndServe() })
	})

	tasks.Start("php-fastcgi", func(ctx context.Context) error {
		// php-cgi would've been simpler, but it crapped itself when Thunderbird made too many requests.
		// there might be a problem in gofast (I witnessed large amount of TIME_WAIT connections).
		//
		// php-fpm7 fixed it, so let's go with it for now.
		phpCgi := exec.CommandContext(ctx, "php-fpm7", "--nodaemonize")
		phpCgi.Stdout = os.Stdout
		phpCgi.Stderr = os.Stderr

		return phpCgi.Run()
	})

	return tasks.Wait()
}

func phpEndpoint(routes *mux.Router, phpUrl string, phpCgiBackend gofast.ClientFactory) {
	requestChain := gofast.Chain(
		gofast.BasicParamsMap, // CONTENT_TYPE, REMOTE_ADDR, REQUEST_METHOD etc.
		gofast.MapHeader,      // headers like HTTP_HOST, HTTP_AUTHORIZATION etc.
		lieAboutXForwardedProto,
		gofast.MapEndpoint("html"+phpUrl)) // SCRIPT_NAME, DOCUMENT_ROOT etc.

	phpHandler := gofast.NewHandler(requestChain(gofast.BasicSession), phpCgiBackend)

	// "/dav.php" is also used as "/dav.php/calendars/USERNAME"
	routes.PathPrefix(phpUrl).Handler(phpHandler)

	// "/admin/index.php" => register also "/admin/"
	if strings.HasSuffix(phpUrl, "index.php") {
		routes.Handle(
			strings.TrimSuffix(phpUrl, "index.php"),
			phpHandler)
	}
}

func permanentRedirectTo(location string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, location, http.StatusPermanentRedirect)
	})
}

func lieAboutXForwardedProto(inner gofast.SessionHandler) gofast.SessionHandler {
	return func(client gofast.Client, req *gofast.Request) (*gofast.ResponsePipe, error) {
		// we need to lie we're using HTTPS, because the app for some reason makes absolute URLs
		// and thus it needs to detect the scheme, and behind a proxy it can't know for sure.
		//
		// TODO: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Proto ?
		//       but Go's reverse proxy doesn't seem to add that
		//       https://github.com/yookoala/gofast/issues/47
		//       https://github.com/sabre-io/Baikal/issues/1001
		req.Params["HTTP_X_FORWARDED_PROTO"] = "https" // Go's reverse proxy doesn't set this https://github.com/golang/go/issues/30963

		return inner(client, req)
	}
}
