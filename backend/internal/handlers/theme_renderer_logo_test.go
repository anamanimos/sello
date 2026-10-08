package handlers

import (
	"sello/internal/models"
	"strings"
	"testing"
)

func TestReplaceStoreLogoPriority(t *testing.T) {
	// Template snippet
	template := `<img src="{{#if store.logo_url}}{{store.logo_url}}{{else}}fallback{{/if}}" />`

	// Setup Config
	config := make(map[string]interface{})
	
	// Case 1: Default (No logo, no avatar)
	storeEmpty := models.Store{}
	ownerEmpty := models.User{}
	res1 := replaceStorePlaceholders(template, storeEmpty, config, ownerEmpty)
	if !strings.Contains(res1, "sello-pp.webp") {
		t.Errorf("Case 1 Failed: Expected default fallback, got %s", res1)
	}

	// Case 2: User Avatar only
	ownerWithAvatar := models.User{AvatarURL: "avatar.png"}
	res2 := replaceStorePlaceholders(template, storeEmpty, config, ownerWithAvatar)
	if !strings.Contains(res2, "avatar.png") {
		t.Errorf("Case 2 Failed: Expected avatar, got %s", res2)
	}

	// Case 3: Store Logo (Should override avatar)
	storeWithLogo := models.Store{LogoURL: "store_logo.png"}
	res3 := replaceStorePlaceholders(template, storeWithLogo, config, ownerWithAvatar)
	if !strings.Contains(res3, "store_logo.png") {
		t.Errorf("Case 3 Failed: Expected store logo, got %s", res3)
	}

	// Case 4: Theme Config Logo (Should override everything)
	config["logo"] = "theme_logo.png"
	res4 := replaceStorePlaceholders(template, storeWithLogo, config, ownerWithAvatar)
	if !strings.Contains(res4, "theme_logo.png") {
		t.Errorf("Case 4 Failed: Expected theme config logo, got %s", res4)
	}
}
