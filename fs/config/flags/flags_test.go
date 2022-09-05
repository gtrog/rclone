package flags

import (
	"context"
	"fmt"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fstest/logtest"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const Value = "MY_VALUE"

func TestSetDefaultFromEnvNotLoggingSensitiveValues(t *testing.T) {
	for _, testCase := range []struct {
		redacted bool
		option   fs.Option
	}{
		{redacted: true, option: fs.Option{Name: "access_key_id", IsPassword: true}},
		{redacted: true, option: fs.Option{Name: "region"}},
		{redacted: false, option: fs.Option{Name: "access_key_id", IsPassword: true}},
		{redacted: false, option: fs.Option{Name: "region"}},
	} {
		redacted := testCase.redacted
		option := testCase.option

		var flag string
		flags := &pflag.FlagSet{}
		flags.StringVarP(&flag, "access_key_id", "a", "", "")
		flags.StringVarP(&flag, "region", "r", "", "")
		flags.StringVarP(&flag, "foobar", "f", "", "")
		_ = flags.Parse([]string{"--access_key_id", "--region", "--foobar"})

		t.Run(fmt.Sprintf("redacted=%v,option=%s", redacted, option.Name), func(t *testing.T) {
			ctx := context.Background()
			ctx, config := fs.AddConfig(ctx)

			config.Redacted = redacted
			fs.GetConfig(context.TODO()).LogLevel = fs.LogLevelDebug

			envName := fs.OptionToEnv(option.Name)
			err := os.Setenv(envName, Value)
			assert.NoError(t, err)

			logMessage := logtest.CaptureLogging(func() {
				SetDefaultFromEnv(ctx, flags, option.Name, &option)
			})

			if option.IsPassword && redacted {
				assert.NotContains(t, logMessage, Value)
			} else {
				assert.Contains(t, logMessage, Value)
			}
		})
	}
}
