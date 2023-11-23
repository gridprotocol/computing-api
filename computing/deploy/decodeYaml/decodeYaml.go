/*
This package is used for decoding yaml file into objects.
*/
package decodeYaml

import (
	"os"

	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// get decoder with apimachinery lib
func GetDecoder() runtime.Decoder {
	// https://godoc.org/k8s.io/apimachinery/pkg/runtime#Scheme
	scheme := runtime.NewScheme()
	// https://godoc.org/k8s.io/apimachinery/pkg/runtime/serializer#CodecFactory
	codecFactory := serializer.NewCodecFactory(scheme)
	// https://godoc.org/k8s.io/apimachinery/pkg/runtime#Decoder
	decoder := codecFactory.UniversalDeserializer()

	return decoder
}

// decode deployment from yaml filepath
func DecDeployment(yaml string) (*appsv1.Deployment, error) {
	// get decoder
	decoder := GetDecoder()

	// read yaml
	deploymentYAML, err := os.ReadFile(yaml)
	if err != nil {
		return nil, err
	}

	// decode yaml for deployment
	deployment := new(appsv1.Deployment)
	_, _, err = decoder.Decode(deploymentYAML, nil, deployment)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

// decode namespace from yaml filepath
func DecNamespace(yaml string) (*corev1.Namespace, error) {
	// get decoder
	decoder := GetDecoder()

	// read yaml
	namespaceYAML, err := os.ReadFile(yaml)
	if err != nil {
		return nil, err
	}

	// decode yaml for namespace
	namespace := new(corev1.Namespace)
	_, _, err = decoder.Decode(namespaceYAML, nil, namespace)
	if err != nil {
		return nil, err
	}

	fmt.Printf("namespace: %s\n", namespace.ObjectMeta.GetName())

	return namespace, nil
}
