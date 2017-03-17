package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/coreos-inc/alb-ingress-controller/pkg/cmd/controller"

	ingresscontroller "k8s.io/ingress/core/pkg/ingress/controller"
)

func main() {
	clusterName := os.Getenv("CLUSTER_NAME")
	if clusterName == "" {
		glog.Exit("A CLUSTER_NAME environment variable must be defined")
	}

	noop, _ := strconv.ParseBool(os.Getenv("NOOP"))
	awsDebug, _ := strconv.ParseBool(os.Getenv("AWS_DEBUG"))

	config := &controller.Config{
		ClusterName: clusterName,
		Noop:        noop,
		AWSDebug:    awsDebug,
	}

	if len(clusterName) > 11 {
		glog.Exit("CLUSTER_NAME must be 11 characters or less")
	}

	ac := controller.NewALBController(&aws.Config{MaxRetries: aws.Int(5)}, config)
	ic := ingresscontroller.NewIngressController(ac)
	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PROMETHEUS_PORT")
	if port == "" {
		port = "8080"
	}

	go http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	defer func() {
		glog.Infof("Shutting down ingress controller...")
		ic.Stop()
	}()
	ic.Start()
}
