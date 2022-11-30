/*
Copyright 2022 Matt Wise.

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

package v1alpha1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/diranged/oz/webhook"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var execaccessrequestlog = logf.Log.WithName("execaccessrequest-resource")

func (r *ExecAccessRequest) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if err := webhook.RegisterContextualDefaulter(r, mgr); err != nil {
		panic(err)
	}
	if err := webhook.RegisterContextualValidator(r, mgr); err != nil {
		panic(err)
	}

	// boilerplate
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-crds-wizardofoz-co-v1alpha1-execaccessrequest,mutating=true,failurePolicy=fail,sideEffects=None,groups=crds.wizardofoz.co,resources=execaccessrequests,verbs=create;update,versions=v1alpha1,name=mexecaccessrequest.kb.io,admissionReviewVersions=v1

var _ webhook.IContextuallyDefaultableObject = &ExecAccessRequest{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ExecAccessRequest) Default(req admission.Request) error {
	logger := log.FromContext(context.Background())
	logger.Info("defaulter Well gotcha", "req", ObjectToJSON(req), "self", ObjectToJSON(r))
	return errors.New("junk")
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-crds-wizardofoz-co-v1alpha1-execaccessrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=crds.wizardofoz.co,resources=execaccessrequests,verbs=create;update,versions=v1alpha1,name=vexecaccessrequest.kb.io,admissionReviewVersions=v1

var _ webhook.IContextuallyValidatableObject = &ExecAccessRequest{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ExecAccessRequest) ValidateCreate(req admission.Request) error {
	execaccessrequestlog.Info("validate create", "name", r.Name)

	return fmt.Errorf("faking an error")

	// TODO(user): fill in your validation logic upon object creation.
	// return nil
}

// ValidateUpdate prevents immutable updates to the ExecAccessRequest.
//
// https://stackoverflow.com/questions/70650677/manage-immutable-fields-in-kubebuilder-validating-webhook
func (r *ExecAccessRequest) ValidateUpdate(req admission.Request, old runtime.Object) error {
	execaccessrequestlog.Info("validate update", "name", r.Name)
	oldRequest, _ := old.(*ExecAccessRequest)
	if r.Spec.TargetPod != oldRequest.Spec.TargetPod {
		return fmt.Errorf(
			"error - Spec.TargetPod is an immutable field, create a new PodAccessRequest instead",
		)
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ExecAccessRequest) ValidateDelete(req admission.Request) error {
	execaccessrequestlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

// ObjectToJSON is a quick helper function for pretty-printing an entire K8S object in JSON form.
// Used in certain debug log statements primarily.
func ObjectToJSON(obj any) string {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf("could not marshal json: %s\n", err)
		return ""
	}
	return string(jsonData)
}
