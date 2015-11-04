// +build ignore

package kubernetes

import (
	"fmt"
	"testing"
)

func TestGetPods(t *testing.T) {
	GetPods()
}

func TestDeploy(t *testing.T) {
	deployReq := KubernetesDeployRequest{
		Name:     "cat3",
		Labels:   map[string]string{"name": "cat-label"},
		ImageURL: "pierone.stups.zalan.do/cat/cat-hello-aws:0.0.1",
		Replicas: 2,
		Ports:    []int{8080},
	}
	_, err := Deploy(&deployReq)
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	} else {
		fmt.Print("Success")
	}
}

func TestGetDeployInfo(t *testing.T) {
	req := KubernetesDeployInfoRequest{Name: "guestbook-ucbv2", Labels: map[string]string{}}
	info, err := DeployInfoKubernetes(&req)
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	} else {
		fmt.Println(info)
	}
}

func TestDelete(t *testing.T) {
	deployReq := KubernetesDeleteRequest{Name: "cat3"}
	err := Delete(&deployReq)
	if err != nil {
		t.FailNow()
	} else {
		fmt.Printf("Delete successful")
	}
}
