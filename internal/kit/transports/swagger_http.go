package transports

import (
	"github.com/Minh2009/pv_soa/cfg"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func SwaggerHttpHandler(c cfg.Config) http.Handler {
	pr := mux.NewRouter()
	basePath := c.BasePath

	// Handling & Manipulate swagger.yaml basePath with config-val
	pr.HandleFunc(basePath+"swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		fileBytes, err := os.ReadFile("swagger.yaml")
		if err != nil {
			panic(err)
		}

		//regex, _ := regexp.Compile(`^basePath\s*:\s+.*`)
		//fileBytes = regex.ReplaceAll(fileBytes, []byte("basePath: "+basePrefix))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/yaml")
		_, _ = w.Write(fileBytes)
	})
	opts := middleware.SwaggerUIOpts{SpecURL: "swagger.yaml", BasePath: basePath}
	sh := middleware.SwaggerUI(opts, nil)
	pr.Handle(basePath+"docs", sh)

	//// documentation for share
	opts1 := middleware.RedocOpts{SpecURL: "swagger.yaml", BasePath: basePath, Path: "doc"}
	sh1 := middleware.Redoc(opts1, nil)
	pr.Handle(basePath+"doc", sh1)

	return pr
}
