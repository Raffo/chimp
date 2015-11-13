package api

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	Start()
}

func TestMapMemory(t *testing.T) {
	expected := 2048
	val, err := mapMemory("2048MB")
	if err != nil {
		t.FailNow()
	}
	if val != expected {
		fmt.Println("Expected: " + string(expected) + ", got: " + string(val))
		t.FailNow()
	}
	val, err = mapMemory("2048")
	if err != nil {
		t.FailNow()
	}
	if val != expected {
		fmt.Println("Expected: " + string(expected) + ", got: " + string(val))
		t.FailNow()
	}
}

func TestBuildTeamLabel(t *testing.T) {
	fakeContext := gin.Context{}
	expectedUid := "rdifazio"
	expectedTeam := "TechMonkeys"
	fakeContext.Set("uid", expectedUid)
	fakeContext.Set("team", expectedTeam)
	team, uid := buildTeamLabel(&fakeContext)
	if uid != expectedUid || team != expectedTeam {
		fmt.Printf("Expected: %s - %s, got: %s, %s", expectedUid, expectedTeam, uid, team)
		t.FailNow()
	}
}

func TestCommonDeploy(t *testing.T) {
	fakeContext := gin.Context{}
	fakeContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString(
		"{\"Name\":\"app\", \"Ports\":[8080],\"Labels\":{},\"ImageURL\":\"imagename\",\"Env\":null,\"Replicas\":3}"))
	deployReq, err := commonDeploy(&fakeContext)
	if err != nil {
		t.FailNow()
	}
	if deployReq.Name != "app" {
		fmt.Println("NAME doesn't match, GOT " + deployReq.Name)
		t.FailNow()
	}
	if deployReq.Ports[0] != 8080 {
		fmt.Println()
		t.FailNow()
	}
}

func TestDeployList(t *testing.T) {

}
