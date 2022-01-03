package status

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/fluxcd/pkg/runtime/conditions"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// checkFunc is the function type for all the status check functions.
type checkFunc func(ctx context.Context, scheme *runtime.Scheme, obj conditions.Getter, condns *Conditions) error

// Checker performs all the status checks. It is configured to provide context
// of the target controller.
type Checker struct {
	// K8s client, to fetch the latest version of an object.
	client.Client
	// scheme is the scheme of the target objects.
	scheme *runtime.Scheme
	// conditions is the conditions context of the target controller.
	conditions *Conditions
	// failChecks contains all the strict checks.
	failChecks []checkFunc
	// warnChecks contains all the checks that result in warnings.
	warnChecks []checkFunc
	// DisableFetch disables fetching the latest state of an object using the
	// client. This can be used in unit-tests, while passing an object with
	// all the properties to be checked.
	DisableFetch bool
	// Stdout of the checker.
	Stdout io.Writer
	// Stderr of the checker.
	Stderr io.Writer
	// ExcludeChecks contains the checks that should be excluded.
	// TODO: Add support for it in all the checks.
	// ExcludeChecks map[string]bool
}

// NewChecker constructs and returns a new Checker for a controller.
func NewChecker(cli client.Client, scheme *runtime.Scheme, condns *Conditions) *Checker {
	warnChecks := []checkFunc{
		check_WARN0001,
		check_WARN0002,
		check_WARN0003,
		check_WARN0004,
		check_WARN0005,
	}
	failChecks := []checkFunc{
		check_FAIL0001,
		check_FAIL0002,
		check_FAIL0003,
		check_FAIL0004,
		check_FAIL0005,
		check_FAIL0006,
		check_FAIL0007,
		check_FAIL0008,
		check_FAIL0009,
		check_FAIL0010,
	}
	return &Checker{
		Client:     cli,
		scheme:     scheme,
		conditions: condns,
		warnChecks: warnChecks,
		failChecks: failChecks,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	}
}

// CheckErr performs all the warn and fail checks and prints them to stdout and
// stderr, and exits. This is to be used in CLI.
func (c Checker) CheckErr(ctx context.Context, obj conditions.Getter) {
	fail, warn := c.Check(ctx, obj)
	if warn != nil {
		fmt.Fprintf(c.Stdout, "[Check-WARN]: %v\n", warn)
	}
	if fail != nil {
		fmt.Fprintf(c.Stderr, "[Check-FAIL]: %v\n", fail)
		os.Exit(1)
	}
}

// Check performs all the warn and fail checks and returns the results.
func (c Checker) Check(ctx context.Context, obj conditions.Getter) (fail, warn error) {
	if c.conditions == nil {
		return fmt.Errorf("no conditions context provided"), nil
	}
	// Fetch the latest version of the object.
	if !c.DisableFetch {
		if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
			return err, nil
		}
	}
	warnErrs := []error{}
	for _, check := range c.warnChecks {
		if err := check(ctx, c.scheme, obj, c.conditions); err != nil {
			warnErrs = append(warnErrs, err)
		}
	}
	warn = kerrors.NewAggregate(warnErrs)
	failErr := []error{}
	for _, check := range c.failChecks {
		if err := check(ctx, c.scheme, obj, c.conditions); err != nil {
			failErr = append(failErr, err)
		}
	}
	fail = kerrors.NewAggregate(failErr)
	return fail, warn
}
