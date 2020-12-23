/*
Copyright 2020 Red Hat, Inc. jcope@redhat.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/redhat-et/caliper/pkg/dbhandler"
	"github.com/spf13/pflag"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	routev1 "github.com/openshift/api/route/v1"
	routeClient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	//"github.com/redhat-et/caliper/pkg/dbhandler"
	"github.com/redhat-et/caliper/pkg/top"
)

func hasBearerToken(cfg *rest.Config) bool {
	if len(cfg.BearerToken) == 0 && len(cfg.BearerTokenFile) == 0 {
		return false
	}
	return true
}

func prometheusHost(r *routev1.Route) string {
	return fmt.Sprintf("https://%s", r.Spec.Host)
}

func handleError(e error) {
	if e != nil {
		klog.ExitDepth(1, e)
	}
}

const (
	promNamespace = `openshift-monitoring`
	promRoute     = `prometheus-k8s`
)

func main() {
	pflag.Parse()
	defer klog.Flush()

	klog.Infof("initializing openshift client from KUBECONFIG=%s", kubeconfig)
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	handleError(err)

	if !hasBearerToken(cfg) {
		klog.Exit("error: bearer token not found, required access to prometheus oauth access.  login to cluster with 'oc'")
	}

	rc := routeClient.NewForConfigOrDie(cfg)
	klog.Infof("fetching prometheus route")
	route, err := rc.Routes(promNamespace).Get(context.Background(), promRoute, metav1.GetOptions{})
	handleError(err)

	transport, err := rest.TransportFor(cfg)
	handleError(err)

	host := prometheusHost(route)
	klog.Infof("initializing connection for host: %s", host)
	conn, err := promapi.NewClient(promapi.Config{
		Address:      host,
		RoundTripper: transport,
	})

	klog.Info("creating prometheus api client")
	pc := promv1.NewAPI(conn)
	handleError(err)

	result, err := top.Top(top.Config{
		Range:            queryRange,
		Context:          context.Background(),
		PrometheusClient: pc,
	})
	handleError(err)

	switch outFormat {
	case "postgres":
		err = streamToDatabase(result)
	case "csv":
		fmt.Printf("%s\n", result.MarshalCSV())
	default:
		printToStdout(result)
	}
	handleError(err)
}

//Postgres compatible time format, required for converting query timestamps
func streamToDatabase(metrics top.PodMetrics) error {
	klog.Infoln("init postgres db client")
	db, err := dbhandler.NewPostgresClient()
	if err != nil {
		return fmt.Errorf("failed to send to db: %v", err)
	}
	sqIns := squirrel.
		Insert("collated_metrics").
		Columns(dbhandler.ColumnsHeaders()...).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(db)

	t0 := time.Now().Format(dbhandler.TimestampFormat)

	for _, m := range metrics {
		sqIns = sqIns.Values(
			version,
			m.Metric,
			m.Pod,
			m.Namespace,
			m.LabelApp,
			t0,
			m.AvgValue,
			"",
			m.Q95Value,
			m.MaxValue,
			m.MinValue,
		)

	}
	resp, err := sqIns.Exec()
	if err != nil {
		return fmt.Errorf("unable to generate sql insert script: %v", err)
	}
	nrows, _ := resp.RowsAffected()
	klog.Infof("insert success, updated %d rows", nrows)
	return nil
}

func printToStdout(podMetrics []*top.PodMetric) {
	klog.Infof("got %d results", len(podMetrics))
	for _, pm := range podMetrics {
		klog.Info(pm)
	}
}
