<#
    .SYNOPSIS
        Provisions VM as a Kubernetes agent.

    .DESCRIPTION
        Provisions VM as a Kubernetes agent.

        The parameters passed in are required, and will vary per-deployment.

        Notes on modifying this file:
        - This file extension is PS1, but it is actually used as a template from pkg/engine/template_generator.go
        - All of the lines that have braces in them will be modified. Please do not change them here, change them in the Go sources
        - Single quotes are forbidden, they are reserved to delineate the different members for the ARM template concat() call
        - windowscsehelper.ps1 contains basic util functions. It will be compressed to a zip file and then be converted to base64 encoding
          string and stored in $zippedFiles. Reason: This script is a template and has some limitations.
        - All other scripts will be packaged and published in a single package. It will be downloaded in provisioning VM.
          Reason: CustomData has length limitation 87380.
        - ProvisioningScriptsPackage contains scripts to start kubelet, kubeproxy, etc. The source is https://github.com/Azure/aks-engine/tree/master/staging/provisioning/windows
#>
[CmdletBinding(DefaultParameterSetName="Standard")]
param(
    [string]
    [ValidateNotNullOrEmpty()]
    $MasterIP,

    [parameter()]
    [ValidateNotNullOrEmpty()]
    $KubeDnsServiceIp,

    [parameter(Mandatory=$true)]
    [ValidateNotNullOrEmpty()]
    $MasterFQDNPrefix,

    [parameter(Mandatory=$true)]
    [ValidateNotNullOrEmpty()]
    $Location,

    [parameter(Mandatory=$true)]
    [ValidateNotNullOrEmpty()]
    $AgentKey,

    [parameter(Mandatory=$true)]
    [ValidateNotNullOrEmpty()]
    $AADClientId,

    [parameter(Mandatory=$true)]
    [ValidateNotNullOrEmpty()]
    $AADClientSecret, # base64

    [parameter(Mandatory=$true)]
    [ValidateNotNullOrEmpty()]
    $NetworkAPIVersion,

    [parameter(Mandatory=$true)]
    [ValidateNotNullOrEmpty()]
    $TargetEnvironment,

    # C:\AzureData\provision.complete
    # MUST keep generating this file when CSE is done and do not change the name
    #  - It is used to avoid running CSE multiple times
    #  - Some customers use this file to check if CSE is done
    [parameter(Mandatory=$true)]
    [ValidateNotNullOrEmpty()]
    $CSEResultFilePath,

    [string]
    $UserAssignedClientID
)
# Do not parse the start time from $LogFile to simplify the logic
$StartTime=Get-Date
$global:ExitCode=0
$global:ErrorMessage=""

# These globals will not change between nodes in the same cluster, so they are not
# passed as powershell parameters

## SSH public keys to add to authorized_keys
$global:SSHKeys = @( {{ GetSshPublicKeysPowerShell }} )

## Certificates generated by aks-engine
$global:CACertificate = "{{GetParameter "caCertificate"}}"
$global:AgentCertificate = "{{GetParameter "clientCertificate"}}"

## Download sources provided by aks-engine
$global:KubeBinariesPackageSASURL = "{{GetParameter "kubeBinariesSASURL"}}"
$global:WindowsKubeBinariesURL = "{{GetParameter "windowsKubeBinariesURL"}}"
$global:KubeBinariesVersion = "{{GetParameter "kubeBinariesVersion"}}"
$global:ContainerdUrl = "{{GetParameter "windowsContainerdURL"}}"
$global:ContainerdSdnPluginUrl = "{{GetParameter "windowsSdnPluginURL"}}"

## Docker Version
$global:DockerVersion = "{{GetParameter "windowsDockerVersion"}}"

## ContainerD Usage
$global:DefaultContainerdWindowsSandboxIsolation = "{{GetParameter "defaultContainerdWindowsSandboxIsolation"}}"
$global:ContainerdWindowsRuntimeHandlers = "{{GetParameter "containerdWindowsRuntimeHandlers"}}"

## VM configuration passed by Azure
$global:WindowsTelemetryGUID = "{{GetParameter "windowsTelemetryGUID"}}"
{{if eq GetIdentitySystem "adfs"}}
$global:TenantId = "adfs"
{{else}}
$global:TenantId = "{{GetVariable "tenantID"}}"
{{end}}
$global:SubscriptionId = "{{GetVariable "subscriptionId"}}"
$global:ResourceGroup = "{{GetVariable "resourceGroup"}}"
$global:VmType = "{{GetVariable "vmType"}}"
$global:SubnetName = "{{GetVariable "subnetName"}}"
# NOTE: MasterSubnet is still referenced by `kubeletstart.ps1` and `windowsnodereset.ps1`
# for case of Kubenet
$global:MasterSubnet = ""
$global:SecurityGroupName = "{{GetVariable "nsgName"}}"
$global:VNetName = "{{GetVariable "virtualNetworkName"}}"
$global:RouteTableName = "{{GetVariable "routeTableName"}}"
$global:PrimaryAvailabilitySetName = "{{GetVariable "primaryAvailabilitySetName"}}"
$global:PrimaryScaleSetName = "{{GetVariable "primaryScaleSetName"}}"

$global:KubeClusterCIDR = "{{GetParameter "kubeClusterCidr"}}"
$global:KubeServiceCIDR = "{{GetParameter "kubeServiceCidr"}}"
$global:VNetCIDR = "{{GetParameter "vnetCidr"}}"
{{if IsKubernetesVersionGe "1.16.0"}}
$global:KubeletNodeLabels = "{{GetAgentKubernetesLabels . }}"
{{else}}
$global:KubeletNodeLabels = "{{GetAgentKubernetesLabelsDeprecated . }}"
{{end}}
$global:KubeletConfigArgs = @( {{GetKubeletConfigKeyValsPsh}} )
$global:KubeproxyConfigArgs = @( {{GetKubeproxyConfigKeyValsPsh}} )

$global:KubeproxyFeatureGates = @( {{GetKubeProxyFeatureGatesPsh}} )

$global:UseManagedIdentityExtension = "{{GetVariable "useManagedIdentityExtension"}}"
$global:UseInstanceMetadata = "{{GetVariable "useInstanceMetadata"}}"

$global:LoadBalancerSku = "{{GetVariable "loadBalancerSku"}}"
$global:ExcludeMasterFromStandardLB = "{{GetVariable "excludeMasterFromStandardLB"}}"

$global:PrivateEgressProxyAddress = "{{GetPrivateEgressProxyAddress}}"

# Windows defaults, not changed by aks-engine
$global:CacheDir = "c:\akse-cache"
$global:KubeDir = "c:\k"
$global:HNSModule = [Io.path]::Combine("$global:KubeDir", "hns.v2.psm1")

$global:KubeDnsSearchPath = "svc.cluster.local"

$global:CNIPath = [Io.path]::Combine("$global:KubeDir", "cni")
$global:NetworkMode = "L2Bridge"
$global:CNIConfig = [Io.path]::Combine($global:CNIPath, "config", "`$global:NetworkMode.conf")
$global:CNIConfigPath = [Io.path]::Combine("$global:CNIPath", "config")


$global:AzureCNIDir = [Io.path]::Combine("$global:KubeDir", "azurecni")
$global:AzureCNIBinDir = [Io.path]::Combine("$global:AzureCNIDir", "bin")
$global:AzureCNIConfDir = [Io.path]::Combine("$global:AzureCNIDir", "netconf")

# Azure cni configuration
# $global:NetworkPolicy = "{{GetParameter "networkPolicy"}}" # BUG: unused
$global:NetworkPlugin = "{{GetParameter "networkPlugin"}}"
$global:VNetCNIPluginsURL = "{{GetParameter "vnetCniWindowsPluginsURL"}}"
$global:IsDualStackEnabled = {{if IsIPv6DualStackFeatureEnabled}}$true{{else}}$false{{end}}
$global:IsAzureCNIOverlayEnabled = {{if IsAzureCNIOverlayFeatureEnabled}}$true{{else}}$false{{end}}

# Kubelet credential provider
$global:CredentialProviderURL = "{{GetParameter "windowsCredentialProviderURL"}}"

# CSI Proxy settings
$global:EnableCsiProxy = [System.Convert]::ToBoolean("{{GetVariable "windowsEnableCSIProxy" }}");
$global:CsiProxyUrl = "{{GetVariable "windowsCSIProxyURL" }}";

# Hosts Config Agent settings
$global:EnableHostsConfigAgent = [System.Convert]::ToBoolean("{{ EnableHostsConfigAgent }}");

# These scripts are used by cse
$global:CSEScriptsPackageUrl = "{{GetVariable "windowsCSEScriptsPackageURL" }}";

# The windows nvidia gpu driver related url is used by windows cse
$global:GpuDriverURL = "{{GetVariable "windowsGpuDriverURL" }}";

# PauseImage
$global:WindowsPauseImageURL = "{{GetVariable "windowsPauseImageURL" }}";
$global:AlwaysPullWindowsPauseImage = [System.Convert]::ToBoolean("{{GetVariable "alwaysPullWindowsPauseImage" }}");

# Calico
$global:WindowsCalicoPackageURL = "{{GetVariable "windowsCalicoPackageURL" }}";

## GPU install
$global:ConfigGPUDriverIfNeeded = [System.Convert]::ToBoolean("{{GetVariable "configGPUDriverIfNeeded" }}");

# GMSA
$global:WindowsGmsaPackageUrl = "{{GetVariable "windowsGmsaPackageUrl" }}";

# TLS Bootstrap Token
$global:TLSBootstrapToken = "{{GetTLSBootstrapTokenForKubeConfig}}"

# Disable OutBoundNAT in Azure CNI configuration
$global:IsDisableWindowsOutboundNat = [System.Convert]::ToBoolean("{{GetVariable "isDisableWindowsOutboundNat" }}");

# Base64 representation of ZIP archive
$zippedFiles = "{{ GetKubernetesWindowsAgentFunctions }}"

$global:KubeClusterConfigPath = "c:\k\kubeclusterconfig.json"
$fipsEnabled = [System.Convert]::ToBoolean("{{ FIPSEnabled }}")

# HNS remediator
$global:HNSRemediatorIntervalInMinutes = [System.Convert]::ToUInt32("{{GetHnsRemediatorIntervalInMinutes}}");

# Log generator
$global:LogGeneratorIntervalInMinutes = [System.Convert]::ToUInt32("{{GetLogGeneratorIntervalInMinutes}}");

$global:EnableIncreaseDynamicPortRange = $false

$global:RebootNeeded = $false

$global:IsSkipCleanupNetwork = [System.Convert]::ToBoolean("{{GetVariable "isSkipCleanupNetwork" }}");

$global:EnableKubeletServingCertificateRotation = [System.Convert]::ToBoolean("{{EnableKubeletServingCertificateRotation}}")

# Extract cse helper script from ZIP
[io.file]::WriteAllBytes("scripts.zip", [System.Convert]::FromBase64String($zippedFiles))
Expand-Archive scripts.zip -DestinationPath "C:\\AzureData\\" -Force

# Dot-source windowscsehelper.ps1 with functions that are called in this script
. c:\AzureData\windows\windowscsehelper.ps1
# util functions only can be used after this line, for example, Write-Log

$global:OperationId = New-Guid

try
{
    Logs-To-Event -TaskName "AKS.WindowsCSE.ExecuteCustomDataSetupScript" -TaskMessage ".\CustomDataSetupScript.ps1 -MasterIP $MasterIP -KubeDnsServiceIp $KubeDnsServiceIp -MasterFQDNPrefix $MasterFQDNPrefix -Location $Location -AADClientId $AADClientId -NetworkAPIVersion $NetworkAPIVersion -TargetEnvironment $TargetEnvironment -CSEResultFilePath $CSEResultFilePath"

    # Exit early if the script has been executed
    if (Test-Path -Path $CSEResultFilePath -PathType Leaf) {
        Write-Log "The script has been executed before, will exit without doing anything."
        return
    }

    # This involes using proxy, log the config before fetching packages
    Write-Log "private egress proxy address is '$global:PrivateEgressProxyAddress'"
    # TODO update to use proxy

    $WindowsCSEScriptsPackage = "aks-windows-cse-scripts-v0.0.50.zip"
    Write-Log "CSEScriptsPackageUrl is $global:CSEScriptsPackageUrl"
    Write-Log "WindowsCSEScriptsPackage is $WindowsCSEScriptsPackage"
    # Old AKS RP sets the full URL (https://acs-mirror.azureedge.net/aks/windows/cse/aks-windows-cse-scripts-v0.0.11.zip) in CSEScriptsPackageUrl
    # but it is better to set the CSE package version in Windows CSE in AgentBaker
    # since most changes in CSE package also need the change in Windows CSE in AgentBaker
    # In future, AKS RP only sets the endpoint with the pacakge name, for example, https://acs-mirror.azureedge.net/aks/windows/cse/
    if ($global:CSEScriptsPackageUrl.EndsWith("/")) {
        $global:CSEScriptsPackageUrl = $global:CSEScriptsPackageUrl + $WindowsCSEScriptsPackage
        Write-Log "CSEScriptsPackageUrl is set to $global:CSEScriptsPackageUrl"
    }

    # Download CSE function scripts
    Logs-To-Event -TaskName "AKS.WindowsCSE.DownloadAndExpandCSEScriptPackageUrl" -TaskMessage "Start to get CSE scripts. CSEScriptsPackageUrl: $global:CSEScriptsPackageUrl"
    $tempfile = 'c:\csescripts.zip'
    DownloadFileOverHttp -Url $global:CSEScriptsPackageUrl -DestinationPath $tempfile -ExitCode $global:WINDOWS_CSE_ERROR_DOWNLOAD_CSE_PACKAGE
    Expand-Archive $tempfile -DestinationPath "C:\\AzureData\\windows" -Force
    Remove-Item -Path $tempfile -Force
    
    # Dot-source cse scripts with functions that are called in this script
    . c:\AzureData\windows\azurecnifunc.ps1
    . c:\AzureData\windows\calicofunc.ps1
    . c:\AzureData\windows\configfunc.ps1
    . c:\AzureData\windows\containerdfunc.ps1
    . c:\AzureData\windows\kubeletfunc.ps1
    . c:\AzureData\windows\kubernetesfunc.ps1
    . c:\AzureData\windows\nvidiagpudriverfunc.ps1

    # Install OpenSSH if SSH enabled
    $sshEnabled = [System.Convert]::ToBoolean("{{ WindowsSSHEnabled }}")

    if ( $sshEnabled ) {
        Write-Log "Install OpenSSH"
        Install-OpenSSH -SSHKeys $SSHKeys
    }

    Set-TelemetrySetting -WindowsTelemetryGUID $global:WindowsTelemetryGUID

    Resize-OSDrive
    
    Initialize-DataDisks
    
    Initialize-DataDirectories
    
    Logs-To-Event -TaskName "AKS.WindowsCSE.GetProvisioningAndLogCollectionScripts" -TaskMessage "Start to get provisioning scripts and log collection scripts"
    Create-Directory -FullPath "c:\k"
    Write-Log "Remove `"NT AUTHORITY\Authenticated Users`" write permissions on files in c:\k"
    icacls.exe "c:\k" /inheritance:r
    icacls.exe "c:\k" /grant:r SYSTEM:`(OI`)`(CI`)`(F`)
    icacls.exe "c:\k" /grant:r BUILTIN\Administrators:`(OI`)`(CI`)`(F`)
    icacls.exe "c:\k" /grant:r BUILTIN\Users:`(OI`)`(CI`)`(RX`)
    Write-Log "c:\k permissions: "
    icacls.exe "c:\k"
    Get-ProvisioningScripts
    Get-LogCollectionScripts

    # NOTE: this function MUST be called before Write-KubeClusterConfig since it has the potential
    # to mutate both kubelet config args and kubelet node labels.
    Configure-KubeletServingCertificateRotation
    
    Write-KubeClusterConfig -MasterIP $MasterIP -KubeDnsServiceIp $KubeDnsServiceIp

    Install-CredentialProvider -KubeDir $global:KubeDir -CustomCloudContainerRegistryDNSSuffix {{if IsAKSCustomCloud}}"{{ AKSCustomCloudContainerRegistryDNSSuffix }}"{{else}}""{{end}} 

    Get-KubePackage -KubeBinariesSASURL $global:KubeBinariesPackageSASURL
    
    $cniBinPath = $global:AzureCNIBinDir
    $cniConfigPath = $global:AzureCNIConfDir
    if ($global:NetworkPlugin -eq "kubenet") {
        $cniBinPath = $global:CNIPath
        $cniConfigPath = $global:CNIConfigPath
    }

    Install-Containerd-Based-On-Kubernetes-Version -ContainerdUrl $global:ContainerdUrl -CNIBinDir $cniBinPath -CNIConfDir $cniConfigPath -KubeDir $global:KubeDir -KubernetesVersion $global:KubeBinariesVersion
    
    Retag-ImagesForAzureChinaCloud -TargetEnvironment $TargetEnvironment
    
    # For AKSClustomCloud, TargetEnvironment must be set to AzureStackCloud
    Write-AzureConfig `
        -KubeDir $global:KubeDir `
        -AADClientId $AADClientId `
        -AADClientSecret $([System.Text.Encoding]::ASCII.GetString([System.Convert]::FromBase64String($AADClientSecret))) `
        -TenantId $global:TenantId `
        -SubscriptionId $global:SubscriptionId `
        -ResourceGroup $global:ResourceGroup `
        -Location $Location `
        -VmType $global:VmType `
        -SubnetName $global:SubnetName `
        -SecurityGroupName $global:SecurityGroupName `
        -VNetName $global:VNetName `
        -RouteTableName $global:RouteTableName `
        -PrimaryAvailabilitySetName $global:PrimaryAvailabilitySetName `
        -PrimaryScaleSetName $global:PrimaryScaleSetName `
        -UseManagedIdentityExtension $global:UseManagedIdentityExtension `
        -UserAssignedClientID $UserAssignedClientID `
        -UseInstanceMetadata $global:UseInstanceMetadata `
        -LoadBalancerSku $global:LoadBalancerSku `
        -ExcludeMasterFromStandardLB $global:ExcludeMasterFromStandardLB `
        -TargetEnvironment {{if IsAKSCustomCloud}}"AzureStackCloud"{{else}}$TargetEnvironment{{end}} 

    # we borrow the logic of AzureStackCloud to achieve AKSCustomCloud. 
    # In case of AKSCustomCloud, customer cloud env will be loaded from azurestackcloud.json 
    {{if IsAKSCustomCloud}}
    $azureStackConfigFile = [io.path]::Combine($global:KubeDir, "azurestackcloud.json")
    $envJSON = "{{ GetBase64EncodedEnvironmentJSON }}"
    [io.file]::WriteAllBytes($azureStackConfigFile, [System.Convert]::FromBase64String($envJSON))

    Get-CACertificates
    {{end}}

    Write-CACert -CACertificate $global:CACertificate `
        -KubeDir $global:KubeDir
    
    if ($global:EnableCsiProxy) {
        New-CsiProxyService -CsiProxyPackageUrl $global:CsiProxyUrl -KubeDir $global:KubeDir
    }

    if ($global:TLSBootstrapToken) {
        Write-BootstrapKubeConfig -CACertificate $global:CACertificate `
            -KubeDir $global:KubeDir `
            -MasterFQDNPrefix $MasterFQDNPrefix `
            -MasterIP $MasterIP `
            -TLSBootstrapToken $global:TLSBootstrapToken
        
        # NOTE: we need kubeconfig to setup calico even if TLS bootstrapping is enabled
        #       This kubeconfig will deleted after calico installation.
        # TODO(hbc): once TLS bootstrap is fully enabled, remove this if block
        Write-Log "Write temporary kube config"
    } else {
        Write-Log "Write kube config"
    }

    Write-KubeConfig -CACertificate $global:CACertificate `
        -KubeDir $global:KubeDir `
        -MasterFQDNPrefix $MasterFQDNPrefix `
        -MasterIP $MasterIP `
        -AgentKey $AgentKey `
        -AgentCertificate $global:AgentCertificate
    
    if ($global:EnableHostsConfigAgent) {
        New-HostsConfigService
    }

    Write-Log "Configuring networking with NetworkPlugin:$global:NetworkPlugin"

    # Configure network policy.
    Get-HnsPsm1 -HNSModule $global:HNSModule
    Import-Module $global:HNSModule
    
    Install-VnetPlugins -AzureCNIConfDir $global:AzureCNIConfDir `
        -AzureCNIBinDir $global:AzureCNIBinDir `
        -VNetCNIPluginsURL $global:VNetCNIPluginsURL
    
    Set-AzureCNIConfig -AzureCNIConfDir $global:AzureCNIConfDir `
        -KubeDnsSearchPath $global:KubeDnsSearchPath `
        -KubeClusterCIDR $global:KubeClusterCIDR `
        -KubeServiceCIDR $global:KubeServiceCIDR `
        -VNetCIDR $global:VNetCIDR `
        -IsDualStackEnabled $global:IsDualStackEnabled `
        -IsAzureCNIOverlayEnabled $global:IsAzureCNIOverlayEnabled
    
    if ($TargetEnvironment -ieq "AzureStackCloud") {
        GenerateAzureStackCNIConfig `
            -TenantId $global:TenantId `
            -SubscriptionId $global:SubscriptionId `
            -ResourceGroup $global:ResourceGroup `
            -AADClientId $AADClientId `
            -KubeDir $global:KubeDir `
            -AADClientSecret $([System.Text.Encoding]::ASCII.GetString([System.Convert]::FromBase64String($AADClientSecret))) `
            -NetworkAPIVersion $NetworkAPIVersion `
            -AzureEnvironmentFilePath $([io.path]::Combine($global:KubeDir, "azurestackcloud.json")) `
            -IdentitySystem "{{ GetIdentitySystem }}"
    }

    New-ExternalHnsNetwork -IsDualStackEnabled $global:IsDualStackEnabled
    
    Install-KubernetesServices `
        -KubeDir $global:KubeDir

    Set-Explorer
    Adjust-PageFileSize
    Logs-To-Event -TaskName "AKS.WindowsCSE.PreprovisionExtension" -TaskMessage "Start preProvisioning script"
    PREPROVISION_EXTENSION
    Update-ServiceFailureActions
    Adjust-DynamicPortRange
    Register-LogsCleanupScriptTask
    Register-NodeResetScriptTask
    Update-DefenderPreferences

    $windowsVersion = Get-WindowsVersion
    if ($windowsVersion -ne "1809") {
        Logs-To-Event -TaskName "AKS.WindowsCSE.EnableSecureTLS" -TaskMessage "Skip secure TLS protocols for Windows version: $windowsVersion"
    } else {
        Logs-To-Event -TaskName "AKS.WindowsCSE.EnableSecureTLS" -TaskMessage "Start to enable secure TLS protocols"
        try {
            . C:\k\windowssecuretls.ps1
            Enable-SecureTls
        }
        catch {
            Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_ENABLE_SECURE_TLS -ErrorMessage $_
        }
    }

    Enable-FIPSMode -FipsEnabled $fipsEnabled
    if ($global:WindowsGmsaPackageUrl) {
        Install-GmsaPlugin -GmsaPackageUrl $global:WindowsGmsaPackageUrl
    }

    Check-APIServerConnectivity -MasterIP $MasterIP

    if ($global:WindowsCalicoPackageURL) {
        Start-InstallCalico -RootDir "c:\" -KubeServiceCIDR $global:KubeServiceCIDR -KubeDnsServiceIp $KubeDnsServiceIp
    }

    Start-InstallGPUDriver -EnableInstall $global:ConfigGPUDriverIfNeeded -GpuDriverURL $global:GpuDriverURL
    
    if (Test-Path $CacheDir)
    {
        Write-Log "Removing aks cache directory"
        Remove-Item $CacheDir -Recurse -Force
    }

    if ($global:TLSBootstrapToken) {
        Write-Log "Removing temporary kube config"
        $kubeConfigFile = [io.path]::Combine($KubeDir, "config")
        Remove-Item $kubeConfigFile
    }

    Enable-GuestVMLogs -IntervalInMinutes $global:LogGeneratorIntervalInMinutes

    if ($global:RebootNeeded) {
        Logs-To-Event -TaskName "AKS.WindowsCSE.RestartComputer" -TaskMessage "Setup Complete, calling Postpone-RestartComputer with reboot"
        Postpone-RestartComputer
    } else {
        Logs-To-Event -TaskName "AKS.WindowsCSE.StartScheduledTask" -TaskMessage "Setup Complete, start NodeResetScriptTask to register Windows node without reboot"
        Start-ScheduledTask -TaskName "k8s-restart-job"

        $timeout = 180 ##  seconds
        $timer = [Diagnostics.Stopwatch]::StartNew()
        while ((Get-ScheduledTask -TaskName 'k8s-restart-job').State -ne 'Ready') {
            # The task `k8s-restart-job` needs ~8 seconds.
            if ($timer.Elapsed.TotalSeconds -gt $timeout) {
                Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_START_NODE_RESET_SCRIPT_TASK -ErrorMessage "NodeResetScriptTask is not finished after [$($timer.Elapsed.TotalSeconds)] seconds"
            }

            Write-Log -Message "Waiting on NodeResetScriptTask..."
            Start-Sleep -Seconds 3
        }
        $timer.Stop()
        Write-Log -Message "We waited [$($timer.Elapsed.TotalSeconds)] seconds on NodeResetScriptTask"
    }
}
catch
{
    # Set-ExitCode will exit with the specified ExitCode immediately and not be caught by this catch block
    # Ideally all exceptions will be handled and no exception will be thrown.
    Set-ExitCode -ExitCode $global:WINDOWS_CSE_ERROR_UNKNOWN -ErrorMessage $_
}
finally
{
    # Generate CSE result so it can be returned as the CSE response in csecmd.ps1
    $ExecutionDuration=$(New-Timespan -Start $StartTime -End $(Get-Date))
    Write-Log "CSE ExecutionDuration: $ExecutionDuration. ExitCode: $global:ExitCode"

    Logs-To-Event -TaskName "AKS.WindowsCSE.cse_main" -TaskMessage "ExitCode: $global:ExitCode. ErrorMessage: $global:ErrorMessage." 

    Write-Log "Start to enable full support of CimFS"
    c:\AzureData\windows\StagingTool.exe /enable 44354424 47273241 47273202 48004466 48230524
    Postpone-RestartComputer

    # Please not use Write-Log or Logs-To-Events after Stop-Transcript
    Stop-Transcript

    # $CSEResultFilePath is used to avoid running CSE multiple times
    if ($global:ExitCode -ne 0) {
        # $JsonString = "ExitCode: |{0}|, Output: |{1}|, Error: |{2}|"
        # Max length of the full error message returned by Windows CSE is ~256. We use 240 to be safe.
        $errorMessageLength = "ExitCode: |$global:ExitCode|, Output: |$($global:ErrorCodeNames[$global:ExitCode])|, Error: ||".Length
        $turncatedErrorMessage = $global:ErrorMessage.Substring(0, [Math]::Min(240 - $errorMessageLength, $global:ErrorMessage.Length))
        Set-Content -Path $CSEResultFilePath -Value "ExitCode: |$global:ExitCode|, Output: |$($global:ErrorCodeNames[$global:ExitCode])|, Error: |$turncatedErrorMessage|"
    }
    else {
        Set-Content -Path $CSEResultFilePath -Value $global:ExitCode -Force
    }

    if ($global:ExitCode -eq $global:WINDOWS_CSE_ERROR_DOWNLOAD_CSE_PACKAGE) {
        Write-Log "Do not call Upload-GuestVMLogs because there is no cse script package downloaded"
    }
    else {
        Upload-GuestVMLogs -ExitCode $global:ExitCode
    }
}
