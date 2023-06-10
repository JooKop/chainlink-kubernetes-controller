/*
Copyright 2023.

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

package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	oraclev1alpha1 "github.com/JooKop/chainlink-kubernetes-operator/api/v1alpha1"
)

// ChainlinkJobReconciler reconciles a ChainlinkJob object
type ChainlinkJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type Jar struct {
	lk      sync.Mutex
	cookies map[string][]*http.Cookie
}

func NewJar() *Jar {
	jar := new(Jar)
	jar.cookies = make(map[string][]*http.Cookie)
	return jar
}

// SetCookies handles the receipt of the cookies in a reply for the
// given URL.  It may or may not choose to save the cookies, depending
// on the jar's policy and implementation.
func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lk.Lock()
	jar.cookies[u.Host] = cookies
	jar.lk.Unlock()
}

// Cookies returns the cookies to send in a request for the given URL.
// It is up to the implementation to honor the standard cookie use
// restrictions such as in RFC 6265.
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
}

type PostBody struct {
	Email    string `json:"email"`
	Password string `json:"PASSWORD"`
}

type JobBody struct {
	OperationName string         `json:"operationName"`
	Query         string         `json:"query"`
	Variables     VariableStruct `json:"variables"`
}

type VariableStruct struct {
	Input InputStruct `json:"input"`
}

type InputStruct struct {
	Toml string `json:"TOML"`
}

//+kubebuilder:rbac:groups=oracle.example.com,resources=chainlinkjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oracle.example.com,resources=chainlinkjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oracle.example.com,resources=chainlinkjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ChainlinkJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ChainlinkJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	//Fetch the changed ChainlinkJob resource
	chainlinkJob := &oraclev1alpha1.ChainlinkJob{}
	err := r.Get(ctx, req.NamespacedName, chainlinkJob)

	if err != nil {
		log.Error(err, "An error occurred")

		if apierrors.IsNotFound(err) {
			log.Info("ChainlinkJob resource not found. Ignoring error because it means the resource was deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the updated object. Retry.
		log.Error(err, "Failed to get ChainlinkJob")
		return ctrl.Result{}, err
	}

	// New job appeared. Authenticate with the spec.chainlinkNode node and add the new job.
	log.Info("Alright, let's post a new job")
	jar := NewJar()
	client := http.Client{Transport: nil, CheckRedirect: nil, Jar: jar, Timeout: 0}

	// Prepare request body
	body := PostBody{
		Email:    "test@example.com",
		Password: "mysecretpassword",
	}
	bodyBytes, err := json.Marshal(&body)
	if err != nil {
		log.Error(err, "Failed to marshal json body")
	}
	reader := bytes.NewReader(bodyBytes)
	resp, err := client.Post("http://"+chainlinkJob.Spec.ChainlinkNode+"-service."+chainlinkJob.Namespace+"/sessions", "application/json", reader)
	if err != nil {
		log.Error(err, "An error")
	}
	b, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Println(string(b))
	//Authenticated. Post a new job.

	c := &JobBody{
		OperationName: "CreateJob",
		Query: `mutation CreateJob($input: CreateJobInput!) {
	createJob(input: $input) {
	  ... on CreateJobSuccess {
		job {
		  id
		  __typename
		}
		__typename
	  }
	  ... on InputErrors {
		errors {
		  path
		  message
		  code
		  __typename
		}
		__typename
	  }
	  __typename
	}
  }
  `,
		Variables: VariableStruct{
			Input: InputStruct{
				Toml: chainlinkJob.Spec.JobSpec,
			},
		}}
	bodyBytes, err = json.Marshal(&c)
	if err != nil {
		log.Error(err, "Failed to marshal json body")
	}
	reader = bytes.NewReader(bodyBytes)
	resp, err = client.Post("http://"+chainlinkJob.Spec.ChainlinkNode+"-service."+chainlinkJob.Namespace+"/query", "application/json", reader)
	if err != nil {
		log.Error(err, "An error occurred")
	}
	b, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println(string(b))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChainlinkJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oraclev1alpha1.ChainlinkJob{}).
		Complete(r)
}
