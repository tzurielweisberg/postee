package webserver

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tzurielweisberg/postee/v2/log"
	"github.com/tzurielweisberg/postee/v2/router"
)

func (ctx *WebServer) tenantHandler(w http.ResponseWriter, r *http.Request) {
	route, ok := mux.Vars(r)["route"]
	if !ok || len(route) == 0 {
		log.Logger.Errorf("Failed route: %q", route)
		ctx.writeResponse(w, http.StatusBadRequest, "failed route")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Logger.Errorf("Failed ioutil.ReadAll: %s", err)
		ctx.writeResponseError(w, http.StatusInternalServerError, err)
		return
	}

	defer r.Body.Close()
	log.Logger.Debugf("%s\n\n", string(body))
	router.Instance().HandleRoute(route, body)
	ctx.writeResponse(w, http.StatusOK, "")
}
