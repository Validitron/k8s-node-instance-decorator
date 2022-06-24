/*

node-instance-decorator
Centre for Digital Transformation of Health
Copyright Kit Huckvale 2022.

*/

//lint:file-ignore ST1005 Override golang logging/error formatting conventions (use Validitron standard which is 'Sentence case with punctuation.')

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NodeReconciler reconciles built-in Node objects.
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	NAME_TEMPLATE           = "eks-%s-%s-workerNode-%s"
	CLUSTER_TAG_NAME        = "eks:cluster-name"
	NODE_GROUP_TAG_NAME     = "eks:nodegroup-name"
	NAME_TAG_NAME           = "Name"
	NAME_TAG_MAX_LENGTH     = 256
	DEFAULT_REQUEUE_LATENCY = 15 * time.Second
)

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Tells the controller which object type this reconciler will handle.
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)

	node := &corev1.Node{}
	if err := r.Get(ctx, req.NamespacedName, node); err != nil {
		if !k8serr.IsNotFound(err) {
			log.Error(err, "Unable to retrieve node.")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Object is marked for deletion - nothing to do (since the EC2 instance will be removed automatically.)
	if !node.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("Node is marked for deletion: nothing to do.")
		return ctrl.Result{}, nil
	}

	// Retrieve node hostname
	hostName := node.Labels[corev1.LabelHostname]
	if hostName == "" {
		log.Info("No hostname label for node.")
		return ctrl.Result{RequeueAfter: DEFAULT_REQUEUE_LATENCY}, nil
	}

	// Connect to EC2 and retrieve instance details
	// The AWS go library automatically retrieves region, service account-linked role ARN and web identity token from environment variables. See https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/
	// These will be automatically set for the pod in which the operator is running as long as the K8s service account is configured appropriately, see the project README and optionally https://docs.aws.amazon.com/eks/latest/userguide/specify-service-account-role.html
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Error(err, "Failed to load AWS configuration.")
		return ctrl.Result{}, err
	}

	ec2Client := ec2.NewFromConfig(cfg)

	filters := []types.Filter{
		{
			Name: aws.String("network-interface.private-dns-name"),
			Values: []string{
				hostName,
			},
		},
	}
	input := ec2.DescribeInstancesInput{Filters: filters}
	instances, err := ec2Client.DescribeInstances(context.TODO(), &input)
	if err != nil {
		log.Error(err, "Failed to describe EC2 instance.")
		return ctrl.Result{}, err
	}
	if len(instances.Reservations) == 0 {
		log.Info("EC2 found no matching instance reservations.")
		return ctrl.Result{RequeueAfter: DEFAULT_REQUEUE_LATENCY}, nil
	}
	if len(instances.Reservations) > 1 {
		log.Info("EC2 found ambiguous matching instance reservations.")
		return ctrl.Result{RequeueAfter: DEFAULT_REQUEUE_LATENCY}, nil
	}

	reservation := instances.Reservations[0]

	if len(reservation.Instances) == 0 {
		log.Info("EC2 found no matching instances.")
		return ctrl.Result{RequeueAfter: DEFAULT_REQUEUE_LATENCY}, nil
	}
	if len(reservation.Instances) > 1 {
		log.Info("EC2 found ambiguous matching instances.")
		return ctrl.Result{RequeueAfter: DEFAULT_REQUEUE_LATENCY}, nil
	}
	instance := reservation.Instances[0]

	clusterName, err := r.GetTag(&instance, CLUSTER_TAG_NAME)
	if err != nil {
		log.Error(err, fmt.Sprintf("Could not retrieve tag %s.", CLUSTER_TAG_NAME))
		return ctrl.Result{RequeueAfter: DEFAULT_REQUEUE_LATENCY}, err
	}

	nodeGroupName, err := r.GetTag(&instance, NODE_GROUP_TAG_NAME)
	if err != nil {
		log.Error(err, fmt.Sprintf("Could not retrieve tag %s.", NODE_GROUP_TAG_NAME))
		return ctrl.Result{RequeueAfter: DEFAULT_REQUEUE_LATENCY}, err
	}

	ipAddress := instance.PrivateIpAddress

	compositeName := fmt.Sprintf(NAME_TEMPLATE, clusterName, nodeGroupName, *ipAddress)

	runes := []rune(compositeName)
	if len(runes) > NAME_TAG_MAX_LENGTH {
		runes = runes[:NAME_TAG_MAX_LENGTH]
		compositeName = string(runes)
	}

	existingName, _ := r.GetTag(&instance, NAME_TAG_NAME)
	if compositeName == existingName {
		log.Info("EC2 name already set: nothing to do.")
		return ctrl.Result{}, nil
	}

	log.Info(fmt.Sprintf("Setting name '%s' for instance...", compositeName))

	tagInput := ec2.CreateTagsInput{
		Resources: []string{*instance.InstanceId},
		Tags: []types.Tag{
			{
				Key:   aws.String(NAME_TAG_NAME),
				Value: aws.String(compositeName),
			},
		},
	}
	_, err = ec2Client.CreateTags(context.TODO(), &tagInput)
	if err != nil {
		log.Error(err, "Failed to tag EC2 instance name.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NodeReconciler) GetTag(instance *types.Instance, tagKey string) (string, error) {

	for _, r := range instance.Tags {
		if *r.Key == tagKey {
			return *r.Value, nil
		}
	}

	return "", fmt.Errorf("tag '%s' not found", tagKey)
}
