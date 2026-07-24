package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cgi" //nolint:gosec // deliberately implementing a CGI handler
)

/*
REQUEST_METHOD=GET \
QUERY_STRING= \
CONTENT_LENGTH=0 \
SERVER_PROTOCOL=HTTP/1.1 \
make run.
*/
func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := fmt.Fprintln(w, "ok"); err != nil {
			log.Println(err)
			return
		}
		if _, err := fmt.Fprintln(w, r.Method); err != nil { //nolint:gosec // method is echoed as text/plain, not rendered as HTML
			log.Println(err)
		}
	})
	if err := cgi.Serve(handler); err != nil {
		log.Println(err)
	}
}
