// Code generated by go-bindata.
// sources:
// templates/page.gotemplate
// templates/table.gotemplate
// DO NOT EDIT!

package abu

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _templatesPageGotemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x54\xc1\x6e\xdb\x38\x10\xbd\xfb\x2b\x66\xb9\x3e\x86\x16\xb2\xd8\x43\x36\xa0\x04\x6c\x5b\xe4\xd4\x02\x45\x9b\x7c\x00\x2d\x8e\x2c\x22\x34\xa5\x52\x63\xc5\x01\xcb\x7f\x2f\x48\x49\x91\xac\x18\xf1\xc1\x16\x4d\xce\x7b\x4f\x6f\x1e\x47\xd4\x74\x34\x60\xa4\x3d\xe4\x0c\x2d\x2b\x36\xa2\x46\xa9\x8a\x0d\x00\x80\xf7\x5b\xb4\x0a\xee\x73\x60\xde\xa3\x55\x21\xb0\x10\xd2\x8e\x38\x22\x49\x28\x6b\xe9\x3a\xa4\x9c\x3d\x3d\x3e\xf0\x3b\x96\x15\x8b\x3d\x2b\x8f\x98\xb3\x5e\xe3\x4b\xdb\x38\x62\x50\x36\x96\xd0\x52\xce\x5e\xb4\xa2\x3a\x57\xd8\xeb\x12\x79\x5a\xdc\x80\xb6\x9a\xb4\x34\xbc\x2b\xa5\xc1\xfc\x76\x42\xf2\x9e\xbd\x68\xaa\x61\xf7\x45\x92\x64\xf0\x1b\x0e\xd4\x9a\x49\x01\x69\x32\x58\x78\xbf\xfb\x2e\x9d\x3c\x76\xbb\xc7\xb8\x0e\x41\x64\xc3\xc6\x08\xc0\x61\x9b\x84\x0f\x35\x46\xdb\x67\x70\x68\x72\xd6\xd1\xab\xc1\xae\x46\x24\x06\xb5\xc3\x2a\x67\x33\xd2\xa7\xa6\xa1\x8e\x9c\x6c\x9f\x7e\x7c\x0d\x21\x8a\x11\xd9\x60\x8a\xd8\x37\xea\xb5\xd8\x88\xbf\x38\x07\x6a\x5a\xe0\xbc\xd8\x44\x0e\x5d\xc1\x76\xaa\xfe\x86\xf6\x34\xf1\x59\xd9\x43\x69\x64\xd7\xe5\xcc\xca\x7e\x2f\x1d\x0c\x3f\x1c\xcf\xad\xb4\x8a\x9b\x03\xec\x0f\x5c\x49\xf7\xcc\x06\xc1\xa9\x4a\xe9\x55\xd5\x78\x9c\x81\x56\xd3\x5f\x3f\x4f\x6d\xf4\x15\xd5\xe7\xc1\xd7\x45\x7d\xc2\x38\x99\x15\x44\x94\x72\x74\x5c\x9e\xa8\x59\x9d\x9d\x8c\x72\xd2\x1e\x10\xb6\xc9\xbe\x1b\xd8\x3a\xac\x62\xe7\xaf\xbd\xd7\xfa\x23\x8c\x5e\xb0\x71\x4d\x78\x04\xef\x75\x05\xf8\x6b\xae\xff\xbf\x24\xdd\x4f\xf8\x21\xc8\xb4\x9c\x62\xf5\x5e\xd1\x1b\xb6\x5c\x42\xc7\x06\x32\x48\xcd\xcb\x59\xd9\x98\xc6\xdd\xc3\xdf\xd5\x5d\xf5\x5f\x25\xe7\x3e\x46\xe9\x11\xd3\xfb\x89\x4c\x64\xf2\x3a\x83\xc8\x8c\xbe\xee\xc6\x9c\x9a\xf9\xec\xc9\x2c\xba\x94\x29\xdd\x8f\x81\xcf\xac\xec\x87\x20\x0c\x55\xcb\x0e\xc6\xdc\x4b\x6d\xd1\xb1\x39\x92\xba\x82\x55\x6a\x67\xd8\xbd\xcb\x66\x92\x8f\x2e\x40\x3a\x5d\xdf\x4e\x3c\x84\x67\xe2\x25\x5a\x8a\x4c\x57\x6e\x45\x7d\xbb\x84\xbd\xb8\x16\x97\xaf\x3b\x2b\xf8\xf8\xfa\xc9\xbd\xc1\x37\xf2\xb4\x48\xdf\xbc\x23\xa7\x5b\x54\xec\x92\x6e\x0c\x97\xb6\x0a\xcf\x37\xb0\xad\x34\x9a\x34\x58\x06\x81\xdd\xda\x6a\x72\x13\xb4\xe2\x95\xc1\xf3\x95\x80\x08\xaa\x67\x8f\x0d\x3f\x2a\xfe\x2f\xc4\x07\x73\xe0\xff\xa4\xde\x27\x8e\x34\x10\xea\x6b\xd5\xaa\xf0\x9e\x79\xcf\x42\xd8\x79\x9f\x74\xc1\x76\xf7\x10\x6b\xba\x51\x67\x08\x69\x9b\x25\x0c\xb5\xba\x60\x19\xb9\xcb\x37\x5c\x38\x98\x25\x27\xde\x8d\xa0\x31\x32\x22\xc5\xb7\xd8\x6c\x44\x36\x3e\x89\x6c\x9c\x2b\x59\x1c\xc6\xc5\x9f\x00\x00\x00\xff\xff\x23\x3d\x90\x9b\x93\x05\x00\x00")

func templatesPageGotemplateBytes() ([]byte, error) {
	return bindataRead(
		_templatesPageGotemplate,
		"templates/page.gotemplate",
	)
}

func templatesPageGotemplate() (*asset, error) {
	bytes, err := templatesPageGotemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/page.gotemplate", size: 1427, mode: os.FileMode(436), modTime: time.Unix(1545298939, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _templatesTableGotemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x56\xdb\x6e\xe3\x36\x13\xbe\xf7\x53\xcc\xcf\x3f\xe8\xd5\xd2\x6a\xee\xb6\x29\xe5\xa2\x07\x2c\x50\x20\xdd\x2e\xda\xdd\x07\xa0\xcd\x91\xc5\x86\xa2\x54\x6a\xac\x24\x50\xf9\xee\x05\x49\xeb\x60\x4b\x71\x7b\x51\x5d\x58\xa2\x66\xe6\x9b\xd3\x37\x23\x8b\x92\x2a\x03\x46\xda\x63\xce\xd0\xb2\xdd\x46\x94\x28\xd5\x6e\x03\x00\x20\x2a\x24\x09\x87\x52\xba\x16\x29\x67\x5f\x3e\x7f\xe0\xef\x59\x36\x97\x59\x59\x61\xce\x3a\x8d\xcf\x4d\xed\x88\xc1\xa1\xb6\x84\x96\x72\xf6\xac\x15\x95\xb9\xc2\x4e\x1f\x90\xc7\xc3\x3b\xd0\x56\x93\x96\x86\xb7\x07\x69\x30\xbf\x1f\x91\x48\x93\xc1\x5d\xdf\x6f\x3f\x49\x27\xab\x76\xfb\x39\x9c\xbd\x17\x59\x12\x24\x25\xa3\xed\x13\x38\x34\x39\x6b\xe9\xd5\x60\x5b\x22\x12\x83\xd2\x61\x91\xb3\xc9\xf4\x87\xba\xa6\x96\x9c\x6c\xbe\xfc\xf6\xe8\x7d\xf0\x20\xb2\x94\x8f\xd8\xd7\xea\x75\xb7\xe9\xfb\x3b\xb4\x0a\x1e\x72\x60\x7d\x8f\x56\x79\xcf\xbc\xdf\x88\xff\x71\x0e\x54\x37\xc0\x79\x50\xe1\xa0\x0b\xb8\x1b\x30\x7f\x41\x7b\xf2\x3e\x45\x61\x65\x07\x07\x23\xdb\x36\x67\x56\x76\x7b\xe9\x20\xdd\x38\xbe\x34\xd2\x2a\x6e\x8e\xb0\x3f\x72\x25\xdd\x13\x4b\x71\x47\x2b\xa5\xaf\xac\xce\xea\x0c\xb4\x1a\x5e\xfd\x7e\x6a\x42\x09\x51\xfd\x98\x4a\x38\xb3\x8f\x18\x27\x73\x05\x11\x42\xa9\x1c\x97\x27\xaa\xaf\x74\xc3\x15\x92\x70\xd2\x1e\x11\xee\x62\x15\xdf\xc1\x9d\xc3\x22\xe4\xbd\x96\xd7\xf5\x25\x8c\x9e\x79\xe3\x9a\xb0\x82\xbe\xd7\x05\xe0\x9f\x93\xfd\xf7\x07\xd2\xdd\x80\xef\xbd\x8c\xc7\xa1\xa8\xcb\x88\x46\x6c\x39\x87\x0e\x6d\x65\x10\x5b\x9a\xb3\x43\x6d\x6a\xf7\x00\xff\x2f\xde\x17\xdf\x14\x72\xea\x6e\x08\x3d\x60\xf6\xfd\xe0\x4c\x64\x72\xdd\x83\xc8\x8c\x5e\xaf\x46\x8c\xeb\xb2\xa6\xd9\xc9\xcc\xba\x94\x29\xdd\x9d\xc9\x96\x59\xd9\x25\x22\x24\xab\x79\x07\x03\xc5\xa5\xb6\xe8\xce\x39\x9e\xe9\x72\x45\xde\x09\x76\xef\xb2\x99\x93\xf2\x7e\xc0\x21\x7c\x21\x7e\x40\x4b\x01\x69\x85\xfc\xe5\xfd\x84\x3f\xc5\x3e\xc1\x09\x92\x7b\x83\x23\x5a\x3c\xc4\x5f\xde\x92\xd3\x0d\x2a\xb8\x26\xf2\xcf\x84\xd5\xa3\xb6\x4f\xde\x27\xbd\xb2\xee\xd0\x2d\x1b\x26\x68\xda\x00\xe9\xec\x2e\x2b\xda\xf7\x67\x66\x15\x1a\x4d\x9c\xa5\x14\x75\xbb\xc2\x26\x41\x65\x68\x5b\xd4\x8c\x23\x5d\x5e\x63\x5d\xf6\x45\x64\x73\x6f\x41\xff\x2a\x96\x34\xc6\x93\x39\x1b\x78\xbe\xfd\x49\x92\x64\xf0\x17\x1c\xa9\x31\x73\xc4\x65\xf8\xe3\x68\x4c\x09\x7c\x08\x4f\x6b\x09\xbc\x5d\xc4\x75\xfe\x91\xba\x49\xfd\x91\xd2\x4b\xbc\xc0\x02\xd6\xf7\xcc\xfb\xed\x54\xb1\x78\x66\x37\x09\xbf\xe6\x31\x92\xc6\xb4\x78\x23\xca\x9b\xce\xde\xc4\x5c\x0c\xd1\xf2\xed\x65\x0b\xd3\xc2\xbd\x14\x4f\x3d\x14\x59\xa4\x62\x58\xd3\x71\xfa\x6e\x4e\x9a\x88\x53\x39\xe2\x4c\x1b\xb1\x91\x47\x6d\x25\xe9\xda\xc2\x1f\xa7\x96\x74\xf1\xca\xcf\x1f\xa2\x71\xc0\x2e\x07\x7f\x5a\x6f\x8d\x3c\xe2\xb0\xdf\x52\x3d\x74\x01\x06\x61\xfb\x6b\x51\xb4\x48\xf0\xf5\x50\x13\xa5\xdb\x10\xa9\x1a\xf2\x59\xd9\x70\xd3\x66\x8b\xa8\x69\xb5\xa5\x86\x7f\x57\x47\xb8\xbc\xef\xd9\xf6\x93\xc3\x6e\xe0\x29\x78\xff\x95\xd1\x95\x4e\x92\xc7\xf0\x34\x13\xb1\xb5\xe6\x91\xdc\x6b\xab\xf0\x25\x67\xfc\x9e\xed\x02\x98\xae\x4f\xed\x82\x1f\xcb\x45\xf8\x56\xd6\x21\x61\x82\xed\xc7\x53\x05\x8b\x08\xfe\xbb\xac\x3f\xe2\x0b\xfd\xcb\xac\x77\x41\xf7\x1f\x12\x9a\x76\xf7\x79\x59\x8f\x04\x8a\xdf\x92\x24\x22\x05\xfd\x68\xf1\x5c\x3b\xc5\xf7\x0e\xe5\xd3\x03\xc4\x1b\x97\xc6\x7c\x1b\xc5\x7e\xb3\xd9\x88\xec\x6c\x28\xb2\x44\x4f\x91\x85\x7f\x46\xbb\xbf\x03\x00\x00\xff\xff\x20\xeb\x5a\xe3\x20\x09\x00\x00")

func templatesTableGotemplateBytes() ([]byte, error) {
	return bindataRead(
		_templatesTableGotemplate,
		"templates/table.gotemplate",
	)
}

func templatesTableGotemplate() (*asset, error) {
	bytes, err := templatesTableGotemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/table.gotemplate", size: 2336, mode: os.FileMode(436), modTime: time.Unix(1545229954, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"templates/page.gotemplate":  templatesPageGotemplate,
	"templates/table.gotemplate": templatesTableGotemplate,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"templates": &bintree{nil, map[string]*bintree{
		"page.gotemplate":  &bintree{templatesPageGotemplate, map[string]*bintree{}},
		"table.gotemplate": &bintree{templatesTableGotemplate, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
