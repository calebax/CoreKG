package version

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/logs"
)

var (
	version    string
	gitCommit  string
	builtAt    string
	deployMode string
)

// GetVersion returns the version of the application.
func GetVersion() string {
	return version
}

// GitCommit returns the git commit of the application.
func GitCommit() string {
	return gitCommit
}

// BuiltAt returns the date and time when the application was built.
func BuiltAt() string {
	return builtAt
}

// Completed .
func Completed() string {
	return fmt.Sprintf("%s-%s", version, gitCommit)
}

func DeployMode() string {
	return deployMode
}

func SetDeployMode(mode string) {
	deployMode = mode
}

// Short returns the short version of the application.
func Short(ctx context.Context) string {
	if version == "" {
		return ""
	}
	sv, err := semver.NewVersion(version)
	if err != nil {
		logs.WarnContextf(ctx, "parse version error: %s", err)
		return version
	}
	return fmt.Sprintf("v%d.%d.%d", sv.Major(), sv.Minor(), sv.Patch())
}

// VersionCmd .
func VersionCmd(name string) *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  `All software has versions`,
		Run: func(cmd *cobra.Command, args []string) {
			if short {
				fmt.Println(Short(cmd.Context()))
				return
			}
			fmt.Printf("%s(%s) version: %s\n", name, deployMode, GetVersion())
			fmt.Printf("git commit: %s\n", GitCommit())
			fmt.Printf("built at: %s\n", BuiltAt())
		},
	}
	cmd.Flags().BoolVarP(&short, "short", "s", false, "Print the short version number")
	return cmd
}
