package marathon

import (
	"github.com/stretchr/testify/assert"
	"github.com/zalando-techmonkeys/chimp/backend"
	"testing"
	"time"
)

var m *MarathonBackend

func Test_GetAppNames(t *testing.T) {
	m = New()
	apps, err := m.GetAppNames()
	if err != nil {
		t.Errorf("%s", err.Error())
		t.FailNow()
	}
	if !(len(apps) > 0) {
		t.Errorf("Should be at least one application which is running")
	}
	t.Log(apps[0])
}

func Test_Deploy(t *testing.T) {
	m = New()
	var beReq *backend.CreateRequest = &backend.CreateRequest{BaseRequest: backend.BaseRequest{
		Name:        "appname",
		Ports:       []int{8080},
		CPULimit:    2,
		MemoryLimit: 2048,
		ImageURL:    "pierone.stups.zalan.do/cat/cat-hello-aws:0.0.1",
		Env:         map[string]string{"foo": "bar"},
		Labels:      map[string]string{"environment": "live", "team": "shop"},
		Replicas:    1,
	}}

	app, err := m.Deploy(beReq)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	} else {
		assert.Equal(t, "/appname", app)
		t.Logf("Application successfully deployed")
	}
}

func Test_GetApp(t *testing.T) {
	m = New()
	time.Sleep(20 * time.Second)
	appid := &backend.ArtifactRequest{Name: "appname"}
	app, err := m.GetApp(appid)
	if err != nil {
		t.Errorf("Should not return any error message, error: %s", err.Error())
	}
	expect := "/appname"
	assert.Equal(t, app.Name, expect)
	t.Logf("Application information: \n")
	t.Logf("%v", app)
}

func Test_Scale(t *testing.T) {
	m = New()
	time.Sleep(20 * time.Second)
	scale := &backend.ScaleRequest{Name: "appname", Replicas: 2}
	_, err := m.Scale(scale)
	if err != nil {
		t.Errorf("Expect could not scale application")
	} else {
		t.Logf("Application is scaled")
	}
}

func Test_Delete(t *testing.T) {
	time.Sleep(20 * time.Second)
	app := backend.ArtifactRequest{Name: "appname"}
	_, err := m.Delete(&app)
	if err != nil {
		t.Errorf("Expect could not delete application")
	} else {
		t.Logf("Application is deleted")
	}
}
