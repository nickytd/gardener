// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeQuotas implements QuotaInterface
type FakeQuotas struct {
	Fake *FakeCoreV1beta1
	ns   string
}

var quotasResource = v1beta1.SchemeGroupVersion.WithResource("quotas")

var quotasKind = v1beta1.SchemeGroupVersion.WithKind("Quota")

// Get takes name of the quota, and returns the corresponding quota object, and an error if there is any.
func (c *FakeQuotas) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.Quota, err error) {
	emptyResult := &v1beta1.Quota{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(quotasResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.Quota), err
}

// List takes label and field selectors, and returns the list of Quotas that match those selectors.
func (c *FakeQuotas) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.QuotaList, err error) {
	emptyResult := &v1beta1.QuotaList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(quotasResource, quotasKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.QuotaList{ListMeta: obj.(*v1beta1.QuotaList).ListMeta}
	for _, item := range obj.(*v1beta1.QuotaList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested quotas.
func (c *FakeQuotas) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(quotasResource, c.ns, opts))

}

// Create takes the representation of a quota and creates it.  Returns the server's representation of the quota, and an error, if there is any.
func (c *FakeQuotas) Create(ctx context.Context, quota *v1beta1.Quota, opts v1.CreateOptions) (result *v1beta1.Quota, err error) {
	emptyResult := &v1beta1.Quota{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(quotasResource, c.ns, quota, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.Quota), err
}

// Update takes the representation of a quota and updates it. Returns the server's representation of the quota, and an error, if there is any.
func (c *FakeQuotas) Update(ctx context.Context, quota *v1beta1.Quota, opts v1.UpdateOptions) (result *v1beta1.Quota, err error) {
	emptyResult := &v1beta1.Quota{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(quotasResource, c.ns, quota, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.Quota), err
}

// Delete takes name of the quota and deletes it. Returns an error if one occurs.
func (c *FakeQuotas) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(quotasResource, c.ns, name, opts), &v1beta1.Quota{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeQuotas) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(quotasResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.QuotaList{})
	return err
}

// Patch applies the patch and returns the patched quota.
func (c *FakeQuotas) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.Quota, err error) {
	emptyResult := &v1beta1.Quota{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(quotasResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.Quota), err
}
