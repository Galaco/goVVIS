package vvis

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps/datatypes/node"
	"github.com/galaco/bsp/lumps/datatypes/face"
	"log"
	"os"
)


func main() {
	// Load Arguments
	Args := ParseCmdArguments()

	if Args.LowPriority {
		//SetLowPriority()
	}
	// Load bsp file
	file := ImportBSP(Args.File)

	if len(file.GetLump(bsp.LUMP_NODES).GetContents().GetData().([]node.Node)) == 0 ||
		len(file.GetLump(bsp.LUMP_FACES).GetContents().GetData().([]face.Face)) == 0 {
		log.Fatal("Empty map.")
	}
}

func ImportBSP(filename string) bsp.Bsp {
	file,err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	fi,err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	fileData := make([]byte, fi.Size())
	file.Read(fileData)
	file.Close()

	Reader := bsp.NewReader()
	Reader.SetBuffer(fileData)

	return Reader.Read()
}
