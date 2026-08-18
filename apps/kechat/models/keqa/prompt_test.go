package keqa

import (
	"strings"
	"testing"
)

var allModes = []string{"normal", "concise", "study", "explanation", "formal"}

func TestGetKeQAPrompts_ReplaceRoleName(t *testing.T) {
	roleName := "小科"
	prompts := GetKeQAPrompts(roleName)

	for _, mode := range allModes {
		result, ok := prompts[mode]
		if !ok {
			t.Errorf("mode %s not found in prompts map", mode)
			continue
		}
		if !strings.Contains(result, roleName) {
			t.Errorf("mode %s: expected result to contain roleName '%s', but it doesn't", mode, roleName)
		}
	}
}

func TestGetKeQAPrompts_NoResidualPlaceholder(t *testing.T) {
	roleName := "小科"
	prompts := GetKeQAPrompts(roleName)

	for _, mode := range allModes {
		result, ok := prompts[mode]
		if !ok {
			t.Errorf("mode %s not found in prompts map", mode)
			continue
		}
		if strings.Contains(result, roleNamePlaceholder) {
			t.Errorf("mode %s: result still contains placeholder '%s'", mode, roleNamePlaceholder)
		}
	}
}

func TestGetKeQAPrompts_EmptyRoleName(t *testing.T) {
	roleName := ""
	prompts := GetKeQAPrompts(roleName)

	for _, mode := range allModes {
		result, ok := prompts[mode]
		if !ok {
			t.Errorf("mode %s not found in prompts map", mode)
			continue
		}
		if strings.Contains(result, roleNamePlaceholder) {
			t.Errorf("mode %s: result still contains placeholder '%s' after empty roleName replacement", mode, roleNamePlaceholder)
		}
	}
}

func TestGetKeQAPrompts_AllModesPresent(t *testing.T) {
	prompts := GetKeQAPrompts("小科")

	if len(prompts) != len(allModes) {
		t.Errorf("expected %d modes, got %d", len(allModes), len(prompts))
	}

	for _, mode := range allModes {
		if _, ok := prompts[mode]; !ok {
			t.Errorf("expected mode '%s' not found in prompts map", mode)
		}
	}
}

func TestGetKeQAPrompts_SpecialRoleName(t *testing.T) {
	roleName := "科<问答>助手"
	prompts := GetKeQAPrompts(roleName)

	for _, mode := range allModes {
		result, ok := prompts[mode]
		if !ok {
			t.Errorf("mode %s not found in prompts map", mode)
			continue
		}
		if strings.Contains(result, roleNamePlaceholder) {
			t.Errorf("mode %s: result still contains placeholder '%s'", mode, roleNamePlaceholder)
		}
		if !strings.Contains(result, roleName) {
			t.Errorf("mode %s: expected result to contain roleName '%s'", mode, roleName)
		}
	}
}