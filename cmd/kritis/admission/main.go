/*
Copyright 2018 Google LLC

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
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/grafeas/kritis/cmd/kritis/version"
	"github.com/grafeas/kritis/pkg/kritis/admission"
	"github.com/grafeas/kritis/pkg/kritis/cron"
	kubernetesutil "github.com/grafeas/kritis/pkg/kritis/kubernetes"
	"github.com/grafeas/kritis/pkg/kritis/metadata/containeranalysis"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"

	// Initialize all known client auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	tlsCertFile  string
	tlsKeyFile   string
	cronInterval string
	showVersion  bool
)

const (
	Addr = ":443"
)

func main() {
	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Fatal(err.Error())
	}

	flag.StringVar(&tlsCertFile, "tls-cert-file", "/var/tls/tls.crt", "TLS certificate file.")
	flag.StringVar(&tlsKeyFile, "tls-key-file", "/var/tls/tls.key", "TLS key file.")
	flag.BoolVar(&showVersion, "version", false, "kritis-server version")
	flag.StringVar(&cronInterval, "cron-interval", "1h", "Cron Job time interval as Duration e.g. 1h, 2s")
	flag.Parse()

	if showVersion {
		fmt.Println(version.Commit)
		os.Exit(0)
	}

	// Kick off back ground cron job.
	if err := StartCronJob(); err != nil {
		glog.Fatal(errors.Wrap(err, "starting background job"))
	}

	// Start the Kritis Server.
	glog.Info("Running the server")
	http.HandleFunc("/", admission.ReviewHandler)
	httpsServer := NewServer(Addr)
	glog.Fatal(httpsServer.ListenAndServeTLS(tlsCertFile, tlsKeyFile))
}

func NewServer(addr string) *http.Server {
	return &http.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			// TODO: Change this to tls.RequireAndVerifyClientCert
			ClientAuth: tls.NoClientCert,
		},
	}
}

func StartCronJob() error {
	checkInterval, err := time.ParseDuration(cronInterval)
	if err != nil {
		return err
	}
	ctx := context.Background()
	ki, err := kubernetesutil.GetClientset()
	if err != nil {
		return err
	}
	kcs := ki.(*kubernetes.Clientset)
	metadataClient, err := containeranalysis.NewContainerAnalysisClient()
	if err != nil {
		return err
	}
	go cron.Start(ctx, *cron.NewCronConfig(kcs, *metadataClient), checkInterval)
	return nil
}
