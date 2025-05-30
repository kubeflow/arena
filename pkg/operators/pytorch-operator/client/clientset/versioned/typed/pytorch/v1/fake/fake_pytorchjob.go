// Copyright 2020 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	pytorchv1 "github.com/kubeflow/arena/pkg/operators/pytorch-operator/apis/pytorch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePyTorchJobs implements PyTorchJobInterface
type FakePyTorchJobs struct {
	Fake *FakeKubeflowV1
	ns   string
}

var pytorchjobsResource = schema.GroupVersionResource{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}

var pytorchjobsKind = schema.GroupVersionKind{Group: "kubeflow.org", Version: "v1", Kind: "PyTorchJob"}

// Get takes name of the pyTorchJob, and returns the corresponding pyTorchJob object, and an error if there is any.
func (c *FakePyTorchJobs) Get(name string, options v1.GetOptions) (result *pytorchv1.PyTorchJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(pytorchjobsResource, c.ns, name), &pytorchv1.PyTorchJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*pytorchv1.PyTorchJob), err
}

// List takes label and field selectors, and returns the list of PyTorchJobs that match those selectors.
func (c *FakePyTorchJobs) List(opts v1.ListOptions) (result *pytorchv1.PyTorchJobList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(pytorchjobsResource, pytorchjobsKind, c.ns, opts), &pytorchv1.PyTorchJobList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &pytorchv1.PyTorchJobList{ListMeta: obj.(*pytorchv1.PyTorchJobList).ListMeta}
	for _, item := range obj.(*pytorchv1.PyTorchJobList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested pyTorchJobs.
func (c *FakePyTorchJobs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(pytorchjobsResource, c.ns, opts))

}

// Create takes the representation of a pyTorchJob and creates it.  Returns the server's representation of the pyTorchJob, and an error, if there is any.
func (c *FakePyTorchJobs) Create(pyTorchJob *pytorchv1.PyTorchJob) (result *pytorchv1.PyTorchJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(pytorchjobsResource, c.ns, pyTorchJob), &pytorchv1.PyTorchJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*pytorchv1.PyTorchJob), err
}

// Update takes the representation of a pyTorchJob and updates it. Returns the server's representation of the pyTorchJob, and an error, if there is any.
func (c *FakePyTorchJobs) Update(pyTorchJob *pytorchv1.PyTorchJob) (result *pytorchv1.PyTorchJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(pytorchjobsResource, c.ns, pyTorchJob), &pytorchv1.PyTorchJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*pytorchv1.PyTorchJob), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePyTorchJobs) UpdateStatus(pyTorchJob *pytorchv1.PyTorchJob) (*pytorchv1.PyTorchJob, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(pytorchjobsResource, "status", c.ns, pyTorchJob), &pytorchv1.PyTorchJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*pytorchv1.PyTorchJob), err
}

// Delete takes name of the pyTorchJob and deletes it. Returns an error if one occurs.
func (c *FakePyTorchJobs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(pytorchjobsResource, c.ns, name), &pytorchv1.PyTorchJob{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePyTorchJobs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(pytorchjobsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &pytorchv1.PyTorchJobList{})
	return err
}

// Patch applies the patch and returns the patched pyTorchJob.
func (c *FakePyTorchJobs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *pytorchv1.PyTorchJob, err error) {
	obj, err := c.Fake.
		// delete param pt because this function of current k8s.io/client-go/testing doesn't have this param, same as tf
		Invokes(testing.NewPatchSubresourceAction(pytorchjobsResource, c.ns, name, pt, data, subresources...), &pytorchv1.PyTorchJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*pytorchv1.PyTorchJob), err
}
