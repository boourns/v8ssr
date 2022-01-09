package v8ssr

import (
	"encoding/json"
	"io/ioutil"
	"os"
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

	r.Shutdown()
}

func TestScriptReload(t *testing.T) {
	src1 := `
const render = () => {
	return "script1"
}
`

	tempFile := writeTempFile(src1, "")
	defer os.Remove(tempFile)

	r := NewRendererFromFile(tempFile, RendererConfig{ReloadOnChange: true})
	result := r.Render("blah")
	if result.Output != "script1" || result.Error != nil {
		t.Fatalf("Render scr1 failed: %v", result)
	}

	src2 := `
const render = () => {
	return "script2"
}
`
	writeTempFile(src2, tempFile)
	result = r.Render("blah")
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