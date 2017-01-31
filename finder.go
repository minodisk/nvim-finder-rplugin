package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/minodisk/go-nvim/buffer"
	"github.com/minodisk/go-nvim/nvim"
	"github.com/minodisk/go-nvim/window"
	tree "github.com/minodisk/go-tree"
)

const (
	ConfigBufferName = "finder_buffer_name"
	ConfigFileType   = "finder_file_type"
	ConfigWidth      = "finder_width"

	DefaultBufferName = "finder"
	DefaultFileType   = "finder"
	DefaultWidth      = 30
)

type Finder struct {
	nvim       *nvim.Nvim
	bufferName string
	fileType   string
	width      int
	tree       *tree.Tree
}

func New(v *nvim.Nvim, index int, context *tree.Context) (*Finder, error) {
	f := &Finder{nvim: v}
	cw := v.NearestDirectory()
	{
		bufferName, err := v.VarString(ConfigBufferName)
		if err != nil || bufferName == "" {
			bufferName = DefaultBufferName
		}
		f.bufferName = filepath.Join(cw, fmt.Sprintf("%s-%d", bufferName, index))
	}
	{
		fileType, err := v.VarString(ConfigFileType)
		if err != nil || fileType == "" {
			fileType = DefaultFileType
		}
		f.fileType = fileType
	}
	{
		width, err := v.VarInt(ConfigWidth)
		if err != nil || width == 0 {
			width = DefaultWidth
		}
		f.width = width
	}

	{
		var err error
		f.tree, err = tree.New(cw, context)
		if err != nil {
			return f, err
		}
	}

	w, err := v.CreateWindowLeft(f.bufferName)
	if err != nil {
		return f, err
	}
	if err := w.Focus(); err != nil {
		return f, err
	}
	if err := w.SetOption(window.Option{
		FoldColumn:  0,
		FoldEnable:  false,
		List:        false,
		Spell:       false,
		WinFixWidth: true,
		Wrap:        false,
	}); err != nil {
		return f, err
	}
	if err := w.SetWidth(f.width); err != nil {
		return f, err
	}

	b, err := w.Buffer()
	if err != nil {
		return f, err
	}
	if err := b.SetOption(buffer.Option{
		BufHidden:  "hide",
		BufListed:  false,
		BufType:    "nofile",
		ReadOnly:   false,
		SwapFile:   false,
		Modifiable: false,
		Modified:   false,
	}); err != nil {
		return f, err
	}
	if err := b.SetFileType(f.fileType); err != nil {
		return f, err
	}

	if err := f.tree.Open(f.Render); err != nil {
		return f, err
	}
	return f, nil
}

func (f *Finder) Closed() bool {
	ws, err := f.Windows()
	if err != nil {
		return false
	}
	return len(ws) == 0
}

func (f *Finder) Windows() ([]*window.Window, error) {
	ws, err := f.nvim.Windows()
	if err != nil {
		return nil, err
	}
	res := []*window.Window{}
	for _, w := range ws {
		b, err := w.Buffer()
		if err != nil {
			return nil, err
		}
		n, err := b.Name()
		if err != nil {
			continue
		}
		if n != f.bufferName {
			continue
		}
		res = append(res, w)
	}
	return res, nil
}

func (f *Finder) Buffer() (*buffer.Buffer, error) {
	bs, err := f.nvim.Buffers()
	if err != nil {
		return nil, err
	}
	for _, b := range bs {
		n, err := b.Name()
		if err != nil {
			continue
		}
		if n != f.bufferName {
			continue
		}
		focused, err := b.Focused()
		if err != nil {
			return nil, err
		}
		if focused {
			return b, nil
		}
	}
	return nil, fmt.Errorf("finder buffer not found")
}

func (f *Finder) Cursor() (int, error) {
	b, err := f.Buffer()
	if err != nil {
		return 0, err
	}
	p, err := b.CurrentCursor()
	if err != nil {
		return 0, err
	}
	return p.Y(), nil
}

func (f *Finder) SetCursor(at int) error {
	b, err := f.Buffer()
	if err != nil {
		return err
	}
	p, err := b.CurrentCursor()
	if err != nil {
		return err
	}
	p.SetY(at)
	return b.SetCurrentCursor(p)
}

func (f *Finder) Render(lines [][]byte) error {
	b, err := f.Buffer()
	if err != nil {
		return err
	}
	// var mem runtime.MemStats
	// runtime.ReadMemStats(&mem)
	// f.nvim.Printf("rendered (%droutines, %dbytes, %dallocs, %dtallocs)\n", runtime.NumGoroutine(), mem.HeapAlloc, mem.Alloc, mem.TotalAlloc)
	return b.Write(lines)
}

func (f *Finder) OpenFile(file *tree.File) error {
	ws, err := f.nvim.Windows()
	if err != nil {
		return err
	}
	for _, w := range ws {
		b, err := w.Buffer()
		if err != nil {
			return err
		}
		ft, err := b.FileType()
		if err != nil {
			return err
		}
		if ft != f.fileType {
			if err := w.Open(file.Path()); err != nil {
				return err
			}
			return w.Focus()
		}
	}

	w, err := f.nvim.CreateWindowRight(file.Path())
	if err != nil {
		return err
	}
	if err := f.ResetWindowWidth(); err != nil {
		return err
	}
	return w.Focus()
}

func (f *Finder) ResetWindowWidth() error {
	ws, err := f.Windows()
	if err != nil {
		return err
	}
	for _, w := range ws {
		if err := w.SetWidth(f.width); err != nil {
			return err
		}
	}
	return nil
}

func (f *Finder) RegisterYank(text string) error {
	return f.nvim.SetRegisterYank(text)
}

// Commands

func (f *Finder) Close() error {
	ws, err := f.Windows()
	if err != nil {
		return err
	}
	for _, w := range ws {
		if err := w.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (f *Finder) CD() error {
	return f.tree.CD(func() (string, error) {
		dir, err := f.nvim.InputString("Enter the destination directory", "", nvim.CompletionDir)
		if err != nil {
			return "", err
		}
		if err := f.nvim.SetCurrentDirectory(dir); err != nil {
			return "", err
		}
		dir, err = f.nvim.CurrentDirectory()
		if err != nil {
			return "", err
		}
		return dir, nil
	}, f.Render)
}

func (f *Finder) Root() error {
	return f.tree.Root(f.Render)
}

func (f *Finder) Home() error {
	return f.tree.Home(f.Render)
}

func (f *Finder) Trash() error {
	return f.tree.Trash(f.Render)
}

func (f *Finder) Project() error {
	return f.tree.Project(f.Render)
}

func (f *Finder) Up() error {
	return f.tree.Up(f.Cursor, f.Render)
}

func (f *Finder) Down() error {
	return f.tree.Down(f.Cursor, f.OpenFile, f.Render)
}

func (f *Finder) Select() error {
	return f.tree.Select(f.Cursor, f.SetCursor, f.Render)
}

func (f *Finder) ReverseSelected() error {
	return f.tree.ReverseSelected(f.Render)
}

func (f *Finder) Toggle() error {
	return f.tree.Toggle(f.Cursor, f.Render)
}

func (f *Finder) ToggleRec() error {
	return f.tree.ToggleRec(f.Cursor, f.Render)
}

func (f *Finder) CreateDir() error {
	return f.tree.CreateDir(f.Cursor, func() ([]string, error) {
		return f.nvim.InputStrings("Enter the directory names to create", nil, nvim.CompletionNone)
	}, f.Render)
}

func (f *Finder) CreateFile() error {
	return f.tree.CreateFile(f.Cursor, func() ([]string, error) {
		return f.nvim.InputStrings("Enter the file names to create", nil, nvim.CompletionNone)
	}, f.Render)
}

func (f *Finder) Rename() error {
	return f.tree.Rename(f.Cursor, func(o tree.Operator) (string, error) {
		return f.nvim.InputString(fmt.Sprintf("Rename the %s '%s' to", tree.Type(o), o.Name()), o.Name(), nvim.CompletionNone)
	}, func(os tree.Operators) ([]string, error) {
		names := make([]string, len(os))
		for i, o := range os {
			names[i] = o.Name()
		}
		return f.nvim.InputStrings("Rename the objects to", names, nvim.CompletionNone)
	}, func() error {
		return f.nvim.Printf("Renaming has been canceled.\n")
	}, f.Render)
}

func (f *Finder) Move() error {
	return f.tree.Move(f.Cursor, func(os tree.Operators) (string, error) {
		if len(os) == 1 {
			o := os[0]
			return f.nvim.InputString(fmt.Sprintf("Enter the destination to move the %s '%s'", tree.Type(o), o.Name()), "", nvim.CompletionDir)
		}
		return f.nvim.InputString("Enter the destination to move the selected files", "", nvim.CompletionDir)
	}, func() error {
		return f.nvim.Printf("Moving has been canceled.\n")
	}, f.Render)
}

func (f *Finder) OpenExternally() error {
	return f.tree.OpenExternally(f.Cursor, f.Render)
}

func (f *Finder) OpenDirExternally() error {
	return f.tree.OpenDirExternally(f.Cursor, f.Render)
}

func (f *Finder) Remove() error {
	return f.tree.Remove(f.Cursor, func(os ...tree.Operator) (bool, error) {
		if len(os) == 1 {
			o := os[0]
			return f.nvim.InputBool(fmt.Sprintf("Are you sure you want to remove the %s '%s'?", tree.Type(o), o.Name()))
		}
		return f.nvim.InputBool("Are you sure you want to remove the selected objects?")
	}, func() error {
		return f.nvim.Printf("Remove has been canceled.\n")
	}, f.Render)
}

func (f *Finder) Restore() error {
	return f.tree.Restore(f.Cursor, func(os ...tree.Operator) (bool, error) {
		if len(os) == 1 {
			o := os[0]
			return f.nvim.InputBool(fmt.Sprintf("Are you sure you want to restore the %s '%s'?", tree.Type(o), tree.OriginalPath(o)))
		}
		return f.nvim.InputBool("Are you sure you want to restore the selected objects?")
	}, func() error {
		return f.nvim.Printf("Remove has been canceled.\n")
	}, f.Render)
}

func (f *Finder) RemovePermanently() error {
	return f.tree.RemovePermanently(f.Cursor, func(os ...tree.Operator) (bool, error) {
		if len(os) == 1 {
			o := os[0]
			return f.nvim.InputBool(fmt.Sprintf("Are you sure you want to permanently remove the %s '%s'?", tree.Type(o), o.Name()))
		}
		return f.nvim.InputBool("Are you sure you want to permanently remove the selected objects?")
	}, func() error {
		return f.nvim.Printf("Remove permanently has been canceled.\n")
	}, f.Render)
}

func (f *Finder) Copy() error {
	return f.tree.Copy(f.Cursor)
}

func (f *Finder) CopiedList() error {
	return f.tree.CopiedList(func(os tree.Operators) error {
		pathes := make([]string, len(os))
		for i, o := range os {
			pathes[i] = o.Path()
		}
		return f.nvim.Printf("%s\n", strings.Join(pathes, "\n"))
	})
}

func (f *Finder) Paste() error {
	return f.tree.Paste(f.Cursor, f.Render)
}

func (f *Finder) Yank() error {
	return f.tree.Yank(f.Cursor, f.RegisterYank)
}
