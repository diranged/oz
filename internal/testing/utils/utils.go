// Package testutil provides common utilities used during end to end tests
package testutil

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	//revive:disable:dot-imports
	. "github.com/onsi/ginkgo/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Run executes the provided command within this context
func Run(cmd *exec.Cmd) (string, error) {
	dir, _ := GetProjectDir()
	cmd.Dir = dir
	fmt.Fprintf(GinkgoWriter, "running dir: %s\n", cmd.Dir)

	// To allow make commands be executed from the project directory which is subdir on SDK repo
	// TODO:(user) You might not need the following code
	if err := os.Chdir(cmd.Dir); err != nil {
		fmt.Fprintf(GinkgoWriter, "chdir dir: %s\n", err)
	}

	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	command := strings.Join(cmd.Args, " ")
	fmt.Fprintf(GinkgoWriter, "running: %s\n", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(
				output,
			), fmt.Errorf(
				"%s failed with error: (%v) %s",
				command,
				err,
				string(output),
			)
	}

	return string(output), nil
}

// GetProjectDir will return the directory where the project is
func GetProjectDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return wd, err
	}
	wd = strings.Replace(wd, "/internal/testing/e2e", "", -1)
	return wd, nil
}

// LoadImageToKindClusterWithName loads a local docker image to the kind cluster
func LoadImageToKindClusterWithName(name string) error {
	cluster := "kind"
	if v, ok := os.LookupEnv("KIND_CLUSTER"); ok {
		cluster = v
	}
	kindOptions := []string{"load", "docker-image", name, "--name", cluster}
	cmd := exec.Command("kind", kindOptions...)
	_, err := Run(cmd)
	return err
}

// GetNonEmptyLines converts given command output string into individual objects
// according to line breakers, and ignores the empty elements in it.
func GetNonEmptyLines(output string) []string {
	var res []string
	elements := strings.Split(output, "\n")
	for _, element := range elements {
		if element != "" {
			res = append(res, element)
		}
	}

	return res
}

// RandomString is a function for generating a random string for certain tests
func RandomString(length int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	random.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}

// FindUnstructuredByOwner returns a list of Unstructured Objects based on the
// OwnerReference supplied. Used to dynamically search for resource types that
// we expect the to have been created by our Builders, making sure that they
//
// NOTE: Unused right now. TODO, Maybe remove.
func FindUnstructuredByOwner(
	ctx context.Context,
	cl client.Client,
	namespace string,
	parentUID types.UID,
	list *unstructured.UnstructuredList,
) (*unstructured.Unstructured, error) {
	opts := []client.ListOption{client.InNamespace(namespace)}
	if err := cl.List(ctx, list, opts...); err != nil {
		return nil, err
	}
	for _, obj := range list.Items {
		refs := obj.GetOwnerReferences()

		if len(refs) > 0 {
			if refs[0].UID == parentUID {
				return &obj, nil
			}
		}
	}
	return nil, fmt.Errorf("Not found")
}
