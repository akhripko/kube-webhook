package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_patchObject_Extend(t *testing.T) {
	var p patchObject
	assert.Nil(t, p)

	p.prepareCap(4)
	assert.Equal(t, 0, len(p))
	assert.Equal(t, 4, cap(p))

	p = append(p, patchOperation{})
	p.prepareCap(3)
	assert.Equal(t, 1, len(p))
	assert.Equal(t, 4, cap(p))

	for i := 0; i < 3; i++ {
		p = append(p, patchOperation{})
	}
	assert.Equal(t, 4, len(p))
	assert.Equal(t, 4, cap(p))
	p.prepareCap(5)
	assert.Equal(t, 4, len(p))
	assert.Equal(t, 9, cap(p))
}

func Test_patchObject_new(t *testing.T) {
	p := newPatchObject(
		withCap(10),
		withAnnotations(&AdmissionRequestObject{ObjectMeta: metav1.ObjectMeta{}}, annotations{
			"abc": "123",
		}),
	)
	assert.Equal(t, 1, len(p))
	assert.Equal(t, 10, cap(p))
}

func Test_patchObject_withCap(t *testing.T) {
	p := newPatchObject(
		withCap(10),
	)
	assert.Equal(t, 0, len(p))
	assert.Equal(t, 10, cap(p))

}

func Test_patchObject_withAnnotations(t *testing.T) {
	pReplace := newPatchObject(
		withAnnotations(&AdmissionRequestObject{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: annotations{
					"zxc": "890",
				},
			},
		}, annotations{
			"zxc": "456",
		}),
	)
	assert.Equal(t, 1, len(pReplace))
	assert.Equal(t, patchObject{
		patchOperation{
			Op:    "replace",
			Path:  "/metadata/annotations/zxc",
			Value: "456",
		},
	}, pReplace)
	pAdd := newPatchObject(
		withAnnotations(&AdmissionRequestObject{ObjectMeta: metav1.ObjectMeta{}}, annotations{
			"abc": "123",
		}),
	)
	assert.Equal(t, 1, len(pAdd))
	assert.Equal(t, patchObject{
		patchOperation{
			Op:   "add",
			Path: "/metadata/annotations",
			Value: map[string]string{
				"abc": "123",
			},
		},
	}, pAdd)
}

func Test_Annotations_withOwner(t *testing.T) {
	exp := annotations{
		ownersKey: "userA",
	}
	res := newAnnotations(
		withOwners([]string{"userA"}),
	)
	assert.Equal(t, exp, res)

	exp = annotations{
		ownersKey: "userA,userB",
	}
	res = newAnnotations(
		withOwners([]string{"userA", "userB"}),
	)
	assert.Equal(t, exp, res)

}
