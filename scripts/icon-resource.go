//go:build ignore

// Writes the exe's icon resource from the main icon, so Explorer, shortcuts,
// Task Manager and the Startup apps list show it. From the repository root:
//
//	go run scripts/icon-resource.go
package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

const iconPath = "cmd/assets/windows-claude-swap-icon-v2.ico"

func main() {
	ico, err := os.ReadFile(iconPath)
	if err != nil {
		log.Fatal(err)
	}
	if len(ico) < 6 || binary.LittleEndian.Uint16(ico[2:]) != 1 {
		log.Fatal(iconPath + " is not an icon")
	}
	count := int(binary.LittleEndian.Uint16(ico[4:]))
	var images [][]byte
	var group bytes.Buffer
	binary.Write(&group, binary.LittleEndian, [3]uint16{0, 1, uint16(count)})
	for i := range count {
		entry := ico[6+16*i : 22+16*i]
		size := binary.LittleEndian.Uint32(entry[8:])
		offset := binary.LittleEndian.Uint32(entry[12:])
		images = append(images, ico[offset:offset+size])
		// A group entry is the file's entry with the image's resource ID in
		// place of its file offset.
		group.Write(entry[:12])
		binary.Write(&group, binary.LittleEndian, uint16(i+1))
	}
	targets := []struct {
		arch           string
		machine, reloc uint16
	}{
		{"amd64", 0x8664, 0x0003},
		{"arm64", 0xaa64, 0x0002},
	}
	for _, target := range targets {
		data := object(target.machine, target.reloc, append(images, group.Bytes()))
		if err := os.WriteFile("rsrc_windows_"+target.arch+".syso", data, 0o644); err != nil {
			log.Fatal(err)
		}
	}
}

// object lays out a COFF object with one .rsrc section holding RT_ICON 1..n
// and RT_GROUP_ICON 1, the last blob. Each data entry carries its offset in
// the section; an ADDR32NB relocation lets the linker add the section's
// address.
func object(machine, reloc uint16, blobs [][]byte) []byte {
	const (
		rtIcon      = 3
		rtGroupIcon = 14
		english     = 0x0409
		directory   = 16
		entry       = 8
	)
	icons := len(blobs) - 1
	iconTypes := directory + 2*entry
	groupTypes := iconTypes + directory + icons*entry
	languages := groupTypes + directory + entry
	dataEntries := languages + len(blobs)*(directory+entry)
	offsets := make([]int, len(blobs))
	next := dataEntries + len(blobs)*16
	for i, blob := range blobs {
		next = (next + 7) &^ 7
		offsets[i] = next
		next += len(blob)
	}

	var section bytes.Buffer
	put := func(values ...any) {
		for _, value := range values {
			binary.Write(&section, binary.LittleEndian, value)
		}
	}
	header := func(ids int) { put(uint32(0), uint32(0), uint16(0), uint16(0), uint16(0), uint16(ids)) }
	branch := func(id, offset int) { put(uint32(id), uint32(offset)|0x80000000) }

	header(2)
	branch(rtIcon, iconTypes)
	branch(rtGroupIcon, groupTypes)
	header(icons)
	for i := range icons {
		branch(i+1, languages+i*(directory+entry))
	}
	header(1)
	branch(1, languages+icons*(directory+entry))
	for i := range blobs {
		header(1)
		put(uint32(english), uint32(dataEntries+i*16))
	}
	for i, blob := range blobs {
		put(uint32(offsets[i]), uint32(len(blob)), uint32(0), uint32(0))
	}
	for i, blob := range blobs {
		section.Write(make([]byte, offsets[i]-section.Len()))
		section.Write(blob)
	}

	const fileHeader, sectionHeader = 20, 40
	data := section.Bytes()
	relocations := fileHeader + sectionHeader + len(data)
	symbols := relocations + len(blobs)*10
	name := [8]byte{'.', 'r', 's', 'r', 'c'}
	var out bytes.Buffer
	write := func(values ...any) {
		for _, value := range values {
			binary.Write(&out, binary.LittleEndian, value)
		}
	}
	write(machine, uint16(1), uint32(0), uint32(symbols), uint32(1), uint16(0), uint16(0))
	write(name, uint32(0), uint32(0), uint32(len(data)), uint32(fileHeader+sectionHeader), uint32(relocations), uint32(0), uint16(len(blobs)), uint16(0), uint32(0x40000040))
	out.Write(data)
	for i := range blobs {
		write(uint32(dataEntries+i*16), uint32(0), reloc)
	}
	write(name, uint32(0), int16(1), uint16(0), uint8(3), uint8(0))
	write(uint32(4))
	return out.Bytes()
}
