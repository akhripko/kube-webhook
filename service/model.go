package service

import (
	"strings"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ownersKey = "owners"

type ErrorForbidden string

func (e ErrorForbidden) Error() string {
	return "Forbidden: " + string(e)
}

// see types "k8s.io/api/apps/v1"
type AdmissionRequestObject struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            struct {
		Template struct {
			meta.ObjectMeta `json:"metadata,omitempty"`
		} `json:"template"`
	} `json:"spec,omitempty"`
}

type annotations map[string]string

type annotationsOption func(a annotations)

func newAnnotations(opts ...annotationsOption) annotations {
	a := annotations{}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func withOwners(owners []string) annotationsOption {
	return func(a annotations) {
		a[ownersKey] = strings.Join(owners, ",")
	}
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type patchObject []patchOperation

type pathObjectOption func(object *patchObject)

func newPatchObject(opts ...pathObjectOption) patchObject {
	p := patchObject{}
	for _, opt := range opts {
		opt(&p)
	}
	return p
}

func (p *patchObject) prepareCap(require int) {
	if p == nil {
		*p = make(patchObject, 0, require)
		return
	}
	if cap(*p)-len(*p) >= require {
		return
	}
	current := *p
	*p = make(patchObject, len(current), cap(current)+require)
	copy(*p, current)
}

func withCap(cap int) pathObjectOption {
	return func(object *patchObject) {
		object.prepareCap(cap)
	}
}

func withAnnotations(obj *AdmissionRequestObject, data annotations) pathObjectOption {
	return func(p *patchObject) {
		// prepare patch object memory
		p.prepareCap(len(data))
		// in case if object is without annotations add all annotations
		if obj.Annotations == nil {
			for key, value := range data {
				*p = append(*p, patchOperation{
					Op:   "add",
					Path: "/metadata/annotations",
					Value: map[string]string{
						key: value,
					},
				})
			}
			return
		}
		// in case if object has annotation -> replace, if not -> add
		for key, value := range data {
			if _, exist := obj.Annotations[key]; exist {
				*p = append(*p, patchOperation{
					Op:    "replace",
					Path:  "/metadata/annotations/" + key,
					Value: value,
				})
				continue
			}
			*p = append(*p, patchOperation{
				Op:    "add",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			})
		}
	}
}

func withSpecTemplateAnnotations(obj *AdmissionRequestObject, data annotations) pathObjectOption {
	return func(p *patchObject) {
		if !(obj.Kind == "Deployment" || obj.Kind == "ReplicaSet") {
			return
		}
		// prepare patch object memory
		p.prepareCap(len(data))
		annots := obj.Spec.Template.Annotations
		// in case if object is without annotations add all annotations
		if annots == nil {
			for key, value := range data {
				*p = append(*p, patchOperation{
					Op:   "add",
					Path: "/spec/template/metadata/annotations",
					Value: map[string]string{
						key: value,
					},
				})
			}
			return
		}
		// in case if object has annotation -> replace, if not -> add
		for key, value := range data {
			if _, exist := annots[key]; exist {
				*p = append(*p, patchOperation{
					Op:    "replace",
					Path:  "/spec/template/metadata/annotations/" + key,
					Value: value,
				})
				continue
			}
			*p = append(*p, patchOperation{
				Op:    "add",
				Path:  "/spec/template/metadata/annotations/" + key,
				Value: value,
			})
		}
	}
}
