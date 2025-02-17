package msgservice

import (
	"log"
	"os"
	"testing"

	"github.com/tzurielweisberg/postee/v2/actions"
	"github.com/tzurielweisberg/postee/v2/data"
	"github.com/tzurielweisberg/postee/v2/dbservice"
	"github.com/tzurielweisberg/postee/v2/routes"
)

func TestAggregateByTimeout(t *testing.T) {
	const aggregationSeconds = 3

	dbPathReal := dbservice.DbPath
	savedRunScheduler := RunScheduler
	schedulerInvctCnt := 0
	defer func() {
		os.Remove(dbservice.DbPath)
		dbservice.ChangeDbPath(dbPathReal)
		RunScheduler = savedRunScheduler
	}()
	RunScheduler = func(
		route *routes.InputRoute,
		fnSend func(plg actions.Action, cnt map[string]string),
		fnAggregate func(outputName string, currentContent map[string]string, counts int, ignoreLength bool) []map[string]string,
		inpteval data.Inpteval,
		name *string,
		output actions.Action,
	) {
		log.Printf("Mocked Scheduler is activated for route %q. Period: %d sec", route.Name, route.Plugins.AggregateTimeoutSeconds)
		route.StartScheduler()

		schedulerInvctCnt++
	}

	dbservice.ChangeDbPath("test_webhooks.db")
	dbservice.DbPath = "test_webhooks.db"

	demoRoute := &routes.InputRoute{
		Name: "demo-route1",
		Plugins: routes.Plugins{
			AggregateTimeoutSeconds: aggregationSeconds,
		},
	}

	demoEmailPlg := &DemoEmailAction{}

	demoInptEval := &DemoInptEval{}

	srvUrl := ""

	srv1 := new(MsgService)
	srv1.MsgHandling([]byte(mockScan1), demoEmailPlg, demoRoute, demoInptEval, &srvUrl)
	srv1.MsgHandling([]byte(mockScan2), demoEmailPlg, demoRoute, demoInptEval, &srvUrl)
	srv1.MsgHandling([]byte(mockScan3), demoEmailPlg, demoRoute, demoInptEval, &srvUrl)

	expectedSchedulerInvctCnt := 1

	if schedulerInvctCnt != expectedSchedulerInvctCnt {
		t.Errorf("Unexpected plugin invocation count %d, expected %d \n", schedulerInvctCnt, expectedSchedulerInvctCnt)
	}

	demoRoute.StopScheduler()
}
