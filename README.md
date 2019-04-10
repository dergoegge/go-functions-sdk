## Importer beware this is a **WORK IN PROGRESS**

# Unofficial golang cloud functions sdk

This projects aims to simplify the deployment process of golang cloud functions.

Code the functions and deploy with `gocf deploy`.

## Install the functions package
```
go get -u github.com/dergoegge/go-functions-sdk
```

## Function examples

```golang
package funcpackage

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dergoegge/go-functions-sdk/pkg/functions"
)

var HTTPFuncName = functions.New().
	RunWith(functions.RuntimeOptions{
		Timeout: 70, // Timeout in seconds
		Memory:  "128MB",
	}).
	Region("us-central1").
	RetryOnFailure(true).
	OnRequest(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Yeet")
	})

var FirestoreFuncName = functions.New().
	RunWith(functions.RuntimeOptions{
		Timeout: 70,
		Memory:  "1GB",
	}).
	Region("europe-west2").
	Firestore().
	Document("some-collection/{docID}").
	OnWrite(func(ctx context.Context, event functions.FirestoreEvent) error {
		fmt.Println(event.Value.Fields)
		return nil
	})

```

Currently only http, firestore and storage triggered functions are supported.

## Install the deployment tool

```sh
gcloud auth login
gcloud config set project <project-id>

go install github.com/dergoegge/go-functions-sdk/cmd/gocf
```

## Comands
gocf deploy [--only "comma seperated list of functions to deploy"]

## Todos

* support more function triggers
* add custom oauth2 flow and use cloudfunctions api instead of the gcloud util
* manage deletion of functions
* run `go mod init/tidy` for function modules
* run `go mod vendor` for function modules
* add examples
* write tests