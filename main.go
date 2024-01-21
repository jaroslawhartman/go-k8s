/*
Copyright 2016 The Kubernetes Authors.

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

// Note: the example only works with the code within the same release/branch.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"

	restclient "k8s.io/client-go/rest"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// ExecCmd exec command on specific pod and wait the command's output.
func ExecCmdExample(client kubernetes.Interface, config *restclient.Config, podName string, namespace string,
	command string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	req := client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(namespace).SubResource("exec")
	option := &v1.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	if stdin == nil {
		option.Stdin = false
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return err
	}

	return nil
}

func checkOutput(s strings.Builder) {
	r := bufio.NewReader(strings.NewReader(s.String()))

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "media") {
			fmt.Println("Read line:", line)
		}
	}

}

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	podSpec := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-pod",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "nginx",
					Image: "nginx",
				},
			},
		},
	}

	fmt.Println("Creating pod")
	pod, err := clientset.CoreV1().Pods("nginx").Create(context.TODO(), podSpec, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(pod.GetObjectMeta().GetName())

	fmt.Println("Waiting for pod readiness")
	for {
		pod, err := clientset.CoreV1().Pods("nginx").Get(context.TODO(), "new-pod", metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		phase := pod.Status.Phase
		fmt.Println("Pod phase ", phase)

		if phase == v1.PodRunning {
			break
		}

		time.Sleep(time.Second * 1)
	}

	var buffer strings.Builder
	writer := io.Writer(&buffer)

	err = ExecCmdExample(clientset, config, "new-pod", "nginx", "ls -l", os.Stdin, writer, os.Stderr)
	if err != nil {
		panic(err.Error())
	}

	checkOutput(buffer)

	fmt.Println("Deleting pod")
	err = clientset.CoreV1().Pods("nginx").Delete(context.TODO(), "new-pod", metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
}
