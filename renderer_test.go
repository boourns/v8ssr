package v8ssr

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	v8 "rogchap.com/v8go"
	"testing"
)

func TestBasicRenderer(t *testing.T) {
	ctx := context.Background()

	src := `
const entry = () => {
	return "hello from javascript, " + params 
}
`
	r := NewRenderer(src, RendererConfig{}, map[string]RendererCallback{})

	result := r.Render(ctx, "blah")
	if result.Error != nil {
		t.Fatalf("Render returned error: %v", result.Error)
	}

	if result.Output != "hello from javascript, blah" {
		t.Fatalf("Incorrect output, received %v", result.Output)
	}

	r.Shutdown()

	src = "const entry = () => {let p = JSON.parse(params); return `${p.things.length} things, url ${p.url}`}"

	r = NewRenderer(src, RendererConfig{}, map[string]RendererCallback{})

	data, err := json.Marshal(map[string]interface{}{
		"things": []interface{}{"apple", "banana"},
		"url": "http://localhost:8000/blah",
	})
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	result = r.Render(ctx, string(data))

	if result.Error != nil {
		t.Fatalf("Render returned error: %v", result.Error)
	}

	if result.Output != "2 things, url http://localhost:8000/blah" {
		t.Fatalf("Incorrect output, received %v", result.Output)
	}

	r.Shutdown()
}


func TestRendererCallbacks(t *testing.T) {
	ctx := context.Background()

	src := `
const entry = () => {
	return "Calling callback() - " + callback("stringParam")
}
`
	r := NewRenderer(src, RendererConfig{}, map[string]RendererCallback{
		"callback": func(ctx context.Context, values []*v8.Value) interface{} {
			if !values[0].IsString() || values[0].String() != "stringParam" {
				return "fails"
			}
			return "works!"
		},
	})

	result := r.Render(ctx, "blah")
	if result.Error != nil {
		t.Fatalf("Render returned error: %v", result.Error)
	}

	if result.Output != "Calling callback() - works!" {
		t.Fatalf("Incorrect output, received %v", result.Output)
	}

	r.Shutdown()
}

func TestScriptReload(t *testing.T) {
	ctx := context.Background()

	src1 := `
const entry = () => {
	return "script1"
}
`

	tempFile := writeTempFile(src1, "")
	defer os.Remove(tempFile)

	r := NewRendererFromFile(tempFile, RendererConfig{ReloadOnChange: true}, map[string]RendererCallback{})
	result := r.Render(ctx, "blah")
	if result.Output != "script1" || result.Error != nil {
		t.Fatalf("Render scr1 failed: %v", result)
	}

	src2 := `
const entry = () => {
	return "script2"
}
`
	writeTempFile(src2, tempFile)
	result = r.Render(ctx, "blah")
	if result.Output != "script2" || result.Error != nil {
		t.Fatalf("Render scr2 failed: %v", result)
	}

}

func writeTempFile(content string, filename string) string {
	var file *os.File
	var err error

	if filename == "" {
		file, err = ioutil.TempFile("", "v8ssr-test-*")
	} else {
		file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	}

	if err != nil {
		panic(err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		panic(err)
	}

	file.Close()

	return file.Name()
}