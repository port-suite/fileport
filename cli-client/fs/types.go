package fs

import (
	"fmt"
)

type Inode interface {
	Type() InodeType
	print(string, bool)
	Print()
}

type InodeType int

const (
	Directory InodeType = iota
	File
)

type DirectoryInode struct {
	Name  string  `json:"dir_name"`
	Nodes []Inode `json:"dir_nodes"`
}

func (d *DirectoryInode) Type() InodeType {
	return Directory
}

func (d *DirectoryInode) Print() {
	d.print("", true)
}

func (d *DirectoryInode) print(indent string, lastINode bool) {
	fmt.Printf("%s+- %s (directory) \n", indent, d.Name)

	if lastINode {
		indent += "   "
	} else {
		indent += "|  "
	}

	for i, node := range d.Nodes {
		node.print(indent, i == len(d.Nodes)-1)
	}
}

type FileInode struct {
	Name string `json:"file_name"`
}

func (f *FileInode) Type() InodeType {
	return File
}

func (f *FileInode) Print() {
	f.print("", true)
}

func (f *FileInode) print(indent string, lastINode bool) {
	fmt.Printf("%s+- %s\n", indent, f.Name)
}

func NewDirectoryInode(name string, nodes []Inode) *DirectoryInode {
	return &DirectoryInode{
		Name:  name,
		Nodes: nodes,
	}
}

func NewFileINode(name string) *FileInode {
	return &FileInode{Name: name}
}

func (d *DirectoryInode) AddINode(node Inode) {
	d.Nodes = append(d.Nodes, node)
}

func isString(v any) bool {
	switch v.(type) {
	case string:
		return true
	default:
		return false
	}
}

func MapToDirectoryInode(m map[string]any) *DirectoryInode {
	dirName, exists := m["dir_name"]
	if !exists {
		return nil
	}
	if !isString(dirName) {
		return nil
	}
	dir := NewDirectoryInode(dirName.(string), make([]Inode, 0))
	dirNodes, exists := m["dir_nodes"]
	if !exists {
		return nil
	}
	dirContent := []Inode{}
	for _, node := range dirNodes.([]any) {
		dir, isDir := node.(map[string]any)["dir_name"]
		if isDir {
			if !isString(dir) {
				return nil
			}
			dirContent = append(dirContent, &DirectoryInode{Name: dir.(string)})
			continue
		}
		file, isFile := node.(map[string]any)["file_name"]
		if isFile {
			if !isString(file) {
				return nil
			}
			dirContent = append(dirContent, &FileInode{Name: file.(string)})
		}
	}
	dir.Nodes = dirContent

	return dir
}

func MapToDirectoryInodeR(m map[string]any) *DirectoryInode {
	dirName, exists := m["dir_name"]
	if !exists {
		return nil
	}
	if !isString(dirName) {
		return nil
	}
	dir := NewDirectoryInode(dirName.(string), make([]Inode, 0))
	dirNodes, exists := m["dir_nodes"]
	if !exists {
		return nil
	}
	dirContent := []Inode{}
	for _, node := range dirNodes.([]any) {
		dir, isDir := node.(map[string]any)["dir_name"]
		if isDir {
			if !isString(dir) {
				return nil
			}
			dirContent = append(dirContent, MapToDirectoryInodeR(node.(map[string]any)))
			continue
		}
		file, isFile := node.(map[string]any)["file_name"]
		if isFile {
			if !isString(file) {
				return nil
			}
			dirContent = append(dirContent, &FileInode{Name: file.(string)})
		}
	}
	dir.Nodes = dirContent

	return dir
}
