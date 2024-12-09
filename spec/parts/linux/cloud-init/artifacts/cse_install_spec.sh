#!/bin/bash
lsb_release() {
    echo "mock lsb_release"
}

readPackage() {
    local packageName=$1
    package=$(jq ".Packages" "spec/parts/linux/cloud-init/artifacts/test_components.json" | jq ".[] | select(.name == \"$packageName\")")
    echo "$package"
}

Describe 'cse_install.sh'
    Include "./parts/linux/cloud-init/artifacts/cse_install.sh"
    Include "./parts/linux/cloud-init/artifacts/cse_helpers.sh"
    Describe 'installContainerRuntime'
        logs_to_events() {
            echo "mock logs to events calling with $1"
        }
        NEEDS_CONTAINERD="true"
        COMPONENTS_FILEPATH="spec/parts/linux/cloud-init/artifacts/test_components.json"
        It 'returns expected output for successful installation of fake containerd in UBUNTU 20.04'
            UBUNTU_RELEASE="20.04"
            containerdPackage=$(readPackage "containerd")
            When call installContainerRuntime 
            The variable containerdMajorMinorPatchVersion should equal "1.2.3"
            The variable containerdHotFixVersion should equal ""
            The output line 3 should equal "mock logs to events calling with AKS.CSE.installContainerRuntime.installStandaloneContainerd"
            The output line 4 should equal "in installContainerRuntime - CONTAINERD_VERSION = 1.2.3"
        End
        It 'returns expected output for successful installation of containerd in Mariner'
            UBUNTU_RELEASE="" # mocking Mariner doesn't have command `lsb_release -cs`
            OS="MARINER"
            containerdPackage=$(readPackage "containerd")
            When call installContainerRuntime 
            The variable containerdMajorMinorPatchVersion should equal "1.2.3"
            The variable containerdHotFixVersion should equal "5.fake"
            The output line 3 should equal "mock logs to events calling with AKS.CSE.installContainerRuntime.installStandaloneContainerd"
            The output line 4 should equal "in installContainerRuntime - CONTAINERD_VERSION = 1.2.3-5.fake"
        End
        It 'skips the containerd installation for Mariner with Kata'
            UBUNTU_RELEASE="" # mocking Mariner doesn't have command `lsb_release -cs`
            OS="MARINER"
            containerdPackage=$(readPackage "containerd")
            IS_KATA="true"
            When call installContainerRuntime
            The output line 3 should equal "INFO: containerd package versions array is either empty or the first element is <SKIP>. Skipping containerd installation."   
        End         
        It 'returns expected output for successful installation of containerd in AzureLinux'
            UBUNTU_RELEASE="" # mocking AzureLinux doesn't have command `lsb_release -cs`
            OS="AZURELINUX"
            containerdPackage=$(readPackage "containerd")
            When call installContainerRuntime
            The variable containerdMajorMinorPatchVersion should equal "1.7.13"
            The variable containerdHotFixVersion should equal "3.fake"
            The output line 3 should equal "mock logs to events calling with AKS.CSE.installContainerRuntime.installStandaloneContainerd"
            The output line 4 should equal "in installContainerRuntime - CONTAINERD_VERSION = 1.7.13-3.fake"
        End
        It 'skips validation if components.json file is not found'
            COMPONENTS_FILEPATH="non_existent_file.json"
            installContainerdWithManifestJson() {
                echo "mock installContainerdWithManifestJson calling"
            }
            When call installContainerRuntime 
            The output line 2 should equal "Package \"containerd\" does not exist in $COMPONENTS_FILEPATH."
            The output line 3 should equal "mock installContainerdWithManifestJson calling"
        End
    End

    Describe 'updateKubeBinaryRegistryURL'
        logs_to_events() {
            echo "mock logs to events calling with $1"
        }
        It 'returns KUBE_BINARY_URL if it is already registry url'
            KUBE_BINARY_URL="mcr.microsoft.com/oss/binaries/kubernetes/kubernetes-node:v1.30.0-linux-amd64"
            KUBERNETES_VERSION="1.30.0"
            updateKubeBinaryRegistryURL
            When call updateKubeBinaryRegistryURL
            The variable KUBE_BINARY_REGISTRY_URL should equal "$KUBE_BINARY_URL"
            The output line 1 should equal "KUBE_BINARY_URL is a registry url, will use it to pull the kube binary"
        End
        It 'returns expected output from KUBE_BINARY_URL'
            KUBE_BINARY_URL="https://acs-mirror.azureedge.net/kubernetes/v1.30.0-hotfix20241209/binaries/kubernetes-nodes-linux-amd64.tar.gz"
            BOOTSTRAP_PROFILE_CONTAINER_REGISTRY_SERVER="mcr.microsoft.com"
            KUBERNETES_VERSION="1.30.0"
            CPU_ARCH="amd64"
            updateKubeBinaryRegistryURL
            When call updateKubeBinaryRegistryURL
            The variable KUBE_BINARY_REGISTRY_URL should equal "$BOOTSTRAP_PROFILE_CONTAINER_REGISTRY_SERVER/oss/binaries/kubernetes/kubernetes-node:v1.30.0-hotfix20241209-linux-amd64"
            The output line 1 should equal "Extracted version: v1.30.0-hotfix20241209 from KUBE_BINARY_URL: $KUBE_BINARY_URL"
        End
        It 'returns expected output for moonckae acs-mirror'
            KUBE_BINARY_URL="https://acs-mirror.azureedge.cn/kubernetes/v1.30.0-alpha/binaries/kubernetes-nodes-linux-amd64.tar.gz"
            BOOTSTRAP_PROFILE_CONTAINER_REGISTRY_SERVER="mcr.microsoft.com"
            KUBERNETES_VERSION="1.30.0"
            CPU_ARCH="amd64"
            updateKubeBinaryRegistryURL
            When call updateKubeBinaryRegistryURL
            The variable KUBE_BINARY_REGISTRY_URL should equal "$BOOTSTRAP_PROFILE_CONTAINER_REGISTRY_SERVER/oss/binaries/kubernetes/kubernetes-node:v1.30.0-alpha-linux-amd64"
            The output line 1 should equal "Extracted version: v1.30.0-alpha from KUBE_BINARY_URL: $KUBE_BINARY_URL"
        End
        It 'uses KUBENETES_VERSION if KUBE_BINARY_URL is invalid'
            KUBE_BINARY_URL="https://invalidpath/v1.30.0-lts100/binaries/kubernetes-nodes-linux-amd64.tar.gz"
            BOOTSTRAP_PROFILE_CONTAINER_REGISTRY_SERVER="mcr.microsoft.com"
            KUBERNETES_VERSION="1.30.0"
            CPU_ARCH="amd64"
            updateKubeBinaryRegistryURL
            When call updateKubeBinaryRegistryURL
            The variable KUBE_BINARY_REGISTRY_URL should equal "$BOOTSTRAP_PROFILE_CONTAINER_REGISTRY_SERVER/oss/binaries/kubernetes/kubernetes-node:v1.30.0-linux-amd64"
            The output line 1 should equal "KUBE_BINARY_URL is formatted unexpectedly, will use the kubernetes version as binary version: v$KUBERNETES_VERSION"
        End
    End
End