package config

import "testing"

func TestPreviewUsesAPIOrigin(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("PORT", "18085")
	t.Setenv("BASE_URL", "https://frontend.example")
	for _, c := range []struct{ explicit, render, want string }{
		{"", "", "http://localhost:18085"},
		{"", "https://api.onrender.com", "https://api.onrender.com"},
		{"https://custom-api.example/", "https://api.onrender.com", "https://custom-api.example"},
	} {
		t.Setenv("API_BASE_URL", c.explicit)
		t.Setenv("RENDER_EXTERNAL_URL", c.render)
		cfg := Load()
		if cfg.APIBaseURL != c.want || cfg.BaseURL != "https://frontend.example" {
			t.Fatal("API and frontend origins were not kept separate")
		}
	}
}
