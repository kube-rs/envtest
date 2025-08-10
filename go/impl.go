package main

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/versioning"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var (
	codec          runtime.Codec
	environments   Map[string, envtest.Environment]
	jsonSerializer *json.Serializer
)

type Envtest struct{}

func init() {
	EnvTestImpl = Envtest{}
}

func init() {
	ctrl.SetLogger(klog.Background())

	jsonSerializer = json.NewSerializerWithOptions(json.DefaultMetaFactory, latest.Scheme, latest.Scheme, json.SerializerOptions{})
	codec = versioning.NewDefaultingCodecForScheme(
		latest.Scheme,
		jsonSerializer,
		jsonSerializer,
		schema.GroupVersion{Version: latest.Version},
		runtime.InternalGroupVersioner,
	)
}

// FromEnvTestConfig returns a new Kubeconfig string form when running in envtest.
func FromEnvTestConfig(cfg *rest.Config) (string, error) {
	contextName := fmt.Sprintf("%s@envtest", cfg.Username)
	c := api.Config{
		Clusters: map[string]*api.Cluster{
			"envtest": {
				Server:                   cfg.Host,
				CertificateAuthorityData: cfg.CAData,
			},
		},
		Contexts: map[string]*api.Context{
			contextName: {
				Cluster:  "envtest",
				AuthInfo: cfg.Username,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			cfg.Username: {
				ClientKeyData:         cfg.KeyData,
				ClientCertificateData: cfg.CertData,
			},
		},
		CurrentContext: contextName,
	}

	data, err := runtime.Encode(codec, &c)

	return string(data), err
}

// create implements EnvTest.
func (e Envtest) create(req *Environment) (resp CreateResponse) {
	env := &envtest.Environment{
		BinaryAssetsDirectory:        req.binary_assets_settings.binary_assets_directory,
		DownloadBinaryAssetsVersion:  req.binary_assets_settings.download_binary_assets_version,
		DownloadBinaryAssetsIndexURL: req.binary_assets_settings.download_binary_assets_index_url,
		DownloadBinaryAssets:         req.binary_assets_settings.download_binary_assets,
		CRDInstallOptions: envtest.CRDInstallOptions{
			Paths:              req.crd_install_options.paths,
			ErrorIfPathMissing: req.crd_install_options.error_if_path_missing,
		},
	}

	storeErr := func(err error) CreateResponse {
		resp.err = append(resp.err, err.Error())

		return resp
	}

	destroy := func(env *envtest.Environment) {
		if err := env.Stop(); err != nil {
			storeErr(err)
		}
	}

	gvk := apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition")
	for _, data := range req.crd_install_options.crds {
		crd := &apiextensionsv1.CustomResourceDefinition{}
		if _, _, err := jsonSerializer.Decode([]byte(data), &gvk, crd); err != nil {
			return storeErr(err)
		}

		env.CRDs = append(env.CRDs, crd)
	}

	config, err := env.Start()
	if err != nil {
		return storeErr(err)
	}

	kubeconfig, err := FromEnvTestConfig(config)
	if err != nil {
		defer destroy(env)

		return storeErr(err)
	}

	environments.Store(kubeconfig, *env)

	resp.server = Server{kubeconfig: kubeconfig}

	return
}

// destroy implements EnvTest.
func (e Envtest) destroy(kubeconfig *string) (resp DestroyResponse) {
	if kubeconfig == nil || *kubeconfig == "" {
		return
	}

	env, ok := environments.Load(*kubeconfig)
	if !ok {
		return
	}

	storeErr := func(err error) DestroyResponse {
		resp.err = append(resp.err, err.Error())

		return resp
	}

	err := env.Stop()
	if err != nil {
		return storeErr(err)
	}

	return
}
