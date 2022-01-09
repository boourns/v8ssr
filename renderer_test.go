package v8ssr

import (
	"encoding/json"
	"testing"
)

func TestBasicRenderer(t *testing.T) {
	src := `
const render = () => {
	return "hello from javascript, " + params 
}
`
	r := NewRenderer(src, RendererConfig{})

	result := r.Render("blah")
	if result.Error != nil {
		t.Fatalf("Render returned error: %v", result.Error)
	}

	if result.Output != "hello from javascript, blah" {
		t.Fatalf("Incorrect output, received %v", result.Output)
	}

	r.Shutdown()

	src = "const render = () => {let p = JSON.parse(params); return `${p.things.length} things, url ${p.url}`}"

	r = NewRenderer(src, RendererConfig{})

	data, err := json.Marshal(map[string]interface{}{
		"things": []interface{}{"apple", "banana"},
		"url": "http://localhost:8000/blah",
	})
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	result = r.Render(string(data))

	if result.Error != nil {
		t.Fatalf("Render returned error: %v", result.Error)
	}

	if result.Output != "2 things, url http://localhost:8000/blah" {
		t.Fatalf("Incorrect output, received %v", result.Output)
	}

}
