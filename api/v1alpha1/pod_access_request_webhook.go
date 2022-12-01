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
var podaccessrequestlog = logf.Log.WithName("podaccessrequest-resource")

// SetupWebhookWithManager configures the webhook service in the Manager to
// accept MutatingWebhookConfiguration and ValidatingWebhookConfiguration calls
// from the Kubernetes API server.
func (r *PodAccessRequest) SetupWebhookWithManager(mgr ctrl.Manager) error {
	// if err := webhook.RegisterContextualDefaulter(r, mgr); err != nil {
	// 	panic(err)
	// }
	if err := webhook.RegisterContextualValidator(r, mgr); err != nil {
		panic(err)
	}

	// boilerplate
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-crds-wizardofoz-co-v1alpha1-podaccessrequest,mutating=true,failurePolicy=fail,sideEffects=None,groups=crds.wizardofoz.co,resources=podaccessrequests,verbs=create;update,versions=v1alpha1,name=mpodaccessrequest.kb.io,admissionReviewVersions=v1

// var _ webhook.IContextuallyDefaultableObject = &PodAccessRequest{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
//func (r *PodAccessRequest) Default(req admission.Request) error {
//  return nil
//}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-crds-wizardofoz-co-v1alpha1-podaccessrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=crds.wizardofoz.co,resources=podaccessrequests,verbs=create;update,versions=v1alpha1,name=vpodaccessrequest.kb.io,admissionReviewVersions=v1

var _ webhook.IContextuallyValidatableObject = &PodAccessRequest{}

// ValidateCreate implements webhook.IContextuallyValidatableObject so a webhook will be registered for the type
func (r *PodAccessRequest) ValidateCreate(req admission.Request) error {
	if req.UserInfo.Username != "" {
		podaccessrequestlog.Info(
			fmt.Sprintf("Create PodAccessRequest from %s", req.UserInfo.Username),
		)
	} else {
		// TODO: Make this fail, after we have confidence in the code in a live environment.
		podaccessrequestlog.Info("WARNING - Create ExecAccessRequest with missing user identity")
	}
	return nil
}

// ValidateUpdate implements webhook.IContextuallyValidatableObject so a webhook will be registered for the type
func (r *PodAccessRequest) ValidateUpdate(req admission.Request, _ runtime.Object) error {
	if req.UserInfo.Username != "" {
		podaccessrequestlog.Info(
			fmt.Sprintf("Update PodAccessRequest from %s", req.UserInfo.Username),
		)
	} else {
		// TODO: Make this fail, after we have confidence in the code in a live environment.
		podaccessrequestlog.Info("WARNING - Update ExecAccessRequest with missing user identity")
	}
	return nil
}

// ValidateDelete implements webhook.IContextuallyValidatableObject so a webhook will be registered for the type
func (r *PodAccessRequest) ValidateDelete(req admission.Request) error {
	podaccessrequestlog.Info(
		fmt.Sprintf("Delete PodAccessRequest from %s", req.UserInfo.Username),
	)
	return nil
}
