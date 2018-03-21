package pas

import (
	"fmt"
	"github.com/galaco/bsp/lumps"
	"log"
)

func Calculate(numPortalClusters int) {
	var bitbyte int32
	var dest, src *int64
	var scan *byte
	var uncompressed [lumps.MAX_MAP_LEAFS/8]byte
	var compressed [lumps.MAX_MAP_LEAFS/8]byte

	fmt.Printf("Building PAS...\n")

	count := 0
	for i := 0; i < numPortalClusters; i++ {
		scan = uncompressedvis + i*leafbytes
		memcpy (uncompressed, scan, leafbytes)

		for j := 0; j < leafbytes; j++ {
			bitbyte = scan[j]
			if bitbyte < 1 {
				continue
			}
			for k := 0; k < 8; k++ {
				if !(bitbyte & (1<<k)) {
					continue
				}
				// OR this pvs row into the phs
				index := (j<<3)+k
				if index >= numPortalClusters {
					log.Fatal("Bad bit in PVS")	// pad bits should be 0
				}
				src = (long *)(uncompressedvis + index*leafbytes)
				dest = (long *)uncompressed
				for l := 0; l < leaflongs; l++ {
					((long *)uncompressed)[l] |= src[l]
				}
			}
		}
		for j := 0; j < numPortalClusters; j++ {
			if CheckBit( uncompressed, j ) {
				count++
			}
		}

		//
		// compress the bit string
		//
		j := CompressVis (uncompressed, compressed)

		dest = (long *)vismap_p
		vismap_p += j

		if (vismap_p > vismap_end)
		log.Fatal("Vismap expansion overflow")

		dvis->bitofs[i][DVIS_PAS] = (byte *)dest-vismap

		memcpy (dest, compressed, j)
	}

	fmt.Printf("Average clusters audible: %i\n", count / numPortalClusters)
}