package recipes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/newrelic/newrelic-cli/internal/install/execution"
	"github.com/newrelic/newrelic-cli/internal/install/types"
)

func TestRecommend_CustomScript_Success(t *testing.T) {
	recipe := types.OpenInstallationRecipe{
		Name: "test-recipe",
		PreInstall: types.OpenInstallationPreInstallConfiguration{
			RequireAtDiscovery: "echo 1234",
		},
	}

	m := &types.DiscoveryManifest{}

	r := NewRecipeFilterRunner(types.InstallerContext{}, &execution.InstallStatus{})

	result := r.RunFilter(context.Background(), &recipe, m)
	require.NoError(t, result)
}

func TestRecommend_CustomScript_Failure(t *testing.T) {
	recipe := types.OpenInstallationRecipe{
		Name: "test-recipe",
		PreInstall: types.OpenInstallationPreInstallConfiguration{
			RequireAtDiscovery: "bogus command",
		},
	}

	m := &types.DiscoveryManifest{}

	r := NewRecipeFilterRunner(types.InstallerContext{}, &execution.InstallStatus{})

	result := r.RunFilter(context.Background(), &recipe, m)
	require.Error(t, result)
}

func TestShouldGetRecipeFirstNameValid(t *testing.T) {
	recipe := types.OpenInstallationRecipe{
		Name:        "test-recipe",
		DisplayName: "MongoDB installation something",
	}

	name := getRecipeFirstName(recipe)
	require.True(t, name == "MongoDB")
}

func TestShouldGetRecipeFirstNameValidWhole(t *testing.T) {
	recipe := types.OpenInstallationRecipe{
		Name:        "test-recipe",
		DisplayName: "MongoDB-single-word",
	}

	name := getRecipeFirstName(recipe)
	require.True(t, name == "MongoDB-single-word")
}

func TestShouldGetRecipeFirstNameInvalid(t *testing.T) {
	recipe := types.OpenInstallationRecipe{
		Name: "test-recipe",
	}

	name := getRecipeFirstName(recipe)
	require.True(t, name == "")
}

func TestRecipeFilterRunner_ShouldMatchRecipe(t *testing.T) {
	recipe := types.OpenInstallationRecipe{
		Name:         "test-recipe",
		ProcessMatch: []string{"php-fpm"},
		PreInstall: types.OpenInstallationPreInstallConfiguration{
			RequireAtDiscovery: "exit 0",
		},
	}

	matchedProcess := mockProcess{
		cmdline: "php-fpm",
		name:    `php-fpm`,
		pid:     int32(1234),
	}

	m := types.DiscoveryManifest{
		DiscoveredProcesses: []types.GenericProcess{matchedProcess},
	}

	installStatus := &execution.InstallStatus{
		DiscoveryManifest: m,
	}

	r := NewRecipeFilterRunner(types.InstallerContext{}, installStatus)

	result := r.RunFilter(context.Background(), &recipe, &m)
	require.NoError(t, result) // ensure recipe is not filtered out
}

func TestRecipeFilterRunner_ShouldFilterOutRecipeWithPreInstallError(t *testing.T) {
	recipe := types.OpenInstallationRecipe{
		Name:         "test-recipe",
		ProcessMatch: []string{"apache2"},
		PreInstall: types.OpenInstallationPreInstallConfiguration{
			RequireAtDiscovery: "exit 132", // simulate failed preinstall check (should be filtered out)
		},
	}

	matchedProcess := mockProcess{
		cmdline: "apache2",
		name:    `apache2`,
		pid:     int32(1234),
	}

	m := types.DiscoveryManifest{
		DiscoveredProcesses: []types.GenericProcess{matchedProcess},
	}

	installStatus := &execution.InstallStatus{
		DiscoveryManifest: m,
	}

	r := NewRecipeFilterRunner(types.InstallerContext{}, installStatus)

	result := r.RunFilter(context.Background(), &recipe, &m)
	require.Error(t, result)
}

func TestRecipeFilterRunner_ShouldFailPreInstallWithDetected(t *testing.T) {
	recipe := types.OpenInstallationRecipe{
		Name:         "test-recipe",
		ProcessMatch: []string{"apache2"},
		PreInstall: types.OpenInstallationPreInstallConfiguration{
			RequireAtDiscovery: "exit 132", // exit 132 should report DETECTED
		},
	}

	matchedProcess := mockProcess{
		cmdline: "apache2",
		name:    `apache2`,
		pid:     int32(1234),
	}

	m := types.DiscoveryManifest{
		DiscoveredProcesses: []types.GenericProcess{matchedProcess},
	}
	mockReporter := execution.NewMockStatusReporter()
	statusSubscribers := []execution.StatusSubscriber{mockReporter}
	platformLinkGenerator := execution.NewPlatformLinkGenerator()
	installStatus := execution.NewInstallStatus(statusSubscribers, platformLinkGenerator)

	r := NewRecipeFilterRunner(types.InstallerContext{}, installStatus)

	result := r.RunFilter(context.Background(), &recipe, &m)
	require.Error(t, result)

	require.Equal(t, 1, mockReporter.RecipeDetectedCallCount)
	require.Equal(t, 0, mockReporter.RecipeUnsupportedCallCount)
}

func TestRecipeFilterRunner_ShouldFailPreInstallWithUnsupported(t *testing.T) {
	recipe := types.OpenInstallationRecipe{
		Name:         "test-recipe",
		ProcessMatch: []string{"apache2"},
		PreInstall: types.OpenInstallationPreInstallConfiguration{
			RequireAtDiscovery: "exit 1", // exit 1 should report UNSUPPORTED
		},
	}

	matchedProcess := mockProcess{
		cmdline: "apache2",
		name:    `apache2`,
		pid:     int32(1234),
	}

	m := types.DiscoveryManifest{
		DiscoveredProcesses: []types.GenericProcess{matchedProcess},
	}
	mockReporter := execution.NewMockStatusReporter()
	statusSubscribers := []execution.StatusSubscriber{mockReporter}
	platformLinkGenerator := execution.NewPlatformLinkGenerator()
	installStatus := execution.NewInstallStatus(statusSubscribers, platformLinkGenerator)

	r := NewRecipeFilterRunner(types.InstallerContext{}, installStatus)

	result := r.ConfirmCompatibleRecipes(context.Background(), []types.OpenInstallationRecipe{recipe}, &m)
	require.Error(t, result)

	require.Equal(t, 0, mockReporter.RecipeDetectedCallCount)
	require.Equal(t, 1, mockReporter.RecipeUnsupportedCallCount)
}
