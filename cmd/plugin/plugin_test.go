package plugincmd

import (
	"strings"
	"testing"

	pluginioc "github.com/Duke1616/ecmdb/cmd/plugin/ioc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestResolveTenantConfigFromConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	readTenantConfig(t)

	cfg, err := pluginioc.ProvideBuiltinTenantConfig()
	require.NoError(t, err)
	require.Equal(t, "config-access", cfg.AccessKey)
	require.Equal(t, "config-secret", cfg.SecretKey)
}

func TestResolveTenantConfigPrefersFlags(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	readTenantConfig(t)

	cmd := newTenantConfigCommand(t)
	require.NoError(t, cmd.Flags().Set("access-key", "flag-access"))
	require.NoError(t, cmd.Flags().Set("secret-key", "flag-secret"))

	cfg, err := pluginioc.ProvideBuiltinTenantConfig()
	require.NoError(t, err)
	require.Equal(t, "flag-access", cfg.AccessKey)
	require.Equal(t, "flag-secret", cfg.SecretKey)
}

func newTenantConfigCommand(t *testing.T) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "import-builtin"}
	cmd.Flags().String("access-key", "", "")
	cmd.Flags().String("secret-key", "", "")

	_ = viper.BindPFlag("plugin.builtin.tenant.access_key", cmd.Flags().Lookup("access-key"))
	_ = viper.BindPFlag("plugin.builtin.tenant.secret_key", cmd.Flags().Lookup("secret-key"))

	return cmd
}

func readTenantConfig(t *testing.T) {
	t.Helper()

	viper.SetConfigType("yaml")
	require.NoError(t, viper.ReadConfig(strings.NewReader(`
plugin:
  builtin:
    tenant:
      access_key: config-access
      secret_key: config-secret
`)))
}
