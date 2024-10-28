package config

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/go-armbalancer"
)

// modeled after e2e/config/azure.go

type AzureClient struct {
	Core                    *azcore.Client
	ResourceGroup           *armresources.ResourceGroupsClient
	VirtualMachinesClient   *armcompute.VirtualMachinesClient
	NetworkInterfacesClient *armnetwork.InterfacesClient
	VNetClient              *armnetwork.VirtualNetworksClient
	SubNetClient            *armnetwork.SubnetsClient
	ResourceGroupsClient    *armresources.ResourceGroupsClient
}

func mustNewAzureClient(subscription string) *AzureClient {
	client, err := NewAzureClient(subscription)
	if err != nil {
		panic(err)
	}
	return client
}

func NewAzureClient(subscription string) (*AzureClient, error) {
	httpClient := &http.Client{
		// use a bunch of connections for load balancing
		// ensure all timeouts are defined and reasonable
		// ensure TLS1.2+ and HTTP2
		Transport: armbalancer.New(armbalancer.Options{
			PoolSize: 100,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		}),
	}

	logger := runtime.NewLogPolicy(&policy.LogOptions{
		IncludeBody: true,
	})

	opts := &arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Transport: httpClient,
			PerCallPolicies: []policy.Policy{
				logger,
			},
			Retry: DefaultRetryOpts(),
		},
	}

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	plOpts := runtime.PipelineOptions{}
	clOpts := &azcore.ClientOptions{
		Transport: httpClient,
		PerCallPolicies: []policy.Policy{
			runtime.NewBearerTokenPolicy(credential, []string{"https://management.azure.com/.default"}, nil),
			logger,
		},
		Retry: DefaultRetryOpts(),
	}

	cloud := &AzureClient{}
	cloud.Core, err = azcore.NewClient("agentbakere2e.e2e_test", "v0.0.0", plOpts, clOpts)
	if err != nil {
		return nil, fmt.Errorf("create core client: %w", err)
	}

	// purely for telemetry, entirely unused today
	cloud.Core, err = azcore.NewClient("agentbaker_production_vms.production_vms_test", "v0.0.0", plOpts, clOpts)
	if err != nil {
		return nil, fmt.Errorf("create core client: %w", err)
	}

	cloud.ResourceGroup, err = armresources.NewResourceGroupsClient(subscription, credential, opts)
	if err != nil {
		return nil, fmt.Errorf("create resource group client: %w", err)
	}

	cloud.VirtualMachinesClient, err = armcompute.NewVirtualMachinesClient(subscription, credential, opts)
	if err != nil {
		return nil, fmt.Errorf("create virtual machines client: %w", err)
	}

	cloud.NetworkInterfacesClient, err = armnetwork.NewInterfacesClient(subscription, credential, opts)
	if err != nil {
		return nil, fmt.Errorf("create network interfaces client: %w", err)
	}

	cloud.VNetClient, err = armnetwork.NewVirtualNetworksClient(subscription, credential, opts)
	if err != nil {
		return nil, fmt.Errorf("create virtual network client: %w", err)
	}

	cloud.SubNetClient, err = armnetwork.NewSubnetsClient(subscription, credential, opts)
	if err != nil {
		return nil, fmt.Errorf("create subnet client: %w", err)
	}

	cloud.ResourceGroupsClient, err = armresources.NewResourceGroupsClient(subscription, credential, opts)
	if err != nil {
		return nil, fmt.Errorf("create resource groups client: %w", err)
	}

	return cloud, nil
}

func DefaultRetryOpts() policy.RetryOptions {
	return policy.RetryOptions{
		MaxRetries: 3,
		RetryDelay: time.Second * 5,
		StatusCodes: []int{
			http.StatusRequestTimeout,      // 408
			http.StatusTooManyRequests,     // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout,      // 504
			http.StatusNotFound,            // 404
		},
	}
}
