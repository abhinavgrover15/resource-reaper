package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	ResourceReaperAnnotation = "resource-reaper/ttl"
)

// ResourceReaper reconciles objects with TTL annotations
type ResourceReaper struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceReaper) SetupWithManager(mgr ctrl.Manager) error {
	// Create a slice of GVKs we want to watch
	gvks := []schema.GroupVersionKind{
		// Core resources
		{Version: "v1", Kind: "Pod"},
		// Apps resources
		{Group: "apps", Version: "v1", Kind: "Deployment"},
		{Group: "apps", Version: "v1", Kind: "StatefulSet"},
		{Group: "apps", Version: "v1", Kind: "DaemonSet"},
		// Core resources
		{Version: "v1", Kind: "Service"},
		{Version: "v1", Kind: "ConfigMap"},
		{Version: "v1", Kind: "Secret"},
		// Batch resources
		{Group: "batch", Version: "v1", Kind: "Job"},
		{Group: "batch", Version: "v1", Kind: "CronJob"},
	}

	// Create a predicate for TTL annotation
	ttlPredicate := predicate.NewPredicateFuncs(func(object client.Object) bool {
		_, exists := object.GetAnnotations()[ResourceReaperAnnotation]
		return exists
	})

	// Create a controller for each GVK
	for _, gvk := range gvks {
		obj := &metav1.PartialObjectMetadata{}
		obj.SetGroupVersionKind(gvk)

		builder := ctrl.NewControllerManagedBy(mgr).
			WithEventFilter(ttlPredicate).
			For(obj)

		if err := builder.Complete(r); err != nil {
			return fmt.Errorf("failed to add watch for %v: %w", gvk, err)
		}
	}

	return nil
}

// setupScheme registers all the resource types with the scheme
func (r *ResourceReaper) setupScheme(scheme *runtime.Scheme) error {
	if err := corev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error adding core types to scheme: %w", err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error adding apps types to scheme: %w", err)
	}
	if err := batchv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error adding batch types to scheme: %w", err)
	}
	if err := metav1.AddMetaToScheme(scheme); err != nil {
		return fmt.Errorf("error adding meta types to scheme: %w", err)
	}
	return nil
}

// parseDuration parses duration strings like "24h", "30m", etc.
func parseDuration(value string) (time.Duration, error) {
	return time.ParseDuration(value)
}

// Reconcile handles the reconciliation loop for TTL-annotated resources
func (r *ResourceReaper) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("request", req)

	// Get the resource using PartialObjectMetadata
	obj := &metav1.PartialObjectMetadata{}

	// Try different GVKs until we find the right one
	gvks := []schema.GroupVersionKind{
		// Core resources
		{Version: "v1", Kind: "Pod"},
		{Version: "v1", Kind: "Service"},
		{Version: "v1", Kind: "ConfigMap"},
		{Version: "v1", Kind: "Secret"},
		// Apps resources
		{Group: "apps", Version: "v1", Kind: "Deployment"},
		{Group: "apps", Version: "v1", Kind: "StatefulSet"},
		{Group: "apps", Version: "v1", Kind: "DaemonSet"},
		// Batch resources
		{Group: "batch", Version: "v1", Kind: "Job"},
		{Group: "batch", Version: "v1", Kind: "CronJob"},
	}

	var foundResource bool
	for _, gvk := range gvks {
		obj.SetGroupVersionKind(gvk)
		if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
			if !errors.IsNotFound(err) {
				continue
			}
			// If we've tried all GVKs and none worked, return
			if gvk == gvks[len(gvks)-1] {
				return ctrl.Result{}, nil
			}
			continue
		}
		foundResource = true
		break
	}

	if !foundResource {
		return ctrl.Result{}, nil
	}

	// Check if resource has TTL annotation
	ttlStr, exists := obj.GetAnnotations()[ResourceReaperAnnotation]
	if !exists {
		return ctrl.Result{}, nil
	}

	// Parse TTL duration
	ttl, err := parseDuration(ttlStr)
	if err != nil {
		logger.Error(err, "Invalid TTL annotation value", "ttl", ttlStr)
		return ctrl.Result{}, fmt.Errorf("invalid TTL value: %s", ttlStr)
	}

	// Handle negative TTL
	if ttl < 0 {
		err := fmt.Errorf("negative TTL is invalid: %s", ttlStr)
		logger.Error(err, "Negative TTL is invalid", "ttl", ttlStr)
		return ctrl.Result{}, err
	}

	// Handle zero or positive TTL
	creationTime := obj.GetCreationTimestamp().Time
	expirationTime := creationTime.Add(ttl)
	now := time.Now()

	// If TTL is zero or resource has expired, delete it
	if ttl == 0 || now.After(expirationTime) {
		logger.Info("Deleting resource due to TTL expiration",
			"name", obj.GetName(),
			"namespace", obj.GetNamespace(),
			"kind", obj.GetObjectKind().GroupVersionKind().Kind,
			"group", obj.GetObjectKind().GroupVersionKind().Group,
			"ttl", ttlStr,
			"creationTime", creationTime,
			"expirationTime", expirationTime,
			"currentTime", now)

		if err := r.Delete(ctx, obj); err != nil {
			if !errors.IsNotFound(err) {
				logger.Error(err, "Unable to delete resource")
				return ctrl.Result{}, err
			}
		}
		logger.Info("Successfully deleted resource",
			"name", obj.GetName(),
			"namespace", obj.GetNamespace(),
			"kind", obj.GetObjectKind().GroupVersionKind().Kind,
			"group", obj.GetObjectKind().GroupVersionKind().Group)
		return ctrl.Result{}, nil
	}

	logger.Info("Resource not expired yet",
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
		"kind", obj.GetObjectKind().GroupVersionKind().Kind,
		"group", obj.GetObjectKind().GroupVersionKind().Group,
		"ttl", ttlStr,
		"creationTime", creationTime,
		"expirationTime", expirationTime,
		"timeUntilExpiration", expirationTime.Sub(now))

	// Resource hasn't expired yet, requeue at expiration time
	return ctrl.Result{RequeueAfter: expirationTime.Sub(now)}, nil
}
