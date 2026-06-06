package main

import (
	"fmt"
	"net/http"
	"net/http/cgi"
)

/*
REQUEST_METHOD=GET \
QUERY_STRING= \
CONTENT_LENGTH=0 \
SERVER_PROTOCOL=HTTP/1.1 \
make run
*/
func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
		fmt.Fprintln(w, r.Method)
	})
	cgi.Serve(handler)
}
