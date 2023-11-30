/*
This package is used for decoding yaml file into objects.
*/
package decodeYaml

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

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

// decode yaml data into a deployment, data can be read from a file or an url
func DecDeployment(data []byte) (*appsv1.Deployment, error) {
	// get decoder
	decoder := GetDecoder()

	// decode yaml for deployment
	deployment := new(appsv1.Deployment)
	_, _, err := decoder.Decode(data, nil, deployment)
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

// get yaml data from a file
func ReadYamlFile(file string) ([]byte, error) {
	// read yaml
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// get yaml data from an url
func ReadYamlUrl(urlPath string) ([]byte, error) {
	// check ext
	u, err := url.Parse(urlPath)
	if err != nil {
		return nil, err
	}
	fileExt := path.Ext(u.Path)
	if fileExt != ".yaml" && fileExt != ".yml" {
		return nil, fmt.Errorf("file ext in url is invalid, should be .yaml or .yml file")
	}

	// check size
	resp, err := http.Head(urlPath)
	if err != nil {
		return nil, err
	}
	fsize := resp.ContentLength
	if fsize > 1024*1024 {
		return nil, fmt.Errorf("yaml file is too big, limited to 1 Mib")
	}

	// get yaml data from url
	resp, err = http.Get(urlPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read into a buffer
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
