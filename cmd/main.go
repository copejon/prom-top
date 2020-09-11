/*
Copyright 2020 Jonathan Cope jcope@redhat.com

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
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
)

var (
	kubeconfig string
	context    string
)

func main() {

	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}
	_ = config
}

const (
	kubeconfigEnv = "KUBECONFIG"
)

func init() {

	kubeconfig = os.Getenv(kubeconfigEnv)

	pflag.StringVarP(&kubeconfig, "kubeconfig", "", kubeconfig, "path to kubeconfig file")
	pflag.StringVarP(&context, "context", "", "", "")
	pflag.Parse()
}
