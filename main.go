/*
Copyright The Helm Authors.
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
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Azure/azure-storage-blob-go/2018-03-28/azblob"
)

// TODO(mattfarina): Add testing

var errCounter = 0

// Aggregated logs are generally safe to share publicly
// Individual logs contain details such as the username and IP the user used
// This information should generally not be shared publicly. This project handles
// aggregated logs.
func main() {
	// Sending all info to stdout and stderr so that Kubernetes or other tool
	// can pull them in.
	fmt.Println("Starting up Query and Store Quay Logs")

	// Part 1: Query Quay for logs
	// Docs on the Quay API can be found at https://docs.quay.io/api/

	// The account that queries for the logs needs the repo:admin scope. This
	// is a bit much in practice and it would be nice if there was a more
	// granular permission. So, guard the creds and consider a regular
	// rotation of them.
	quayToken := envOrErr("QUAY_TOKEN")

	// Construct a URL
	tod := time.Now()
	yesterday := tod.AddDate(0, 0, -1)

	// The / in the date needs to be escaped for query strings
	datestring := url.QueryEscape(yesterday.Format("1/2/2006"))

	// TODO(mattfarina): Support Quay enterprise
	// Note, this is getting the aggregated logs. The raw query logs should be kept
	// private because they have details, such as IPs, for every request.
	uQuay := "https://quay.io/api/v1/repository/helmpack/chart-testing/aggregatelogs"

	// Note, when the end date is the same as the start date it will increment it
	// by a day for you.
	u, err := url.Parse(fmt.Sprintf("%s?starttime=%s&endtime=%s", uQuay, datestring, datestring))
	handleErr(err)

	client := &http.Client{}
	req, err := http.NewRequest("GET", u.String(), nil)
	handleErr(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", quayToken))
	req.Header.Set("User-Agent", "query-store-quay-logs/0.0.1") // sending a user agent for Quays benefit

	// Query and handle the things
	resp, err := client.Do(req)
	handleErr(err)

	if resp.StatusCode != 200 {
		// This is likely and auth error
		fmt.Fprintf(os.Stderr, "ERROR querying Quay. Received a response code of: %s\n", resp.Status)
		os.Exit(1)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	handleErr(err)

	fmt.Println("Received aggregated logs from Quay for", yesterday.Format("1/2/2006"))

	// Put logs in Object Storage (in this case Azure)
	// TODO(mattfarina): Support systems other than Azure
	fmt.Println("Preparing to store logs in Azure Blog Store")

	// The interactions are inspired by
	// https://github.com/Azure-Samples/storage-blobs-go-quickstart/
	accountName := envOrErr("AZURE_STORAGE_ACCOUNT")
	accountKey := envOrErr("AZURE_STORAGE_ACCESS_KEY")
	containerName := envOrErr("AZURE_CONTAINER")

	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	handleErr(err)
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// Get a URL for the new blob
	u, err = url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))
	handleErr(err)
	containerURL := azblob.NewContainerURL(*u, p)
	// Saving the json data in a format that Year-Month-Day so it's in an
	// easily searchable/filterable manner
	blobURL := containerURL.NewBlockBlobURL(yesterday.Format("2006-01-02.json"))

	fmt.Printf("Uploading %s to %s\n", yesterday.Format("2006-01-02.json"), u.String())

	ctx := context.Background()
	_, err = azblob.UploadBufferToBlockBlob(ctx, content, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16})
	handleErr(err)

	fmt.Println("Completed uploading file")
}

func handleErr(err error) {
	errCounter++
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR copying logs: %s\n", err)

		os.Exit(errCounter)
	}
}

func envOrErr(name string) string {
	val := os.Getenv(name)
	if len(val) == 0 {
		fmt.Fprintf(os.Stderr, "ERROR Missing %q\n", name)
		os.Exit(1)
	}

	return val
}
