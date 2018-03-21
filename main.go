package main

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps/datatypes/node"
	"github.com/galaco/bsp/lumps/datatypes/face"
	"github.com/galaco/vmf"
	"log"
	"os"
	"bytes"
	"fmt"
	"strconv"
	"github.com/galaco/bsp/lumps/datatypes/leaf"
	"strings"
	"path/filepath"
	"github.com/galaco/vvis/portals"
	"github.com/galaco/vvis/pas"
	"github.com/galaco/bsp/lumps"
)

var Leafs []leaf.Leaf


func main() {
	fmt.Println("DormantLemon/Galaco's VVIS compiler")
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

	entities := ParseEntities(file.GetLump(bsp.LUMP_ENTITIES).GetContents().GetData().(string))

	if Args.UseRadius == false {
		radius := DetermineVisRadius(&entities)
		if radius > 0.0 {
			Args.UseRadius = true
			Args.VisRadius = radius * radius
		}
	}

	if Args.UseRadius {
		Leafs = file.GetLump(bsp.LUMP_LEAFS).GetContents().GetData().([]leaf.Leaf)
		MarkLeavesAsRadial(&Leafs)
	}

	portalFilename := Args.File
	// Does this even do anything except print?
	if Args.TmpIn != "" {
		// sprintf ( portalfile, "%s%s", inbase, argv[i] );
		// Q_StripExtension( portalfile, portalfile, sizeof( portalfile ) );
	}
	portalFilename = strings.TrimSuffix(portalFilename, filepath.Ext(portalFilename))
	portalFilename += ".prt"
	fmt.Println("Reading " + portalFilename)

	portalInfo := portals.Load(portalFilename, false, file.GetLump(bsp.LUMP_VISIBILITY).(*lumps.Visibility))

	if Args.TraceClusterStart < 0 {
		// CalcVis()
		pas.Calculate()

		// BuildClusterTable()
		// CalcVisibleFogVolumes()
		// CalcDistanceFromLeavesToWater()

		// visdatasize = vismap_p - dvisdata;
		fmt.Println("visdatasize:%i  compressed from %i\n", visdatasize, originalvismapsize*2)

		fmt.Println("writing %s\n", Args.File)
		// WriteBSPFile (mapFile);
	} else {
		if Args.TraceClusterStart < 0 ||
			Args.TraceClusterStart >= portalInfo.PortalClusters ||
			Args.TraceClusterStop < 0 ||
			Args.TraceClusterStop >= portalInfo.PortalClusters {
			log.Fatalf("Invalid cluster trace: %d to %d, valid range is 0 to %d\n", Args.TraceClusterStart, Args.TraceClusterStop, portalInfo.PortalClusters-1 );
		}
		// CalcVisTrace ();
		// WritePortalTrace(source);
	}
}

// Read and parse BSP created by vbsp
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

func ParseEntities(entdata string) vmf.Node {
	buffer := bytes.NewBufferString(entdata)

	reader := vmf.NewReader(buffer)
	parsed,err := reader.Read()
	if err != nil {
		FatalError("Unable to parse entities")
	}
	return parsed.Unclassified
}

func FatalError(text string) {
	log.Fatal(text)
}

// Determine max leaf Radius from fog controller farz
func DetermineVisRadius(entities *vmf.Node) float64 {
	radius := -1.0
	for _,i := range *entities.GetAllValues() {
		n := i.(vmf.Node)
		if n.GetProperty("classname") == "env_fog_controller" {
			s := n.GetProperty("farz")
			if s == "" {
				return radius
			}
			f,err := strconv.ParseFloat(s, 64)
			if err != nil {
				if f == 0.0 {
					radius = -1.0
				} else {
					radius = f
				}
			}
			break
		}
	}
	return radius
}

//Mark leaves as radius
func MarkLeavesAsRadial(leafs *[]leaf.Leaf) {
	for _,l := range *leafs {
		f := l.Flags() | leaf.LEAF_FLAGS_RADIAL
		l.SetFlags(f)
	}
}