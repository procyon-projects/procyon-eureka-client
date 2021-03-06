package eureka

import (
	core "github.com/procyon-projects/procyon-core"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
)

type InstanceInfoProvider interface {
	GetInstanceInfo() *InstanceInfo
}

type DefaultInstanceInfoProvider struct {
	instanceProperties InstanceProperties
	instanceInfo       *InstanceInfo
	instanceInfoMu     sync.RWMutex
	environment        core.Environment
}

func newDefaultInstanceInfoProvider(instanceProperties InstanceProperties, environment core.Environment) *DefaultInstanceInfoProvider {
	return &DefaultInstanceInfoProvider{
		instanceProperties: instanceProperties,
		environment:        environment,
	}
}

func (provider *DefaultInstanceInfoProvider) GetInstanceInfo() *InstanceInfo {
	provider.instanceInfoMu.Lock()
	defer provider.instanceInfoMu.Unlock()
	if provider.instanceInfo != nil {
		return provider.instanceInfo
	}

	port := provider.environment.GetProperty("server.port", "8080").(string)
	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	instanceInfo := &InstanceInfo{
		InstanceId:     provider.instanceProperties.InstanceId,
		AppName:        strings.ToUpper(provider.instanceProperties.ApplicationName),
		AppGroupName:   provider.instanceProperties.ApplicationGroupName,
		IpAddr:         provider.instanceProperties.IpAddr,
		HomePageUrl:    provider.getUrl(false, hostName, port, provider.instanceProperties.HomePageUrl),
		StatusPageUrl:  provider.getUrl(false, hostName, port, provider.instanceProperties.StatusPageUrl),
		HealthCheckUrl: provider.getUrl(false, hostName, port, provider.instanceProperties.HealthCheckUrl),
		DataCenterInfo: &DataCenterInfo{
			Name:  provider.instanceProperties.DataCenterInfo.Name,
			Class: provider.instanceProperties.DataCenterInfo.Class,
		},
		HostName:         provider.instanceProperties.Hostname,
		CountryId:        1,
		OverriddenStatus: "UNKNOWN",
		LeaseInfo: &LeaseInfo{
			RenewalIntervalInSecs: 30,
			DurationInSecs:        90,
		},
		LastDirtyTimestamp: "0",
	}

	instanceInfo.VipAddress = provider.instanceProperties.ApplicationName
	instanceInfo.Port = &PortWrapper{
		Enabled: strconv.FormatBool(provider.instanceProperties.NonSecurePortEnabled),
		Port:    provider.instanceProperties.NonSecurePort,
	}

	instanceInfo.SecureVipAddress = provider.instanceProperties.ApplicationName
	instanceInfo.SecurePort = &PortWrapper{
		Enabled: strconv.FormatBool(provider.instanceProperties.SecurePortEnabled),
		Port:    provider.instanceProperties.SecurePort,
	}
	instanceInfo.SecureHealthCheckUrl = provider.getUrl(true, hostName, port, provider.instanceProperties.HealthCheckUrl)
	instanceInfo.Status = InstanceStatusUp

	provider.instanceInfo = instanceInfo
	return instanceInfo
}

func (provider *DefaultInstanceInfoProvider) getUrl(isSecure bool, hostName string, port string, urlPath string) string {
	scheme := "http"
	if isSecure {
		scheme = "https"
	}

	stringPort := ""
	if port != "80" && port != "443" {
		stringPort = ":" + port
	}
	host := hostName + stringPort

	result := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   urlPath,
	}
	return result.String()
}
