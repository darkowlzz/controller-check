package status

import (
	"context"
	"testing"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/darkowlzz/controller-check/testdata"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name         string
		disableFetch bool
		wantWarn     bool
		wantFail     bool
	}{
		{
			name:     "with fetch",
			wantWarn: true,
		},
		{
			name:         "without fetch",
			disableFetch: true,
			wantFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			obj := &testdata.Fake{}
			obj.Name = "TestObj"
			obj.Namespace = "TestNS"
			obj.Generation = 4
			obj.Status.ObservedGeneration = 3

			// Create a copy of the initial version of the object.
			objOld := obj.DeepCopy()

			// Update the object with new conditions and add it to the fake client.
			conditions.MarkFalse(obj, meta.ReadyCondition, "SomeReason", "SomeMsg")
			conditions.MarkTrue(obj, "TestCondition1", "Rsn", "Msg")

			scheme := runtime.NewScheme()
			g.Expect(testdata.AddFakeToScheme(scheme)).To(Succeed())
			builder := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(obj)

			// Register negative polarity conditions with the checker.
			conditions := &Conditions{NegativePolarity: []string{"TestCondition1", "TestCondition2"}}
			checker := NewChecker(builder.Build(), scheme, conditions)
			checker.DisableFetch = tt.disableFetch

			fail, warn := checker.Check(context.TODO(), objOld)
			g.Expect(warn != nil).To(Equal(tt.wantWarn))
			g.Expect(fail != nil).To(Equal(tt.wantFail))
		})
	}
}
