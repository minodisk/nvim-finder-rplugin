package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	caseconv "github.com/minodisk/go-caseconv"
	cnvim "github.com/neovim/go-client/nvim"
	cplugin "github.com/neovim/go-client/nvim/plugin"
)

type Callback func(*cnvim.Nvim) error

type Function struct {
	Name     string
	Callback Callback
	Keymaps  []string
}

var (
	Functions = []Function{
		{
			Name:     "TogglePane",
			Callback: TogglePane,
		},
		{
			Name:     "OpenPane",
			Callback: OpenPane,
		},
		{
			Name:     "ClosePane",
			Callback: ClosePane,
			Keymaps:  []string{"q"},
		},
		{
			Name:     "CloseAllPanes",
			Callback: CloseAllPanes,
			Keymaps:  []string{"Q"},
		},
		{
			Name:     "GoToRoot",
			Callback: GoToRoot,
			Keymaps:  []string{"\\"},
		},
		{
			Name:     "GoToHome",
			Callback: GoToHome,
			Keymaps:  []string{"~"},
		},
		{
			Name:     "GoToTrash",
			Callback: GoToTrash,
			Keymaps:  []string{"$"},
		},
		{
			Name:     "GoToProject",
			Callback: GoToProject,
			Keymaps:  []string{"^"},
		},
		{
			Name:     "GoToUpper",
			Callback: GoToUpper,
			Keymaps:  []string{"h"},
		},
		{
			Name:     "GoToLowerOrOpen",
			Callback: GoToLowerOrOpen,
			Keymaps:  []string{"l", "e", "<CR>"},
		},
		{
			Name:     "GoTo",
			Callback: GoTo,
			Keymaps:  []string{">"},
		},
		{
			Name:     "Select",
			Callback: Select,
			Keymaps:  []string{"<Space>"},
		},
		{
			Name:     "ReverseSelected",
			Callback: ReverseSelected,
			Keymaps:  []string{"*"},
		},
		{
			Name:     "Toggle",
			Callback: Toggle,
			Keymaps:  []string{"t"},
		},
		{
			Name:     "ToggleRecursively",
			Callback: ToggleRecursively,
			Keymaps:  []string{"T"},
		},
		{
			Name:     "CreateDir",
			Callback: CreateDir,
			Keymaps:  []string{"K"},
		},
		{
			Name:     "CreateFile",
			Callback: CreateFile,
			Keymaps:  []string{"N"},
		},
		{
			Name:     "Rename",
			Callback: Rename,
			Keymaps:  []string{"r"},
		},
		{
			Name:     "Move",
			Callback: Move,
			Keymaps:  []string{"m"},
		},
		{
			Name:     "OpenExternally",
			Callback: OpenExternally,
			Keymaps:  []string{"x"},
		},
		{
			Name:     "OpenDirExternally",
			Callback: OpenDirExternally,
			Keymaps:  []string{"X"},
		},
		{
			Name:     "RemovePermanently",
			Callback: RemovePermanently,
			Keymaps:  []string{"D"},
		},
		{
			Name:     "Remove",
			Callback: Remove,
			Keymaps:  []string{"d"},
		},
		{
			Name:     "Restore",
			Callback: Restore,
			Keymaps:  []string{"R"},
		},
		{
			Name:     "ShowCopiedList",
			Callback: ShowCopiedList,
			Keymaps:  []string{"C"},
		},
		{
			Name:     "Copy",
			Callback: Copy,
			Keymaps:  []string{"c"},
		},
		{
			Name:     "Paste",
			Callback: Paste,
			Keymaps:  []string{"p"},
		},
		{
			Name:     "Yank",
			Callback: Yank,
			Keymaps:  []string{"y"},
		},
	}
)

var (
	manifest       bool
	manifestOutput string
)

func main() {
	if len(os.Args) > 2 {
		switch os.Args[1] {
		case "manifest":
			manifest = true
			manifestOutput = os.Args[2]
			cplugin.Main(plug)
			return
		case "keymap":
			printKeymap(os.Args[2])
			return
		}
	}
	cplugin.Main(plug)
}

func printKeymap(path string) {
	var b bytes.Buffer

	for _, f := range Functions {
		fmt.Fprintf(&b, "noremap <Plug>(finder-%s) :<C-u>call Finder%s()<CR>\n", caseconv.LowerHyphens(f.Name), f.Name)
	}
	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "augroup finder\n")
	fmt.Fprintf(&b, "  autocmd!\n")
	for _, f := range Functions {
		if len(f.Keymaps) == 0 {
			continue
		}
		for _, k := range f.Keymaps {
			fmt.Fprintf(&b, "  autocmd FileType finder nnoremap <buffer> %s <Plug>(finder-%s)<CR>\n", k, caseconv.LowerHyphens(f.Name))
		}
	}
	fmt.Fprintf(&b, "augroup END\n")
	ioutil.WriteFile(path, b.Bytes(), 0644)
}

func plug(p *cplugin.Plugin) error {
	p.HandleCommand(&cplugin.CommandOptions{
		Name:  "Finder",
		NArgs: "0",
	}, OnFinder)

	for _, f := range Functions {
		p.HandleFunction(&cplugin.FunctionOptions{
			Name: fmt.Sprintf("Finder%s", f.Name),
		}, f.Callback)
	}

	if manifest {
		var b bytes.Buffer
		fmt.Fprintf(&b, `" nvim-finder

if exists('g:finder_manifest_loaded')
    finish
endif
let g:finder_manifest_loaded = 1

`)
		b.Write(p.Manifest("finder"))
		ioutil.WriteFile(manifestOutput, b.Bytes(), 0644)
		os.Exit(0)
	}

	return nil
}

func OnFinder(v *cnvim.Nvim, args []string) error {
	return TogglePane(v)
}
