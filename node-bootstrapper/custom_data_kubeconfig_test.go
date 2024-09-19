package main

import (
	"fmt"
	"github.com/Azure/agentbaker/pkg/agent/datamodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	bootstrapConfigFile        = "/var/lib/kubelet/bootstrap-kubeconfig"
	kubeConfigFile             = "/var/lib/kubelet/kubeconfig"
	bootstrapConfigFileWindows = "c:\\k\\bootstrap-config"
	kubeConfigFileWindows      = "c:\\k\\config"
	arcTokenSh                 = "/opt/azure/bootstrap/arc-token.sh"
	azureTokenSh               = "/opt/azure/bootstrap/azure-token.sh"
)

func assertKubeconfig(t *testing.T, nbc *datamodel.NodeBootstrappingConfiguration, expected string) {
	t.Helper()
	files, err := customData(nil, nbc)
	require.NoError(t, err)
	require.NotContains(t, files, bootstrapConfigFile)
	var configFile = kubeConfigFile
	if nbc.AgentPoolProfile.IsWindows() {
		configFile = kubeConfigFileWindows
	}
	actual := getFile(t, nbc, configFile, 0644)
	assert.YAMLEq(t, expected, actual)
}

func assertBootstrapKubeconfig(t *testing.T, nbc *datamodel.NodeBootstrappingConfiguration, expected string) {
	t.Helper()
	files, err := customData(nil, nbc)
	require.NoError(t, err)
	require.NotContains(t, files, kubeConfigFile)
	var configFile = bootstrapConfigFile
	if nbc.AgentPoolProfile.IsWindows() {
		configFile = bootstrapConfigFileWindows
	}
	actual := getFile(t, nbc, configFile, 0644)
	assert.YAMLEq(t, expected, actual)
}

func assertArcTokenSh(t *testing.T, nbc *datamodel.NodeBootstrappingConfiguration, aadAppId string) {
	t.Helper()
	files, err := customData(nil, nbc)
	require.NoError(t, err)
	require.NotContains(t, files, azureTokenSh)
	actual := getFile(t, nbc, arcTokenSh, 0755)
	expected := fmt.Sprintf(`#!/bin/bash

# Fetch an AAD token from Azure Arc HIMDS and output it in the ExecCredential format
# https://learn.microsoft.com/azure/azure-arc/servers/managed-identity-authentication

TOKEN_URL="http://127.0.0.1:40342/metadata/identity/oauth2/token?api-version=2019-11-01&resource=%s"
EXECCREDENTIAL='''
{
  "kind": "ExecCredential",
  "apiVersion": "client.authentication.k8s.io/v1",
  "spec": {
    "interactive": false
  },
  "status": {
    "expirationTimestamp": .expires_on | tonumber | todate,
    "token": .access_token
  }
}
'''

# Arc IMDS requires a challenge token from a file only readable by root for security
CHALLENGE_TOKEN_PATH=$(curl -s -D - -H Metadata:true $TOKEN_URL | grep Www-Authenticate | cut -d "=" -f 2 | tr -d "[:cntrl:]")
CHALLENGE_TOKEN=$(cat $CHALLENGE_TOKEN_PATH)
if [ $? -ne 0 ]; then
    echo "Could not retrieve challenge token, double check that this command is run with root privileges."
    exit 255
fi

curl -s -H Metadata:true -H "Authorization: Basic $CHALLENGE_TOKEN" $TOKEN_URL | jq "$EXECCREDENTIAL"
`, aadAppId)
	assert.Equal(t, expected, actual)
}

func assertAzureTokenSh(t *testing.T, nbc *datamodel.NodeBootstrappingConfiguration, aadAppId string) {
	t.Helper()
	files, err := customData(nil, nbc)
	require.NoError(t, err)
	require.NotContains(t, files, arcTokenSh)
	actual := getFile(t, nbc, azureTokenSh, 0755)
	expected := fmt.Sprintf(`#!/bin/bash

TOKEN_URL="http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=%s"
EXECCREDENTIAL='''
{
  "kind": "ExecCredential",
  "apiVersion": "client.authentication.k8s.io/v1",
  "spec": {
    "interactive": false
  },
  "status": {
    "expirationTimestamp": .expires_on | tonumber | todate,
    "token": .access_token
  }
}
'''

curl -s -H Metadata:true $TOKEN_URL | jq "$EXECCREDENTIAL"
`, aadAppId)
	assert.Equal(t, expected, actual)
}

func TestKubeConfigGeneratedCorrectly(t *testing.T) {

	t.Run("kubeconfig", func(t *testing.T) {
		nbc := validNBC()
		assertKubeconfig(t, nbc, `
apiVersion: v1
kind: Config
clusters:
- name: localcluster
  cluster:
    certificate-authority: /etc/kubernetes/certs/ca.crt
    server: https://:443
users:
- name: client
  user:
    client-certificate: /etc/kubernetes/certs/client.crt
    client-key: /etc/kubernetes/certs/client.key
contexts:
- context:
    cluster: localcluster
    user: client
  name: localclustercontext
current-context: localclustercontext
`)
	})

	t.Run("bootstrap-kubeconfig", func(t *testing.T) {
		nbc := validNBC()
		nbc.KubeletClientTLSBootstrapToken = Ptr("test-token")
		assertBootstrapKubeconfig(t, nbc, `apiVersion: v1
clusters:
    - cluster:
        certificate-authority: /etc/kubernetes/certs/ca.crt
        server: https://:443
      name: localcluster
contexts:
    - context:
        cluster: localcluster
        user: kubelet-bootstrap
      name: bootstrap-context
current-context: bootstrap-context
kind: Config
users:
    - name: kubelet-bootstrap
      user:
        token: test-token
`)
	})

	t.Run("secureTlsBootstrapKubeConfig sets bootstrap-kubeconfig correctly", func(t *testing.T) {
		nbc := validNBC()
		nbc.EnableSecureTLSBootstrapping = true
		assertBootstrapKubeconfig(t, nbc, `apiVersion: v1
clusters:
    - cluster:
        certificate-authority: /etc/kubernetes/certs/ca.crt
        server: https://:443
      name: localcluster
contexts:
    - context:
        cluster: localcluster
        user: kubelet-bootstrap
      name: bootstrap-context
current-context: bootstrap-context
kind: Config
users:
- name: kubelet-bootstrap
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1
      args:
        - bootstrap
        - --next-proto=aks-tls-bootstrap
        - --aad-resource=test-app-id
      command: /opt/azure/tlsbootstrap/tls-bootstrap-client
      interactiveMode: Never
      provideClusterInfo: true
`)
	})

	t.Run("BootstrappingMethod=UseSecureTlsBootstrapping sets bootstrap-kubeconfig correctly", func(t *testing.T) {
		nbc := validNBC()
		nbc.BootstrappingMethod = datamodel.UseSecureTlsBootstrapping
		assertBootstrapKubeconfig(t, nbc, `apiVersion: v1
clusters:
    - cluster:
        certificate-authority: /etc/kubernetes/certs/ca.crt
        server: https://:443
      name: localcluster
contexts:
    - context:
        cluster: localcluster
        user: kubelet-bootstrap
      name: bootstrap-context
current-context: bootstrap-context
kind: Config
users:
- name: kubelet-bootstrap
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1
      args:
        - bootstrap
        - --next-proto=aks-tls-bootstrap
        - --aad-resource=test-app-id
      command: /opt/azure/tlsbootstrap/tls-bootstrap-client
      interactiveMode: Never
      provideClusterInfo: true
`)
	})

	t.Run("BootstrappingMethod=UseTlsBootstrapToken sets bootstrap-kubeconfig correctly", func(t *testing.T) {
		nbc := validNBC()
		nbc.BootstrappingMethod = datamodel.UseTlsBootstrapToken
		nbc.KubeletClientTLSBootstrapToken = Ptr("test-token-value")
		assertBootstrapKubeconfig(t, nbc, `apiVersion: v1
clusters:
    - cluster:
        certificate-authority: /etc/kubernetes/certs/ca.crt
        server: https://:443
      name: localcluster
contexts:
    - context:
        cluster: localcluster
        user: kubelet-bootstrap
      name: bootstrap-context
current-context: bootstrap-context
kind: Config
users:
    - name: kubelet-bootstrap
      user:
        token: test-token-value
`)
	})

	t.Run("BootstrappingMethod=UseArcMsiToMakeCSR sets bootstrap-kubeconfig correctly", func(t *testing.T) {
		nbc := validNBC()
		nbc.BootstrappingMethod = datamodel.UseArcMsiToMakeCSR
		assertBootstrapKubeconfig(t, nbc, `apiVersion: v1
clusters:
    - cluster:
        certificate-authority: /etc/kubernetes/certs/ca.crt
        server: https://:443
      name: localcluster
contexts:
    - context:
        cluster: localcluster
        user: kubelet-bootstrap
      name: bootstrap-context
current-context: bootstrap-context
kind: Config
users:
    - name: kubelet-bootstrap
      user:
        exec:
          apiVersion: client.authentication.k8s.io/v1
          command: /opt/azure/bootstrap/arc-token.sh
          interactiveMode: Never
          provideClusterInfo: false
`)
	})

	t.Run("BootstrappingMethod=UseArcMsiToMakeCSR sets token.sh correctly with the AKS AAD App ID", func(t *testing.T) {
		nbc := validNBC()
		nbc.CustomSecureTLSBootstrapAADServerAppID = ""
		nbc.BootstrappingMethod = datamodel.UseArcMsiToMakeCSR
		assertArcTokenSh(t, nbc, "6dae42f8-4368-4678-94ff-3960e28e3630")
	})

	t.Run("BootstrappingMethod=UseArcMsiToMakeCSR sets token.sh correctly with a different AKS AAD App ID", func(t *testing.T) {
		nbc := validNBC()
		nbc.CustomSecureTLSBootstrapAADServerAppID = "different_app_id"
		nbc.BootstrappingMethod = datamodel.UseArcMsiToMakeCSR
		assertArcTokenSh(t, nbc, "different_app_id")
	})

	t.Run("BootstrappingMethod=UseArcMsiDirectly sets kubeconfig correctly", func(t *testing.T) {
		nbc := validNBC()
		nbc.BootstrappingMethod = datamodel.UseArcMsiDirectly
		assertKubeconfig(t, nbc, `
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /etc/kubernetes/certs/ca.crt
    server: https://:443
  name: localcluster
contexts:
- context:
    cluster: localcluster
    user: client
  name: localclustercontext
current-context: localclustercontext
kind: Config
users:
- name: default-auth
  user:
    exec:
       apiVersion: client.authentication.k8s.io/v1
       command: /opt/azure/bootstrap/arc-token.sh
       provideClusterInfo: false
`)
	})

	t.Run("BootstrappingMethod=UseArcMsiDirectly sets token.sh correctly with the AKS AAD App ID", func(t *testing.T) {
		nbc := validNBC()
		nbc.CustomSecureTLSBootstrapAADServerAppID = ""
		nbc.BootstrappingMethod = "UseArcMsiDirectly"
		assertArcTokenSh(t, nbc, "6dae42f8-4368-4678-94ff-3960e28e3630")
	})

	t.Run("BootstrappingMethod=UseArcMsiDirectly sets token.sh correctly with a different AKS AAD App ID", func(t *testing.T) {
		nbc := validNBC()
		nbc.CustomSecureTLSBootstrapAADServerAppID = "different_app_id"
		nbc.BootstrappingMethod = "UseArcMsiDirectly"
		assertArcTokenSh(t, nbc, "different_app_id")
	})

	t.Run("BootstrappingMethod=UseAzureMsiDirectly sets kubeconfig correctly", func(t *testing.T) {
		nbc := validNBC()
		nbc.BootstrappingMethod = datamodel.UseAzureMsiDirectly
		assertKubeconfig(t, nbc, `
apiVersion: v1
clusters:
   - cluster:
       certificate-authority: /etc/kubernetes/certs/ca.crt
       server: https://:443
     name: localcluster
contexts:
   - context:
       cluster: localcluster
       user: client
     name: localclustercontext
current-context: localclustercontext
kind: Config
users:
   - name: default-auth
     user:
       exec:
         apiVersion: client.authentication.k8s.io/v1
         command: /opt/azure/bootstrap/azure-token.sh
         provideClusterInfo: false
`)
	})

	t.Run("BootstrappingMethod=UseAzureMsiDirectly sets token.sh correctly with the AKS AAD App ID", func(t *testing.T) {
		nbc := validNBC()
		nbc.CustomSecureTLSBootstrapAADServerAppID = ""
		nbc.BootstrappingMethod = datamodel.UseAzureMsiDirectly
		assertAzureTokenSh(t, nbc, "6dae42f8-4368-4678-94ff-3960e28e3630")
	})

	t.Run("BootstrappingMethod=UseAzureMsiDirectly sets token.sh correctly with a different AKS AAD App ID", func(t *testing.T) {
		nbc := validNBC()
		nbc.CustomSecureTLSBootstrapAADServerAppID = "different_app_id"
		nbc.BootstrappingMethod = datamodel.UseAzureMsiDirectly
		assertAzureTokenSh(t, nbc, "different_app_id")
	})

	t.Run("BootstrappingMethod=UseAzureMsiDirectly and windows sets kubeconfig correctly", func(t *testing.T) {
		nbc := validNBC()
		nbc.AgentPoolProfile.OSType = datamodel.Windows
		nbc.BootstrappingMethod = datamodel.UseAzureMsiDirectly
		assertKubeconfig(t, nbc, `
apiVersion: v1
clusters:
   - cluster:
       certificate-authority: c:/k/ca.crt
       server: https://:443
     name: localcluster
contexts:
   - context:
       cluster: localcluster
       user: client
     name: localclustercontext
current-context: localclustercontext
kind: Config
users:
   - name: default-auth
     user:
       exec:
         apiVersion: client.authentication.k8s.io/v1
         command: kubelogin
         args:
         - get-token
         - --environment
         - AzurePublicCloud
         - --server-id
         - test-app-id
         - --login
         - msi
         - --client-id
         - 5f0b9406-fbf1-4e1c-8a61-b6f4a6702057
         provideClusterInfo: false
`)
	})

	t.Run("BootstrappingMethod=UseAzureMsiDirectly and windows has no token-azure or token-arc.sh", func(t *testing.T) {
		nbc := validNBC()
		nbc.AgentPoolProfile.OSType = datamodel.Windows
		nbc.BootstrappingMethod = datamodel.UseAzureMsiDirectly

		files, err := customData(nil, nbc)
		require.NoError(t, err)
		require.NotContains(t, files, arcTokenSh)
		require.NotContains(t, files, azureTokenSh)
	})

	t.Run("BootstrappingMethod=UseAzureMsiDirectly sets kubeconfig correctly", func(t *testing.T) {
		nbc := validNBC()
		nbc.BootstrappingMethod = datamodel.UseAzureMsiToMakeCSR
		assertBootstrapKubeconfig(t, nbc, `apiVersion: v1
clusters:
    - cluster:
        certificate-authority: /etc/kubernetes/certs/ca.crt
        server: https://:443
      name: localcluster
contexts:
    - context:
        cluster: localcluster
        user: kubelet-bootstrap
      name: bootstrap-context
current-context: bootstrap-context
kind: Config
users:
    - name: kubelet-bootstrap
      user:
        exec:
          apiVersion: client.authentication.k8s.io/v1
          command: /opt/azure/bootstrap/azure-token.sh
          interactiveMode: Never
          provideClusterInfo: false
`)
	})

	t.Run("BootstrappingMethod=UseAzureMsiDirectly sets token.sh correctly with the AKS AAD App ID", func(t *testing.T) {
		nbc := validNBC()
		nbc.CustomSecureTLSBootstrapAADServerAppID = ""
		nbc.BootstrappingMethod = datamodel.UseAzureMsiToMakeCSR
		assertAzureTokenSh(t, nbc, "6dae42f8-4368-4678-94ff-3960e28e3630")
	})

	t.Run("BootstrappingMethod=UseAzureMsiDirectly sets token.sh correctly with a different AKS AAD App ID", func(t *testing.T) {
		nbc := validNBC()
		nbc.CustomSecureTLSBootstrapAADServerAppID = "different_app_id"
		nbc.BootstrappingMethod = datamodel.UseAzureMsiToMakeCSR
		assertAzureTokenSh(t, nbc, "different_app_id")
	})

	t.Run("BootstrappingMethod=UseAzureMsiToMakeCSR and windows sets bootstrap kubeconfig correctly", func(t *testing.T) {
		nbc := validNBC()
		nbc.AgentPoolProfile.OSType = datamodel.Windows
		nbc.BootstrappingMethod = datamodel.UseAzureMsiToMakeCSR
		assertBootstrapKubeconfig(t, nbc, `
apiVersion: v1
clusters:
   - cluster:
       certificate-authority: c:/k/ca.crt
       server: https://:443
     name: localcluster
contexts:
   - context:
       cluster: localcluster
       user: kubelet-bootstrap
     name: bootstrap-context
current-context: bootstrap-context
kind: Config
users:
   - name: kubelet-bootstrap
     user:
       exec:
         apiVersion: client.authentication.k8s.io/v1
         command: kubelogin
         args:
         - get-token
         - --environment
         - AzurePublicCloud
         - --server-id
         - test-app-id
         - --login
         - msi
         - --client-id
         - 5f0b9406-fbf1-4e1c-8a61-b6f4a6702057
         provideClusterInfo: false
         interactiveMode: Never
`)
	})

	t.Run("BootstrappingMethod=UseAzureMsiToMakeCSR and windows has no token-azure or token-arc.sh", func(t *testing.T) {
		nbc := validNBC()
		nbc.AgentPoolProfile.OSType = datamodel.Windows
		nbc.BootstrappingMethod = datamodel.UseAzureMsiToMakeCSR

		files, err := customData(nil, nbc)
		require.NoError(t, err)
		require.NotContains(t, files, arcTokenSh)
		require.NotContains(t, files, azureTokenSh)
	})
}
