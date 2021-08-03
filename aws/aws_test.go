package aws

import (
	"testing"

	"github.com/sensu/sensu-plugin-sdk/sensu"
	"github.com/stretchr/testify/assert"
)

type Config struct {
	//base Sensu plugin configs
	sensu.PluginConfig
	//common AWS Config configs
	AWSPluginConfig
	//Specific Check Configs
	Tags    []string
	Verbose bool
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "aws-service-test",
			Short:    "AWS Service Test",
			Keyspace: "sensu.io/plugins/aws-service-test/test/config",
		},
	}
)

func TestGetAWSAopts(t *testing.T) {

	options := plugin.GetAWSOpts()
	assert.NotNil(t, options)

}

func TestCheckAWSCredsNoCredsDefined(t *testing.T) {
	status, err := plugin.CheckAWSCreds()
	assert.Error(t, err)
	assert.Equal(t, 2, status)
}

func TestCheckAWSCredsWithCredsFile(t *testing.T) {
	plugin.AWSCredentialsFiles = []string{
		"./testing/missing-credentials",
	}
	status, err := plugin.CheckAWSCreds()
	assert.Error(t, err)
	assert.Equal(t, 2, status)
	plugin.AWSCredentialsFiles = []string{
		"./testing/credentials",
	}
	status, err = plugin.CheckAWSCreds()
	assert.NoError(t, err)
	assert.Equal(t, 0, status)
}

func TestCheckAWSCredsWithConfigFile(t *testing.T) {
	plugin.AWSConfigFiles = []string{
		"./testing/missing-config",
	}
	status, err := plugin.CheckAWSCreds()
	assert.Error(t, err)
	assert.Equal(t, 2, status)
	plugin.AWSConfigFiles = []string{
		"./testing/config",
	}
	status, err = plugin.CheckAWSCreds()
	assert.NoError(t, err)
	assert.Equal(t, 0, status)
}
