package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func (c *WorkspaceClient) Workspaces(namespace string) WorkspaceInterface {
	return &workspaceClient{
		client: c.RestClient,
		ns:     namespace,
	}
}

type WorkspaceClient struct {
	RestClient rest.Interface
}

type WorkspaceInterface interface {
	Create(obj *Workspace) (*Workspace, error)
	Update(obj *Workspace) (*Workspace, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string) (*Workspace, error)
}

type workspaceClient struct {
	client rest.Interface
	ns     string
}

func (c *workspaceClient) Create(obj *Workspace) (*Workspace, error) {
	result := &Workspace{}
	err := c.client.Post().
		Namespace(c.ns).Resource(WorkspacePlural).
		Body(obj).Do().Into(result)
	return result, err
}

func (c *workspaceClient) Update(obj *Workspace) (*Workspace, error) {
	result := &Workspace{}
	err := c.client.Put().
		Namespace(c.ns).Resource(WorkspacePlural).
		Body(obj).Do().Into(result)
	return result, err
}

func (c *workspaceClient) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).Resource(WorkspacePlural).
		Name(name).Body(options).Do().
		Error()
}

func (c *workspaceClient) Get(name string) (*Workspace, error) {
	result := &Workspace{}
	err := c.client.Get().
		Namespace(c.ns).Resource(WorkspacePlural).
		Name(name).Do().Into(result)
	return result, err
}
