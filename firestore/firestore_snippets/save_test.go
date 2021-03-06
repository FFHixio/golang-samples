// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"os"
	"reflect"
	"runtime"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/golang-samples/internal/testutil"
)

func TestSave(t *testing.T) {
	testutil.EndToEndTest(t)
	// TODO: revert this to testutil.SystemTest(t).ProjectID
	// when datastore and firestore can co-exist in a project.
	projectID := os.Getenv("GOLANG_SAMPLES_FIRESTORE_PROJECT")
	if projectID == "" {
		t.Skip("Skipping firestore test. Set GOLANG_SAMPLES_FIRESTORE_PROJECT.")
	}

	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		t.Fatal(err)
	}

	must := func(f func(context.Context, *firestore.Client) error) {
		err := f(ctx, client)
		if err != nil {
			fn := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
			t.Fatalf("%s: %v", fn, err)
		}
	}

	// TODO(someone): check values of docs to make sure data is being manipulated properly.
	must(addDocAsMap)
	must(addDocDataTypes)
	must(addDocAsEntity)
	must(addDocWithID)
	must(addDocWithoutID)
	must(addDocAfterAutoGeneratedID)
	must(updateDoc)
	must(updateDocCreateIfMissing)
	must(updateDocMultiple)
	must(updateDocNested)
	if value, _, err := getField(ctx, client, "users", "frank", "favorites"); err != nil {
		t.Fatal(err)
	} else {
		favorites := value.(map[string]interface{})
		if got, want := favorites["color"], "Red"; got != want {
			t.Errorf("users/frank/favorites.color = %#v; want %#v", got, want)
		}
		if got, want := favorites["food"], "Pizza"; got != want {
			t.Errorf("users/frank/favorites.age = %#v; want %#v", got, want)
		}
	}

	must(deleteDoc)

	if _, exists, err := getField(ctx, client, "cities", "BJ", "capital"); err != nil {
		t.Fatal(err)
	} else if !exists {
		t.Error("Expected 'cities/BJ/capital' to be present")
	}
	must(deleteField)
	if _, exists, err := getField(ctx, client, "cities", "BJ", "capital"); err != nil {
		t.Fatal(err)
	} else if exists {
		t.Error("Expected 'cities/BJ/capital' to be deleted")
	}

	must(runSimpleTransaction)
	must(infoTransaction)
	must(batchWrite)
}

func getField(ctx context.Context, client *firestore.Client, collection, doc, field string) (value interface{}, exists bool, err error) {
	dsnap, err := client.Collection(collection).Doc(doc).Get(ctx)
	if err != nil {
		return nil, false, err
	}
	val, ok := dsnap.Data()[field]
	return val, ok, nil
}
