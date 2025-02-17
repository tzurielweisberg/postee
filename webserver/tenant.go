package webserver

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tzurielweisberg/postee/v2/router"
	"github.com/tzurielweisberg/postee/v2/utils"
)

func (ctx *WebServer) tenantHandler(w http.ResponseWriter, r *http.Request) {
	route, ok := mux.Vars(r)["route"]
	if !ok || len(route) == 0 {
		log.Printf("Failed route: %q", route)
		ctx.writeResponse(w, http.StatusBadRequest, "failed route")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed ioutil.ReadAll: %s", err)
		ctx.writeResponseError(w, http.StatusInternalServerError, err)
		return
	}

	defer r.Body.Close()
	utils.Debug("%s\n\n", string(body))
	router.Instance().HandleRoute(route, body)
	ctx.writeResponse(w, http.StatusOK, "")
}
