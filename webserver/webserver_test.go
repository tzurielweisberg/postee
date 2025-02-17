package webserver

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tzurielweisberg/postee/v2/router"
)

func TestWebServer_eventsHandler(t *testing.T) {
	rtr := router.Instance()
	rtr.Send([]byte(`{"SigMetadata":{"ID":"TRC-2", "hostname":"postee-0"}}`))
	rtr.Send([]byte(`{"SigMetadata":{"ID":"TRC-3", "hostname":"postee-0"}}`))

	ws := WebServer{}
	w := httptest.NewRecorder()
	var r *http.Request
	ws.eventsHandler(w, r)

	resp := w.Result()
	defer func() {
		_ = resp.Body.Close()
	}()
	got, _ := ioutil.ReadAll(resp.Body)

	assert.JSONEq(t, `[
   {
      "SigMetadata":{
         "ID":"TRC-2",
         "hostname":"postee-0"
      }
   },
   {
      "SigMetadata":{
         "ID":"TRC-3",
         "hostname":"postee-0"
      }
   }
]`, string(got))
}
