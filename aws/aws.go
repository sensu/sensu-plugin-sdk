package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

type AWSPluginConfig struct {
	//Common AWS elements
	AWSRegion           string
	AWSProfile          string
	AWSCredentialsFiles []string
	AWSConfigFiles      []string
	AWSAccessKeyID      string
	AWSSecretAccessKey  string
	AWSConfig           *aws.Config
	AWSCredentials      *aws.Credentials
}

func (plugin *AWSPluginConfig) GetAWSOpts() []*sensu.PluginConfigOption {
	awsOpts := []*sensu.PluginConfigOption{
		&sensu.PluginConfigOption{
			Path:      "region",
			Env:       "AWS_REGION",
			Argument:  "region",
			Shorthand: "",
			Default:   "",
			Usage:     "AWS Region to use, (or set envvar AWS_REGION)",
			Value:     &plugin.AWSRegion,
		},
		&sensu.PluginConfigOption{
			Path:      "profile",
			Env:       "AWS_PROFILE",
			Argument:  "profile",
			Shorthand: "",
			Default:   "",
			Usage:     "AWS Credential Profile (for security use envvar AWS_PROFILE)",
			Secret:    false,
			Value:     &plugin.AWSProfile,
		},
		&sensu.PluginConfigOption{
			Path:      "config-files",
			Env:       "",
			Argument:  "config-files",
			Shorthand: "",
			Default:   []string{},
			Usage:     "comma separated list of AWS config files",
			Secret:    false,
			Value:     &plugin.AWSConfigFiles,
		},
		&sensu.PluginConfigOption{
			Path:      "credentials-files",
			Env:       "",
			Argument:  "credentials-files",
			Shorthand: "",
			Default:   []string{},
			Usage:     "comma separated list of AWS Credential files",
			Secret:    false,
			Value:     &plugin.AWSCredentialsFiles,
		},
	}
	return awsOpts
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (plugin *AWSPluginConfig) CheckAWSCreds() (int, error) {
	var err error
	// Common arg checking that should be done for all AWS plugins
	for _, f := range plugin.AWSCredentialsFiles {
		if !fileExists(f) {
			return sensu.CheckStateCritical, fmt.Errorf("Credential file missing: %s", f)
		}
	}
	for _, f := range plugin.AWSConfigFiles {
		if !fileExists(f) {
			return sensu.CheckStateCritical, fmt.Errorf("Config file missing: %s", f)
		}
	}

	// Note: slight workaround here as sdk wont let me pass an array of arguments
	// due to a type mismatch
	// workaround for now is to pass the same function pointer multiple times in some cases
	regionArg := config.WithRegion(plugin.AWSRegion)
	configArg := regionArg
	if len(plugin.AWSConfigFiles) > 0 {
		configArg = config.WithSharedConfigFiles(plugin.AWSConfigFiles)
	}
	credsArg := regionArg
	if len(plugin.AWSCredentialsFiles) > 0 {
		credsArg = config.WithSharedCredentialsFiles(plugin.AWSCredentialsFiles)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		regionArg,
		configArg,
		credsArg,
	)
	if err != nil {
		return sensu.CheckStateCritical, err
	}
	plugin.AWSConfig = &cfg
	creds, err := plugin.AWSConfig.Credentials.Retrieve(context.Background())
	if err != nil {
		return sensu.CheckStateCritical, err
	}
	plugin.AWSCredentials = &creds
	return sensu.CheckStateOK, nil
}
