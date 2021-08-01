package webhookserver

import (
	"context"
	"encoding/json"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/inject-pod-agent,mutating=true,sideEffects=none,failurePolicy=ignore,groups="",resources=pods,verbs=create,versions=v1,admissionReviewVersions=v1,name=pod-agent.k8s.tarian.io

type PodAgentInjector struct {
	Client  client.Client
	decoder *admission.Decoder
	config  PodAgentContainerConfig
}

type PodAgentContainerConfig struct {
	Name        string
	Image       string
	LogEncoding string
	Host        string
	Port        string
}

// podAnnotator adds an annotation to every incoming pods.
func (p *PodAgentInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := p.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if pod.Annotations == nil {
		return admission.Allowed("no annotation found")
	}

	if _, ok := pod.Annotations["pod-agent.k8s.tarian.io/inject"]; !ok {
		return admission.Allowed("annotation pod-agent.k8s.tarian.io/inject not found")
	}

	podAgentContainer := corev1.Container{
		Name:  p.config.Name,
		Image: p.config.Image,
		Env: []corev1.EnvVar{
			{
				Name: "NAMESPACE", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
			},
		},
		Args: []string{
			"--log-encoding=" + p.config.LogEncoding,
			"run",
			"--host=" + p.config.Host,
			"--port=" + p.config.Port,
			"--namespace=$(NAMESPACE)",
			"--pod-labels-file==/etc/podinfo/labels",
		},
		VolumeMounts: []corev1.VolumeMount{{Name: "podinfo", MountPath: "/etc/podinfo"}},
	}
	pod.Spec.Containers = append(pod.Spec.Containers, podAgentContainer)
	pod.Spec.ShareProcessNamespace = pointer.BoolPtr(true)

	podInfoVolume := corev1.Volume{
		Name: "podinfo",
		VolumeSource: corev1.VolumeSource{
			DownwardAPI: &corev1.DownwardAPIVolumeSource{
				Items: []corev1.DownwardAPIVolumeFile{{Path: "labels", FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels"}}},
			},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, podInfoVolume)

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// InjectDecoder injects the decoder.
func (p *PodAgentInjector) InjectDecoder(d *admission.Decoder) error {
	p.decoder = d
	return nil
}
