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
	"fmt"

	"github.com/diranged/oz/webhook"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var execaccessrequestlog = logf.Log.WithName("execaccessrequest-resource")

// SetupWebhookWithManager configures the webhook service in the Manager to
// accept MutatingWebhookConfiguration and ValidatingWebhookConfiguration calls
// from the Kubernetes API server.
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

//+kubebuilder:webhook:path=/mutate-crds-wizardofoz-co-v1alpha1-execaccessrequest,mutating=true,failurePolicy=fail,sideEffects=None,groups=crds.wizardofoz.co,resources=execaccessrequests,verbs=create;update,versions=v1alpha1,name=mexecaccessrequest.kb.io,admissionReviewVersions=v1

var _ webhook.IContextuallyDefaultableObject = &ExecAccessRequest{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ExecAccessRequest) Default(_ admission.Request) error {
	return nil
}

//+kubebuilder:webhook:path=/validate-crds-wizardofoz-co-v1alpha1-execaccessrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=crds.wizardofoz.co,resources=execaccessrequests,verbs=create;update;delete,versions=v1alpha1,name=vexecaccessrequest.kb.io,admissionReviewVersions=v1

var _ webhook.IContextuallyValidatableObject = &ExecAccessRequest{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ExecAccessRequest) ValidateCreate(req admission.Request) error {
	if req.UserInfo.Username != "" {
		execaccessrequestlog.Info(
			fmt.Sprintf("Create ExecAccessRequest from %s", req.UserInfo.Username),
		)
	} else {
		// TODO: Make this fail, after we have confidence in the code in a live environment.
		execaccessrequestlog.Info("WARNING - Create ExecAccessRequest with missing user identity")
	}
	return nil
}

// ValidateUpdate prevents immutable updates to the ExecAccessRequest.
func (r *ExecAccessRequest) ValidateUpdate(_ admission.Request, old runtime.Object) error {
	execaccessrequestlog.Info("validate update", "name", r.Name)

	// https://stackoverflow.com/questions/70650677/manage-immutable-fields-in-kubebuilder-validating-webhook
	oldRequest, _ := old.(*ExecAccessRequest)
	if r.Spec.TargetPod != oldRequest.Spec.TargetPod {
		return fmt.Errorf(
			"error - Spec.TargetPod is an immutable field, create a new PodAccessRequest instead",
		)
	}
	return nil
}

// ValidateDelete implements webhook.IContextuallyValidatableObject so a webhook will be registered for the type
func (r *ExecAccessRequest) ValidateDelete(req admission.Request) error {
	execaccessrequestlog.Info(
		fmt.Sprintf("Delete ExecAccessRequest from %s", req.UserInfo.Username),
	)
	return nil
}
