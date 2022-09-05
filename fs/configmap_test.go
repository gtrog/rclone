package fs

import (
	"context"
	"fmt"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fstest/logtest"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func testSensitiveValues(
	t *testing.T,
	getterFactory func(fsInfo *RegInfo) configmap.Getter,
	envNameMapper func(fsInfo *RegInfo, opt *Option) string,
) {
	for _, testCase := range []struct {
		redacted bool
		option   Option
	}{
		{redacted: true, option: Option{Name: "access_key_id", IsPassword: true}},
		{redacted: true, option: Option{Name: "region"}},
		{redacted: false, option: Option{Name: "access_key_id", IsPassword: true}},
		{redacted: false, option: Option{Name: "region"}},
	} {
		redacted := testCase.redacted
		option := testCase.option

		t.Run(fmt.Sprintf("redacted=%v,option=%s", redacted, option.Name), func(t *testing.T) {
			globalConfig = &ConfigInfo{Redacted: redacted}
			GetConfig(context.TODO()).LogLevel = LogLevelDebug

			fsInfo := &RegInfo{
				Prefix: "s3",
				Options: []Option{
					{Name: "other"},
					option,
				},
			}

			getter := getterFactory(fsInfo)

			envName := envNameMapper(fsInfo, &option)
			err := os.Setenv(envName, Value)
			assert.NoError(t, err)

			logMessage := logtest.CaptureLogging(func() {
				getter.Get(option.Name)
			})

			if option.IsPassword && redacted {
				assert.NotContains(t, logMessage, Value)
			} else {
				assert.Contains(t, logMessage, Value)
			}
		})
	}
}

const Value = "MY_VALUE"
const Section = "s3"

func TestConfigEnvVarsNotLoggingSensitiveValues(t *testing.T) {
	testSensitiveValues(
		t,
		func(fsInfo *RegInfo) configmap.Getter {
			return configEnvVars{Section, fsInfo}
		},
		func(_ *RegInfo, opt *Option) string {
			return ConfigToEnv(Section, opt.Name)
		},
	)
}

func TestOptionEnvVarsNotLoggingSensitiveValues(t *testing.T) {
	testSensitiveValues(
		t,
		func(fsInfo *RegInfo) configmap.Getter {
			return optionEnvVars{fsInfo}
		},
		func(fsInfo *RegInfo, opt *Option) string {
			return OptionToEnv(fsInfo.Prefix + "-" + opt.Name)
		},
	)
}
