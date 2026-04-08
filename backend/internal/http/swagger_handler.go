package httpx

import (
	"net/http"
)

func SwaggerUIHandler() http.Handler {
	return http.FileServer(http.Dir("internal/http/swagger-ui"))
}
