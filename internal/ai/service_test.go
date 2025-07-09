package ai

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

func TestNewService(t *testing.T) {
	service := NewService("test-key", "gpt-4", 2000)
	
	assert.NotNil(t, service)
	
	// Type assertion to access internal fields for testing
	aiService, ok := service.(*Service)
	require.True(t, ok)
	
	assert.NotNil(t, aiService.client)
	assert.Equal(t, "gpt-4", aiService.model)
	assert.Equal(t, 2000, aiService.maxTokens)
	assert.NotNil(t, aiService.log)
}

func TestBuildConflictResolutionPrompt(t *testing.T) {
	service := &Service{}
	
	conflict := interfaces.GitConflict{
		File:    "test.go",
		Content: "<<<<<<< HEAD\nour code\n=======\ntheir code\n>>>>>>> branch",
		Ours:    "our code",
		Theirs:  "their code",
	}
	
	prompt := service.buildConflictResolutionPrompt(conflict)
	
	assert.Contains(t, prompt, "test.go")
	assert.Contains(t, prompt, "our code")
	assert.Contains(t, prompt, "their code")
	assert.Contains(t, prompt, "resolve this conflict")
	assert.Contains(t, prompt, "Return only the resolved code")
}

func TestBuildCommitMessagePrompt(t *testing.T) {
	service := &Service{}
	
	t.Run("with changes", func(t *testing.T) {
		changes := []string{"file1.go", "file2.go"}
		prompt := service.buildCommitMessagePrompt(changes)
		
		assert.Contains(t, prompt, "file1.go, file2.go")
		assert.Contains(t, prompt, "conventional commit")
		assert.Contains(t, prompt, "AI-assisted rebase")
		assert.Contains(t, prompt, "under 50 characters")
	})
	
	t.Run("without changes", func(t *testing.T) {
		changes := []string{}
		prompt := service.buildCommitMessagePrompt(changes)
		
		assert.Contains(t, prompt, "AI-assisted rebase")
		assert.Contains(t, prompt, "conflict resolution")
	})
}

func TestBuildPRDescriptionPrompt(t *testing.T) {
	service := &Service{}
	
	commits := []string{"feat: add new feature", "fix: resolve bug"}
	conflicts := []interfaces.GitConflict{
		{File: "file1.go"},
		{File: "file2.go"},
	}
	
	prompt := service.buildPRDescriptionPrompt(commits, conflicts)
	
	assert.Contains(t, prompt, "AI-assisted rebase")
	assert.Contains(t, prompt, "feat: add new feature")
	assert.Contains(t, prompt, "fix: resolve bug")
	assert.Contains(t, prompt, "2 files")
	assert.Contains(t, prompt, "file1.go")
	assert.Contains(t, prompt, "file2.go")
	assert.Contains(t, prompt, "Rebased the internal repository")
	assert.Contains(t, prompt, "Used AI to resolve")
	assert.Contains(t, prompt, "Ran all configured tests")
}

func TestBuildPRDescriptionPrompt_EmptyInputs(t *testing.T) {
	service := &Service{}
	
	commits := []string{}
	conflicts := []interfaces.GitConflict{}
	
	prompt := service.buildPRDescriptionPrompt(commits, conflicts)
	
	assert.Contains(t, prompt, "AI-assisted rebase")
	assert.Contains(t, prompt, "automated rebase operation")
	assert.NotContains(t, prompt, "Recent commits:")
	assert.NotContains(t, prompt, "Conflicts resolved")
}

// Integration test that requires OpenAI API key
func TestResolveConflict_Integration(t *testing.T) {
	// Skip if no API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test - no OpenAI API key provided")
	}
	
	service := NewService(apiKey, "gpt-3.5-turbo", 1000)
	
	conflict := interfaces.GitConflict{
		File:    "example.go",
		Content: "package main\n\n<<<<<<< HEAD\nfunc hello() {\n    fmt.Println(\"Hello World\")\n}\n=======\nfunc hello() {\n    fmt.Println(\"Hello Universe\")\n}\n>>>>>>> feature-branch\n",
		Ours:    "func hello() {\n    fmt.Println(\"Hello World\")\n}",
		Theirs:  "func hello() {\n    fmt.Println(\"Hello Universe\")\n}",
	}
	
	ctx := context.Background()
	resolution, err := service.ResolveConflict(ctx, conflict)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, resolution)
	assert.NotContains(t, resolution, "<<<<<<<")
	assert.NotContains(t, resolution, "=======")
	assert.NotContains(t, resolution, ">>>>>>>")
}

// Integration test that requires OpenAI API key
func TestGenerateCommitMessage_Integration(t *testing.T) {
	// Skip if no API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test - no OpenAI API key provided")
	}
	
	service := NewService(apiKey, "gpt-3.5-turbo", 100)
	
	changes := []string{"main.go", "utils.go"}
	
	ctx := context.Background()
	message, err := service.GenerateCommitMessage(ctx, changes)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, message)
	assert.True(t, len(message) < 100, "Commit message should be reasonably short")
}

// Integration test that requires OpenAI API key
func TestGeneratePRDescription_Integration(t *testing.T) {
	// Skip if no API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test - no OpenAI API key provided")
	}
	
	service := NewService(apiKey, "gpt-3.5-turbo", 1000)
	
	commits := []string{"feat: add new feature", "fix: resolve bug"}
	conflicts := []interfaces.GitConflict{
		{File: "main.go"},
		{File: "utils.go"},
	}
	
	ctx := context.Background()
	description, err := service.GeneratePRDescription(ctx, commits, conflicts)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, description)
	assert.Contains(t, description, "##") // Should contain markdown headers
}

// Test with realistic coreboot-style Kconfig conflicts
func TestResolveConflict_KconfigConflict_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test - no OpenAI API key provided")
	}
	
	service := NewService(apiKey, "gpt-4", 2000)
	
	conflict := interfaces.GitConflict{
		File: "src/Kconfig",
		Content: `choice
	prompt "Compiler to use"
<<<<<<< HEAD
	default COMPILER_GCC
	help
	  Use GCC toolchain for compilation
=======
	default COMPILER_LLVM_CLANG
	help
	  Use LLVM Clang toolchain for compilation with better optimization
>>>>>>> feature-branch

config COMPILER_GCC
	bool "GCC"`,
		Ours:   "default COMPILER_GCC\n\thelp\n\t  Use GCC toolchain for compilation",
		Theirs: "default COMPILER_LLVM_CLANG\n\thelp\n\t  Use LLVM Clang toolchain for compilation with better optimization",
	}
	
	ctx := context.Background()
	resolution, err := service.ResolveConflict(ctx, conflict)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, resolution)
	assert.NotContains(t, resolution, "<<<<<<<")
	assert.NotContains(t, resolution, "=======")
	assert.NotContains(t, resolution, ">>>>>>>")
	assert.Contains(t, resolution, "choice")
	assert.Contains(t, resolution, "prompt")
}

// Test with realistic coreboot-style register definition conflicts
func TestResolveConflict_RegisterDefinition_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test - no OpenAI API key provided")
	}
	
	service := NewService(apiKey, "gpt-4", 2000)
	
	conflict := interfaces.GitConflict{
		File: "src/soc/intel/common/block/gpio/gpio.c",
		Content: `#define GPIO_BASE_ADDRESS 0x48000000

static const struct gpio_config gpio_table[] = {
<<<<<<< HEAD
	{GPIO_PIN_0, GPIO_MODE_INPUT, GPIO_PULL_UP, GPIO_DRIVE_DEFAULT},
	{GPIO_PIN_1, GPIO_MODE_OUTPUT, GPIO_PULL_NONE, GPIO_DRIVE_HIGH},
	{GPIO_PIN_2, GPIO_MODE_ALTERNATE, GPIO_PULL_DOWN, GPIO_DRIVE_DEFAULT},
=======
	{GPIO_PIN_0, GPIO_MODE_INPUT, GPIO_PULL_UP, GPIO_DRIVE_STRONG},
	{GPIO_PIN_1, GPIO_MODE_OUTPUT, GPIO_PULL_NONE, GPIO_DRIVE_HIGH},  
	{GPIO_PIN_2, GPIO_MODE_ALTERNATE, GPIO_PULL_DOWN, GPIO_DRIVE_STRONG},
	{GPIO_PIN_3, GPIO_MODE_INPUT, GPIO_PULL_UP, GPIO_DRIVE_DEFAULT},
>>>>>>> feature-branch
};`,
		Ours:   "{GPIO_PIN_0, GPIO_MODE_INPUT, GPIO_PULL_UP, GPIO_DRIVE_DEFAULT},\n\t{GPIO_PIN_1, GPIO_MODE_OUTPUT, GPIO_PULL_NONE, GPIO_DRIVE_HIGH},\n\t{GPIO_PIN_2, GPIO_MODE_ALTERNATE, GPIO_PULL_DOWN, GPIO_DRIVE_DEFAULT},",
		Theirs: "{GPIO_PIN_0, GPIO_MODE_INPUT, GPIO_PULL_UP, GPIO_DRIVE_STRONG},\n\t{GPIO_PIN_1, GPIO_MODE_OUTPUT, GPIO_PULL_NONE, GPIO_DRIVE_HIGH},\n\t{GPIO_PIN_2, GPIO_MODE_ALTERNATE, GPIO_PULL_DOWN, GPIO_DRIVE_STRONG},\n\t{GPIO_PIN_3, GPIO_MODE_INPUT, GPIO_PULL_UP, GPIO_DRIVE_DEFAULT},",
	}
	
	ctx := context.Background()
	resolution, err := service.ResolveConflict(ctx, conflict)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, resolution)
	assert.NotContains(t, resolution, "<<<<<<<")
	assert.NotContains(t, resolution, "=======")
	assert.NotContains(t, resolution, ">>>>>>>")
	assert.Contains(t, resolution, "GPIO_PIN_")
	assert.Contains(t, resolution, "gpio_table")
}

// Test with realistic coreboot-style device tree conflicts
func TestResolveConflict_DeviceTree_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test - no OpenAI API key provided")
	}
	
	service := NewService(apiKey, "gpt-4", 2000)
	
	conflict := interfaces.GitConflict{
		File: "src/mainboard/google/reef/devicetree.cb",
		Content: `chip soc/intel/apollolake
	device pci 00.0 on end # Host Bridge
	device pci 02.0 on
<<<<<<< HEAD
		register "panel_cfg" = "{
			.up_delay_ms = 200,
			.down_delay_ms = 50,
			.cycle_delay_ms = 500,
			.backlight_pwm_hz = 200,
		}"
=======
		register "panel_cfg" = "{
			.up_delay_ms = 210,
			.down_delay_ms = 60,
			.cycle_delay_ms = 600,
			.backlight_pwm_hz = 1000,
			.backlight_disable_tx_link_sync = 1,
		}"
>>>>>>> feature-branch
	end # Integrated Graphics Device
end`,
		Ours:   `register "panel_cfg" = "{\n\t\t\t.up_delay_ms = 200,\n\t\t\t.down_delay_ms = 50,\n\t\t\t.cycle_delay_ms = 500,\n\t\t\t.backlight_pwm_hz = 200,\n\t\t}"`,
		Theirs: `register "panel_cfg" = "{\n\t\t\t.up_delay_ms = 210,\n\t\t\t.down_delay_ms = 60,\n\t\t\t.cycle_delay_ms = 600,\n\t\t\t.backlight_pwm_hz = 1000,\n\t\t\t.backlight_disable_tx_link_sync = 1,\n\t\t}"`,
	}
	
	ctx := context.Background()
	resolution, err := service.ResolveConflict(ctx, conflict)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, resolution)
	assert.NotContains(t, resolution, "<<<<<<<")
	assert.NotContains(t, resolution, "=======")
	assert.NotContains(t, resolution, ">>>>>>>")
	assert.Contains(t, resolution, "panel_cfg")
	assert.Contains(t, resolution, "delay_ms")
}

// Test descriptive commit message generation with Kconfig conflicts
func TestGenerateCommitMessageWithConflicts_KconfigConflict_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test - no OpenAI API key provided")
	}
	
	service := NewService(apiKey, "gpt-4", 150)
	
	changes := []string{"src/Kconfig"}
	conflicts := []interfaces.GitConflict{
		{
			File: "src/Kconfig",
			Content: `choice
	prompt "Compiler to use"
<<<<<<< HEAD
	default COMPILER_GCC
=======
	default COMPILER_LLVM_CLANG
>>>>>>> feature-branch`,
			Ours:   "default COMPILER_GCC",
			Theirs: "default COMPILER_LLVM_CLANG",
		},
	}
	
	ctx := context.Background()
	message, err := service.GenerateCommitMessageWithConflicts(ctx, changes, conflicts)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, message)
	assert.NotContains(t, message, "resolve merge conflicts")
	assert.Contains(t, message, ":")
	// Should describe the nature of the conflict
	assert.True(t, strings.Contains(message, "config") || strings.Contains(message, "compiler") || strings.Contains(message, "kconfig"))
}

// Test descriptive commit message generation with GPIO conflicts
func TestGenerateCommitMessageWithConflicts_GPIOConflict_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test - no OpenAI API key provided")
	}
	
	service := NewService(apiKey, "gpt-4", 150)
	
	changes := []string{"src/soc/intel/common/block/gpio/gpio.c"}
	conflicts := []interfaces.GitConflict{
		{
			File: "src/soc/intel/common/block/gpio/gpio.c",
			Content: `static const struct gpio_config gpio_table[] = {
<<<<<<< HEAD
	{GPIO_PIN_0, GPIO_MODE_INPUT, GPIO_PULL_UP, GPIO_DRIVE_DEFAULT},
=======
	{GPIO_PIN_0, GPIO_MODE_INPUT, GPIO_PULL_UP, GPIO_DRIVE_STRONG},
>>>>>>> feature-branch
};`,
			Ours:   "{GPIO_PIN_0, GPIO_MODE_INPUT, GPIO_PULL_UP, GPIO_DRIVE_DEFAULT}",
			Theirs: "{GPIO_PIN_0, GPIO_MODE_INPUT, GPIO_PULL_UP, GPIO_DRIVE_STRONG}",
		},
	}
	
	ctx := context.Background()
	message, err := service.GenerateCommitMessageWithConflicts(ctx, changes, conflicts)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, message)
	assert.NotContains(t, message, "resolve merge conflicts")
	assert.Contains(t, message, ":")
	// Should describe the nature of the conflict
	assert.True(t, strings.Contains(message, "gpio") || strings.Contains(message, "drive") || strings.Contains(message, "pin"))
}

// Test descriptive commit message generation with device tree conflicts
func TestGenerateCommitMessageWithConflicts_DeviceTreeConflict_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test - no OpenAI API key provided")
	}
	
	service := NewService(apiKey, "gpt-4", 150)
	
	changes := []string{"src/mainboard/google/reef/devicetree.cb"}
	conflicts := []interfaces.GitConflict{
		{
			File: "src/mainboard/google/reef/devicetree.cb",
			Content: `register "panel_cfg" = "{
<<<<<<< HEAD
	.up_delay_ms = 200,
	.backlight_pwm_hz = 200,
=======
	.up_delay_ms = 210,
	.backlight_pwm_hz = 1000,
>>>>>>> feature-branch
}"`,
			Ours:   ".up_delay_ms = 200,\n\t.backlight_pwm_hz = 200,",
			Theirs: ".up_delay_ms = 210,\n\t.backlight_pwm_hz = 1000,",
		},
	}
	
	ctx := context.Background()
	message, err := service.GenerateCommitMessageWithConflicts(ctx, changes, conflicts)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, message)
	assert.NotContains(t, message, "resolve merge conflicts")
	assert.Contains(t, message, ":")
	// Should describe the nature of the conflict
	assert.True(t, strings.Contains(message, "devicetree") || strings.Contains(message, "panel") || strings.Contains(message, "timing") || strings.Contains(message, "delay"))
}